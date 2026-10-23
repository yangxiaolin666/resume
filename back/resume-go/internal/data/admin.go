package data

import (
	"database/sql"
	"strings"

	"resume-backend/internal/model"
)

func (d *DB) AdminStudents(name string, page, pageSize int) ([]model.Student, int, error) {
	where := " WHERE 1=1"
	args := []any{}
	if name != "" {
		where += " AND name LIKE CONCAT('%', ?, '%')"
		args = append(args, name)
	}
	var total int
	if err := d.QueryRow(`SELECT COUNT(*) FROM student`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	listArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	rows, err := d.Query(`SELECT `+studentColumns+` FROM student`+where+` ORDER BY id DESC LIMIT ? OFFSET ?`, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []model.Student{}
	for rows.Next() {
		s, err := scanStudent(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *s)
	}
	return list, total, rows.Err()
}

func (d *DB) AdminCompanies(name string, page, pageSize int) ([]model.Company, int, error) {
	where := " WHERE 1=1"
	args := []any{}
	if name != "" {
		where += " AND com_name LIKE CONCAT('%', ?, '%')"
		args = append(args, name)
	}
	var total int
	if err := d.QueryRow(`SELECT COUNT(*) FROM company`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	listArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	rows, err := d.Query(`SELECT `+companyColumns+` FROM company`+where+` ORDER BY id DESC LIMIT ? OFFSET ?`, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []model.Company{}
	for rows.Next() {
		c, err := scanCompany(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *c)
	}
	return list, total, rows.Err()
}

func (d *DB) InsertStudentTx(tx *sql.Tx, s model.Student) error {
	_, err := tx.Exec(`INSERT INTO student (stu_id, name, age, gender, nation, major, degree,
		graduation_year, internship, user_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.StuID, s.Name, s.Age, s.Gender, s.Nation, s.Major, s.Degree, s.GraduationYear, s.Internship, s.UserID)
	return err
}

func (d *DB) InsertCompanyTx(tx *sql.Tx, c model.Company) error {
	_, err := tx.Exec(`INSERT INTO company (com_name, com_region, user_id) VALUES (?, ?, ?)`,
		c.ComName, c.ComRegion, c.UserID)
	return err
}

func (d *DB) DeleteStudentsTx(tx *sql.Tx, ids []int) (int64, error) {
	res, err := tx.Exec(`DELETE FROM student WHERE id IN (`+placeholders(len(ids))+`)`, toArgs(ids)...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *DB) DeleteCompaniesTx(tx *sql.Tx, ids []int) (int64, error) {
	res, err := tx.Exec(`DELETE FROM company WHERE id IN (`+placeholders(len(ids))+`)`, toArgs(ids)...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *DB) UpdateStudentTx(tx *sql.Tx, s model.StudentInput) (int64, error) {
	sets := []string{}
	args := []any{}
	if s.StuID != nil {
		sets = append(sets, "stu_id = ?")
		args = append(args, *s.StuID)
	}
	if s.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *s.Name)
	}
	if s.Age != nil {
		sets = append(sets, "age = ?")
		args = append(args, *s.Age)
	}
	if s.Gender != nil {
		sets = append(sets, "gender = ?")
		args = append(args, *s.Gender)
	}
	if s.Nation != nil {
		sets = append(sets, "nation = ?")
		args = append(args, *s.Nation)
	}
	if s.Major != nil {
		sets = append(sets, "major = ?")
		args = append(args, *s.Major)
	}
	if s.Degree != nil {
		sets = append(sets, "degree = ?")
		args = append(args, *s.Degree)
	}
	if s.GraduationYear != nil {
		sets = append(sets, "graduation_year = ?")
		args = append(args, *s.GraduationYear)
	}
	if s.Internship != nil {
		sets = append(sets, "internship = ?")
		args = append(args, *s.Internship)
	}
	if len(sets) == 0 {
		return 0, nil
	}
	args = append(args, *s.ID)
	res, err := tx.Exec(`UPDATE student SET `+strings.Join(sets, ", ")+` WHERE id = ?`, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *DB) UpdateCompanyTx(tx *sql.Tx, c model.CompanyInput) (int64, error) {
	sets := []string{}
	args := []any{}
	if c.ComName != nil {
		sets = append(sets, "com_name = ?")
		args = append(args, *c.ComName)
	}
	if c.ComRegion != nil {
		sets = append(sets, "com_region = ?")
		args = append(args, *c.ComRegion)
	}
	if len(sets) == 0 {
		return 0, nil
	}
	args = append(args, *c.ID)
	res, err := tx.Exec(`UPDATE company SET `+strings.Join(sets, ", ")+` WHERE id = ?`, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *DB) GetAllLanguages() ([]string, error) {
	rows, err := d.Query(`SELECT DISTINCT language_name FROM language ORDER BY language_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []string{}
	for rows.Next() {
		var name sql.NullString
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		result = append(result, str(name))
	}
	return result, rows.Err()
}

func (d *DB) AddLanguage(name string) error {
	_, err := d.Exec(`INSERT INTO language (language_name) VALUES (?)`, name)
	return err
}

func (d *DB) DeleteLanguage(name string) error {
	_, err := d.Exec(`DELETE FROM language WHERE language_name = ?`, name)
	return err
}

func (d *DB) GetAllCertificates() ([]string, error) {
	rows, err := d.Query(`SELECT DISTINCT certificate_name FROM certificate ORDER BY certificate_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []string{}
	for rows.Next() {
		var name sql.NullString
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		result = append(result, str(name))
	}
	return result, rows.Err()
}

func (d *DB) AddCertificate(name string) error {
	_, err := d.Exec(`INSERT INTO certificate (certificate_name) VALUES (?)`, name)
	return err
}

func (d *DB) DeleteCertificate(name string) error {
	_, err := d.Exec(`DELETE FROM certificate WHERE certificate_name = ?`, name)
	return err
}
