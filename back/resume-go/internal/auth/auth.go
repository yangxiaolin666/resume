// Package auth 负责「身份凭证」这一层：签发/校验 JWT、哈希/校验密码、解析请求头里的 Token。
//
// 【这一层负责什么】
//   - 生成 access token（短有效期）和 refresh token（长有效期）；
//   - 校验 token 的三段结构、HS256 签名和过期时间；
//   - 用 bcrypt 对密码做不可逆哈希，登录时比对；
//   - 从 Authorization 请求头里取出 Bearer Token。
//
// 【与 api/data/cache/mq/rpc 的关系】
//   - api：Gin 中间件调用 TokenFromHeader + Parse，把解析出的 Claims 放进请求上下文，
//     后续 Handler 就能直接拿到 userID/userType，不需要每个接口再查一次库；
//   - service：登录、刷新 token 时调用 AccessToken/RefreshToken，
//     注册、改密、Excel 导入建账号时调用 HashPassword；
//   - data/cache/mq/rpc：本层完全不依赖它们，是纯内存计算，方便单独阅读和测试。
//
// 【对应简历能力】JWT 双 Token 登录鉴权、bcrypt 密码加密。
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

// Claims 是 JWT 第二段（payload）里携带的业务字段。
// 与 Java 版 JwtUtil 生成的 payload 保持一致，这样新旧两套后端签发的 Token 可以互相识别。
//
// 字段含义：
//   - UserID/Username/UserType：告诉后端「这是谁、什么角色」，避免每次请求都查用户表；
//   - TokenType：区分 access / refresh。刷新接口会强制要求 refresh（见 service.RefreshToken），
//     业务接口目前只验签、不强制 tokenType，所以它更像「用途标记」而非硬隔离；
//     如果要做严格隔离，应在鉴权中间件里再校验 tokenType == "access"；
//   - Exp：过期时间，Unix 秒（不是毫秒），解析时会显式校验。
//
// 注意：JWT 的 payload 只是 Base64URL 编码，并没有加密，任何人都能解开看内容，
// 所以这里绝不能放密码、手机号等敏感信息。
type Claims struct {
	UserID    int    `json:"userId"`
	Username  string `json:"username"`
	UserType  string `json:"userType"`
	TokenType string `json:"tokenType"`
	Exp       int64  `json:"exp"`
}

const (
	// AccessTokenHours：access token 只有 2 小时。
	// 它会被前端频繁携带，泄露风险相对更高，所以有效期要短。
	AccessTokenHours = 2
	// RefreshTokenDays：refresh token 7 天，只在「换新 access token」时使用，
	// 相当于给用户一个「7 天内免登录」的凭证。
	RefreshTokenDays = 7
	// RefreshSoonWindow：距离过期不足 30 分钟时，前端应主动刷新。
	// 留出窗口是为了避免请求正好卡在过期边界上，出现「刚打开页面就 401」的体验问题。
	RefreshSoonWindow = 30 * time.Minute
)

// Manager 持有签名密钥，负责签发和校验 Token。
// secret 来自配置（环境变量），必须满足两点：
//  1. 不能写死在代码仓库里，否则任何人都能伪造 Token；
//  2. 多实例部署时所有实例必须使用同一个 secret，否则 A 实例签发的 Token 到 B 实例会验签失败。
type Manager struct {
	secret []byte
}

// NewManager 把字符串密钥转成字节切片。
// 这里不做长度校验，属于「配置错误由部署方负责」的取舍；
// 生产环境建议至少 32 字节的随机密钥。
func NewManager(secret string) *Manager {
	return &Manager{secret: []byte(secret)}
}

// Generate 手动拼出一个标准 JWT。
//
// JWT 由三段组成，用「.」连接：header.payload.signature
//  1. header：声明签名算法（HS256）和类型（JWT）；
//  2. payload：业务字段（Claims），其中 exp 是过期时间；
//  3. signature：用密钥对「header.payload」做 HMAC-SHA256，再做 Base64URL 编码。
//
// 【为什么用 HS256】
// HS256 是对称签名：签发和校验用同一个 secret，实现简单，
// 适合「单体后端自己签发、自己校验」的场景。
// 如果以后要拆成多个互不信任的服务，可以换成 RS256（私钥签名、公钥验签）。
//
// 【为什么用 RawURLEncoding】
// JWT 要求三段都是 URL 安全字符，且不能带 Base64 的 "=" 填充，
// 否则放进 URL/Header 时要额外转义。Raw 就是不补 "=" 的版本。
//
// 返回值不返回 error：header/claims 都是简单结构体，json.Marshal 不会失败；
// 这里忽略错误是为了让调用方代码更简洁（实际项目中也可以选择 panic 或返回 error）。
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

