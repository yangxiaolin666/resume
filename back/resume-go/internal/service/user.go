package service

import (
	"strings"
	"unicode/utf8"

	"resume-backend/internal/auth"
	"resume-backend/internal/model"
)

const maxDisplayNameLength = 20

func (s *Service) Login(username, password, userType string) (*model.LoginVO, error) {
	user, err := s.DB.FindUserByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, model.NewAppError(40005, "用户名或密码错误")
	}
	if user.UserType != userType {
		return nil, model.NewAppError(401, "用户身份不匹配，请选择正确的身份登录")
	}
	if !auth.CheckPassword(password, user.Password) {
		return nil, model.NewAppError(40005, "用户名或密码错误")
	}
	displayName, err := s.DB.ChatUserName(user.UserID, user.UserType)
	if err != nil {
		return nil, err
	}
	if displayName == "" {
		displayName = user.Username
	}
	return &model.LoginVO{
		UserID:       user.UserID,
		Username:     user.Username,
		UserType:     user.UserType,
		DisplayName:  displayName,
		Token:        s.Auth.AccessToken(user.UserID, user.Username, user.UserType),
		RefreshToken: s.Auth.RefreshToken(user.UserID, user.Username, user.UserType),
	}, nil
}

// Me 返回当前登录用户的账号信息与昵称，供前端刷新右上角显示。
func (s *Service) Me(userID int) (*model.UserProfile, error) {
	user, err := s.DB.FindUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, model.NewAppError(404, "用户不存在")
	}
	displayName, err := s.DB.ChatUserName(user.UserID, user.UserType)
	if err != nil {
		return nil, err
	}
	if displayName == "" {
		displayName = user.Username
	}
	return &model.UserProfile{
		UserID:      user.UserID,
		Username:    user.Username,
		UserType:    user.UserType,
		DisplayName: displayName,
	}, nil
}

// Rename 修改自己的昵称：学生改姓名，企业改企业名称，登录账号不变。
func (s *Service) Rename(userID int, name string) (*model.UserProfile, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, model.NewAppError(400, "昵称不能为空")
	}
	if utf8.RuneCountInString(name) > maxDisplayNameLength {
		return nil, model.NewAppError(400, "昵称不能超过20个字符")
	}

	user, err := s.DB.FindUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, model.NewAppError(404, "用户不存在")
	}
	if err := s.DB.UpdateDisplayName(userID, user.UserType, name); err != nil {
		return nil, err
	}
	return &model.UserProfile{
		UserID:      user.UserID,
		Username:    user.Username,
		UserType:    user.UserType,
		DisplayName: name,
	}, nil
}

func (s *Service) RefreshToken(refreshToken string) (string, error) {
	claims, err := s.Auth.Parse(refreshToken)
	if err != nil {
		return "", model.NewAppError(401, "Refresh Token无效或已过期")
	}
	if claims.TokenType != "refresh" {
		return "", model.NewAppError(401, "Token类型错误")
	}
	return s.Auth.AccessToken(claims.UserID, claims.Username, claims.UserType), nil
}
