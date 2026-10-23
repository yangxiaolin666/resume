package api

import (
	"net/http"
	"strconv"
	"strings"

	"resume-backend/internal/model"
)

func (a *API) publishJob(w http.ResponseWriter, r *http.Request) {
	var input model.JobPostingInput
	if err := decodeJSON(r, &input); err != nil {
		a.fail(w, err)
		return
	}
	job, err := a.Service.PublishJob(userID(r), input)
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, job)
}

func (a *API) updateJob(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respond(w, model.Error(400, "招聘信息ID格式错误"))
		return
	}
	var input model.JobPostingInput
	if err := decodeJSON(r, &input); err != nil {
		a.fail(w, err)
		return
	}
	job, svcErr := a.Service.UpdateJob(userID(r), id, input)
	if svcErr != nil {
		a.fail(w, svcErr)
		return
	}
	ok(w, job)
}

func (a *API) deleteJob(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respond(w, model.Error(400, "招聘信息ID格式错误"))
		return
	}
	success, svcErr := a.Service.DeleteJob(userID(r), id)
	if svcErr != nil {
		a.fail(w, svcErr)
		return
	}
	message := "删除失败"
	if success {
		message = "招聘信息已删除"
	}
	ok(w, map[string]any{"success": success, "message": message})
}

func (a *API) jobPostings(w http.ResponseWriter, r *http.Request) {
	pageNum := queryInt(r, "pageNum", 1)
	pageSize := queryInt(r, "pageSize", 10)
	status := queryString(r, "status")
	keyword := queryString(r, "keyword")
	var comID *int
	if raw := queryString(r, "comId"); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			comID = &value
		}
	}
	page, err := a.Service.JobPostings(status, keyword, comID, pageNum, pageSize)
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, page)
}

func (a *API) jobDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respond(w, model.Error(400, "招聘信息ID格式错误"))
		return
	}
	job, svcErr := a.Service.JobDetail(id)
	if svcErr != nil {
		a.fail(w, svcErr)
		return
	}
	ok(w, job)
}

func (a *API) myJobPostings(w http.ResponseWriter, r *http.Request) {
	status := queryString(r, "status")
	list, err := a.Service.MyJobPostings(userID(r), status)
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, list)
}

func (a *API) updateJobStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respond(w, model.Error(400, "招聘信息ID格式错误"))
		return
	}
	var params map[string]string
	if err := decodeJSON(r, &params); err != nil {
		a.fail(w, err)
		return
	}
	status := strings.TrimSpace(params["status"])
	if status == "" {
		respond(w, model.Error(400, "状态不能为空"))
		return
	}
	job, svcErr := a.Service.UpdateJobStatus(userID(r), id, status)
	if svcErr != nil {
		a.fail(w, svcErr)
		return
	}
	ok(w, job)
}
