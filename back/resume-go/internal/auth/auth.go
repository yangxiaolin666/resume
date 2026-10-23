package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Claims 与 Java 版 JwtUtil 生成的 payload 保持一致。
type Claims struct {
	UserID    int    `json:"userId"`
	Username  string `json:"username"`
	UserType  string `json:"userType"`
	TokenType string `json:"tokenType"`
	Exp       int64  `json:"exp"`
}

const (
	AccessTokenHours  = 2
	RefreshTokenDays  = 7
	RefreshSoonWindow = 30 * time.Minute
)

type Manager struct {
	secret []byte
}

func NewManager(secret string) *Manager {
	return &Manager{secret: []byte(secret)}
}

func (m *Manager) Generate(userID int, username, userType, tokenType string, ttl time.Duration) string {
	header := map[string]any{"alg": "HS256", "typ": "JWT"}
	claims := Claims{
		UserID:    userID,
		Username:  username,
		UserType:  userType,
		TokenType: tokenType,
		Exp:       time.Now().Add(ttl).Unix(),
	}
	hb, _ := json.Marshal(header)
	cb, _ := json.Marshal(claims)
	unsigned := b64(hb) + "." + b64(cb)
	sig := m.sign(unsigned)
	return unsigned + "." + sig
}

func (m *Manager) AccessToken(userID int, username, userType string) string {
	return m.Generate(userID, username, userType, "access", AccessTokenHours*time.Hour)
}

func (m *Manager) RefreshToken(userID int, username, userType string) string {
	return m.Generate(userID, username, userType, "refresh", RefreshTokenDays*24*time.Hour)
}

// Parse 校验签名与过期时间，返回 payload。
func (m *Manager) Parse(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token")
	}
	unsigned := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(m.sign(unsigned)), []byte(parts[2])) {
		return nil, errors.New("invalid signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	if claims.Exp == 0 || time.Now().Unix() >= claims.Exp {
		return nil, errors.New("token expired")
	}
	return &claims, nil
}

// ExpiringSoon 判断 Token 是否将在 30 分钟内过期。
func ExpiringSoon(claims *Claims) bool {
	return time.Until(time.Unix(claims.Exp, 0)) < RefreshSoonWindow
}

func (m *Manager) sign(unsigned string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(unsigned))
	return b64(mac.Sum(nil))
}

func b64(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func HashPassword(raw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPassword(raw, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw)) == nil
}

// TokenFromHeader 从 Authorization: Bearer xxx 中提取 Token。
func TokenFromHeader(header string) string {
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	return ""
}
