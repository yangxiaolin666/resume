// 本文件负责 student 主表及其各张「一对多」子表的读写：
// stu_language（语言）、stu_region（意向地区）、stu_certificate（证书）、
// stu_internship（实习经历）、stu_speciality（技能特长）、stu_job（求职意向）、
// stu_resume（简历文件）。
//
// 业务定位：学生在简历页填写/修改资料、上传简历；管理员查看统计报表；
// 详情页按 student.id 把各子表拼成完整简历。调用方是 internal/service/student.go 与 admin.go。
//
// 写入模式的统一约定：简历编辑采用「全量替换」——
// 先 DELETE 该学生的全部子表记录，再逐条 INSERT 新数据，整个过程放在一个事务里。
// 好处是不用做复杂的 diff（找哪些新增/删除/修改），逻辑简单；
// 代价是每次编辑都会产生删除+插入，且 id 会变化（自增），
// 所以业务上不能把子表 id 当作稳定标识暴露给前端。

package data

import (
	"database/sql"
	"errors"

	"resume-backend/internal/model"
)

// rowScanner 抽象了「能 Scan 一行」的最小能力。
//
// *sql.Row（QueryRow 的返回值）和 *sql.Rows（Query 的返回值）都实现了
// Scan(dest ...any) error，所以把参数声明成这个接口后，
// scanStudent / scanCompany / scanJob 等映射函数就能同时服务于单条查询和列表查询，
// 不用写两份几乎相同的 Scan 代码。这是 Go 里「面向接口做小抽象」的典型用法。
type rowScanner interface {
	Scan(dest ...any) error
}

// studentColumns 学生主表的列清单常量。
// 抽出来是为了让 FindStudentBy* 和后台分页查询共用同一份列顺序，
// 与 scanStudent 的 Scan 参数严格对应，避免“加了字段忘了改某处”的低级错误。
const studentColumns = `id, stu_id, name, age, gender, nation, major, degree, graduation_year, internship, user_id`

// scanStudent 把一行 student 记录映射成 model.Student。
//
// 学生资料里的绝大多数列都允许 NULL（注册时只填了账号，资料是后补的），
// 所以文本列 Scan 进 sql.NullString、数字列 Scan 进 sql.NullInt64，
// 再统一转换：
//   - 文本字段用 str() 转成空字符串（前端直接展示，不用判 null）；
//   - age / graduation_year / internship / user_id 用 intPtr() 转成 *int，
//     保留“未填写 = null”这一业务语义。
//
// ErrNoRows 翻译成 (nil, nil)，调用方据此返回 404 或走“新建资料”的分支。
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

// FindStudentByStuID 按学号 stu_id 查学生。
// 注意 stu_id 是「学号」（业务字段，可能是学校给的字符串型编号），
// 而 id 是本表的自增主键——收藏、聊天等内部关联一律用 id，
// 只有需要展示/录入学号时才用 stu_id。两者不要混用。
func (d *DB) FindStudentByStuID(stuID int) (*model.Student, error) {
	return scanStudent(d.QueryRow(`SELECT `+studentColumns+` FROM student WHERE stu_id = ?`, stuID))
}

// FindStudentByID 按主键查学生，详情页、收藏详情都用它。
// 主键等值查询走聚簇索引，最多一行。
func (d *DB) FindStudentByID(id int) (*model.Student, error) {
	return scanStudent(d.QueryRow(`SELECT `+studentColumns+` FROM student WHERE id = ?`, id))
}

// FindStudentByUserID 由登录账号 user_id 反查学生资料，用于「当前登录学生」的会话。
// 学生注册时 user 和 student 是一对一写入的，所以正常只会返回一行。
func (d *DB) FindStudentByUserID(userID int) (*model.Student, error) {
	return scanStudent(d.QueryRow(`SELECT `+studentColumns+` FROM student WHERE user_id = ?`, userID))
}

