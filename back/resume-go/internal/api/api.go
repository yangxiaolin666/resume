package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"resume-backend/internal/config"
	"resume-backend/internal/model"
	"resume-backend/internal/service"
)

// API 聚合 HTTP 层需要的依赖。
type API struct {
	Cfg     config.Config
	Service *service.Service
}

// NewRouter 构建 Gin 引擎。
//
// 【Gin 的核心概念】
//   - Engine：整个 HTTP 服务，负责路由匹配；
//   - RouterGroup：路由分组，可以给一组路由统一挂中间件；
//   - Context：一次请求的上下文，参数、响应、中间件数据都通过它传递。
//
// 【和标准库的区别】
// 之前用的是 net/http 的 ServeMux，需要自己写 CORS、鉴权、路径参数解析；
// Gin 把这些都内置了，并且路由分组能让「公开接口 / 登录接口 / 管理员接口」层次更清晰。
func NewRouter(cfg config.Config, svc *service.Service) *gin.Engine {
	gin.SetMode(gin.ReleaseMode) // 减少调试日志；需要排查路由时改成 gin.DebugMode
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), corsMiddleware())

	api := &API{Cfg: cfg, Service: svc}
	registerChatWSHandler(svc) // WebSocket 收到消息后的业务回调

	api.registerRoutes(engine)
	return engine
}

// registerRoutes 集中声明所有路由。
// 路由分组的好处：一眼能看出哪些接口需要登录、哪些仅管理员可用。
func (a *API) registerRoutes(engine *gin.Engine) {
	// ---------- 公开接口（无需 Token） ----------
	engine.POST("/login", a.login)
	engine.POST("/refresh-token", a.refreshToken)
	engine.POST("/validateToken", a.validateToken)
	engine.GET("/stu/statistics", a.statistics) // 首页统计数据，允许游客查看

	// WebSocket 握手无法携带 Authorization 头，单独在 handler 里校验 ?token=
	engine.GET("/chat/ws", a.chatWebSocket)

	// ---------- 登录后可用 ----------
	authorized := engine.Group("/", a.authMiddleware())
	{
		authorized.GET("/user/me", a.me)
		authorized.PUT("/user/name", a.renameSelf)

		// 学生端
		authorized.GET("/stu/getAllSkill", a.getAllSkill)
		authorized.PUT("/stu/completeInfo", a.completeInfo)
		authorized.POST("/stu/uploadResume", a.uploadResume)
		authorized.GET("/stu/resume/latest", a.latestResume)
		authorized.PATCH("/stu/updatePwd", a.updatePwd)
		authorized.POST("/stu/applications", a.applyJob)
		authorized.GET("/stu/applications", a.myApplications)
		authorized.DELETE("/stu/applications/:id", a.withdrawApplication)

		// 企业端
		authorized.GET("/com/QueryComIdByUserId", a.queryComID)
		authorized.POST("/com/LikeAdd", a.likeAdd)
		authorized.POST("/com/LikeQuery", a.likeQuery)
		authorized.DELETE("/com/LikeDelete", a.likeDelete)
		authorized.GET("/com/LikeIf", a.likeIf)
		authorized.GET("/com/LikeIfByRedis", a.likeIfByRedis)
		authorized.POST("/com/search", a.searchStudents)
		authorized.POST("/com/detailSearch", a.searchStudentDetails)
		authorized.POST("/com/jobPosting", a.publishJob)
		authorized.PUT("/com/jobPosting/:id", a.updateJob)
		authorized.DELETE("/com/jobPosting/:id", a.deleteJob)
		authorized.GET("/com/jobPostings", a.jobPostings)
		authorized.GET("/com/jobPosting/:id", a.jobDetail)
		authorized.GET("/com/myJobPostings", a.myJobPostings)
		authorized.PATCH("/com/jobPosting/:id/status", a.updateJobStatus)
		authorized.GET("/com/applications", a.companyApplications)
		authorized.PATCH("/com/applications/:id/status", a.updateApplicationStatus)

		// 聊天（学生 <-> 企业）
		authorized.GET("/chat/peers", a.chatPeers)
		authorized.GET("/chat/conversations", a.chatConversations)
		authorized.GET("/chat/messages", a.chatHistory)
		authorized.GET("/chat/profile", a.chatPeerProfile)
		authorized.POST("/chat/messages", a.chatSend)
		authorized.POST("/chat/uploadImage", a.chatUploadImage)
		authorized.POST("/chat/read", a.chatRead)
		authorized.GET("/chat/resume-permission", a.chatResumePermission)
		authorized.GET("/chat/notifications", a.chatNotifications)
		authorized.POST("/chat/notifications/read", a.chatNotificationsRead)
		authorized.GET("/chat/notifications/unread", a.chatUnreadCount)
	}

	// ---------- 管理端（先登录，再校验 admin 角色） ----------
	admin := engine.Group("/admin", a.authMiddleware(), a.adminMiddleware())
	{
		admin.PUT("/updateStuPwd/:stu_id", a.updateStuPwd)
		admin.POST("/addStu", a.importStudents)
		admin.POST("/addCom", a.importCompanies)
		admin.DELETE("/delStu", a.deleteStudents)
		admin.DELETE("/delCom", a.deleteCompanies)
		admin.PUT("/updateStu", a.updateStudents)
		admin.PUT("/updateCom", a.updateCompanies)
		admin.GET("/searchStu", a.adminSearchStudents)
		admin.GET("/searchCom", a.adminSearchCompanies)
		admin.POST("/searchStuDetail", a.adminSearchStudentDetails)
		admin.GET("/getAllLanguage", a.getAllLanguages)
		admin.POST("/addLanguage", a.addLanguage)
		admin.POST("/delLanguage", a.delLanguage)
		admin.GET("/getAllCertificate", a.getAllCertificates)
		admin.POST("/addCertificate", a.addCertificate)
		admin.POST("/delCertificate", a.delCertificate)
		admin.POST("/addSkillAll", a.addSkillAll)
		admin.GET("/getAllSkill", a.getAllSkill)
		admin.DELETE("/deleteSkill", a.deleteSkill)
	}

	// ---------- 静态资源 ----------
	// 简历文件在 uploads/resume，聊天图片在 uploads/chat；
	// 这里把整个 uploads 目录映射出去，同时保留旧的 /resume/uploads 前缀。
	root := a.Cfg.StaticRoot()
	engine.Static("/uploads", root)
	engine.Static("/resume/uploads", root)

	engine.NoRoute(func(c *gin.Context) {
		respond(c, model.Error(404, "接口不存在"))
	})
}

