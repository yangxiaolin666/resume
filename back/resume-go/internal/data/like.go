package data

import (
	"database/sql"
	"errors"
	"strings"

	"resume-backend/internal/model"
)

type LikePair struct {
	ComID int
	StuID int
}

func (d *DB) ExistsStudent(id int) (bool, error) {
	var exists bool
	err := d.QueryRow(`SELECT COUNT(*) > 0 FROM student WHERE id = ?`, id).Scan(&exists)
	return exists, err
}

func (d *DB) ExistsCompany(id int) (bool, error) {
	var exists bool
	err := d.QueryRow(`SELECT COUNT(*) > 0 FROM company WHERE id = ?`, id).Scan(&exists)
	return exists, err
}

func (d *DB) ExistingLikes(pairs []LikePair) ([]LikePair, error) {
	if len(pairs) == 0 {
		return []LikePair{}, nil
	}
	values := make([]string, len(pairs))
	args := make([]any, 0, len(pairs)*2)
	for i, p := range pairs {
		values[i] = "(?, ?)"
		args = append(args, p.StuID, p.ComID)
	}
	query := `SELECT stu_id, com_id FROM com_like_stu WHERE (stu_id, com_id) IN (` +
		strings.Join(values, ",") + `)`
	rows, err := d.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []LikePair{}
	for rows.Next() {
		var p LikePair
		if err := rows.Scan(&p.StuID, &p.ComID); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (d *DB) InsertLikesTx(tx *sql.Tx, pairs []LikePair) error {
	for _, p := range pairs {
		if _, err := tx.Exec(`INSERT INTO com_like_stu (stu_id, com_id, create_time) VALUES (?, ?, NOW())`,
			p.StuID, p.ComID); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) DeleteLikesTx(tx *sql.Tx, pairs []LikePair) error {
	if len(pairs) == 0 {
		return nil
	}
	values := make([]string, len(pairs))
	args := make([]any, 0, len(pairs)*2)
	for i, p := range pairs {
		values[i] = "(?, ?)"
		args = append(args, p.StuID, p.ComID)
	}
	_, err := tx.Exec(`DELETE FROM com_like_stu WHERE (stu_id, com_id) IN (`+
		strings.Join(values, ",")+`)`, args...)
	return err
}

func (d *DB) QueryComIDByUserID(userID int) (*int, error) {
	var id int
	err := d.QueryRow(`SELECT id FROM company WHERE user_id = ?`, userID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &id, nil
}

const favoriteSelect = `SELECT f.stu_id AS stuId,
	stu.user_id AS userId,
	stu.gender AS gender,
	stu.major AS major,
	stu.degree AS degree,
	stu.graduation_year AS graduationYear,
	stu.name AS studentName,
	lang.languages AS languages,
	lang.level AS level,
	spec.specialities AS specialities,
	job.post AS post
	FROM com_like_stu f
	LEFT JOIN (SELECT id, user_id, name, gender, major, degree, graduation_year FROM student) stu ON stu.id = f.stu_id
	LEFT JOIN (SELECT stu_id, GROUP_CONCAT(language SEPARATOR '、') AS languages,
		GROUP_CONCAT(level SEPARATOR '、') AS level FROM stu_language GROUP BY stu_id) lang ON lang.stu_id = f.stu_id
	LEFT JOIN (SELECT stu_id, GROUP_CONCAT(speciality SEPARATOR '、') AS specialities
		FROM stu_speciality GROUP BY stu_id) spec ON spec.stu_id = f.stu_id
	LEFT JOIN (SELECT stu_id, GROUP_CONCAT(post SEPARATOR '、') AS post
		FROM stu_job GROUP BY stu_id) job ON job.stu_id = f.stu_id
	WHERE f.com_id = ?`

func scanFavorite(sc rowScanner) (*model.StuFavorite, error) {
	var f model.StuFavorite
	var userID sql.NullInt64
	var gender, major, degree, graduationYear, languages, level, specialities, post, studentName sql.NullString
	err := sc.Scan(&f.StuID, &userID, &gender, &major, &degree, &graduationYear, &studentName,
		&languages, &level, &specialities, &post)
	if err != nil {
		return nil, err
	}
	if userID.Valid {
		f.UserID = int(userID.Int64)
	}
	f.Gender = str(gender)
	f.Major = str(major)
	f.Degree = str(degree)
	f.GraduationYear = str(graduationYear)
	f.StudentName = str(studentName)
	f.Languages = str(languages)
	f.Level = str(level)
	f.Specialities = str(specialities)
	f.Post = str(post)
	return &f, nil
}

func (d *DB) FavoriteStudents(comID, pageNum, pageSize int) ([]model.StuFavorite, int, error) {
	var total int
	if err := d.QueryRow(`SELECT COUNT(*) FROM (`+favoriteSelect+`) t`, comID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := d.Query(favoriteSelect+` ORDER BY f.create_time DESC LIMIT ? OFFSET ?`,
		comID, pageSize, (pageNum-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []model.StuFavorite{}
	for rows.Next() {
		item, err := scanFavorite(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *item)
	}
	return list, total, rows.Err()
}
