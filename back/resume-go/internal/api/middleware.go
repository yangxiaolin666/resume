package api

import (
	"context"
	"net/http"
	"strings"

	"resume-backend/internal/auth"
	"resume-backend/internal/model"
)

type ctxKey string

const (
	ctxUserID   ctxKey = "userId"
	ctxUsername ctxKey = "username"
	ctxUserType ctxKey = "userType"
)

func (a *API) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Max-Age", "3600")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// withAuth 合并了原 Java 的 TokenRefreshInterceptor + AuthInterceptor + AdminInterceptor。
func (a *API) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		protected := strings.HasPrefix(path, "/com/") ||
			strings.HasPrefix(path, "/stu/") ||
			strings.HasPrefix(path, "/admin/") ||
			strings.HasPrefix(path, "/user/") ||
			strings.HasPrefix(path, "/chat/")
		// WebSocket 无法携带 Authorization 请求头，改在 handler 内用 ?token= 校验
		excluded := path == "/login" || path == "/refresh-token" ||
			path == "/stu/statistics" || path == "/chat/ws"
		if !protected || excluded {
			next.ServeHTTP(w, r)
			return
		}

		token := auth.TokenFromHeader(r.Header.Get("Authorization"))
		if token == "" {
			writeJSON(w, http.StatusUnauthorized, model.Error(401, "未提供Token"))
			return
		}
		claims, err := a.Service.Auth.Parse(token)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, model.Error(401, "Token无效或已过期"))
			return
		}

		if auth.ExpiringSoon(claims) {
			w.Header().Set("New-Token", a.Service.Auth.AccessToken(claims.UserID, claims.Username, claims.UserType))
		}

		if strings.HasPrefix(path, "/admin/") && claims.UserType != "admin" {
			writeJSON(w, http.StatusForbidden, model.Error(403, "权限不足，需要管理员权限"))
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID)
		ctx = context.WithValue(ctx, ctxUsername, claims.Username)
		ctx = context.WithValue(ctx, ctxUserType, claims.UserType)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func userID(r *http.Request) int {
	id, _ := r.Context().Value(ctxUserID).(int)
	return id
}

func userType(r *http.Request) string {
	value, _ := r.Context().Value(ctxUserType).(string)
	return value
}

func username(r *http.Request) string {
	value, _ := r.Context().Value(ctxUsername).(string)
	return value
}
