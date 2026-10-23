// 本文件负责 job_postings 表（企业发布的职位）的增删改查。
//
// 业务定位：企业端「发布/编辑/上下架职位」、学生端「职位列表与搜索」、
// 后台「职位管理」。调用方是 internal/service 下的 job 服务。
//
// job_postings.com_id 指向 company.id（业务主键），不是 user_id；
// 查询时 LEFT JOIN company 既为了拿到企业名，也为了把 c.user_id 带出来做权限判断
// （service 层用它校验「这条职位是不是当前登录企业发的」）。

package data

import (
	"database/sql"
	"errors"
	"strings"

	"resume-backend/internal/model"
)

// JobFilter 是职位列表的筛选条件。
//
// 所有字段都用指针而不是值类型，是为了区分两种语义：
//   - nil  -> 这个条件根本没传，不要出现在 WHERE 里；
//   - 非 nil -> 用户明确传了条件，即使传的是空字符串也要按条件处理（见 buildJobWhere 里的判断）。
//
// 如果用 string 值类型，就没法区分「没传 status」和「传了 status=""」。
type JobFilter struct {
	Status  *string
	Keyword *string
	ComID   *int
}

// jobSelect 职位列表/详情的公共查询主体。
//
// 使用 LEFT JOIN 而不是 INNER JOIN：即使 company 里对应记录缺失（脏数据），
// 职位本身仍然能查出来，只是企业名称为空，不会让整条记录凭空消失。
const jobSelect = `SELECT jp.id, jp.com_id, c.user_id, c.com_name, jp.title, jp.description,
	jp.external_link, jp.status, jp.created_at, jp.updated_at
	FROM job_postings jp LEFT JOIN company c ON jp.com_id = c.id`

// scanJob 把一行结果映射成 model.JobPosting。
//
// 因为用了 LEFT JOIN，c.user_id / c.com_name 可能是 NULL（企业不存在），
// 所以必须先 Scan 进 sql.NullInt64 / sql.NullString，
// 否则会报 "converting NULL to int is unsupported"。
// com_id / com_user_id 用 if Valid 判断后赋零值，其余文本字段用 str() 兜底成空串。
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

// buildJobWhere 根据筛选条件动态拼出 WHERE 子句，并同步收集对应的参数值。
//
// 这是「动态条件查询」的标准写法，两个要点：
//
//  1. 以 "1=1" 打底，后面统一用 " AND xxx" 追加。
//     这样不用判断「是不是第一个条件该用 WHERE 还是 AND」，
//     拼接逻辑简单且不容易出错（优化器会直接消掉 1=1，不产生额外开销）。
//
//  2. SQL 片段和参数值必须严格一一对应、顺序一致：
//     clauses 追加一个 "?"，args 就追加一个值，最终按顺序展开。
//     SQL 文本里永远是 ?，用户输入永远只出现在 args 里——这就是参数化查询，
//     它把「代码」和「数据」彻底分开，从根本上杜绝 SQL 注入。
//
// 返回 (where 子句, 参数切片)，注意返回的 where 以 " WHERE " 开头，
// 因为调用方普遍是 jobSelect + where 这样直接拼接。
//
// LIKE 特意写成 CONCAT('%', ?, '%')：把通配符放在 SQL 侧，
// 避免在 Go 里拼字符串（那样既不安全也容易漏转义）。
// 代价是前置 % 会让 idx_status_created 之外的 title 索引失效，
// 数据量大时应改用全文索引或搜索引擎，这里数据规模小可以接受。
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

