package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"resume-backend/internal/cache"
	"resume-backend/internal/data"
	"resume-backend/internal/model"
)

// ==================================================================
// 职位业务层（含 Redis 热点缓存）
// ==================================================================
// 简历对应点：「结合 Redis 缓存热点职位数据，减少重复查询」。
//
// 【与其它层的关系】
//   - api：接收筛选条件（status/keyword/comId/分页）并返回统一分页结构；
//   - data：真正的 SQL 查询，只有缓存未命中时才会被调用（回源）；
//   - cache：提供 GetOrLoad（读缓存 + singleflight + 回源）和版本号；
//     本层不直接操作 redis 客户端，缓存细节都封装在 cache.Store 里。
//
// 【缓存策略】
//   - 职位列表：TTL 30 秒，key 里带完整查询条件，不同筛选互不干扰；
//   - 职位详情：TTL 5 分钟，单个职位被高频访问时收益明显；
//   - 一致性：职位发生增删改时让「缓存版本号 +1」，
//     读请求自然切换到新版本 key，旧数据等 TTL 到期自动淘汰。
//
// 为什么不直接 DEL 具体 key？
// 列表的查询条件组合几乎无穷多（状态、关键词、公司、页码……），
// 提前删不干净；版本号相当于给整类缓存打了一个「代际」标记，
// 是工程上更简单可靠的方案。
//
// 【缓存三大问题在本项目的处理】
//   - 穿透（查不存在的数据，请求每次都打到 MySQL）：
//     详情不存在时直接返回业务错误，不做空值缓存，避免“不存在的 id”占满缓存；
//   - 击穿（热点 key 过期瞬间大量并发回源）：
//     cache.GetOrLoad 内部用 singleflight，同一个 key 同时只有一个请求查库；
//   - 雪崩（大量 key 同一时刻过期）：
//     cache.Set 给 TTL 加最多 10% 的随机抖动，把过期时间打散。
// ==================================================================

const (
	jobListCacheTTL   = 30 * time.Second
	jobDetailCacheTTL = 5 * time.Minute
)

// jobListCache 是列表缓存的序列化载体。
// 不直接缓存 model.PageInfo，是因为它的 List 字段是 any，
// JSON 反序列化后会变成 []interface{}，类型信息丢失；
// 用一个具体结构体缓存，读出来仍然是 []model.JobPosting。
type jobListCache struct {
	List  []model.JobPosting `json:"list"`
	Total int                `json:"total"`
}

// jobCompany 把「登录用户」翻译成「企业档案」，多个方法都要用，所以抽出来。
// 用户存在但没有企业档案（例如账号类型不对、资料未初始化）时返回 404。
func (s *Service) jobCompany(userID int) (*model.Company, error) {
	company, err := s.DB.FindCompanyByUserID(userID)
	if err != nil {
		return nil, err
	}
	if company == nil {
		return nil, model.NewAppError(404, "企业信息不存在")
	}
	return company, nil
}

// PublishJob 发布职位。
//
// 业务规则：
//   - 必须是已建档的企业用户；
//   - 招聘标题必填（去空格后不能为空），职位描述、外链选填；
//   - 不传 status 时默认 active，即发布后立即可见。
//
// 成功后必须 BumpJobVersion：列表和详情缓存都带版本号，
// 版本 +1 后旧缓存立即“逻辑失效”，避免企业刚发布的职位查不到。
func (s *Service) PublishJob(ctx context.Context, userID int, input model.JobPostingInput) (*model.JobPosting, error) {
	company, err := s.DB.FindCompanyByUserID(userID)
	if err != nil {
		return nil, err
	}
	if company == nil {
		return nil, model.NewAppError(404, fmt.Sprintf("企业信息不存在（userId: %d）", userID))
	}
	if input.Title == nil || strings.TrimSpace(*input.Title) == "" {
		return nil, model.NewAppError(400, "招聘标题不能为空")
	}
	status := "active"
	if input.Status != nil {
		status = *input.Status
	}
	description, link := "", ""
	if input.Description != nil {
		description = *input.Description
	}
	if input.ExternalLink != nil {
		link = *input.ExternalLink
	}
	id, err := s.DB.InsertJobPosting(company.ID, *input.Title, description, link, status)
	if err != nil {
		return nil, err
	}
	// 新增职位后让缓存整体失效
	s.Cache.BumpJobVersion(ctx)
	return s.DB.FindJobPostingByID(id)
}

// UpdateJob 修改职位。
//
// 关键校验是「归属权」：先查出职位，确认 existing.ComID 属于当前企业，
// 否则任何企业都能通过猜 id 改别人的职位（越权修改），因此返回 403。
// input 里的字段都是指针，只更新传了的字段；若显式传了空标题则拒绝。
func (s *Service) UpdateJob(ctx context.Context, userID, id int, input model.JobPostingInput) (*model.JobPosting, error) {
	company, err := s.jobCompany(userID)
	if err != nil {
		return nil, err
	}
	existing, err := s.DB.FindJobPostingByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, model.NewAppError(500, "服务器内部错误: 招聘信息不存在")
	}
	if existing.ComID != company.ID {
		return nil, model.NewAppError(403, "无权限修改该招聘信息")
	}
	if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
		return nil, model.NewAppError(400, "招聘标题不能为空")
	}
	if err := s.DB.UpdateJobPosting(id, input); err != nil {
		return nil, err
	}
	s.Cache.BumpJobVersion(ctx)
	return s.DB.FindJobPostingByID(id)
}

