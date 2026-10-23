// Package chat 负责 WebSocket 实时聊天的连接管理。
//
// 学习重点（对照你给的参考项目）：
//  1. Hub 是“连接注册中心”，内部用 map 保存所有在线连接；
//  2. 同一用户允许多个连接（多标签页），所以 map 的值是“连接集合”；
//  3. 所有对 map 的读写都必须用 sync.RWMutex 保护，否则并发时会 panic；
//  4. Hub 不直接依赖 service 包，而是暴露 OnMessage 回调，
//     由上层把业务逻辑注入进来，避免循环依赖。
package chat

import (
	"encoding/json"
	"log"
	"strconv"
	"sync"
)

// Hub 管理所有在线 WebSocket 连接。
//
// 【数据结构为什么是 map 套 map】
//
//	clients["student:11"] = {连接A, 连接B}   // 同一个人在电脑和手机上各开了一个页面
//
// 外层 key 是「用户类型:用户ID」，内层是连接集合：
//   - 一个用户可能有多个连接（多标签页 / 多设备），只存一个 *Client 会互相覆盖；
//   - 推送时遍历集合，给该用户的每个连接都发一份，保证每个页面都能收到；
//   - 断开时只删掉自己那一个连接，最后一个连接移除后才删除外层 key，避免泄漏空集合。
//
// 【为什么必须有 RWMutex】
// WebSocket 的读 goroutine、HTTP 请求 goroutine 会同时访问这张 map。
// Go 的原生 map 不是并发安全的，同时读写会直接 panic（fatal error: concurrent map read and map write）。
// 因此：写操作（注册 / 注销）加 Lock，读操作（推送 / 查在线）加 RLock。
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Client]struct{} // key = "student:11" / "company:12"

	// OnMessage 由上层注入：收到客户端消息后执行真正的业务（校验 + 入库 + 推送）。
	//
	// 【为什么用回调，而不是在 chat 包里直接 import service】
	// api/service 层会反过来使用 chat.Hub 来推送消息（service -> chat）。
	// 如果 chat 再 import service，就形成循环依赖，Go 编译直接报错。
	// 用「函数字段」把业务逻辑从外部塞进来，chat 包只依赖自己的类型，
	// 依赖方向变成单向（service -> chat），这就是依赖倒置在 Go 里最简单的落地方式。
	OnMessage func(client *Client, message IncomingMessage)

	// OnRegister / OnUnregister 由上层注入，用于把在线状态同步到 Redis。
	// Hub 本身不依赖 Redis，保持「连接管理」职责单一。
	OnRegister   func(client *Client)
	OnUnregister func(client *Client)
}

// NewHub 创建 Hub。初始化 clients，避免第一次写入 nil map 触发 panic。
func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[*Client]struct{})}
}

// key 用“类型 + ID”共同标识一个聊天用户。
// 不能只用 ID，因为 user 表中学生和企业的 ID 虽然全局唯一，
// 但显式带上类型可以让代码更清晰，也为以后扩展留空间。
func key(userID int, userType string) string {
	return userType + ":" + strconv.Itoa(userID)
}

// Register 连接建立后注册到 Hub。
//
// 流程：加写锁 -> 若该用户还没有连接集合就先建一个 -> 把当前连接放进集合 ->
// 记录连接数日志 -> 异步通知上层（如把在线状态写入 Redis）。
//
// 用 struct{}{} 作为集合的值：它不占内存，只表达“这个 key 存在”。
// 解锁由 defer 保证，即使中途 return 也不会死锁。
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	k := key(c.UserID, c.UserType)
	if h.clients[k] == nil {
		h.clients[k] = make(map[*Client]struct{})
	}
	h.clients[k][c] = struct{}{}

	log.Printf("[chat] 用户上线 %s(%s)，当前连接数 %d", c.Username, k, len(h.clients[k]))

	// 回调里可能要访问 Redis，放在 goroutine 中异步执行，避免拖慢连接注册。
	if h.OnRegister != nil {
		go h.OnRegister(c)
	}
}

// Unregister 连接断开后从 Hub 移除，避免内存泄漏。
//
// 只删除当前这个连接；如果该用户还有别的连接（比如另一个标签页），
// 外层 key 必须保留，否则会把还在线的连接一起“下线”。
// 只有当集合空了才 delete 外层 key，及时释放内存。
// 回调同样用 go 异步执行，不让 Redis 操作拖慢关闭连接的主流程。
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	k := key(c.UserID, c.UserType)
	if set, ok := h.clients[k]; ok {
		delete(set, c)
		if len(set) == 0 {
			delete(h.clients, k)
		}
	}

	log.Printf("[chat] 用户离线 %s(%s)", c.Username, k)

	if h.OnUnregister != nil {
		go h.OnUnregister(c)
	}
}

// SendToUser 把消息推送给某个用户的所有在线连接。
// 如果用户不在线，消息仍然已经写入数据库，下次打开聊天页会从历史记录中加载。
//
// 实现细节：
//   - 先 json.Marshal 成字节，多个连接复用同一份数据，不用每个连接各序列化一次；
//   - 用读锁（RLock）即可，因为只读 map，不修改结构；
//   - c.push 是非阻塞入队，某个连接卡住不会拖慢其他连接。
func (h *Hub) SendToUser(userID int, userType string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[chat] 消息序列化失败: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.clients[key(userID, userType)] {
		c.push(data)
	}
}

// IsOnline 判断某个用户当前是否有 WebSocket 连接（前端显示在线绿点）。
//
// 注意：这里查的是「当前进程内」的连接表。如果服务部署了多个实例，
// 用户的连接可能在另一台机器上，此时需要以 Redis 中的在线状态为准
// （OnRegister / OnUnregister 回调就是负责同步 Redis 的）。
func (h *Hub) IsOnline(userID int, userType string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[key(userID, userType)]) > 0
}