// AccessToken 签发业务接口使用的短效 Token（2 小时）。
// 登录成功后由 service.Login 调用，和 RefreshToken 一起返回给前端。
func (m *Manager) AccessToken(userID int, username, userType string) string {
	return m.Generate(userID, username, userType, "access", AccessTokenHours*time.Hour)
}

// RefreshToken 签发刷新用的长效 Token（7 天）。
//
// 【双 Token 的意义】
//   - access token 短期有效，即使被窃取，攻击窗口也只有 2 小时；
//   - refresh token 有效期长，但只在 /auth/refresh 一个接口使用，
//     暴露面小；它带 tokenType=refresh，刷新逻辑据此拒绝把 access 当 refresh 用。
//
// 本项目采用「只刷新 access、不轮换 refresh」的简化方案：
// refresh token 7 天内可重复使用。更严格的做法是每次刷新都下发新的 refresh token
// 并让旧的失效（refresh token rotation），需要额外存储状态。
func (m *Manager) RefreshToken(userID int, username, userType string) string {
	return m.Generate(userID, username, userType, "refresh", RefreshTokenDays*24*time.Hour)
}

// Parse 校验签名与过期时间，返回 payload。
//
// 校验顺序很重要，不能调换：
//  1. 先按 "." 切成三段，段数不对直接判定非法（防止随手传个字符串就进入后续逻辑）；
//  2. 重新计算前两段的签名，并与第三段比对——只有验签通过，才能相信 payload 没被篡改；
//  3. 验签通过后才解析 payload；
//  4. 最后检查 exp 是否过期。
//
// 【为什么先验签再解析】
// 如果先解析再验签，攻击者可以塞入超长/畸形 JSON 消耗 CPU 和内存；
// 而且未验签的数据不可信，先解析容易让人误用。
//
// 【为什么用 hmac.Equal 而不是 ==】
// hmac.Equal 是常量时间比较，耗时与「第几个字节不同」无关；
// 用 == 会提前返回，攻击者可以通过响应时间差逐字节猜出正确签名（时序攻击）。
//
// 失败时统一返回 error（invalid token / invalid signature / token expired），
// 由 api 层转换成 401 响应。错误信息不包含内部细节，避免给攻击者提示。
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
// 前端可以据此提前调用刷新接口，避免用户操作到一半突然被踢回登录页。
// 注意这里只做时间判断，不验签——调用方应传入已经 Parse 成功的 Claims。
func ExpiringSoon(claims *Claims) bool {
	return time.Until(time.Unix(claims.Exp, 0)) < RefreshSoonWindow
}

// sign 对 "header.payload" 做 HMAC-SHA256 签名。
// HMAC 的输入是「密钥 + 消息」，没有密钥就无法算出正确签名，
// 因此任何对 payload 的修改都会导致验签失败。
func (m *Manager) sign(unsigned string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(unsigned))
	return b64(mac.Sum(nil))
}

// b64 统一使用 Base64URL 无填充编码，保证三段都符合 JWT 规范。
func b64(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

// HashPassword 用 bcrypt 生成密码哈希。
//
// 【bcrypt 的两个关键点】
//  1. 自动加盐：每次调用都会生成随机 salt 并写进结果字符串，
//     所以同一个密码两次哈希得到的值不同。数据库只存这个哈希，不存明文。
//  2. 故意慢：DefaultCost=10 表示做 2^10 轮计算，单次约几十毫秒。
//     这会让暴力破解和彩虹表成本极高，但正常登录只算一次，用户几乎无感。
//
// 不要用 MD5/SHA1 存密码：它们太快且无盐，撞库和彩虹表可以瞬间还原。
//
// 失败（例如密码超过 bcrypt 72 字节上限）时返回 error，由调用方转成 4xx/5xx。
func HashPassword(raw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	return string(hash), err
}

// CheckPassword 校验明文密码是否与数据库里的 bcrypt 哈希匹配。
//
// 哈希串里已经包含算法版本、cost 和 salt，bcrypt 会自己取出盐重新计算再比对，
// 所以这里不需要额外传盐。内部使用常量时间比较，避免时序攻击。
//
// 返回 bool 而不是 error：对调用方来说「不匹配」就是普通的业务失败，
// 登录接口统一提示「用户名或密码错误」，不暴露具体原因。
func CheckPassword(raw, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw)) == nil
}

// TokenFromHeader 从 Authorization: Bearer xxx 中提取 Token。
// 前缀不匹配时返回空串，由中间件按「未登录」处理。
// 这里只去掉固定前缀并原样返回，不校验格式，真正的合法性交给 Parse。
func TokenFromHeader(header string) string {
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	return ""
}
