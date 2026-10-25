package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"resume-backend/internal/auth"
	"resume-backend/internal/model"
)

// ==================================================================
// Gin 中间件
// ==================================================================
// 中间件就是一个 func(*gin.Context)，在真正的 handler 之前执行：
//   c.Next()       -> 放行，继续执行后面的 handler / 中间件
//   c.Abort()      -> 中断，不再执行后续 handler（鉴权失败时用）
//
// 本项目有三个中间件：
//   1. CORS：解决前后端分离的跨域问题；
//   2. 认证：解析 JWT，把 userId / userType 放进 Context；
//   3. 管理员校验：在认证之后判断角色。
// ==================================================================

const (
	ctxUserID   = "userId"
	ctxUsername = "username"
	ctxUserType = "userType"
)

// corsMiddleware 允许前端跨域访问。
// 注意：Access-Control-Allow-Origin 不能同时写 "*" 和 credentials=true，
// 所以这里回显请求的 Origin。
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		// 浏览器带 credentials 时，Allow-Headers 不能用 *（会被当成字面量），
		// 必须显式列出自定义头，否则 Authorization 的预检会失败。
		// platform 是前端 axios 全局加的自定义头，漏掉它会导致所有请求被浏览器拦截，
		// 表现正是 Axios 报 "Network Error"。
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Refresh-Token, X-Requested-With, platform")
		// 让前端能读到自动续期的新 Token
		c.Header("Access-Control-Expose-Headers", "New-Token")
		c.Header("Access-Control-Max-Age", "3600")

		// 浏览器跨域预检请求，直接返回 204，不进入业务 handler
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// authMiddleware 解析 Authorization 头里的 JWT，同时负责无感续期：
//   - Token 快过期时，通过响应头 New-Token 下发新 Token，前端无感刷新；
//   - Token 无效直接返回 401。
func (a *API) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := auth.TokenFromHeader(c.GetHeader("Authorization"))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.Error(401, "未提供Token"))
			return
		}
		claims, err := a.Service.Auth.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.Error(401, "Token无效或已过期"))
			return
		}

		if auth.ExpiringSoon(claims) {
			fresh := a.Service.Auth.AccessToken(claims.UserID, claims.Username, claims.UserType)
			c.Header("New-Token", fresh)
		}

		// Context 是 Gin 在请求生命周期内共享数据的容器。
		// 必须用 c.Set 而不是把值挂在闭包变量上，否则并发请求会串数据。
		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxUsername, claims.Username)
		c.Set(ctxUserType, claims.UserType)
		c.Next()
	}
}

// adminMiddleware 必须在 authMiddleware 之后使用。
// 路由分组里写成 engine.Group("/admin", a.authMiddleware(), a.adminMiddleware())。
func (a *API) adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if userType(c) != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, model.Error(403, "权限不足，需要管理员权限"))
			return
		}
		c.Next()
	}
}

// 下面三个函数从 Gin Context 里取当前登录用户信息。
// 类型断言失败时返回零值，避免 handler 里到处写 if 判断。

func userID(c *gin.Context) int {
	value, exists := c.Get(ctxUserID)
	if !exists {
		return 0
	}
	id, _ := value.(int)
	return id
}

func userType(c *gin.Context) string {
	value, exists := c.Get(ctxUserType)
	if !exists {
		return ""
	}
	userType, _ := value.(string)
	return userType
}

func username(c *gin.Context) string {
	value, exists := c.Get(ctxUsername)
	if !exists {
		return ""
	}
	name, _ := value.(string)
	return name
}
