package data

import (
	"database/sql"
	"errors"
	"strings"

	"resume-backend/internal/model"
)

type JobFilter struct {
	Status  *string
	Keyword *string
	ComID   *int
}

const jobSelect = `SELECT jp.id, jp.com_id, c.user_id, c.com_name, jp.title, jp.description,
	jp.external_link, jp.status, jp.created_at, jp.updated_at
	FROM job_postings jp LEFT JOIN company c ON jp.com_id = c.id`

func scanJob(sc rowScanner) (*model.JobPosting, error) {
	var j model.JobPosting
	var comID, comUserID sql.NullInt64
	var comName, title, description, link, status sql.NullString
	var createdAt, updatedAt sql.NullTime
	err := sc.Scan(&j.ID, &comID, &comUserID, &comName, &title, &description, &link, &status, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if comID.Valid {
		j.ComID = int(comID.Int64)
	}
	if comUserID.Valid {
		j.ComUserID = int(comUserID.Int64)
	}
	j.ComName = str(comName)
	j.Title = str(title)
	j.Description = str(description)
	j.ExternalLink = str(link)
	j.Status = str(status)
	j.CreatedAt = timePtr(createdAt)
	j.UpdatedAt = timePtr(updatedAt)
	return &j, nil
}

func buildJobWhere(filter JobFilter) (string, []any) {
	clauses := []string{"1=1"}
	args := []any{}
	if filter.ComID != nil {
		clauses = append(clauses, "jp.com_id = ?")
		args = append(args, *filter.ComID)
	}
	if filter.Status != nil && *filter.Status != "" {
		clauses = append(clauses, "jp.status = ?")
		args = append(args, *filter.Status)
	}
	if filter.Keyword != nil && *filter.Keyword != "" {
		clauses = append(clauses, "jp.title LIKE CONCAT('%', ?, '%')")
		args = append(args, *filter.Keyword)
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (d *DB) ListJobPostings(filter JobFilter, pageNum, pageSize int) ([]model.JobPosting, int, error) {
	where, args := buildJobWhere(filter)
	var total int
	if err := d.QueryRow(`SELECT COUNT(*) FROM (`+jobSelect+where+`) t`, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := jobSelect + where + " ORDER BY jp.created_at DESC LIMIT ? OFFSET ?"
	listArgs := append(append([]any{}, args...), pageSize, (pageNum-1)*pageSize)
	rows, err := d.Query(query, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []model.JobPosting{}
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *job)
	}
	return list, total, rows.Err()
}

func (d *DB) ListJobPostingsByComID(status string, comID int) ([]model.JobPosting, error) {
	where, args := buildJobWhere(JobFilter{ComID: &comID, Status: &status})
	rows, err := d.Query(jobSelect+where+" ORDER BY jp.created_at DESC", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []model.JobPosting{}
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *job)
	}
	return list, rows.Err()
}

func (d *DB) FindJobPostingByID(id int) (*model.JobPosting, error) {
	return scanJob(d.QueryRow(jobSelect+` WHERE jp.id = ?`, id))
}

func (d *DB) InsertJobPosting(comID int, title, description, link, status string) (int, error) {
	res, err := d.Exec(`INSERT INTO job_postings (com_id, title, description, external_link, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())`, comID, title, description, link, status)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (d *DB) UpdateJobPosting(id int, input model.JobPostingInput) error {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	if input.Title != nil {
		sets = append(sets, "title = ?")
		args = append(args, *input.Title)
	}
	if input.Description != nil {
		sets = append(sets, "description = ?")
		args = append(args, *input.Description)
	}
	if input.ExternalLink != nil {
		sets = append(sets, "external_link = ?")
		args = append(args, *input.ExternalLink)
	}
	if input.Status != nil {
		sets = append(sets, "status = ?")
		args = append(args, *input.Status)
	}
	args = append(args, id)
	_, err := d.Exec(`UPDATE job_postings SET `+strings.Join(sets, ", ")+` WHERE id = ?`, args...)
	return err
}

func (d *DB) UpdateJobPostingStatus(id int, status string) error {
	_, err := d.Exec(`UPDATE job_postings SET status = ?, updated_at = NOW() WHERE id = ?`, status, id)
	return err
}

func (d *DB) DeleteJobPosting(id int) (int64, error) {
	res, err := d.Exec(`DELETE FROM job_postings WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
