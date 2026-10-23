package data

import (
	"database/sql"
	"errors"

	"resume-backend/internal/model"
)

type rowScanner interface {
	Scan(dest ...any) error
}

const studentColumns = `id, stu_id, name, age, gender, nation, major, degree, graduation_year, internship, user_id`

func scanStudent(sc rowScanner) (*model.Student, error) {
	var s model.Student
	var name, gender, nation, major, degree sql.NullString
	var age, graduation, internship, userID sql.NullInt64
	err := sc.Scan(&s.ID, &s.StuID, &name, &age, &gender, &nation, &major, &degree,
		&graduation, &internship, &userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.Name = str(name)
	s.Age = intPtr(age)
	s.Gender = str(gender)
	s.Nation = str(nation)
	s.Major = str(major)
	s.Degree = str(degree)
	s.GraduationYear = intPtr(graduation)
	s.Internship = intPtr(internship)
	s.UserID = intPtr(userID)
	return &s, nil
}

func (d *DB) FindStudentByStuID(stuID int) (*model.Student, error) {
	return scanStudent(d.QueryRow(`SELECT `+studentColumns+` FROM student WHERE stu_id = ?`, stuID))
}

func (d *DB) FindStudentByID(id int) (*model.Student, error) {
	return scanStudent(d.QueryRow(`SELECT `+studentColumns+` FROM student WHERE id = ?`, id))
}

func (d *DB) FindStudentByUserID(userID int) (*model.Student, error) {
	return scanStudent(d.QueryRow(`SELECT `+studentColumns+` FROM student WHERE user_id = ?`, userID))
}

func (d *DB) UpdateStudentBasicTx(tx *sql.Tx, s *model.Student) (int64, error) {
	res, err := tx.Exec(`UPDATE student SET age = ?, gender = ?, nation = ?, major = ?, degree = ?,
		graduation_year = ?, internship = ? WHERE id = ?`,
		s.Age, s.Gender, s.Nation, s.Major, s.Degree, s.GraduationYear, s.Internship, s.ID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func execDeleteTx(tx *sql.Tx, table string, studentID int) error {
	_, err := tx.Exec(`DELETE FROM `+table+` WHERE stu_id = ?`, studentID)
	return err
}

func (d *DB) DeleteLanguagesTx(tx *sql.Tx, studentID int) error {
	return execDeleteTx(tx, "stu_language", studentID)
}

func (d *DB) DeleteRegionsTx(tx *sql.Tx, studentID int) error {
	return execDeleteTx(tx, "stu_region", studentID)
}

func (d *DB) DeleteCertificatesTx(tx *sql.Tx, studentID int) error {
	return execDeleteTx(tx, "stu_certificate", studentID)
}

func (d *DB) DeleteInternshipsTx(tx *sql.Tx, studentID int) error {
	return execDeleteTx(tx, "stu_internship", studentID)
}

func (d *DB) DeleteSpecialitiesTx(tx *sql.Tx, studentID int) error {
	return execDeleteTx(tx, "stu_speciality", studentID)
}

func (d *DB) DeleteJobsTx(tx *sql.Tx, studentID int) error {
	return execDeleteTx(tx, "stu_job", studentID)
}

func (d *DB) InsertLanguagesTx(tx *sql.Tx, studentID int, items []model.LanguageInfo) error {
	for _, it := range items {
		if _, err := tx.Exec(`INSERT INTO stu_language (stu_id, language, level) VALUES (?, ?, ?)`,
			studentID, it.Language, it.Level); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) InsertRegionsTx(tx *sql.Tx, studentID int, regions []model.RegionDTO) error {
	for _, r := range regions {
		if _, err := tx.Exec(`INSERT INTO stu_region (stu_id, province, city) VALUES (?, ?, ?)`,
			studentID, r.Province, r.City); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) InsertCertificatesTx(tx *sql.Tx, studentID int, items []model.CertificateInfo) error {
	for _, it := range items {
		if _, err := tx.Exec(`INSERT INTO stu_certificate (stu_id, cert_name, level) VALUES (?, ?, ?)`,
			studentID, it.CertName, it.Level); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) InsertInternshipsTx(tx *sql.Tx, studentID int, items []model.InternshipInfo) error {
	for _, it := range items {
		if _, err := tx.Exec(`INSERT INTO stu_internship (stu_id, company, experience) VALUES (?, ?, ?)`,
			studentID, it.Company, it.Experience); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) InsertSpecialitiesTx(tx *sql.Tx, studentID int, items []string) error {
	for _, it := range items {
		if _, err := tx.Exec(`INSERT INTO stu_speciality (stu_id, speciality) VALUES (?, ?)`,
			studentID, it); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) InsertJobsTx(tx *sql.Tx, studentID int, items []string) error {
	for _, it := range items {
		if _, err := tx.Exec(`INSERT INTO stu_job (stu_id, post) VALUES (?, ?)`,
			studentID, it); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) InsertResumeTx(tx *sql.Tx, studentID int, path string) (int, error) {
	res, err := tx.Exec(`INSERT INTO stu_resume (stu_id, resume) VALUES (?, ?)`, studentID, path)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (d *DB) FindResumeByStudentID(studentID int) (string, error) {
	var resume sql.NullString
	err := d.QueryRow(`SELECT resume FROM stu_resume WHERE stu_id = ? ORDER BY id DESC LIMIT 1`, studentID).
		Scan(&resume)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return str(resume), nil
}

func (d *DB) FindLanguagesByStudentID(studentID int) ([]model.LanguageDTO, error) {
	rows, err := d.Query(`SELECT language, level, cert_no, issue_date FROM stu_language
		WHERE stu_id = ? ORDER BY id`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.LanguageDTO{}
	for rows.Next() {
		var language, level, certNo sql.NullString
		var issueDate sql.NullTime
		if err := rows.Scan(&language, &level, &certNo, &issueDate); err != nil {
			return nil, err
		}
		result = append(result, model.LanguageDTO{
			StuID:     studentID,
			Language:  str(language),
			Level:     str(level),
			CertNo:    str(certNo),
			IssueDate: formatDate(issueDate),
		})
	}
	return result, rows.Err()
}

func (d *DB) FindCertificatesByStudentID(studentID int) ([]model.CertificateDTO, error) {
	rows, err := d.Query(`SELECT cert_name, level, cert_no, issue_date FROM stu_certificate
		WHERE stu_id = ? ORDER BY id`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.CertificateDTO{}
	for rows.Next() {
		var certName, level, certNo sql.NullString
		var issueDate sql.NullTime
		if err := rows.Scan(&certName, &level, &certNo, &issueDate); err != nil {
			return nil, err
		}
		result = append(result, model.CertificateDTO{
			StuID:     studentID,
			CertName:  str(certName),
			Level:     str(level),
			CertNo:    str(certNo),
			IssueDate: formatDate(issueDate),
		})
	}
	return result, rows.Err()
}

func (d *DB) FindInternshipsByStudentID(studentID int) ([]model.InternshipDTO, error) {
	rows, err := d.Query(`SELECT company, experience FROM stu_internship WHERE stu_id = ? ORDER BY id`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.InternshipDTO{}
	for rows.Next() {
		var company, experience sql.NullString
		if err := rows.Scan(&company, &experience); err != nil {
			return nil, err
		}
		result = append(result, model.InternshipDTO{StuID: studentID, Company: str(company), Experience: str(experience)})
	}
	return result, rows.Err()
}

func (d *DB) FindJobsByStudentID(studentID int) ([]model.JobDTO, error) {
	rows, err := d.Query(`SELECT post FROM stu_job WHERE stu_id = ? ORDER BY id`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.JobDTO{}
	for rows.Next() {
		var post sql.NullString
		if err := rows.Scan(&post); err != nil {
			return nil, err
		}
		result = append(result, model.JobDTO{StuID: studentID, Post: str(post)})
	}
	return result, rows.Err()
}

func (d *DB) FindSpecialitiesByStudentID(studentID int) ([]model.SpecialityDTO, error) {
	rows, err := d.Query(`SELECT speciality FROM stu_speciality WHERE stu_id = ? ORDER BY id`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.SpecialityDTO{}
	for rows.Next() {
		var speciality sql.NullString
		if err := rows.Scan(&speciality); err != nil {
			return nil, err
		}
		result = append(result, model.SpecialityDTO{StuID: studentID, Speciality: str(speciality)})
	}
	return result, rows.Err()
}

func (d *DB) FindRegionsByStudentID(studentID int) ([]model.RegionDTO, error) {
	rows, err := d.Query(`SELECT province, city FROM stu_region WHERE stu_id = ? ORDER BY id`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.RegionDTO{}
	for rows.Next() {
		var province, city sql.NullString
		if err := rows.Scan(&province, &city); err != nil {
			return nil, err
		}
		result = append(result, model.RegionDTO{Province: strPtr(province), City: strPtr(city)})
	}
	return result, rows.Err()
}

func formatDate(nt sql.NullTime) any {
	if !nt.Valid {
		return nil
	}
	return nt.Time.Format("2006-01-02")
}

// StudentStatistics 汇总 5 类统计数据。
func (d *DB) StudentStatistics() (*model.StudentStatistics, error) {
	stats := &model.StudentStatistics{}

	rows, err := d.Query(`SELECT COUNT(*) AS value, language, level FROM stu_language
		GROUP BY language, level ORDER BY value DESC`)
	if err != nil {
		return nil, err
	}
	stats.Languages = []model.LanguageStat{}
	for rows.Next() {
		var item model.LanguageStat
		var language, level sql.NullString
		if err := rows.Scan(&item.Value, &language, &level); err != nil {
			rows.Close()
			return nil, err
		}
		item.Language, item.Level = str(language), str(level)
		stats.Languages = append(stats.Languages, item)
	}
	rows.Close()

	rows, err = d.Query(`SELECT COUNT(*) AS value, city AS name FROM stu_region
		WHERE city IS NOT NULL GROUP BY city ORDER BY value DESC`)
	if err != nil {
		return nil, err
	}
	stats.Cities = []model.NameStat{}
	for rows.Next() {
		var item model.NameStat
		var name sql.NullString
		if err := rows.Scan(&item.Value, &name); err != nil {
			rows.Close()
			return nil, err
		}
		item.Name = str(name)
		stats.Cities = append(stats.Cities, item)
	}
	rows.Close()

	rows, err = d.Query(`SELECT COUNT(*) AS value, company, MAX(experience) AS experience FROM stu_internship
		WHERE company IS NOT NULL GROUP BY company ORDER BY value DESC`)
	if err != nil {
		return nil, err
	}
	stats.Internships = []model.InternshipStat{}
	for rows.Next() {
		var item model.InternshipStat
		var company, experience sql.NullString
		if err := rows.Scan(&item.Value, &company, &experience); err != nil {
			rows.Close()
			return nil, err
		}
		item.Company, item.Experience = str(company), str(experience)
		stats.Internships = append(stats.Internships, item)
	}
	rows.Close()

	rows, err = d.Query(`SELECT COUNT(*) AS value, speciality AS name FROM stu_speciality
		GROUP BY speciality ORDER BY value DESC`)
	if err != nil {
		return nil, err
	}
	stats.Specialities = []model.NameStat{}
	for rows.Next() {
		var item model.NameStat
		var name sql.NullString
		if err := rows.Scan(&item.Value, &name); err != nil {
			rows.Close()
			return nil, err
		}
		item.Name = str(name)
		stats.Specialities = append(stats.Specialities, item)
	}
	rows.Close()

	rows, err = d.Query(`SELECT COUNT(*) AS value, post AS name FROM stu_job
		GROUP BY post ORDER BY value DESC`)
	if err != nil {
		return nil, err
	}
	stats.Jobs = []model.NameStat{}
	for rows.Next() {
		var item model.NameStat
		var name sql.NullString
		if err := rows.Scan(&item.Value, &name); err != nil {
			rows.Close()
			return nil, err
		}
		item.Name = str(name)
		stats.Jobs = append(stats.Jobs, item)
	}
	rows.Close()

	return stats, nil
}