// DeleteJob 删除职位。
// 同样先校验归属权；只有真正删掉行（rows > 0）才需要让缓存失效，
// 避免无效请求频繁 bump 版本号（每次 bump 都会让所有职位缓存冷启动）。
// 返回 bool 表示是否删除成功，供 api 层决定响应文案。
func (s *Service) DeleteJob(ctx context.Context, userID, id int) (bool, error) {
	company, err := s.jobCompany(userID)
	if err != nil {
		return false, err
	}
	existing, err := s.DB.FindJobPostingByID(id)
	if err != nil {
		return false, err
	}
	if existing == nil {
		return false, model.NewAppError(500, "服务器内部错误: 招聘信息不存在")
	}
	if existing.ComID != company.ID {
		return false, model.NewAppError(403, "无权限删除该招聘信息")
	}
	rows, err := s.DB.DeleteJobPosting(id)
	if err != nil {
		return false, err
	}
	if rows > 0 {
		s.Cache.BumpJobVersion(ctx)
	}
	return rows > 0, nil
}

// JobPostings 分页查询职位列表，走「缓存优先」。
//
// key 的构成：resume:job:v{版本号}:list:{status}|{keyword}|{comID}|{页码}|{页大小}。
//   - 版本号放在 key 里，是实现「整类缓存失效」的关键；
//   - 其余条件全部参与 key，保证不同筛选不会互相串数据；
//   - comID 用 %v 打印，nil 会变成 "<nil>"，同样能作为 key 的一部分（表示不限公司）。
//
// 流程：先读当前版本号 -> 拼 key -> GetOrLoad（未命中时执行闭包查 MySQL）。
// 缓存里存的是 jobListCache 而不是 PageInfo，原因见该结构体的注释。
func (s *Service) JobPostings(ctx context.Context, status, keyword string, comID *int, pageNum, pageSize int) (*model.PageInfo, error) {
	version := s.Cache.JobVersion(ctx)
	key := fmt.Sprintf("resume:job:v%d:list:%s|%s|%v|%d|%d",
		version, status, keyword, comID, pageNum, pageSize)

	filter := data.JobFilter{ComID: comID}
	if status != "" {
		filter.Status = &status
	}
	if keyword != "" {
		filter.Keyword = &keyword
	}

	cached, err := cache.GetOrLoad[jobListCache](ctx, s.Cache, key, jobListCacheTTL, func() (jobListCache, error) {
		list, total, err := s.DB.ListJobPostings(filter, pageNum, pageSize)
		if err != nil {
			return jobListCache{}, err
		}
		return jobListCache{List: list, Total: total}, nil
	})
	if err != nil {
		return nil, err
	}
	page := model.NewPageInfo(cached.List, cached.Total, pageNum, pageSize, len(cached.List))
	return &page, nil
}

// JobDetail 查询职位详情，走「缓存优先」。
//
// 详情是典型的「热点 key」：一个职位被很多人反复打开，缓存收益很高，
// 因此 TTL 设为 5 分钟（比列表长）。
// 详情不存在时返回 AppError，并且不缓存空值——这是防穿透的取舍：
// 空值缓存能挡住恶意刷不存在的 id，但会占用缓存空间，本项目选择简单处理。
func (s *Service) JobDetail(ctx context.Context, id int) (*model.JobPosting, error) {
	version := s.Cache.JobVersion(ctx)
	key := fmt.Sprintf("resume:job:v%d:detail:%d", version, id)

	job, err := cache.GetOrLoad[*model.JobPosting](ctx, s.Cache, key, jobDetailCacheTTL, func() (*model.JobPosting, error) {
		item, err := s.DB.FindJobPostingByID(id)
		if err != nil {
			return nil, err
		}
		if item == nil {
			return nil, model.NewAppError(500, "服务器内部错误: 招聘信息不存在")
		}
		return item, nil
	})
	if err != nil {
		return nil, err
	}
	return job, nil
}

// MyJobPostings 查询「我这家企业」发布的职位，供企业后台管理列表使用。
// 数据量小、且要求看到最新状态，所以直接查库不走缓存，避免自己改完自己看不到。
func (s *Service) MyJobPostings(userID int, status string) ([]model.JobPosting, error) {
	company, err := s.jobCompany(userID)
	if err != nil {
		return nil, err
	}
	return s.DB.ListJobPostingsByComID(status, company.ID)
}

// UpdateJobStatus 上架/下架职位。
//
// status 只允许 active / inactive，属于白名单校验：
// 如果直接把前端传来的字符串写库，拼错或恶意值会污染数据，后续筛选逻辑也会失效。
// 校验顺序是先归属权、再状态值，保证无权用户即使传了非法值也只会得到 403。
func (s *Service) UpdateJobStatus(ctx context.Context, userID, id int, status string) (*model.JobPosting, error) {
	company, err := s.jobCompany(userID)
	if err != nil {
		return nil, err
	}
	existing, err := s.DB.FindJobPostingByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, model.NewAppError(500, "服务器内部错误: 招聘信息不存在")
	}
	if existing.ComID != company.ID {
		return nil, model.NewAppError(403, "无权限修改该招聘信息")
	}
	if status != "active" && status != "inactive" {
		return nil, model.NewAppError(500, "服务器内部错误: 无效的状态值，只能是active或inactive")
	}
	if err := s.DB.UpdateJobPostingStatus(id, status); err != nil {
		return nil, err
	}
	s.Cache.BumpJobVersion(ctx)
	return s.DB.FindJobPostingByID(id)
}
