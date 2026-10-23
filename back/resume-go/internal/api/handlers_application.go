package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"resume-backend/internal/model"
)

// ==================================================================
// 简历投递相关 HTTP 接口
// ==================================================================
// 学生端：
//   POST   /stu/applications          投递职位
//   GET    /stu/applications          我的投递
//   DELETE /stu/applications/:id      撤回投递
//
// 企业端：
//   GET    /com/applications          收到的简历（可按 jobId/status 筛选）
//   PATCH  /com/applications/:id/status  修改投递状态
//
// handler 只负责参数搬运，业务校验都在 service/application.go。
// ==================================================================

func (a *API) applyJob(c *gin.Context) {
	var req model.ApplyJobRequest
	if err := decodeJSON(c, &req); err != nil {
		a.fail(c, err)
		return
	}
	application, err := a.Service.ApplyJob(c.Request.Context(), userID(c), req)
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, application)
}

func (a *API) myApplications(c *gin.Context) {
	list, err := a.Service.MyApplications(userID(c))
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, map[string]any{"list": list})
}

func (a *API) withdrawApplication(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, model.Error(400, "投递记录ID格式错误"))
		return
	}
	if err := a.Service.WithdrawApplication(userID(c), id); err != nil {
		a.fail(c, err)
		return
	}
	ok(c, nil)
}

func (a *API) companyApplications(c *gin.Context) {
	jobID := queryInt(c, "jobId", 0)
	status := queryString(c, "status")
	list, err := a.Service.CompanyApplications(userID(c), jobID, status)
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, map[string]any{"list": list})
}

func (a *API) updateApplicationStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, model.Error(400, "投递记录ID格式错误"))
		return
	}
	var req model.UpdateApplicationStatusRequest
	if err := decodeJSON(c, &req); err != nil {
		a.fail(c, err)
		return
	}
	application, svcErr := a.Service.UpdateApplicationStatus(c.Request.Context(), userID(c), id, req.Status)
	if svcErr != nil {
		a.fail(c, svcErr)
		return
	}
	ok(c, application)
}
