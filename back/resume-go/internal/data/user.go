package data

import (
	"database/sql"
	"errors"

	"resume-backend/internal/model"
)

func (d *DB) FindUserByUsername(username string) (*model.User, error) {
	row := d.QueryRow(`SELECT user_id, username, password, user_type, create_time, update_time
		FROM user WHERE username = ?`, username)
	return scanUser(row)
}

func (d *DB) FindUserByID(userID int) (*model.User, error) {
	row := d.QueryRow(`SELECT user_id, username, password, user_type, create_time, update_time
		FROM user WHERE user_id = ?`, userID)
	return scanUser(row)
}

func scanUser(row *sql.Row) (*model.User, error) {
	var u model.User
	var createTime, updateTime sql.NullTime
	err := row.Scan(&u.UserID, &u.Username, &u.Password, &u.UserType, &createTime, &updateTime)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.CreateTime = timePtr(createTime)
	u.UpdateTime = timePtr(updateTime)
	return &u, nil
}

func (d *DB) UpdateUserPassword(userID int, hash string) error {
	_, err := d.Exec(`UPDATE user SET password = ?, update_time = NOW() WHERE user_id = ?`, hash, userID)
	return err
}

// UpdateDisplayName 修改用户对外展示的名字。
// 学生改 student.name，企业改 company.com_name，账号名（user.username）保持不变。
func (d *DB) UpdateDisplayName(userID int, userType, name string) error {
	var res sql.Result
	var existsQuery string
	var err error
	switch userType {
	case "student":
		res, err = d.Exec(`UPDATE student SET name = ? WHERE user_id = ?`, name, userID)
		existsQuery = `SELECT COUNT(*) FROM student WHERE user_id = ?`
	case "company":
		res, err = d.Exec(`UPDATE company SET com_name = ? WHERE user_id = ?`, name, userID)
		existsQuery = `SELECT COUNT(*) FROM company WHERE user_id = ?`
	default:
		return model.NewAppError(403, "该用户类型不支持修改昵称")
	}
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}
	// affected == 0 有两种可能：资料不存在，或新名字与旧名字相同
	var count int
	if err := d.QueryRow(existsQuery, userID).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return model.NewAppError(404, "用户资料不存在，无法修改昵称")
	}
	return nil
}

// UpdateStudentPasswordByStuID 对应 Java 管理员改密 SQL：
// UPDATE user u INNER JOIN student s ON u.user_id = s.user_id SET u.password = ? WHERE s.stu_id = ?
func (d *DB) UpdateStudentPasswordByStuID(stuID int, hash string) (int64, error) {
	res, err := d.Exec(`UPDATE user u INNER JOIN student s ON u.user_id = s.user_id
		SET u.password = ? WHERE s.stu_id = ?`, hash, stuID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *DB) InsertUserTx(tx *sql.Tx, username, hash, userType string) (int, error) {
	res, err := tx.Exec(`INSERT INTO user (username, password, user_type, create_time, update_time)
		VALUES (?, ?, ?, NOW(), NOW())`, username, hash, userType)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (d *DB) DeleteUsersTx(tx *sql.Tx, userIDs []int) error {
	if len(userIDs) == 0 {
		return nil
	}
	query := `DELETE FROM user WHERE user_id IN (` + placeholders(len(userIDs)) + `)`
	_, err := tx.Exec(query, toArgs(userIDs)...)
	return err
}
