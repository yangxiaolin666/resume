// 本文件负责 company 表（企业资料）的查询。
//
// 业务定位：企业注册后填写公司名/所在地区；学生端职位列表要展示企业名；
// 后台管理要按名称检索企业。调用方主要是 internal/service 下的 auth / job / admin 服务。
//
// company.user_id 关联 user.user_id（一个账号对应一份企业资料），
// company.id 才是职位表 job_postings.com_id 引用的业务主键——
// 这两个 ID 混用是新手最容易出错的地方，读代码时留意区分。

package data

import (
	"database/sql"
	"errors"

	"resume-backend/internal/model"
)

// companyColumns 把列清单抽成常量，保证三处查询的列顺序完全一致，
// 这样 scanCompany 里的 Scan 参数顺序只需维护一份。
// 这里拼接的是代码里写死的常量（不是用户输入），所以不存在注入问题；
// 真正的筛选值一律走 ? 占位符。
const companyColumns = `id, com_name, com_region, user_id`

// scanCompany 把一行 company 记录映射成 model.Company。
//
// com_name / com_region / user_id 在库里都允许 NULL（历史数据可能不全），
// 因此统一先 Scan 进 sql.Null*，再用 str / intPtr 转成 Go 的零值或 nil 指针。
//
// 形参类型是 rowScanner 接口而不是 *sql.Row：
// 这样同一个函数既能处理 QueryRow 的单行结果，也能处理 Query 返回的 Rows 迭代，
// 避免为「列表」和「单条」各写一遍 Scan 逻辑。
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

// FindCompanyByID 按企业业务主键 id 查询，id 就是 job_postings.com_id 引用的那个值。
// 主键等值查询走聚簇索引，代价最低。
func (d *DB) FindCompanyByID(id int) (*model.Company, error) {
	return scanCompany(d.QueryRow(`SELECT `+companyColumns+` FROM company WHERE id = ?`, id))
}

// FindCompanyByUserID 由登录账号 user_id 反查企业资料，用于「当前登录企业」的会话信息。
// 同一个 user_id 理论上只有一份企业资料；若表上没有唯一索引，重复数据会只取第一行，
// 所以业务层注册时必须保证一对一。
func (d *DB) FindCompanyByUserID(userID int) (*model.Company, error) {
	return scanCompany(d.QueryRow(`SELECT `+companyColumns+` FROM company WHERE user_id = ?`, userID))
}

// FindCompanyByName 按公司名精确匹配，主要给注册查重、企业搜索使用。
// 注意这里用的是 = 而不是 LIKE：等值比较可以吃到 com_name 上的索引；
// 如果改成 LIKE '%关键字%'，前置通配符会让 B+ 树索引失效，退化成全表扫描。
func (d *DB) FindCompanyByName(name string) (*model.Company, error) {
	return scanCompany(d.QueryRow(`SELECT `+companyColumns+` FROM company WHERE com_name = ?`, name))
}
