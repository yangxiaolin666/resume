package data

import (
	"database/sql"
	"errors"

	"resume-backend/internal/model"
)

const companyColumns = `id, com_name, com_region, user_id`

func scanCompany(sc rowScanner) (*model.Company, error) {
	var c model.Company
	var name, region sql.NullString
	var userID sql.NullInt64
	err := sc.Scan(&c.ID, &name, &region, &userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.ComName = str(name)
	c.ComRegion = str(region)
	c.UserID = intPtr(userID)
	return &c, nil
}

func (d *DB) FindCompanyByID(id int) (*model.Company, error) {
	return scanCompany(d.QueryRow(`SELECT `+companyColumns+` FROM company WHERE id = ?`, id))
}

func (d *DB) FindCompanyByUserID(userID int) (*model.Company, error) {
	return scanCompany(d.QueryRow(`SELECT `+companyColumns+` FROM company WHERE user_id = ?`, userID))
}

func (d *DB) FindCompanyByName(name string) (*model.Company, error) {
	return scanCompany(d.QueryRow(`SELECT `+companyColumns+` FROM company WHERE com_name = ?`, name))
}
