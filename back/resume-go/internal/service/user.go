package service

import (
	"strings"
	"unicode/utf8"

	"resume-backend/internal/auth"
	"resume-backend/internal/model"
)

// ==================================================================
// 账号业务层：登录、当前用户信息、修改昵称、刷新 Token
// ==================================================================
// 【与其它层的关系】
//   - api：登录/刷新接口不经过鉴权中间件，其余接口从上下文拿到 userID 后调用这里；
//   - auth：密码比对、签发 access/refresh、解析 refresh token；
//   - data：user 表读写，昵称落在 student/company 档案表；
//   - cache/mq/rpc：登录链路不依赖，保证认证只以数据库和密钥为准。
//
// 【对应简历能力】JWT 双 Token 登录、bcrypt 密码校验、账号资料维护。
// ==================================================================

const maxDisplayNameLength = 20

// Login 账号密码登录。
//
// 校验顺序与安全考虑：
//  1. 用户名不存在 / 密码错误，统一返回同一个错误码和文案「用户名或密码错误」。
//     如果分别提示“用户不存在”“密码错误”，攻击者就能批量探测哪些账号真实存在
//     （用户名枚举），所以这里刻意模糊化。
//  2. 身份类型必须匹配：学生、企业、管理员使用不同入口，
//     选错身份时明确提示，属于可接受的体验取舍。
//  3. 密码用 bcrypt 比对，数据库里只存哈希。
//
// 登录成功后同时签发 access（2h）和 refresh（7d）两个 Token；
// 展示名优先取档案里的姓名/企业名，取不到则退回登录用户名，保证前端总有内容可显示。
// 失败时返回 AppError（40005 或 401），由 api 层转成对应 HTTP 状态码。
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
// userID 来自 JWT，不接受前端传入，避免越权查看他人资料。
// 用户被删除但 Token 还没过期时返回 404，前端应清除本地登录态。
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
//
// 规则：
//   - 去首尾空格后不能为空，否则聊天列表会出现空白名字；
//   - 按「字符数」限制 20，而不是字节数，否则中文用户只能填 6 个字；
//   - 只更新档案表的展示名，不动 user.username。
//     登录账号是身份标识，允许随意修改会导致已签发的 Token、历史数据关联全部错乱。
//
// 用户不存在返回 404；更新失败返回底层错误。
// 返回最新的 UserProfile，前端可以直接用它刷新界面，无需再请求一次 /me。
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

// RefreshToken 用 refresh token 换一个新的 access token。
//
// 两道校验缺一不可：
//  1. Parse 验签 + 查过期，保证 Token 没被伪造、也没过期；
//  2. 必须确认 claims.TokenType == "refresh"。
//     否则拿到长效 refresh token 的人可以把它当 access token 直接访问业务接口，
//     “短 access + 长 refresh”的安全设计就失效了。
//
// 成功时只返回新的 access token；refresh token 不变（不轮换），
// 用户 7 天内可以持续刷新。任何失败都返回 401，前端据此跳转登录页。
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
