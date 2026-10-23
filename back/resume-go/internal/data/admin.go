// 本文件是后台管理端的通用数据访问：
//   - student / company 的分页列表、批量删除、局部修改（管理员维护用户资料）；
//   - language（语言字典）、certificate（证书字典）的增删查。
//
// 业务定位：管理员后台的「学生管理」「企业管理」「基础字典维护」三个页面。
// 调用方是 internal/service/admin.go。
//
// 与前台查询的区别：后台不过滤状态、不做复杂筛选，只按名称模糊搜索，
// 因此这里直接内联拼 WHERE，而没有复用 search.go 的 studentWhere。
// 但「值必须走 ? 占位符」这条原则同样适用。

package data

import (
	"database/sql"
	"strings"

	"resume-backend/internal/model"
)

// AdminStudents 分页查询学生列表（后台表格）。
//
// 入参：
//   - name：可选的姓名关键字，空字符串表示不筛选；
//   - page / pageSize：页码（从 1 开始）与每页条数。
//
// 返回 (当前页数据, 总条数)。
//
// WHERE 拼接技巧：先放 " WHERE 1=1"，后面所有条件都以 " AND ..." 追加，
// 这样无论有没有筛选条件，最终 SQL 都合法。
// 关键字用 LIKE CONCAT('%', ?, '%')，值仍然走占位符，
// 但前置 % 会导致 name 上的索引失效——后台数据量有限，可以接受。
//
// 分页用 LIMIT ? OFFSET ?，OFFSET = (page-1)*pageSize。
// 深页码时 OFFSET 需要扫描并丢弃前面所有行，因此页码越大越慢；
// 后台一般只翻前几页，这个实现足够。
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
	// 复制一份 args 再追加分页参数，避免 append 复用底层数组影响原切片；
	// 分页参数同样绑定为 ?，不拼字符串。
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

// AdminCompanies 分页查询企业列表，逻辑与 AdminStudents 完全对称：
// 名称模糊搜索 + 两次查询（COUNT 拿总数、LIMIT/OFFSET 拿当前页）+ 按 id 倒序（新注册的排前面）。
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

// InsertStudentTx 在管理员创建学生的事务里插入 student 主表记录。
//
// 为什么用事务：创建学生要先插 user（账号）、再插 student（资料），
// 两步必须原子完成；tx 由 service 层传入，这里不负责 Begin/Commit。
//
// 所有字段走 ? 绑定；模型里的字段可能为 nil，驱动会把它写成 NULL，
// 这正是我们想要的“未填写”语义。
func (d *DB) InsertStudentTx(tx *sql.Tx, s model.Student) error {
	_, err := tx.Exec(`INSERT INTO student (stu_id, name, age, gender, nation, major, degree,
		graduation_year, internship, user_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.StuID, s.Name, s.Age, s.Gender, s.Nation, s.Major, s.Degree, s.GraduationYear, s.Internship, s.UserID)
	return err
}

// InsertCompanyTx 在事务里插入企业资料，与 InsertStudentTx 对称。
// com_name / com_region 来自后台表单，全部参数化绑定。
func (d *DB) InsertCompanyTx(tx *sql.Tx, c model.Company) error {
	_, err := tx.Exec(`INSERT INTO company (com_name, com_region, user_id) VALUES (?, ?, ?)`,
		c.ComName, c.ComRegion, c.UserID)
	return err
}

// DeleteStudentsTx 批量删除学生主表记录，返回受影响行数。
//
// IN (?,?,?) 的占位符个数由 placeholders 按 ids 长度动态生成，
// 具体的 id 通过 toArgs 绑定。注意这里没有判断空切片：
// 调用方（service）会先保证 ids 非空；若为空会拼出 "IN ()" 而报语法错误，
// 阅读/修改时要注意这个前置条件。
//
// 子表（语言、证书、收藏等）的清理应在同一个事务里由 service 一并完成。
func (d *DB) DeleteStudentsTx(tx *sql.Tx, ids []int) (int64, error) {
	res, err := tx.Exec(`DELETE FROM student WHERE id IN (`+placeholders(len(ids))+`)`, toArgs(ids)...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// DeleteCompaniesTx 批量删除企业主表记录，写法与 DeleteStudentsTx 一致。
// 同样假设调用方已保证 ids 非空。
func (d *DB) DeleteCompaniesTx(tx *sql.Tx, ids []int) (int64, error) {
	res, err := tx.Exec(`DELETE FROM company WHERE id IN (`+placeholders(len(ids))+`)`, toArgs(ids)...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// UpdateStudentTx 后台局部修改学生资料：只更新 input 中非 nil 的字段。
//
// 动态 SET 的拼接要点：
//   - sets 与 args 必须一一对应：加一个 "列 = ?"，就 push 一个值；
//   - 所有列名都是代码里写死的，只有值来自表单，因此值一律 ? 绑定；
//   - 最后 append(*s.ID) 对应 WHERE 子句里的 ?，
//     它必须排在所有 SET 参数之后（占位符按 SQL 中出现顺序绑定）。
//
// len(sets) == 0 时直接返回 (0, nil)：避免生成 "SET  WHERE" 这种非法 SQL，
// 也避免无意义的数据库往返。
//
// 返回值 RowsAffected 交给 service 判断记录是否存在。
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

// UpdateCompanyTx 后台局部修改企业资料，动态 SET 的写法与 UpdateStudentTx 相同。
// 目前只允许改公司名和所在地区，账号相关的字段不在这里维护。
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

// GetAllLanguages 读取语言字典（stu_language 里 language 列的候选值）。
//
// SELECT DISTINCT 在数据库侧去重，配合 language_name 上的主键/唯一约束，
// 结果天然唯一；ORDER BY language_name 让前端下拉按字母/拼音顺序展示。
// 返回 []string（而不是 DTO），因为字典只需要名称本身。
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

// AddLanguage 向语言字典新增一项。
// language.language_name 是主键，重复插入会返回 1062 Duplicate entry 错误，
// 由 service 层转换成“该语言已存在”的友好提示。
func (d *DB) AddLanguage(name string) error {
	_, err := d.Exec(`INSERT INTO language (language_name) VALUES (?)`, name)
	return err
}

// DeleteLanguage 按名称删除语言字典项。
// 字典表只维护候选值，删除不影响学生简历里已经填过的 stu_language 记录。
func (d *DB) DeleteLanguage(name string) error {
	_, err := d.Exec(`DELETE FROM language WHERE language_name = ?`, name)
	return err
}

// GetAllCertificates 读取证书字典，结构与 GetAllLanguages 一致。
// DISTINCT + ORDER BY 让候选列表稳定且无重复。
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

// AddCertificate 新增证书字典项，certificate_name 是主键，重复会报唯一键冲突。
func (d *DB) AddCertificate(name string) error {
	_, err := d.Exec(`INSERT INTO certificate (certificate_name) VALUES (?)`, name)
	return err
}

// DeleteCertificate 按名称删除证书字典项。
func (d *DB) DeleteCertificate(name string) error {
	_, err := d.Exec(`DELETE FROM certificate WHERE certificate_name = ?`, name)
	return err
}
