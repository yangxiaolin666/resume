package data

import (
	"database/sql"
)

func (d *DB) AddSkills(names []string) (int, error) {
	count := 0
	for _, name := range names {
		if _, err := d.Exec(`INSERT INTO admin_all_skills (skillName, create_time) VALUES (?, CURDATE())`, name); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (d *DB) GetSkills() ([]string, error) {
	rows, err := d.Query(`SELECT skillName FROM admin_all_skills ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	seen := make(map[string]struct{})
	result := []string{}
	for rows.Next() {
		var name sql.NullString
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		value := str(name)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (d *DB) DeleteSkills(names []string) (int64, error) {
	if len(names) == 0 {
		return 0, nil
	}
	res, err := d.Exec(`DELETE FROM admin_all_skills WHERE skillName IN (`+placeholders(len(names))+`)`, toArgsStr(names)...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func toArgsStr(values []string) []any {
	args := make([]any, len(values))
	for i, v := range values {
		args[i] = v
	}
	return args
}
