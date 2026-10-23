package api

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"

	"resume-backend/internal/auth"
	"resume-backend/internal/model"
)

// ==================================================================
// 本文件职责：登录、登录态维护与当前用户资料相关的 HTTP 接口。
// ==================================================================
// 路由：
//   POST /login          账号密码登录，签发 access token + refresh token（公开接口）
//   POST /refresh-token  用 Refresh-Token 头换取新的 access token（公开接口）
//   POST /validateToken  校验 token 是否有效（公开接口，老前端在用）
//   GET  /user/me        获取当前登录用户信息（需登录）
//   PUT  /user/name      修改自己的昵称（需登录）
//
// 所有响应依然是 {code,message,data}；成功统一走 ok()，失败走 a.fail()。
// 登录成功后的用户身份不再由前端传参，而是由鉴权中间件解析 token 后写入上下文，
// handler 里用 userID(c) 取出，避免“前端随便传个 userId 就能操作别人数据”。
// ==================================================================

// login 账号密码登录。
//
//	POST /login
//
// 请求体：JSON -> model.LoginDTO（username / password / userType，userType 也接受 user_type）。
// 参数校验：三个字段任一为空就一次性返回所有缺失项，前端可以合并提示。
// 调用：Service.Login(username, password, userType)
// 响应：data 为 model.LoginVO（userId、username、userType、displayName、token、refreshToken）。
func (a *API) login(c *gin.Context) {
	var dto model.LoginDTO
	if err := decodeJSON(c, &dto); err != nil {
		a.fail(c, err)
		return
	}
	missing := []string{}
	if strings.TrimSpace(dto.Username) == "" {
		missing = append(missing, "用户名不能为空")
	}
	if strings.TrimSpace(dto.Password) == "" {
		missing = append(missing, "密码不能为空")
	}
	if strings.TrimSpace(dto.Type()) == "" {
		missing = append(missing, "用户类型不能为空")
	}
	if len(missing) > 0 {
		respond(c, model.Error(400, "处理参数验证异常: "+strings.Join(missing, ", ")))
		return
	}

	vo, err := a.Service.Login(dto.Username, dto.Password, dto.Type())
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, vo)
}

// me 返回当前登录用户的账号名与昵称
//
//	GET /user/me
//
// 请求参数：无（用户身份来自 Authorization 头解析后的上下文）。
// 调用：Service.Me(userID)
// 响应：data 为 model.UserProfile（userId、username、userType、displayName）。
func (a *API) me(c *gin.Context) {
	profile, err := a.Service.Me(userID(c))
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, profile)
}

// renameSelf 修改自己的昵称（学生改姓名，企业改企业名称）
//
//	PUT /user/name
//
// 请求体：{"name":"新昵称"}，这里用匿名结构体接住，只关心一个字段。
// 调用：Service.Rename(userID, name)，具体改 student 表还是 company 表由 service 按用户类型判断。
// 响应：data 为修改后的 model.UserProfile。
func (a *API) renameSelf(c *gin.Context) {
	var dto struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(c, &dto); err != nil {
		a.fail(c, err)
		return
	}
	profile, err := a.Service.Rename(userID(c), dto.Name)
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, profile)
}

// refreshToken 用刷新令牌换新的访问令牌。
//
//	POST /refresh-token
//
// 请求参数：HTTP 头 Refresh-Token（不是请求体）。
// 为什么要单独放请求头：refresh token 不应该出现在 URL / body 日志里，
// 放头部更不容易被意外记录。
// 调用：Service.RefreshToken(refresh)
// 响应：data 为 {"token":"新的 access token"}；头缺失直接返回 401。
func (a *API) refreshToken(c *gin.Context) {
	refresh := c.GetHeader("Refresh-Token")
	if refresh == "" {
		respond(c, model.Error(401, "登录状态已失效，请重新登录"))
		return
	}
	token, err := a.Service.RefreshToken(refresh)
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, map[string]string{"token": token})
}

// validateToken 校验 token 是否仍然有效（给老前端 / 其他调用方使用）。
//
//	POST /validateToken
//
// 请求参数：优先从请求体 {"token":"..."} 读取；也兼容直接把原始 token 当请求体传。
// 之所以先用 GetRawData 再看是不是 JSON，是为了兼容旧调用方，不强制 Content-Type。
// 调用：Service.Auth.Parse(token) 解析 JWT claims，并用 auth.ExpiringSoon 判断是否快过期。
// 响应：data 为 {valid, message, userId, userType, needRefresh}；
// 注意 token 无效时仍然返回 code=200，只是 valid=false，前端据此决定跳登录页。
func (a *API) validateToken(c *gin.Context) {
	token := ""
	// 这里故意不要求 JSON：兼容直接传原始 token 的旧调用方。
	body, _ := c.GetRawData()
	if len(bytes.TrimSpace(body)) > 0 {
		var dto struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal(body, &dto); err == nil {
			token = dto.Token
		}
	}
	if token == "" {
		respond(c, model.Error(401, "Token 不能为空"))
		return
	}

	claims, err := a.Service.Auth.Parse(token)
	if err != nil {
		ok(c, map[string]any{"valid": false, "message": "Token 无效或已过期"})
		return
	}
	ok(c, map[string]any{
		"valid":       true,
		"message":     "Token 有效且未过期",
		"userId":      claims.UserID,
		"userType":    claims.UserType,
		"needRefresh": auth.ExpiringSoon(claims),
	})
}