// ListJobPostings 分页查询职位列表，同时返回总条数和当前页数据。
//
// 为什么分两次查询：
//
//	前端分页组件需要 total 才能算出总页数，而 LIMIT 只返回当前页，
//	拿不到总数，所以必须单独 COUNT 一次。
//
// 为什么 COUNT 要写成 SELECT COUNT(*) FROM (原查询) t：
//
//	原查询带 LEFT JOIN，直接 COUNT(*) 数的是「join 后的行数」；
//	虽然这里 join 的是一对一关系不会放大，但用子查询包一层语义最稳妥，
//	拿到的就是「最终结果集的条数」，和列表逻辑永远保持一致。
//
// 分页原理：
//
//	LIMIT pageSize OFFSET (pageNum-1)*pageSize
//	即跳过前 (pageNum-1) 页，取 pageSize 条。pageNum 从 1 开始。
//
// 深分页问题（面试常问）：
//
//	OFFSET 并不是「直接跳到第 N 行」，MySQL 仍要扫描并丢弃前面所有行，
//	所以页码越大越慢（例如 OFFSET 100000 要白扫 10 万行）。
//	优化手段是「游标分页/keyset pagination」：
//	WHERE (created_at, id) < (上一页最后一条的时间, id) ORDER BY ... LIMIT n，
//	每次都能用索引定位起点，耗时与页码无关。本项目数据量不大，保留 OFFSET 更直观。
//
// ORDER BY jp.created_at DESC 正好能利用 idx_status_created(status, created_at)：
// 当筛选 status 时，索引里 status 相等的一段内部已按 created_at 有序，
// MySQL 可以直接顺序读取、省掉 filesort（索引的「最左前缀 + 有序性」）。
func (d *DB) ListJobPostings(filter JobFilter, pageNum, pageSize int) ([]model.JobPosting, int, error) {
	where, args := buildJobWhere(filter)
	var total int
	if err := d.QueryRow(`SELECT COUNT(*) FROM (`+jobSelect+where+`) t`, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := jobSelect + where + " ORDER BY jp.created_at DESC LIMIT ? OFFSET ?"
	// 先复制一份 args 再追加分页参数：直接 append 到原切片可能复用底层数组，
	// 影响后续（或并发）对 args 的使用。分页参数同样走占位符，不拼字符串。
	listArgs := append(append([]any{}, args...), pageSize, (pageNum-1)*pageSize)
	rows, err := d.Query(query, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	// rows 持有数据库连接，必须关闭，否则连接池很快被耗尽；defer 保证异常路径也能释放。
	defer rows.Close()
	// 初始化成空切片而不是 nil：JSON 序列化时空切片是 []，nil 是 null，
	// 前端就不用额外判空。
	list := []model.JobPosting{}
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *job)
	}
	// 遍历结束后必须再检查 rows.Err()：网络中断等错误只在迭代结束时才暴露，
	// 只看循环有没有报错会漏掉。
	return list, total, rows.Err()
}

// ListJobPostingsByComID 查某个企业下的职位（企业端「我的职位」）。
// 复用 buildJobWhere，把 ComID 和 Status 作为条件传入；不分页，一次取完。
// 条件里的 jp.com_id 可以走 idx_com_id 索引，再按 created_at 排序。
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

// FindJobPostingByID 按主键查职位详情，详情页和「编辑前回显」都用它。
func (d *DB) FindJobPostingByID(id int) (*model.JobPosting, error) {
	return scanJob(d.QueryRow(jobSelect+` WHERE jp.id = ?`, id))
}

// InsertJobPosting 新建职位，返回自增 id。
//
// created_at / updated_at 由数据库 NOW() 生成，保证时间以数据库为准；
// 所有入参（标题、描述、外链、状态）都是用户输入，一律用 ? 绑定。
// external_link 是外部链接文本，即使内容里有引号也不会有注入风险。
func (d *DB) InsertJobPosting(comID int, title, description, link, status string) (int, error) {
	res, err := d.Exec(`INSERT INTO job_postings (com_id, title, description, external_link, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())`, comID, title, description, link, status)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

// UpdateJobPosting 局部更新职位：只更新 input 里非 nil 的字段。
//
// 这是「动态 SET」的标准写法：
//   - sets 收集 "列 = ?" 片段，args 按相同顺序收集值；
//   - 始终带上 updated_at = NOW()，方便排查修改时间；
//   - 最后 append(id) 对应 WHERE id = ?，它必须是 args 的最后一个参数，
//     因为 SQL 文本里 WHERE 的 ? 排在所有 SET 的 ? 之后，占位符按出现顺序绑定。
//
// 如果 input 全是 nil，最终 SQL 仍然合法（只更新 updated_at），
// service 层可以再决定是否直接短路返回。
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

// UpdateJobPostingStatus 只改状态（上架/下架），是 UpdateJobPosting 的特化版本。
// 单独抽出来是因为上下架是高频操作，只更新一列能减少写入量、避免误改其他字段。
func (d *DB) UpdateJobPostingStatus(id int, status string) error {
	_, err := d.Exec(`UPDATE job_postings SET status = ?, updated_at = NOW() WHERE id = ?`, status, id)
	return err
}

// DeleteJobPosting 物理删除职位，返回受影响行数。
// 调用方通过 RowsAffected == 0 判断「职位不存在」，而不是先查一次再删。
func (d *DB) DeleteJobPosting(id int) (int64, error) {
	res, err := d.Exec(`DELETE FROM job_postings WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
