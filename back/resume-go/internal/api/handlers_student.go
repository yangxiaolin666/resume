package api

import (
	"net/http"
	"strings"

	"resume-backend/internal/model"
)

func (a *API) getAllSkill(w http.ResponseWriter, r *http.Request) {
	skills, err := a.Service.GetAllSkills()
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, skills)
}

func (a *API) completeInfo(w http.ResponseWriter, r *http.Request) {
	var dto model.StudentCompleteInfo
	if err := decodeJSON(r, &dto); err != nil {
		a.fail(w, err)
		return
	}
	detail, err := a.Service.CompleteInfo(userID(r), dto)
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, detail)
}

func (a *API) uploadResume(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		respond(w, model.Error(500, "服务器内部错误: "+err.Error()))
		return
	}
	file, header, err := r.FormFile("resumeFile")
	if err != nil {
		respond(w, model.Error(500, "服务器内部错误: Required part 'resumeFile' is not present"))
		return
	}
	defer file.Close()

	vo, svcErr := a.Service.UploadResume(userID(r), header.Filename, header.Size, file)
	if svcErr != nil {
		a.fail(w, svcErr)
		return
	}
	writeJSON(w, http.StatusOK, model.Result{Code: 201, Message: "简历上传成功", Data: vo})
}

func (a *API) statistics(w http.ResponseWriter, r *http.Request) {
	stats, err := a.Service.StudentStatistics()
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, stats)
}

func (a *API) updatePwd(w http.ResponseWriter, r *http.Request) {
	var params map[string]string
	if err := decodeJSON(r, &params); err != nil {
		a.fail(w, err)
		return
	}
	oldPwd, newPwd, rePwd := params["old_pwd"], params["new_pwd"], params["re_pwd"]
	if strings.TrimSpace(oldPwd) == "" || strings.TrimSpace(newPwd) == "" || strings.TrimSpace(rePwd) == "" {
		respond(w, model.Error(400, "参数不完整"))
		return
	}
	if err := a.Service.UpdatePwd(userID(r), oldPwd, newPwd, rePwd); err != nil {
		a.fail(w, err)
		return
	}
	ok(w, nil)
}
