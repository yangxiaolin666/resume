// 本文件负责 com_like_stu 表：企业「收藏/意向」学生，以及企业端「我收藏的学生」列表。
//
// 业务定位：企业浏览学生简历时点「心仪/收藏」，对应简历项目里的「收藏与批量操作」；
// 调用方是 internal/service/like.go（AddLikes / DeleteLikes / FavoriteStudents）。
// 同时在服务端做了一层 Redis 缓存，数据层只保证 MySQL 里的最终状态。
//
// 表上有两个关键索引：
//   - UNIQUE KEY uk_stu_com(stu_id, com_id)：同一企业对同一学生只能收藏一次，
//     也是并发场景下防重复的兜底；
//   - idx_com_time(com_id, create_time)：支撑「某企业的收藏列表按时间倒序分页」。

package data

import (
	"database/sql"
	"errors"
	"strings"

	"resume-backend/internal/model"
)

// LikePair 用「学生 + 企业」这一对 ID 唯一标识一条收藏关系。
// 批量接口传入的是若干 pair，数据层据此做批量查重、批量插入、批量删除。
type LikePair struct {
	ComID int
	StuID int
}

// ExistsStudent 判断学生是否存在，用于收藏前的外键校验，避免写入脏 ID。
//
// SELECT COUNT(*) > 0 在 MySQL 里返回 0/1，驱动可以直接 Scan 进 Go 的 bool。
// 这比 COUNT(*) 再比较更贴合语义，但要注意 COUNT 仍然要扫描匹配行；
// 这里 id 是主键，命中即停，代价很低。
func (d *DB) ExistsStudent(id int) (bool, error) {
	var exists bool
	err := d.QueryRow(`SELECT COUNT(*) > 0 FROM student WHERE id = ?`, id).Scan(&exists)
	return exists, err
}

// ExistsCompany 同上，校验企业 ID 是否真实存在。
func (d *DB) ExistsCompany(id int) (bool, error) {
	var exists bool
	err := d.QueryRow(`SELECT COUNT(*) > 0 FROM company WHERE id = ?`, id).Scan(&exists)
	return exists, err
}

// ExistingLikes 一次性查出 pairs 中「已经存在」的收藏关系。
//
// 为什么不用 for 循环逐个查（N+1 问题）：
//
//	批量收藏 50 个学生就是 50 次数据库往返，网络延迟会被放大 50 倍；
//	这里一条 SQL 就查完，这也是简历里「批量查询优化」的落点。
//
// SQL 用的是「行构造器」语法：(stu_id, com_id) IN ((?, ?), (?, ?), ...)。
//   - 先把每个 pair 转成 "(?, ?)" 文本（只是占位符，不含值）；
//   - 再把所有值按 stu_id、com_id 的顺序依次放进 args。
//
// MySQL 会把行构造器和 (stu_id, com_id) 上的唯一索引 uk_stu_com 做匹配，
// 相当于批量走索引点查，而不是全表扫描。
//
// 返回的 LikePair 只含 StuID/ComID，service 层靠它算出「哪些已存在、哪些需要插入」。
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

// InsertLikesTx 在事务里批量插入收藏记录。
//
// 形参是 *sql.Tx 而不是 *DB，说明「一次批量收藏要么全成功、要么全失败」：
// 事务由 service 层开启（s.DB.Begin），这里只执行语句。
// 典型的调用方式是：
//
//	tx, _ := db.Begin()
//	defer tx.Rollback()          // 提前 return / panic 时自动回滚，已 Commit 后再 Rollback 是无害的
//	... tx.Exec(...) ...
//	tx.Commit()                  // 成功才真正落盘
//
// 注意这里用的是普通 INSERT 而不是 INSERT IGNORE：
// 调用方已经先查过 ExistingLikes 过滤掉重复项，这里如果仍然撞上
// uk_stu_com 唯一索引（并发下可能发生），报错会让整个事务回滚，
// 属于「宁可失败也不写脏数据」的保守策略。
func (d *DB) InsertLikesTx(tx *sql.Tx, pairs []LikePair) error {
	for _, p := range pairs {
		if _, err := tx.Exec(`INSERT INTO com_like_stu (stu_id, com_id, create_time) VALUES (?, ?, NOW())`,
			p.StuID, p.ComID); err != nil {
			return err
		}
	}
	return nil
}

