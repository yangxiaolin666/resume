package api

// ==================================================================
// 本文件职责：学生端的 HTTP 接口（技能字典、完善资料、简历上传、统计、改密）。
// ==================================================================
// 路由：
//   GET   /stu/getAllSkill    获取技能下拉选项（admin 端复用同一 handler）
//   PUT   /stu/completeInfo   一次性提交学生完整资料
//   POST  /stu/uploadResume   上传简历文件（multipart/form-data）
//   GET   /stu/statistics     首页统计图表数据（公开，游客也能看）
//   PATCH /stu/updatePwd      修改自己的密码
//
// 搜索、收藏等企业端接口在另外的文件里。
// 响应仍为 {code,message,data}；上传成功时特意返回 code=201 表示“已创建资源”。
// ==================================================================

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"resume-backend/internal/model"
)

// getAllSkill 获取全部技能选项（新增 / 编辑资料时的下拉框数据）。
//
//	GET /stu/getAllSkill（管理端也有 GET /admin/getAllSkill，指向同一个 handler）
//
// 请求参数：无。
// 调用：Service.GetAllSkills()
// 响应：data 为 []string。
func (a *API) getAllSkill(c *gin.Context) {
	skills, err := a.Service.GetAllSkills()
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, skills)
}

// completeInfo 学生一次性提交完整资料（基本信息 + 语言 + 证书 + 实习 + 地区 + 特长）。
//
//	PUT /stu/completeInfo
//
// 请求体：JSON -> model.StudentCompleteInfo。
// 该结构体的 UnmarshalJSON 兼容 specialities 的两种格式（字符串数组 / 对象数组）。
// 调用：Service.CompleteInfo(userID, dto)，内部用事务保证主表和子表一起更新。
// 响应：data 为更新后的 model.StudentDetailDTO。
func (a *API) completeInfo(c *gin.Context) {
	var dto model.StudentCompleteInfo
	if err := decodeJSON(c, &dto); err != nil {
		a.fail(c, err)
		return
	}
	detail, err := a.Service.CompleteInfo(userID(c), dto)
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, detail)
}

// uploadResume 上传简历文件。
//
//	POST /stu/uploadResume
//	Content-Type: multipart/form-data，文件字段名 resumeFile
//
// 【上传流程与内存控制】
//  1. ParseMultipartForm(12 << 20)：解析整个 multipart 请求，参数表示
//     “内存最多缓存 12MB 的文件内容”。超过阈值的部分不会继续占内存，
//     而是由 Go 标准库写入系统临时目录的文件中，最终避免大文件把进程内存打爆。
//     临时文件会在请求结束后由 net/http 自动清理（MultipartForm.RemoveAll）。
//  2. c.FormFile("resumeFile") 只拿到文件元信息（文件名、大小），并不会把内容读进内存；
//  3. header.Open() 得到一个流式 reader，直接交给 service 边读边写磁盘。
//
// 调用 Service.UploadResume(userID, 文件名, 大小, reader)。
// service 负责大小上限（10MB）、扩展名白名单、生成唯一文件名并落盘。
// 响应：HTTP 200 + {code:201,message:"简历上传成功",data:model.ResumeUploadVO}。
// code=201 表示“新资源已创建”，前端据此判断上传成功。
func (a *API) uploadResume(c *gin.Context) {
	// 限制 multipart 解析占用的内存，超出部分落到临时文件，避免大文件打爆内存。
	if err := c.Request.ParseMultipartForm(12 << 20); err != nil {
		respond(c, model.Error(500, "上传失败，请稍后重试"))
		return
	}
	header, err := c.FormFile("resumeFile")
	if err != nil {
		respond(c, model.Error(500, "请选择要上传的简历文件"))
		return
	}
	file, err := header.Open()
	if err != nil {
		respond(c, model.Error(500, "上传失败，请稍后重试"))
		return
	}
	defer file.Close()

	vo, svcErr := a.Service.UploadResume(userID(c), header.Filename, header.Size, file)
	if svcErr != nil {
		a.fail(c, svcErr)
		return
	}
	c.JSON(http.StatusOK, model.Result{Code: 201, Message: "简历上传成功", Data: vo})
}

// latestResume 返回当前学生最新上传的简历，供聊天里「发送简历」使用。
//
//	GET /stu/resume/latest（需登录）
//
// 响应：data 为 { url, name }；url 可直接用于下载/预览。
func (a *API) latestResume(c *gin.Context) {
	info, err := a.Service.LatestResume(userID(c))
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, info)
}

// statistics 返回首页 / 仪表盘用的学生统计图表数据。
//
//	GET /stu/statistics（公开接口，游客也能访问）
//
// 请求参数：无。
// 调用：Service.StudentStatistics()
// 响应：data 为 model.StudentStatistics，包含 languages / cities / internships 等分组计数。
func (a *API) statistics(c *gin.Context) {
	stats, err := a.Service.StudentStatistics()
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, stats)
}

// updatePwd 修改当前登录用户的密码。
//
//	PATCH /stu/updatePwd
//
// 请求体：{"old_pwd":"...","new_pwd":"...","re_pwd":"..."}（注意是下划线命名）。
// 这里先用 map[string]string 接收，再统一做非空校验；三项缺一都返回 400。
// 调用 Service.UpdatePwd(userID, oldPwd, newPwd, rePwd)；
// service 会校验原密码、新旧是否相同、两次新密码是否一致，并重新做 bcrypt 哈希。
// 响应：成功 data=null；失败由 a.fail 返回具体业务错误码。
func (a *API) updatePwd(c *gin.Context) {
	var params map[string]string
	if err := decodeJSON(c, &params); err != nil {
		a.fail(c, err)
		return
	}
	oldPwd, newPwd, rePwd := params["old_pwd"], params["new_pwd"], params["re_pwd"]
	if strings.TrimSpace(oldPwd) == "" || strings.TrimSpace(newPwd) == "" || strings.TrimSpace(rePwd) == "" {
		respond(c, model.Error(400, "参数不完整"))
		return
	}
	if err := a.Service.UpdatePwd(userID(c), oldPwd, newPwd, rePwd); err != nil {
		a.fail(c, err)
		return
	}
	ok(c, nil)
}
