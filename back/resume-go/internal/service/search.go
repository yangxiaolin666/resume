package service

import "resume-backend/internal/model"

// ==================================================================
// 学生搜索业务层
// ==================================================================
// 提供两种粒度的搜索：
//   - SearchStudents：只要学生基础字段，适合列表页/管理端表格；
//   - SearchStudentDetails：额外带地区汇总和最新简历地址，适合需要展示简历卡片的页面。
//
// 【与其它层的关系】
//   - api：把 query 参数绑定成 StudentQuery；
//   - data：动态拼 WHERE、JOIN 子表、COUNT + LIMIT/OFFSET 分页；
//   - cache：这里不走缓存。搜索条件组合多、结果随数据变动频繁，
//     缓存命中率低且容易读到旧数据；热点职位才值得做缓存。
//
// 【对应简历能力】多条件组合查询与分页封装。
// ==================================================================

// SearchStudents 按条件分页搜索学生基础信息。
//
// Normalize 是必要的：前端/老接口可能传 stu_id 或 stuId、pageNum 或 page_num，
// 统一在这一层归一化，data 层就不用到处兼容。
// 返回统一 PageInfo，total 由 data 层单独 COUNT 得到，
// 保证分页器显示的总数与筛选条件一致。
func (s *Service) SearchStudents(q model.StudentQuery) (*model.PageInfo, error) {
	q.Normalize()
	list, total, err := s.DB.SearchStudents(q)
	if err != nil {
		return nil, err
	}
	page := model.NewPageInfo(list, total, q.PageNum, q.PageSize, len(list))
	return &page, nil
}

// SearchStudentDetails 分页搜索学生详情。
// 相比 SearchStudents，结果里多了省份/城市（多行合并）和最新一份简历的地址，
// 适合企业端浏览候选人卡片的场景。
// 同样先 Normalize 再查库，空结果也会返回 total=0 的正常分页结构，而不是 nil。
func (s *Service) SearchStudentDetails(q model.StudentQuery) (*model.PageInfo, error) {
	q.Normalize()
	list, total, err := s.DB.SearchStudentDetails(q)
	if err != nil {
		return nil, err
	}
	page := model.NewPageInfo(list, total, q.PageNum, q.PageSize, len(list))
	return &page, nil
}
