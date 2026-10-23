package data

import (
	"database/sql"
	"strings"

	"resume-backend/internal/model"
)

func studentWhere(q model.StudentQuery) (string, []any) {
	clauses := []string{"1=1"}
	args := []any{}

	if q.ID != nil {
		clauses = append(clauses, "s.id = ?")
		args = append(args, *q.ID)
	}
	if q.StuID != nil {
		clauses = append(clauses, "s.stu_id = ?")
		args = append(args, *q.StuID)
	}
	if q.Name != nil && *q.Name != "" {
		clauses = append(clauses, "s.name LIKE CONCAT('%', ?, '%')")
		args = append(args, *q.Name)
	}
	if q.Age != nil {
		clauses = append(clauses, "s.age = ?")
		args = append(args, *q.Age)
	}
	if q.Gender != nil && *q.Gender != "" {
		clauses = append(clauses, "s.gender = ?")
		args = append(args, *q.Gender)
	}
	if q.Nation != nil && *q.Nation != "" {
		clauses = append(clauses, "s.nation LIKE CONCAT('%', ?, '%')")
		args = append(args, *q.Nation)
	}
	if q.Major != nil && *q.Major != "" {
		clauses = append(clauses, "s.major LIKE CONCAT('%', ?, '%')")
		args = append(args, *q.Major)
	}
	if q.Degree != nil && *q.Degree != "" {
		clauses = append(clauses, "s.degree = ?")
		args = append(args, *q.Degree)
	}
	if q.GraduationYear != nil {
		clauses = append(clauses, "s.graduation_year = ?")
		args = append(args, *q.GraduationYear)
	}
	if q.Internship != nil {
		clauses = append(clauses, "s.internship >= ?")
		args = append(args, *q.Internship)
	}
	if len(q.Cities) > 0 {
		clauses = append(clauses, "sr.city IN ("+placeholders(len(q.Cities))+")")
		for _, v := range q.Cities {
			args = append(args, v)
		}
	}
	if len(q.Provinces) > 0 {
		clauses = append(clauses, "sr.province IN ("+placeholders(len(q.Provinces))+")")
		for _, v := range q.Provinces {
			args = append(args, v)
		}
	}
	if len(q.Languages) > 0 {
		parts := make([]string, 0, len(q.Languages))
		for _, lang := range q.Languages {
			part := "EXISTS (SELECT 1 FROM stu_language WHERE stu_id = s.id AND language = ?"
			args = append(args, lang.Language)
			if lang.Level != "" {
				part += " AND level = ?"
				args = append(args, lang.Level)
			}
			part += ")"
			parts = append(parts, part)
		}
		clauses = append(clauses, "("+strings.Join(parts, " OR ")+")")
	}
	if len(q.CertNames) > 0 {
		clauses = append(clauses, "EXISTS (SELECT 1 FROM stu_certificate WHERE stu_id = s.id AND cert_name IN ("+
			placeholders(len(q.CertNames))+"))")
		for _, v := range q.CertNames {
			args = append(args, v)
		}
	}
	if q.CertLevel != nil && *q.CertLevel != "" {
		clauses = append(clauses, "s.id IN (SELECT stu_id FROM stu_certificate WHERE level LIKE CONCAT('%', ?, '%'))")
		args = append(args, *q.CertLevel)
	}
	if q.InternshipCompany != nil && *q.InternshipCompany != "" {
		clauses = append(clauses, "s.id IN (SELECT stu_id FROM stu_internship WHERE company LIKE CONCAT('%', ?, '%'))")
		args = append(args, *q.InternshipCompany)
	}
	if len(q.Specialities) > 0 {
		clauses = append(clauses, "EXISTS (SELECT 1 FROM stu_speciality WHERE stu_id = s.id AND speciality IN ("+
			placeholders(len(q.Specialities))+"))")
		for _, v := range q.Specialities {
			args = append(args, v)
		}
	}
	if len(q.Jobs) > 0 {
		clauses = append(clauses, "EXISTS (SELECT 1 FROM stu_job WHERE stu_id = s.id AND post IN ("+
			placeholders(len(q.Jobs))+"))")
		for _, v := range q.Jobs {
			args = append(args, v)
		}
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (d *DB) SearchStudents(q model.StudentQuery) ([]model.Student, int, error) {
	where, args := studentWhere(q)
	base := `SELECT DISTINCT s.id, s.stu_id, s.name, s.age, s.gender, s.nation, s.major, s.degree,
		s.graduation_year, s.internship, s.user_id
		FROM student s
		LEFT JOIN stu_region sr ON s.id = sr.stu_id
		LEFT JOIN stu_language sl ON s.id = sl.stu_id
		LEFT JOIN stu_certificate sc ON s.id = sc.stu_id
		LEFT JOIN stu_internship si ON s.id = si.stu_id` + where

	var total int
	if err := d.QueryRow(`SELECT COUNT(*) FROM (`+base+`) t`, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	listArgs := append(append([]any{}, args...), q.PageSize, (q.PageNum-1)*q.PageSize)
	rows, err := d.Query(base+` ORDER BY s.id LIMIT ? OFFSET ?`, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []model.Student{}
	for rows.Next() {
		student, err := scanStudent(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *student)
	}
	return list, total, rows.Err()
}

func (d *DB) SearchStudentDetails(q model.StudentQuery) ([]model.StudentDetailDTO, int, error) {
	where, args := studentWhere(q)
	base := `SELECT s.id, s.stu_id, s.name, s.age, s.gender, s.nation, s.major, s.degree,
		s.graduation_year, s.internship, s.user_id, sr.province AS province, sr.city AS city, srsm.resume AS resume_url
		FROM student s
		LEFT JOIN (SELECT stu_id,
			GROUP_CONCAT(DISTINCT province SEPARATOR ',') AS province,
			GROUP_CONCAT(DISTINCT city SEPARATOR ',') AS city
			FROM stu_region GROUP BY stu_id) sr ON s.id = sr.stu_id
		LEFT JOIN (SELECT stu_id, resume, ROW_NUMBER() OVER (PARTITION BY stu_id ORDER BY id DESC) AS rn
			FROM stu_resume) srsm ON s.id = srsm.stu_id AND srsm.rn = 1` + where

	var total int
	if err := d.QueryRow(`SELECT COUNT(*) FROM (`+base+`) t`, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	listArgs := append(append([]any{}, args...), q.PageSize, (q.PageNum-1)*q.PageSize)
	rows, err := d.Query(base+` ORDER BY s.id LIMIT ? OFFSET ?`, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	details := []model.StudentDetailDTO{}
	for rows.Next() {
		var item model.StudentDetailDTO
		var name, gender, nation, major, degree, province, city, resume sql.NullString
		var age, graduation, internship, userID sql.NullInt64
		if err := rows.Scan(&item.ID, &item.StuID, &name, &age, &gender, &nation, &major, &degree,
			&graduation, &internship, &userID, &province, &city, &resume); err != nil {
			return nil, 0, err
		}
		item.Name = str(name)
		item.Age = intPtr(age)
		item.Gender = str(gender)
		item.Nation = str(nation)
		item.Major = str(major)
		item.Degree = str(degree)
		item.GraduationYear = intPtr(graduation)
		item.Internship = intPtr(internship)
		item.UserID = intPtr(userID)
		if province.Valid || city.Valid {
			item.Region = &model.RegionDTO{Province: strPtr(province), City: strPtr(city)}
		}
		if resume.Valid && resume.String != "" {
			item.Resume = &model.ResumeDTO{Resume: resume.String}
		}
		details = append(details, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if err := d.fillStudentDetails(details); err != nil {
		return nil, 0, err
	}
	return details, total, nil
}

// fillStudentDetails 批量填充语言、证书、实习、求职、技能等关联信息。
func (d *DB) fillStudentDetails(details []model.StudentDetailDTO) error {
	if len(details) == 0 {
		return nil
	}
	ids := make([]int, len(details))
	for i, item := range details {
		ids[i] = item.ID
		item.Languages = []model.LanguageDTO{}
		item.Certificates = []model.CertificateDTO{}
		item.Internships = []model.InternshipDTO{}
		item.Jobs = []model.JobDTO{}
		item.Specialities = []model.SpecialityDTO{}
		details[i] = item
	}

	index := make(map[int]int, len(details))
	for i, item := range details {
		index[item.ID] = i
	}

	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	in := "(" + placeholders(len(ids)) + ")"

	rows, err := d.Query(`SELECT stu_id, language, level, cert_no, issue_date FROM stu_language
		WHERE stu_id IN `+in+` ORDER BY stu_id, id`, args...)
	if err != nil {
		return err
	}
	for rows.Next() {
		var stuID int
		var language, level, certNo sql.NullString
		var issueDate sql.NullTime
		if err := rows.Scan(&stuID, &language, &level, &certNo, &issueDate); err != nil {
			rows.Close()
			return err
		}
		if i, ok := index[stuID]; ok {
			details[i].Languages = append(details[i].Languages, model.LanguageDTO{
				StuID: stuID, Language: str(language), Level: str(level), CertNo: str(certNo), IssueDate: formatDate(issueDate),
			})
		}
	}
	rows.Close()

	rows, err = d.Query(`SELECT stu_id, cert_name, level, cert_no, issue_date FROM stu_certificate
		WHERE stu_id IN `+in+` ORDER BY stu_id, id`, args...)
	if err != nil {
		return err
	}
	for rows.Next() {
		var stuID int
		var certName, level, certNo sql.NullString
		var issueDate sql.NullTime
		if err := rows.Scan(&stuID, &certName, &level, &certNo, &issueDate); err != nil {
			rows.Close()
			return err
		}
		if i, ok := index[stuID]; ok {
			details[i].Certificates = append(details[i].Certificates, model.CertificateDTO{
				StuID: stuID, CertName: str(certName), Level: str(level), CertNo: str(certNo), IssueDate: formatDate(issueDate),
			})
		}
	}
	rows.Close()

	rows, err = d.Query(`SELECT stu_id, company, experience FROM stu_internship
		WHERE stu_id IN `+in+` ORDER BY stu_id, id`, args...)
	if err != nil {
		return err
	}
	for rows.Next() {
		var stuID int
		var company, experience sql.NullString
		if err := rows.Scan(&stuID, &company, &experience); err != nil {
			rows.Close()
			return err
		}
		if i, ok := index[stuID]; ok {
			details[i].Internships = append(details[i].Internships, model.InternshipDTO{
				StuID: stuID, Company: str(company), Experience: str(experience),
			})
		}
	}
	rows.Close()

	rows, err = d.Query(`SELECT stu_id, post FROM stu_job WHERE stu_id IN `+in+` ORDER BY stu_id, id`, args...)
	if err != nil {
		return err
	}
	for rows.Next() {
		var stuID int
		var post sql.NullString
		if err := rows.Scan(&stuID, &post); err != nil {
			rows.Close()
			return err
		}
		if i, ok := index[stuID]; ok {
			details[i].Jobs = append(details[i].Jobs, model.JobDTO{StuID: stuID, Post: str(post)})
		}
	}
	rows.Close()

	rows, err = d.Query(`SELECT stu_id, speciality FROM stu_speciality
		WHERE stu_id IN `+in+` ORDER BY stu_id, id`, args...)
	if err != nil {
		return err
	}
	for rows.Next() {
		var stuID int
		var speciality sql.NullString
		if err := rows.Scan(&stuID, &speciality); err != nil {
			rows.Close()
			return err
		}
		if i, ok := index[stuID]; ok {
			details[i].Specialities = append(details[i].Specialities, model.SpecialityDTO{StuID: stuID, Speciality: str(speciality)})
		}
	}
	rows.Close()

	return nil
}
