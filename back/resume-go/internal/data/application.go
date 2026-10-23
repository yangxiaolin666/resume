package data

import (
	"database/sql"
	"errors"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"

	"resume-backend/internal/model"
)

// ==================================================================
// 简历投递数据访问层
// ==================================================================
// 对应表 job_application，是「学生投递 -> 企业处理」链路的持久化层。
//
// 【索引与查询】
//   - uk_job_student(job_id, student_id)：唯一索引，
//     既防止重复投递，也让「是否投递过」的等值查询走索引；
//   - idx_student(student_id, id)：支撑学生端「我的投递」列表 + 按 id 倒序；
//   - idx_company_status(com_id, status, id)：支撑企业端按状态筛选并按时间倒序。
// ==================================================================

// ErrDuplicateApplication 表示同一个学生重复投递同一个职位。
// 用哨兵错误（sentinel error）而不是直接返回 MySQL 错误，
// 是为了让 service 层不用关心底层数据库的具体错误码。
var ErrDuplicateApplication = errors.New("duplicate application")

const applicationSelect = `SELECT a.id, a.job_id, COALESCE(jp.title, ''), a.com_id, COALESCE(c.com_name, ''),
	a.student_id, COALESCE(s.user_id, 0), COALESCE(s.name, ''), COALESCE(s.major, ''), COALESCE(s.degree, ''),
	a.resume_url, a.status, a.created_at, a.updated_at
	FROM job_application a
	LEFT JOIN job_postings jp ON jp.id = a.job_id
	LEFT JOIN company c ON c.id = a.com_id
	LEFT JOIN student s ON s.id = a.student_id`

func scanApplication(sc rowScanner) (*model.JobApplication, error) {
	var item model.JobApplication
	var jobTitle, comName, studentName, major, degree, status sql.NullString
	var createdAt, updatedAt sql.NullTime
	err := sc.Scan(&item.ID, &item.JobID, &jobTitle, &item.ComID, &comName,
		&item.StudentID, &item.StudentUserID, &studentName, &major, &degree,
		&item.ResumeURL, &status, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	item.JobTitle = str(jobTitle)
	item.ComName = str(comName)
	item.StudentName = str(studentName)
	item.Major = str(major)
	item.Degree = str(degree)
	item.Status = str(status)
	if createdAt.Valid {
		item.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
	}
	if updatedAt.Valid {
		item.UpdatedAt = updatedAt.Time.Format("2006-01-02 15:04:05")
	}
	return &item, nil
}

// InsertApplication 新增一条投递记录。
//
// 重复投递不靠「先查再插」判断，而是直接插入并捕获唯一键冲突：
//   - 先查再插存在并发窗口（两个请求同时查到“没投过”，然后都插入）；
//   - 唯一索引是数据库层面的最终保证，捕获 1062 错误最可靠。
//
// MySQL 错误码 1062 = ER_DUP_ENTRY（Duplicate entry）。
func (d *DB) InsertApplication(jobID, comID, studentID int, resumeURL string) (int64, error) {
	res, err := d.Exec(`INSERT INTO job_application
		(job_id, com_id, student_id, resume_url, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())`,
		jobID, comID, studentID, resumeURL, model.ApplicationPending)
	if err != nil {
		var mysqlErr *mysqlDriver.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return 0, ErrDuplicateApplication
		}
		return 0, err
	}
	return res.LastInsertId()
}

// FindApplicationByID 查询单条投递详情（用于插入后回填完整数据）。
func (d *DB) FindApplicationByID(id int64) (*model.JobApplication, error) {
	return scanApplication(d.QueryRow(applicationSelect+` WHERE a.id = ?`, id))
}

// ListApplicationsByStudent 学生端「我的投递」。
// 按 id 倒序 = 按投递时间倒序（自增主键与插入顺序一致），
// 可以直接利用 idx_student(student_id, id) 避免额外排序。
func (d *DB) ListApplicationsByStudent(studentID int) ([]model.JobApplication, error) {
	rows, err := d.Query(applicationSelect+` WHERE a.student_id = ? ORDER BY a.id DESC`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.JobApplication{}
	for rows.Next() {
		item, err := scanApplication(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *item)
	}
	return list, rows.Err()
}

// ListApplicationsByCompany 企业端「收到的简历」。
// jobID > 0 时只看某个职位；status 非空时按状态筛选。
// 动态条件仍然全部使用 ? 占位符，值不拼进 SQL。
func (d *DB) ListApplicationsByCompany(comID, jobID int, status string) ([]model.JobApplication, error) {
	clauses := []string{"a.com_id = ?"}
	args := []any{comID}
	if jobID > 0 {
		clauses = append(clauses, "a.job_id = ?")
		args = append(args, jobID)
	}
	if strings.TrimSpace(status) != "" {
		clauses = append(clauses, "a.status = ?")
		args = append(args, strings.TrimSpace(status))
	}
	query := applicationSelect + " WHERE " + strings.Join(clauses, " AND ") + " ORDER BY a.id DESC"

	rows, err := d.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.JobApplication{}
	for rows.Next() {
		item, err := scanApplication(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *item)
	}
	return list, rows.Err()
}

// UpdateApplicationStatus 企业修改投递状态。
//
// WHERE 同时带 com_id：防止 A 企业通过猜测 id 修改 B 企业的投递记录（水平越权）。
// 返回受影响行数，0 表示记录不存在或不属于该企业。
func (d *DB) UpdateApplicationStatus(comID int, id int64, status string) (int64, error) {
	res, err := d.Exec(`UPDATE job_application SET status = ?, updated_at = NOW()
		WHERE id = ? AND com_id = ?`, status, id, comID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// DeleteApplication 学生撤回投递。
// WHERE 带 student_id，保证只能删自己的记录。
func (d *DB) DeleteApplication(studentID int, id int64) (int64, error) {
	res, err := d.Exec(`DELETE FROM job_application WHERE id = ? AND student_id = ?`, id, studentID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
