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
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Client]struct{} // key = "student:11" / "company:12"

	// OnMessage 由上层注入：收到客户端消息后执行真正的业务（校验 + 入库 + 推送）。
	OnMessage func(client *Client, message IncomingMessage)
}

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
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	k := key(c.UserID, c.UserType)
	if h.clients[k] == nil {
		h.clients[k] = make(map[*Client]struct{})
	}
	h.clients[k][c] = struct{}{}

	log.Printf("[chat] 用户上线 %s(%s)，当前连接数 %d", c.Username, k, len(h.clients[k]))
}

// Unregister 连接断开后从 Hub 移除，避免内存泄漏。
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
}

// SendToUser 把消息推送给某个用户的所有在线连接。
// 如果用户不在线，消息仍然已经写入数据库，下次打开聊天页会从历史记录中加载。
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
func (h *Hub) IsOnline(userID int, userType string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[key(userID, userType)]) > 0
}
