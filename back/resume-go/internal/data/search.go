// 本文件实现学生简历的「多条件组合检索」，对应简历项目里的核心亮点功能。
//
// 涉及的表：主表 student，以及它的一对多子表 stu_region（意向地区）、
// stu_language（语言）、stu_certificate（证书）、stu_internship（实习）、
// stu_speciality（技能特长）、stu_job（求职意向）、stu_resume（简历文件）。
//
// 调用方：internal/service/search.go（学生列表 / 企业筛选候选人 / 后台检索）。
//
// 设计取舍：筛选条件组合非常多（十几个维度，每个都可能为空），
// 不可能为每种组合写一条静态 SQL，所以采用「动态拼接 WHERE + 全部参数化绑定」：
// SQL 结构动态生成，但所有用户输入都只以 ? 的形式出现，不会引入注入风险。

package data

import (
	"database/sql"
	"strings"

	"resume-backend/internal/model"
)

// studentWhere 把查询条件 model.StudentQuery 翻译成 WHERE 子句 + 参数列表。
//
// 几个值得展开的学习点：
//
//  1. "1=1" 打底：后面每个条件都统一用 " AND xxx" 追加，
//     不用判断“第一个条件该用 WHERE 还是 AND”，拼接逻辑简单且不出错。
//
//  2. 指针判断即“条件是否传了”：q.ID != nil 表示前端传了这个筛选项；
//     文本条件再额外判空串，避免用户传空字符串时生成 name LIKE '%%' 把全表捞出来。
//
//  3. LIKE CONCAT('%', ?, '%')：通配符放在数据库侧拼，值仍然走占位符。
//     但要注意“前置 %”会让 name/major 这类列上的普通 B+ 树索引失效，
//     只能全表/全索引扫描，所以这种模糊搜索适合数据量不大的场景；
//     大数据量应该上全文索引（FULLTEXT）或 Elasticsearch。
//
//  4. EXISTS 而不是 JOIN：语言/证书/技能/求职意向都是一对多子表。
//     用 JOIN 会让一个学生因为“会 3 门语言”变成 3 行重复数据，还要再 DISTINCT；
//     而 EXISTS 是「半连接」语义——只判断“存不存在满足条件的子记录”，
//     不复制主表行，命中一条即可短路，语义更准确、也避免了行放大。
//
//  5. 多语言条件之间用 OR：业务语义是「会英语 或 会日语」都算命中；
//     如果写成 AND，就变成“同时会多门语言”且必须各自满足等级，与需求不符。
//
//  6. 城市/省份/证书名/技能等列表条件用 IN (?,?,?)：
//     占位符个数由 placeholders(len(...)) 按运行时长度生成，
//     值通过 args 顺序绑定；绝不把列表内容拼进 SQL 文本。
//
// 返回值：以 " WHERE " 开头的 SQL 片段，以及与之严格一一对应的参数切片。
func studentWhere(q model.StudentQuery) (string, []any) {
	clauses := []string{"1=1"}
	args := []any{}

	if q.ID != nil {
		clauses = append(clauses, "s.id = ?")
		args = append(args, *q.ID)
	}
	if q.StuID != nil {
		clauses = append(clauses, "s.stu_id = ?")
		args = append(args, *q.StuID)
	}
	if q.Name != nil && *q.Name != "" {
		clauses = append(clauses, "s.name LIKE CONCAT('%', ?, '%')")
		args = append(args, *q.Name)
	}
	if q.Age != nil {
		clauses = append(clauses, "s.age = ?")
		args = append(args, *q.Age)
	}
	if q.Gender != nil && *q.Gender != "" {
		clauses = append(clauses, "s.gender = ?")
		args = append(args, *q.Gender)
	}
	if q.Nation != nil && *q.Nation != "" {
		clauses = append(clauses, "s.nation LIKE CONCAT('%', ?, '%')")
		args = append(args, *q.Nation)
	}
	if q.Major != nil && *q.Major != "" {
		clauses = append(clauses, "s.major LIKE CONCAT('%', ?, '%')")
		args = append(args, *q.Major)
	}
	if q.Degree != nil && *q.Degree != "" {
		clauses = append(clauses, "s.degree = ?")
		args = append(args, *q.Degree)
	}
	if q.GraduationYear != nil {
		clauses = append(clauses, "s.graduation_year = ?")
		args = append(args, *q.GraduationYear)
	}
	if q.Internship != nil {
		clauses = append(clauses, "s.internship >= ?")
		args = append(args, *q.Internship)
	}
	if len(q.Cities) > 0 {
		clauses = append(clauses, "sr.city IN ("+placeholders(len(q.Cities))+")")
		for _, v := range q.Cities {
			args = append(args, v)
		}
	}
	if len(q.Provinces) > 0 {
		clauses = append(clauses, "sr.province IN ("+placeholders(len(q.Provinces))+")")
		for _, v := range q.Provinces {
			args = append(args, v)
		}
	}
	if len(q.Languages) > 0 {
		// 每个语言条件生成一个独立的 EXISTS，再用 OR 连接：
		// 只要学生"拥有其中任意一门符合条件的语言"就命中。
		// level 为空时表示“只要会这门语言，不限等级”，因此少一个占位符，
		// 对应地也就少 append 一个参数——两者必须同步增删。
		parts := make([]string, 0, len(q.Languages))
		for _, lang := range q.Languages {
			part := "EXISTS (SELECT 1 FROM stu_language WHERE stu_id = s.id AND language = ?"
			args = append(args, lang.Language)
			if lang.Level != "" {
				part += " AND level = ?"
				args = append(args, lang.Level)
			}
			part += ")"
			parts = append(parts, part)
		}
		clauses = append(clauses, "("+strings.Join(parts, " OR ")+")")
	}
	if len(q.CertNames) > 0 {
		// 证书是「任一命中」语义，所以 IN 列表之间的多个值是 OR 关系；
		// 与上面多语言用 OR 连接多个 EXISTS 的差别只是写法，语义一致。
		clauses = append(clauses, "EXISTS (SELECT 1 FROM stu_certificate WHERE stu_id = s.id AND cert_name IN ("+
			placeholders(len(q.CertNames))+"))")
		for _, v := range q.CertNames {
			args = append(args, v)
		}
	}
	if q.CertLevel != nil && *q.CertLevel != "" {
		clauses = append(clauses, "s.id IN (SELECT stu_id FROM stu_certificate WHERE level LIKE CONCAT('%', ?, '%'))")
		args = append(args, *q.CertLevel)
	}
	if q.InternshipCompany != nil && *q.InternshipCompany != "" {
		clauses = append(clauses, "s.id IN (SELECT stu_id FROM stu_internship WHERE company LIKE CONCAT('%', ?, '%'))")
		args = append(args, *q.InternshipCompany)
	}
	if len(q.Specialities) > 0 {
		clauses = append(clauses, "EXISTS (SELECT 1 FROM stu_speciality WHERE stu_id = s.id AND speciality IN ("+
			placeholders(len(q.Specialities))+"))")
		for _, v := range q.Specialities {
			args = append(args, v)
		}
	}
	if len(q.Jobs) > 0 {
		clauses = append(clauses, "EXISTS (SELECT 1 FROM stu_job WHERE stu_id = s.id AND post IN ("+
			placeholders(len(q.Jobs))+"))")
		for _, v := range q.Jobs {
			args = append(args, v)
		}
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

// SearchStudents 分页查询学生列表（精简字段），供列表页使用。
//
// base 里为什么用 SELECT DISTINCT：
//
//	为了支持按地区/语言/证书等子表条件筛选，这里把 4 张子表都 LEFT JOIN 了进来。
//	一对多关系会让一个学生在结果里出现多行（比如有 2 段实习就是 2 行），
//	DISTINCT 负责把 id 相同的重复行压成一条。
//	代价是 MySQL 需要额外做去重（可能用到临时表），数据量大时也可以改写成
//	「只 JOIN 需要筛选的那张表」或把条件改成 EXISTS 子查询来避免行放大。
//
// 分页同样是「COUNT 包一层子查询 + LIMIT/OFFSET」两步：
//   - COUNT 必须基于同一个 base，否则 total 和列表会对不上；
//   - 两次查询的参数是同一份 args（筛选条件），第二次再追加 pageSize/offset。
func (d *DB) SearchStudents(q model.StudentQuery) ([]model.Student, int, error) {
	where, args := studentWhere(q)
	base := `SELECT DISTINCT s.id, s.stu_id, s.name, s.age, s.gender, s.nation, s.major, s.degree,
		s.graduation_year, s.internship, s.user_id
		FROM student s
		LEFT JOIN stu_region sr ON s.id = sr.stu_id
		LEFT JOIN stu_language sl ON s.id = sl.stu_id
		LEFT JOIN stu_certificate sc ON s.id = sc.stu_id
		LEFT JOIN stu_internship si ON s.id = si.stu_id` + where

	var total int
	if err := d.QueryRow(`SELECT COUNT(*) FROM (`+base+`) t`, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	listArgs := append(append([]any{}, args...), q.PageSize, (q.PageNum-1)*q.PageSize)
	rows, err := d.Query(base+` ORDER BY s.id LIMIT ? OFFSET ?`, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []model.Student{}
	for rows.Next() {
		student, err := scanStudent(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *student)
	}
	return list, total, rows.Err()
}

// SearchStudentDetails 分页查询「学生详情卡片」：在学生基础信息之外，
// 再额外带出意向地区（省市）、最新简历地址，并在 fillStudentDetails 里补齐子表明细。
//
// 两处 SQL 技巧：
//
//  1. 地区用子查询 + GROUP_CONCAT(DISTINCT ...) 预先按 stu_id 聚合成一行，
//     避免主查询直接 JOIN stu_region 造成行放大。
//     用 GROUP_CONCAT 而不是“随便取一条”，是因为学生可能有多个意向城市，
//     这里把它们合并成 "北京,上海" 这样的字符串返回给前端展示。
//
//  2. 最新简历用窗口函数 ROW_NUMBER() OVER (PARTITION BY stu_id ORDER BY id DESC)：
//     先给每个学生的简历按 id 倒序编号，再在外层只保留 rn = 1 的那条，
//     即“每个学生最新的一份简历”。这比
//     「LEFT JOIN 一个 MAX(id) 子查询」可读性更好，也更容易扩展（如取最新 3 条）。
//     注意窗口函数需要 MySQL 8.0+。
//
// 排序用 ORDER BY s.id 而不是 created_at：
// id 是主键，排序稳定且能走索引，分页不会出现“同一条记录跨页重复/漏掉”的现象。
// （若按非唯一列排序，相同值的行在不同页之间的相对顺序是不确定的。）
func (d *DB) SearchStudentDetails(q model.StudentQuery) ([]model.StudentDetailDTO, int, error) {
	where, args := studentWhere(q)
	base := `SELECT s.id, s.stu_id, s.name, s.age, s.gender, s.nation, s.major, s.degree,
		s.graduation_year, s.internship, s.user_id, sr.province AS province, sr.city AS city, srsm.resume AS resume_url
		FROM student s
		LEFT JOIN (SELECT stu_id,
			GROUP_CONCAT(DISTINCT province SEPARATOR ',') AS province,
			GROUP_CONCAT(DISTINCT city SEPARATOR ',') AS city
			FROM stu_region GROUP BY stu_id) sr ON s.id = sr.stu_id
		LEFT JOIN (SELECT stu_id, resume, ROW_NUMBER() OVER (PARTITION BY stu_id ORDER BY id DESC) AS rn
			FROM stu_resume) srsm ON s.id = srsm.stu_id AND srsm.rn = 1` + where

	var total int
	if err := d.QueryRow(`SELECT COUNT(*) FROM (`+base+`) t`, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	listArgs := append(append([]any{}, args...), q.PageSize, (q.PageNum-1)*q.PageSize)
	rows, err := d.Query(base+` ORDER BY s.id LIMIT ? OFFSET ?`, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	details := []model.StudentDetailDTO{}
	for rows.Next() {
		var item model.StudentDetailDTO
		var name, gender, nation, major, degree, province, city, resume sql.NullString
		var age, graduation, internship, userID sql.NullInt64
		if err := rows.Scan(&item.ID, &item.StuID, &name, &age, &gender, &nation, &major, &degree,
			&graduation, &internship, &userID, &province, &city, &resume); err != nil {
			return nil, 0, err
		}
		item.Name = str(name)
		item.Age = intPtr(age)
		item.Gender = str(gender)
		item.Nation = str(nation)
		item.Major = str(major)
		item.Degree = str(degree)
		item.GraduationYear = intPtr(graduation)
		item.Internship = intPtr(internship)
		item.UserID = intPtr(userID)
		// 只有至少一个地区字段非 NULL 时才构造 Region 对象，
		// 否则前端会看到一个 {"province":null,"city":null} 的空壳，
		// 与“学生没填意向地区”的语义不符。
		if province.Valid || city.Valid {
			item.Region = &model.RegionDTO{Province: strPtr(province), City: strPtr(city)}
		}
		// 简历同理：地址为空字符串时视为“没上传”，保持 Resume 为 nil。
		if resume.Valid && resume.String != "" {
			item.Resume = &model.ResumeDTO{Resume: resume.String}
		}
		details = append(details, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if err := d.fillStudentDetails(details); err != nil {
		return nil, 0, err
	}
	return details, total, nil
}

// fillStudentDetails 批量填充语言、证书、实习、求职、技能等关联信息。
//
// 这是本文件最值得学的「消除 N+1 查询」写法：
//
//	如果对每个学生分别查 5 张子表，20 条记录就是 20×5 = 100 次查询；
//	这里改成「每张子表只查一次，用 IN (所有学生 id) 批量捞」，
//	总查询次数从 O(N) 降到 O(1)，固定为 5 次。
//
// 回填方式：
//   - 先建 index map[studentID]下标，把 student.id 映射到 details 切片的位置；
//   - 每读出一行子表数据，就通过 map 找到对应学生并 append 到它的切片里；
//   - 因为结果按 ORDER BY stu_id, id 返回，同一学生的明细天然有序。
//
// details 是 []model.StudentDetailDTO（值切片），通过 details[i] 直接修改元素字段
// 是可以生效的；但注意循环里 item := range 拿到的是副本，
// 所以代码里用 details[i] = item 先写回初始化过的空切片，避免出现 nil 切片序列化成 null。
func (d *DB) fillStudentDetails(details []model.StudentDetailDTO) error {
	if len(details) == 0 {
		return nil
	}
	ids := make([]int, len(details))
	for i, item := range details {
		ids[i] = item.ID
		item.Languages = []model.LanguageDTO{}
		item.Certificates = []model.CertificateDTO{}
		item.Internships = []model.InternshipDTO{}
		item.Jobs = []model.JobDTO{}
		item.Specialities = []model.SpecialityDTO{}
		details[i] = item
	}

	index := make(map[int]int, len(details))
	for i, item := range details {
		index[item.ID] = i
	}

	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	in := "(" + placeholders(len(ids)) + ")"

	// 接下来的 5 段查询共用同一个 in 片段：占位符个数一致，参数也复用 args，
	// 但每条 SQL 都重新绑定自己的列清单，互不影响。
	//
	// 这里刻意用显式 rows.Close() 而不是 defer：
	// 同一个函数里有多次 Query，若每段都 defer，5 个 Rows 会一直占着连接直到函数返回，
	// 连接池（上限 20）很容易被这种写法拖满。显式关闭能做到「用完即还」。
	rows, err := d.Query(`SELECT stu_id, language, level, cert_no, issue_date FROM stu_language
		WHERE stu_id IN `+in+` ORDER BY stu_id, id`, args...)
	if err != nil {
		return err
	}
	for rows.Next() {
		var stuID int
		var language, level, certNo sql.NullString
		var issueDate sql.NullTime
		if err := rows.Scan(&stuID, &language, &level, &certNo, &issueDate); err != nil {
			rows.Close()
			return err
		}
		if i, ok := index[stuID]; ok {
			details[i].Languages = append(details[i].Languages, model.LanguageDTO{
				StuID: stuID, Language: str(language), Level: str(level), CertNo: str(certNo), IssueDate: formatDate(issueDate),
			})
		}
	}
	rows.Close()

	rows, err = d.Query(`SELECT stu_id, cert_name, level, cert_no, issue_date FROM stu_certificate
		WHERE stu_id IN `+in+` ORDER BY stu_id, id`, args...)
	if err != nil {
		return err
	}
	for rows.Next() {
		var stuID int
		var certName, level, certNo sql.NullString
		var issueDate sql.NullTime
		if err := rows.Scan(&stuID, &certName, &level, &certNo, &issueDate); err != nil {
			rows.Close()
			return err
		}
		if i, ok := index[stuID]; ok {
			details[i].Certificates = append(details[i].Certificates, model.CertificateDTO{
				StuID: stuID, CertName: str(certName), Level: str(level), CertNo: str(certNo), IssueDate: formatDate(issueDate),
			})
		}
	}
	rows.Close()

	rows, err = d.Query(`SELECT stu_id, company, experience FROM stu_internship
		WHERE stu_id IN `+in+` ORDER BY stu_id, id`, args...)
	if err != nil {
		return err
	}
	for rows.Next() {
		var stuID int
		var company, experience sql.NullString
		if err := rows.Scan(&stuID, &company, &experience); err != nil {
			rows.Close()
			return err
		}
		if i, ok := index[stuID]; ok {
			details[i].Internships = append(details[i].Internships, model.InternshipDTO{
				StuID: stuID, Company: str(company), Experience: str(experience),
			})
		}
	}
	rows.Close()

	rows, err = d.Query(`SELECT stu_id, post FROM stu_job WHERE stu_id IN `+in+` ORDER BY stu_id, id`, args...)
	if err != nil {
		return err
	}
	for rows.Next() {
		var stuID int
		var post sql.NullString
		if err := rows.Scan(&stuID, &post); err != nil {
			rows.Close()
			return err
		}
		if i, ok := index[stuID]; ok {
			details[i].Jobs = append(details[i].Jobs, model.JobDTO{StuID: stuID, Post: str(post)})
		}
	}
	rows.Close()

	rows, err = d.Query(`SELECT stu_id, speciality FROM stu_speciality
		WHERE stu_id IN `+in+` ORDER BY stu_id, id`, args...)
	if err != nil {
		return err
	}
	for rows.Next() {
		var stuID int
		var speciality sql.NullString
		if err := rows.Scan(&stuID, &speciality); err != nil {
			rows.Close()
			return err
		}
		if i, ok := index[stuID]; ok {
			details[i].Specialities = append(details[i].Specialities, model.SpecialityDTO{StuID: stuID, Speciality: str(speciality)})
		}
	}
	rows.Close()

	return nil
}
