package api

// ==================================================================
// 本文件职责：招聘信息（职位）的 HTTP 接口，调用方主要是企业账号。
// ==================================================================
// 路由一览（都挂在登录后的 authorized 分组下）：
//   POST   /com/jobPosting             发布职位
//   PUT    /com/jobPosting/:id         修改职位
//   PATCH  /com/jobPosting/:id/status  上架 / 下架
//   DELETE /com/jobPosting/:id         删除职位
//   GET    /com/jobPostings            分页查询职位列表
//   GET    /com/jobPosting/:id         职位详情
//   GET    /com/myJobPostings          我发布的职位
//
// 【本层的统一套路】
//  1. 解析参数：路径参数用 c.Param，查询串用 queryInt/queryString，
//     请求体用 decodeJSON（内部是 Gin 的 ShouldBindJSON）；
//  2. 取当前用户：userID(c) 来自鉴权中间件写入的上下文，不信任前端传的 userId；
//  3. 调用 service：HTTP 层不写 SQL，只做参数搬运；
//  4. 统一响应：ok() 包装成 {code:200,message,data}，错误统一交给 a.fail()
//     （它会把 *model.AppError 翻译成对应业务码）。
// ==================================================================

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"resume-backend/internal/model"
)

// publishJob 发布招聘信息。
//
//	POST /com/jobPosting
//
// 请求体：JSON，字段兼容 comId/com_id、externalLink/external_link（见 model.JobPostingInput）。
// 调用：Service.PublishJob(ctx, 当前登录 userID, input)
// 响应：data 为新建的 model.JobPosting（序列化时同时输出驼峰与下划线两套字段）。
func (a *API) publishJob(c *gin.Context) {
	var input model.JobPostingInput
	if err := decodeJSON(c, &input); err != nil {
		a.fail(c, err)
		return
	}
	job, err := a.Service.PublishJob(c.Request.Context(), userID(c), input)
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, job)
}

// updateJob 修改指定职位。
//
//	PUT /com/jobPosting/:id
//
// 请求参数：路径 :id（Atoi 失败直接返回 400，避免把非法字符串传给 service）。
// 请求体：同发布接口，字段是指针，没传的字段保持原值。
// 调用：Service.UpdateJob(ctx, userID, id, input)，service 内部还会校验该职位是否属于当前企业。
// 响应：data 为更新后的 model.JobPosting。
func (a *API) updateJob(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respond(c, model.Error(400, "招聘信息ID格式错误"))
		return
	}
	var input model.JobPostingInput
	if err := decodeJSON(c, &input); err != nil {
		a.fail(c, err)
		return
	}
	job, svcErr := a.Service.UpdateJob(c.Request.Context(), userID(c), id, input)
	if svcErr != nil {
		a.fail(c, svcErr)
		return
	}
	ok(c, job)
}

// deleteJob 删除指定职位。
//
//	DELETE /com/jobPosting/:id
//
// 请求参数：仅路径 :id。
// 调用：Service.DeleteJob(ctx, userID, id) 返回 (是否删除成功, error)。
// 响应：data 为 {"success":bool,"message":string}。
// 注意：success=false 时仍然返回 code=200，是否删除成功要看 data.success 和 data.message。
func (a *API) deleteJob(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respond(c, model.Error(400, "招聘信息ID格式错误"))
		return
	}
	success, svcErr := a.Service.DeleteJob(c.Request.Context(), userID(c), id)
	if svcErr != nil {
		a.fail(c, svcErr)
		return
	}
	message := "删除失败"
	if success {
		message = "招聘信息已删除"
	}
	ok(c, map[string]any{"success": success, "message": message})
}

// jobPostings 分页查询职位列表（企业端“职位广场”/管理端查看）。
//
//	GET /com/jobPostings?pageNum=1&pageSize=10&status=&keyword=&comId=
//
// 请求参数：全部来自 URL 查询串；缺省时 pageNum=1、pageSize=10；
// comId 非法或未传时为 nil，表示不按企业过滤。
// 调用：Service.JobPostings(ctx, status, keyword, comID, pageNum, pageSize)
// 响应：data 为 model.PageInfo（total/list/pageNum 等，兼容 PageHelper）。
func (a *API) jobPostings(c *gin.Context) {
	pageNum := queryInt(c, "pageNum", 1)
	pageSize := queryInt(c, "pageSize", 10)
	status := queryString(c, "status")
	keyword := queryString(c, "keyword")
	var comID *int
	if raw := queryString(c, "comId"); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			comID = &value
		}
	}
	page, err := a.Service.JobPostings(c.Request.Context(), status, keyword, comID, pageNum, pageSize)
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, page)
}

// jobDetail 查看职位详情。
//
//	GET /com/jobPosting/:id
//
// 请求参数：路径 :id。
// 调用：Service.JobDetail(ctx, id)
// 响应：data 为 model.JobPosting；职位不存在时 service 返回 *model.AppError，由 a.fail 转成业务错误码。
func (a *API) jobDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respond(c, model.Error(400, "招聘信息ID格式错误"))
		return
	}
	job, svcErr := a.Service.JobDetail(c.Request.Context(), id)
	if svcErr != nil {
		a.fail(c, svcErr)
		return
	}
	ok(c, job)
}

// myJobPostings 查询当前登录企业自己发布的职位。
//
//	GET /com/myJobPostings?status=
//
// 请求参数：查询串 status（可选，用来筛选“招聘中/已下架”）。
// 调用：Service.MyJobPostings(userID, status) —— 注意这里不需要 context，属于轻量查询。
// 响应：data 为 []model.JobPosting（不分页，数量有限）。
func (a *API) myJobPostings(c *gin.Context) {
	status := queryString(c, "status")
	list, err := a.Service.MyJobPostings(userID(c), status)
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, list)
}

// updateJobStatus 上架 / 下架职位（只改状态，不改其他字段）。
//
//	PATCH /com/jobPosting/:id/status
//
// 请求参数：路径 :id；请求体为 {"status":"..."}。
// 这里先用 map[string]string 接住，再做非空校验，最后才调用 service。
// 调用：Service.UpdateJobStatus(ctx, userID, id, status)
// 响应：data 为更新后的 model.JobPosting。
func (a *API) updateJobStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respond(c, model.Error(400, "招聘信息ID格式错误"))
		return
	}
	var params map[string]string
	if err := decodeJSON(c, &params); err != nil {
		a.fail(c, err)
		return
	}
	status := strings.TrimSpace(params["status"])
	if status == "" {
		respond(c, model.Error(400, "状态不能为空"))
		return
	}
	job, svcErr := a.Service.UpdateJobStatus(c.Request.Context(), userID(c), id, status)
	if svcErr != nil {
		a.fail(c, svcErr)
		return
	}
	ok(c, job)
}
