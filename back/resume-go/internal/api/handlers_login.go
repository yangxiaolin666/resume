package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"resume-backend/internal/auth"
	"resume-backend/internal/model"
)

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	var dto model.LoginDTO
	if err := decodeJSON(r, &dto); err != nil {
		a.fail(w, err)
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
		respond(w, model.Error(400, "处理参数验证异常: "+strings.Join(missing, ", ")))
		return
	}

	vo, err := a.Service.Login(dto.Username, dto.Password, dto.Type())
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, vo)
}

// me 返回当前登录用户的账号名与昵称
func (a *API) me(w http.ResponseWriter, r *http.Request) {
	profile, err := a.Service.Me(userID(r))
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, profile)
}

// renameSelf 修改自己的昵称（学生改姓名，企业改企业名称）
func (a *API) renameSelf(w http.ResponseWriter, r *http.Request) {
	var dto struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &dto); err != nil {
		a.fail(w, err)
		return
	}
	profile, err := a.Service.Rename(userID(r), dto.Name)
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, profile)
}

func (a *API) refreshToken(w http.ResponseWriter, r *http.Request) {
	refresh := r.Header.Get("Refresh-Token")
	if refresh == "" {
		respond(w, model.Error(500, "服务器内部错误: Required request header 'Refresh-Token' for method parameter type String is not present"))
		return
	}
	token, err := a.Service.RefreshToken(refresh)
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, map[string]string{"token": token})
}

func (a *API) validateToken(w http.ResponseWriter, r *http.Request) {
	token := ""
	body, _ := io.ReadAll(r.Body)
	r.Body.Close()
	if len(bytes.TrimSpace(body)) > 0 {
		var dto struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal(body, &dto); err == nil {
			token = dto.Token
		}
	}
	if token == "" {
		respond(w, model.Error(401, "Token 不能为空"))
		return
	}

	claims, err := a.Service.Auth.Parse(token)
	if err != nil {
		ok(w, map[string]any{"valid": false, "message": "Token 无效或已过期"})
		return
	}
	ok(w, map[string]any{
		"valid":       true,
		"message":     "Token 有效且未过期",
		"userId":      claims.UserID,
		"userType":    claims.UserType,
		"needRefresh": auth.ExpiringSoon(claims),
	})
}