// ==================================================================
// 统一响应辅助
// ==================================================================
// 所有接口都返回 {code, message, data}，前端只需要判断 code。
// 这样控制器里就不用反复写 c.JSON(...)。

func respond(c *gin.Context, result model.Result) {
	c.JSON(http.StatusOK, result)
}

func ok(c *gin.Context, data any) {
	respond(c, model.Success(data))
}

// fail 把错误翻译成用户可读的响应。
// 业务错误（*model.AppError）直接返回其 code/msg；
// JSON 解析错误给 400；其他未知错误记录日志并返回 500，避免泄露 SQL 等内部信息。
func (a *API) fail(c *gin.Context, err error) {
	var appErr *model.AppError
	if errors.As(err, &appErr) {
		respond(c, model.Error(appErr.Code, appErr.Msg))
		return
	}
	var typeErr *json.UnmarshalTypeError
	var syntaxErr *json.SyntaxError
	if errors.As(err, &typeErr) || errors.As(err, &syntaxErr) || errors.Is(err, io.EOF) {
		respond(c, model.Error(400, "提交的数据格式不正确，请刷新页面后重试"))
		return
	}
	log.Printf("接口处理失败: %v", err)
	respond(c, model.Error(500, "服务器开小差了，请稍后重试"))
}

// decodeJSON 解析请求体。
// Gin 的 ShouldBindJSON 会根据 Content-Type 选择绑定方式，并复用 sync.Pool，性能更好。
func decodeJSON(c *gin.Context, target any) error {
	return c.ShouldBindJSON(target)
}

// queryInt 读取整型查询参数，缺省或非法时返回默认值。
func queryInt(c *gin.Context, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}
	number, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return number
}

func queryString(c *gin.Context, key string) string {
	return strings.TrimSpace(c.Query(key))
}
