package chat

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// ==================================================================
// 单个 WebSocket 连接
// ==================================================================
// 学习重点：
//  1. gorilla/websocket 不允许并发写同一个连接，
//     所以所有写操作都集中到 writePump 这一个 goroutine 里。
//  2. 每个连接有两个 goroutine：
//       readPump  : 读客户端消息
//       writePump : 写消息 + 定时发 Ping 心跳
//  3. 心跳机制：
//       服务端每 54s 发一个 Ping 控制帧
//       浏览器会自动回 Pong
//       readPump 收到 Pong 后刷新读超时
//     如果 60s 内没有收到任何数据，就认为连接已死并关闭。
// ==================================================================

const (
	writeWait      = 10 * time.Second        // 单次写超时
	pongWait       = 60 * time.Second        // 等待 Pong 的最长时间
	pingPeriod     = (pongWait * 9) / 10     // Ping 周期，必须小于 pongWait
	maxMessageSize = 4 * 1024                // 单条消息最大 4KB
	sendBufferSize = 128                     // 发送缓冲队列长度
)

// IncomingMessage 客户端 -> 服务端的消息格式。
//
// 目前只支持两种 type：
//   - message : 发送聊天消息
//   - ping    : 应用层心跳（服务端回 pong）
//
// 示例：
//
//	{"type":"message","receiverId":12,"receiverType":"company","content":"你好"}
//	{"type":"message","receiverId":12,"receiverType":"company","content":"/uploads/chat/x.png","msgType":"image"}
//	{"type":"ping"}
type IncomingMessage struct {
	Type         string `json:"type"`
	ReceiverID   int    `json:"receiverId"`
	ReceiverType string `json:"receiverType"`
	Content      string `json:"content"`
	MsgType      string `json:"msgType"`
}

// Client 代表一个已经完成 JWT 鉴权的 WebSocket 连接。
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte

	UserID   int    // user 表主键
	UserType string // student / company
	Username string // 登录名，仅用于日志和兜底显示
}

func NewClient(hub *Hub, conn *websocket.Conn, userID int, userType, username string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, sendBufferSize),
		UserID:   userID,
		UserType: userType,
		Username: username,
	}
}

// Start 启动读/写两个 goroutine。
// 注意：调用方（HTTP Handler）不能再阻塞等待，否则会占住请求 goroutine。
func (c *Client) Start() {
	c.hub.Register(c)
	go c.writePump()
	go c.readPump()
}

// PushJSON 把任意结构体序列化后放进发送队列。
// 供业务层给某个客户端回错误信息或确认消息。
func (c *Client) PushJSON(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("[chat] 序列化推送消息失败: %v", err)
		return
	}
	c.push(data)
}

// push 非阻塞入队。
// 如果客户端很慢导致队列满，直接丢弃这条消息，避免拖垮整个 Hub。
func (c *Client) push(data []byte) {
	select {
	case c.send <- data:
	default:
		log.Printf("[chat] 发送队列已满，丢弃消息 user=%s", c.Username)
	}
}

// readPump 持续读取客户端发来的消息。
func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister(c)
		close(c.send) // 通知 writePump 退出
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		// 收到浏览器的 Pong 后，续期读超时
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[chat] 连接异常关闭 user=%s err=%v", c.Username, err)
			}
			return
		}

		// 收到任何数据都说明连接还活着
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))

		var msg IncomingMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			c.PushJSON(map[string]any{"type": "error", "message": "消息格式错误"})
			continue
		}

		switch msg.Type {
		case "ping":
			// 应用层心跳：前端据此判断连接是否还活着
			c.PushJSON(map[string]any{"type": "pong"})

		case "message":
			// 真正的业务处理交给上层注入的回调
			if c.hub.OnMessage != nil {
				c.hub.OnMessage(c, msg)
			}

		default:
			c.PushJSON(map[string]any{"type": "error", "message": "未知的消息类型: " + msg.Type})
		}
	}
}

// writePump 是这条连接唯一的写出口。
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case data, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 关闭了 channel，发送关闭帧后退出
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			// 定时发送 WebSocket Ping 控制帧（浏览器自动回 Pong）
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