// DeleteLikesTx 在事务里批量取消收藏。
//
// 与 ExistingLikes 一样使用「行构造器 + IN」的批量语法，
// 一条 DELETE 删掉多组 (stu_id, com_id)，避免逐条 DELETE 的 N 次往返。
// 行数由外层 ExistingLikes 的结果决定，所以这里删的一定是确实存在的记录，
// 返回值（受影响行数）在这里被忽略，因为业务上的「取消数量」由上一层的集合差算出来。
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

// QueryComIDByUserID 由登录账号 user_id 查企业业务主键 company.id。
//
// 返回 *int 而不是 int：查不到时返回 nil，让 service 层能明确区分
// 「这个账号不是企业」和「企业 id 是 0」，避免用 0 当哨兵值。
// 这里把 sql.ErrNoRows 翻译成 (nil, nil)，属于数据层的常见约定。
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

// favoriteSelect 企业端「收藏的学生」列表查询，是本文件最复杂的一段 SQL。
//
// 结构是：主表 com_like_stu 驱动，再去拼学生的各类资料。
//
// 为什么子查询里要先 GROUP BY 再 JOIN：
//
//	一个学生可能有多条语言、技能、求职意向记录。如果直接
//	"com_like_stu JOIN stu_language JOIN stu_speciality ..."，
//	多张一对多表做 JOIN 会产生「笛卡尔积」——
//	3 种语言 × 4 个技能 = 12 行重复数据，而且每列的内容会互相串。
//	这里的做法是先把每张子表按 stu_id 聚合成一行
//	（GROUP_CONCAT 把多行值拼成 "英语、日语" 这样的字符串），
//	再按 stu_id 一对一 JOIN 回主表，行数就不会被放大。
//
// 外层 LEFT JOIN student 而不是 JOIN：
//
//	即使学生资料缺失，收藏记录也要能显示出来（只是字段为空），
//	便于排查脏数据，也不会让企业的收藏数量对不上。
//
// WHERE f.com_id = ? 只筛当前企业，配合 idx_com_time(com_id, create_time)
// 可以直接定位到该企业那一段索引，ORDER BY f.create_time DESC 也能顺着索引读，
// 避免额外的排序开销。
//
// 别名（AS stuId / userId …）沿用驼峰命名，方便与前端字段对齐。
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

// scanFavorite 映射一行收藏学生数据。
//
// 由于 favoriteSelect 里大量使用 LEFT JOIN，几乎所有字段都可能是 NULL，
// 所以全部先 Scan 进 sql.Null* 再转换。
// 注意这里没有处理 sql.ErrNoRows：它只被 Query 的 Rows 调用（列表场景），
// 列表在没数据时根本不会进入循环，不存在 ErrNoRows。
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

// FavoriteStudents 分页查询某企业收藏的学生，返回 (当前页列表, 总数)。
//
// 分页两连查：
//   - 第一次用 SELECT COUNT(*) FROM (favoriteSelect) t 拿总数
//     （包成子查询是因为原 SQL 带多表 JOIN，直接 COUNT 容易把 join 后的行数当成业务条数）；
//   - 第二次在同样的条件上追加 ORDER BY ... LIMIT ? OFFSET ? 取当前页。
//
// 两次查询的 args 分别是 [comID] 和 [comID, pageSize, offset]，
// 占位符按 SQL 中出现的顺序绑定，所以分页参数必须追加在筛选参数之后。
//
// OFFSET = (pageNum-1)*pageSize。深分页时 OFFSET 越大越慢（要扫描并丢弃前面的行），
// 这里是企业收藏列表，量级有限，用 OFFSET 足够；
// 如果数据量真的很大，可以改成基于 create_time/id 的游标分页。
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
