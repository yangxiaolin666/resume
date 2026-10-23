package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"resume-backend/internal/config"
	"resume-backend/internal/model"
	"resume-backend/internal/service"
)

type API struct {
	Cfg     config.Config
	Service *service.Service
}

func NewRouter(cfg config.Config, svc *service.Service) http.Handler {
	api := &API{Cfg: cfg, Service: svc}
	registerChatWSHandler(svc)
	mux := http.NewServeMux()
	api.routes(mux)
	return api.withCORS(api.withAuth(mux))
}

func (a *API) routes(mux *http.ServeMux) {
	// 公共模块
	mux.HandleFunc("POST /login", a.login)
	mux.HandleFunc("POST /refresh-token", a.refreshToken)
	mux.HandleFunc("POST /validateToken", a.validateToken)

	// 当前用户（账号信息 + 昵称）
	mux.HandleFunc("GET /user/me", a.me)
	mux.HandleFunc("PUT /user/name", a.renameSelf)

	// 学生端
	mux.HandleFunc("GET /stu/getAllSkill", a.getAllSkill)
	mux.HandleFunc("PUT /stu/completeInfo", a.completeInfo)
	mux.HandleFunc("POST /stu/uploadResume", a.uploadResume)
	mux.HandleFunc("GET /stu/statistics", a.statistics)
	mux.HandleFunc("PATCH /stu/updatePwd", a.updatePwd)

	// 企业端
	mux.HandleFunc("GET /com/QueryComIdByUserId", a.queryComID)
	mux.HandleFunc("POST /com/LikeAdd", a.likeAdd)
	mux.HandleFunc("POST /com/LikeQuery", a.likeQuery)
	mux.HandleFunc("DELETE /com/LikeDelete", a.likeDelete)
	mux.HandleFunc("GET /com/LikeIf", a.likeIf)
	mux.HandleFunc("GET /com/LikeIfByRedis", a.likeIfByRedis)
	mux.HandleFunc("POST /com/search", a.searchStudents)
	mux.HandleFunc("POST /com/detailSearch", a.searchStudentDetails)
	mux.HandleFunc("POST /com/jobPosting", a.publishJob)
	mux.HandleFunc("PUT /com/jobPosting/{id}", a.updateJob)
	mux.HandleFunc("DELETE /com/jobPosting/{id}", a.deleteJob)
	mux.HandleFunc("GET /com/jobPostings", a.jobPostings)
	mux.HandleFunc("GET /com/jobPosting/{id}", a.jobDetail)
	mux.HandleFunc("GET /com/myJobPostings", a.myJobPostings)
	mux.HandleFunc("PATCH /com/jobPosting/{id}/status", a.updateJobStatus)

	// 管理端
	mux.HandleFunc("PUT /admin/updateStuPwd/{stu_id}", a.updateStuPwd)
	mux.HandleFunc("POST /admin/addStu", a.importStudents)
	mux.HandleFunc("POST /admin/addCom", a.importCompanies)
	mux.HandleFunc("DELETE /admin/delStu", a.deleteStudents)
	mux.HandleFunc("DELETE /admin/delCom", a.deleteCompanies)
	mux.HandleFunc("PUT /admin/updateStu", a.updateStudents)
	mux.HandleFunc("PUT /admin/updateCom", a.updateCompanies)
	mux.HandleFunc("GET /admin/searchStu", a.adminSearchStudents)
	mux.HandleFunc("GET /admin/searchCom", a.adminSearchCompanies)
	mux.HandleFunc("POST /admin/searchStuDetail", a.adminSearchStudentDetails)
	mux.HandleFunc("GET /admin/getAllLanguage", a.getAllLanguages)
	mux.HandleFunc("POST /admin/addLanguage", a.addLanguage)
	mux.HandleFunc("POST /admin/delLanguage", a.delLanguage)
	mux.HandleFunc("GET /admin/getAllCertificate", a.getAllCertificates)
	mux.HandleFunc("POST /admin/addCertificate", a.addCertificate)
	mux.HandleFunc("POST /admin/delCertificate", a.delCertificate)
	mux.HandleFunc("POST /admin/addSkillAll", a.addSkillAll)
	mux.HandleFunc("GET /admin/getAllSkill", a.getAllSkill)
	mux.HandleFunc("DELETE /admin/deleteSkill", a.deleteSkill)

	// 聊天（学生 <-> 企业）
	mux.HandleFunc("GET /chat/peers", a.chatPeers)
	mux.HandleFunc("GET /chat/conversations", a.chatConversations)
	mux.HandleFunc("GET /chat/messages", a.chatHistory)
	mux.HandleFunc("GET /chat/profile", a.chatPeerProfile)
	mux.HandleFunc("POST /chat/messages", a.chatSend)
	mux.HandleFunc("POST /chat/uploadImage", a.chatUploadImage)
	mux.HandleFunc("POST /chat/read", a.chatRead)
	mux.HandleFunc("GET /chat/ws", a.chatWebSocket)

	// 静态资源（简历）
	root := a.Cfg.StaticRoot()
	mux.Handle("/resume/uploads/", http.StripPrefix("/resume/uploads/", http.FileServer(http.Dir(root))))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(root))))
}

// ---------- 响应辅助 ----------

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respond(w http.ResponseWriter, result model.Result) {
	writeJSON(w, http.StatusOK, result)
}

func ok(w http.ResponseWriter, data any) {
	respond(w, model.Success(data))
}

func (a *API) fail(w http.ResponseWriter, err error) {
	var appErr *model.AppError
	if errors.As(err, &appErr) {
		respond(w, model.Error(appErr.Code, appErr.Msg))
		return
	}
	respond(w, model.Error(500, "服务器内部错误: "+err.Error()))
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func queryInt(r *http.Request, key string, fallback int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}

func queryString(r *http.Request, key string) string {
	return strings.TrimSpace(r.URL.Query().Get(key))
}