// UpdateStudentBasicTx 在学生资料编辑事务中更新主表的基础字段。
//
// 入参 *model.Student 里带齐了 age/gender/... 的最终值（service 层已经做过合并），
// 所以这里是一次「整行覆盖式」更新，而不是动态 SET。
// 所有字段都用 ? 绑定；WHERE id = ? 命中主键。
//
// 返回 RowsAffected 供上层判断：0 可能表示学生不存在，
// 也可能表示“提交的内容和库里完全一致”（MySQL 只统计真正变化的行）。
// 事务由 service 层控制 Begin/Commit/Rollback，这里只负责执行语句。
func (d *DB) UpdateStudentBasicTx(tx *sql.Tx, s *model.Student) (int64, error) {
	res, err := tx.Exec(`UPDATE student SET age = ?, gender = ?, nation = ?, major = ?, degree = ?,
		graduation_year = ?, internship = ? WHERE id = ?`,
		s.Age, s.Gender, s.Nation, s.Major, s.Degree, s.GraduationYear, s.Internship, s.ID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// execDeleteTx 删除某个学生在指定子表里的全部记录，是下面 6 个 Delete* 方法的公共实现。
//
// 关于「拼接表名」的安全性问题，这里值得说清楚：
//
//	SQL 标识符（表名、列名）不能用 ? 占位符绑定，只能拼字符串。
//	所以拼接表名是否安全，取决于 table 的来源是否可信。
//	本函数的 table 全部来自本文件里写死的常量（"stu_language" 等），
//	不接受任何外部输入，因此不存在注入风险。
//	反例：如果 table 是 HTTP 参数，就必须用白名单映射，绝不能直接拼。
//
//	而条件里的学生 ID 是运行时数据，必须走 ? 绑定。
//
// 顺带一提：这几张子表都应在 stu_id 上建索引（或联合索引的前缀），
// 否则每次按学生删除都会全表扫描。
func execDeleteTx(tx *sql.Tx, table string, studentID int) error {
	_, err := tx.Exec(`DELETE FROM `+table+` WHERE stu_id = ?`, studentID)
	return err
}

// 下面 6 个 Delete* 是「全量替换」写法的前半段：
// 编辑简历时先把某个学生在这张子表里的旧数据全部清空。
// 它们共享 execDeleteTx，只是把表名固定下来，让调用处语义清晰、也避免手写表名出错。
// 事务由 service 层统一控制：任何一步失败都会整体回滚，旧数据不会丢。
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

// 下面 6 个 Insert* 是「全量替换」的后半段：删除旧数据后逐条写入新数据。
//
// 为什么用循环逐条 INSERT 而不是一条多值 INSERT：
//   - 每条记录字段数量不同（语言有 level，实习有 company/experience），
//     拼成一条语句反而更难写；
//   - 学生填写的条目数量很少（通常个位数），循环带来的性能损耗可以忽略；
//   - 出错时能精确定位是第几条数据有问题。
//
// 所有值都用 ? 绑定。注意这些方法接收的是 *sql.Tx：
// 它们必须和前面的 DELETE 在同一个事务里，否则中途失败会出现
// 「旧数据已删、新数据只写了一半」的不可用状态。
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

// InsertResumeTx 保存一份新简历的文件路径，返回自增 id。
//
// 注意这里不是「覆盖」而是「追加」：同一个学生可以有多个历史版本简历，
// 查询时再按 id 倒序取最新一份（见 FindResumeByStudentID）。
// 保留历史版本的好处是误删/替换后还能找回，代价是文件存储需要定期清理。
func (d *DB) InsertResumeTx(tx *sql.Tx, studentID int, path string) (int, error) {
	res, err := tx.Exec(`INSERT INTO stu_resume (stu_id, resume) VALUES (?, ?)`, studentID, path)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

// FindResumeByStudentID 取某学生最新的一份简历路径。
//
// ORDER BY id DESC LIMIT 1 是「取最新一条」的常用写法：
// id 自增，越大越新；LIMIT 1 让数据库命中索引后立刻停止，不用把全部简历读出来再排序。
// 如果按上传时间排序，遇到同一秒上传的多份就分不出先后，用自增 id 更稳。
//
// 返回 ("", nil) 表示「这个学生还没上传过简历」，属于正常业务状态而不是错误，
// 所以把 ErrNoRows 翻译掉，前端据此显示“暂无简历”。
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

// 下面一批 Find*ByStudentID 是简历详情页的「逐项读取」接口：
// 每个方法对应一张子表，按 stu_id 过滤、按 id 排序（保持用户录入/插入顺序）。
//
// 与 search.go 里的 fillStudentDetails 相比，这里是「单个学生」场景：
// 只有 1 个 ID，不存在 N+1 问题，所以逐个方法查更直观、也方便详情页按需加载。
//
// 每张子表都应在 stu_id 上有索引，否则 ORDER BY id 会演变成全表扫描 + filesort。
// 返回的切片统一初始化为空切片（而不是 nil），保证 JSON 输出是 [] 而不是 null。
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

// formatDate 把可能为 NULL 的日期格式化成 "2006-01-02" 字符串。
//
// 为什么不直接返回 *time.Time：
//   - 前端只需要日期，不需要时区和时分秒；
//   - 返回 any 类型是为了让这个值直接塞进 DTO 字段时既能是 string 也能是 nil，
//     序列化后对应 "2024-06-30" 或 null。
//
// "2006-01-02" 是 Go 特有的时间格式化写法（用参考时间做模板），
// 不能写成 "yyyy-MM-dd"——那样会输出字面量。
func formatDate(nt sql.NullTime) any {
	if !nt.Valid {
		return nil
	}
	return nt.Time.Format("2006-01-02")
}

// StudentStatistics 汇总 5 类统计数据。
//
// 这是后台数据看板的数据来源，5 段 SQL 都是「GROUP BY 分组 + COUNT(*) 计数 + 按数量倒序」：
//   - 语言：按 (language, level) 组合分组，统计“英语-六级”这类组合各有多少人；
//   - 城市：只统计 city 非空的记录（WHERE city IS NOT NULL），避免把“未填写”算成一个大类；
//   - 实习：按公司分组，并顺带用 MAX(experience) 取该公司实习经历里最长的一段做展示；
//   - 技能、求职意向：按名称分组计数。
//
// 为什么用 COUNT(*) AS value 而不是在 Go 里遍历累加：
//
//	聚合计算下推给数据库，只传输「分组数」这么几行结果，网络开销最小。
//
// 注意每段循环结束后都显式 rows.Close()，原因和 search.go 里一样：
// 同一个函数内多次 Query，若都用 defer 会让多个结果集同时占着连接。
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
