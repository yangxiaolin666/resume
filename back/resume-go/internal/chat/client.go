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
//  4. 两个 goroutine 通过带缓冲的 channel（send）通信：
//     业务方只往 channel 里塞数据，真正的网络写由 writePump 串行完成。
//     这样既满足“写不能并发”，又让业务方不必等待网络 IO。
// ==================================================================

// 超时与容量常量。数值之间是有关联的，不能随便改：
//   - pingPeriod 必须小于 pongWait，否则还没发下一次 Ping 就已经超时断开；
//   - 这里取 9/10，给网络抖动留出 6 秒余量。
const (
	writeWait      = 10 * time.Second    // 单次写超时
	pongWait       = 60 * time.Second    // 等待 Pong 的最长时间
	pingPeriod     = (pongWait * 9) / 10 // Ping 周期，必须小于 pongWait
	maxMessageSize = 4 * 1024            // 单条消息最大 4KB
	sendBufferSize = 128                 // 发送缓冲队列长度
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
//
// 注意区分两种心跳：
//   - 这里的 {"type":"ping"} 是应用层 JSON，服务端回 {"type":"pong"}，
//     前端可以据此判断“业务通道”是否正常；
//   - writePump 里定时发送的是 WebSocket 协议层的 Ping 控制帧，
//     由浏览器自动回 Pong，用来维持 TCP 连接不被中间设备断开。
type IncomingMessage struct {
	Type         string `json:"type"`
	ReceiverID   int    `json:"receiverId"`
	ReceiverType string `json:"receiverType"`
	Content      string `json:"content"`
	MsgType      string `json:"msgType"`
}

// Client 代表一个已经完成 JWT 鉴权的 WebSocket 连接。
//
// 字段说明：
//   - hub：反向引用连接中心，断线时用它注销自己；
//   - conn：底层 TCP/WS 连接，读写在两个不同 goroutine 中进行；
//   - send：发送队列。业务方 push 进来，writePump 取出去写网络。
//     它是有缓冲的 channel，起到“削峰”作用：短时间内多条推送先排队，
//     不会让调用方阻塞在网络上。
//   - 三个用户字段来自握手时的 token 解析，后续消息都以它们为准，
//     不信任客户端在消息体里自报的身份。
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte

	UserID   int    // user 表主键
	UserType string // student / company
	Username string // 登录名，仅用于日志和兜底显示
}

// NewClient 创建连接对象，并初始化发送队列。
// 容量 128 的选择：足够吸收一次会话列表刷新带来的突发推送，
// 又不会因为某个“死连接”无限堆积内存。
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
//
// 顺序：
//  1. 先 Register，让 Hub 立刻知道这个用户在线（此时就能收到别人的推送）；
//  2. 再启动 writePump（唯一的写出口，负责发送队列 + 心跳）；
//  3. 最后启动 readPump（唯一读入口，退出时会负责清理）。
//
// 为什么要交给 goroutine：WebSocket 是长连接，handler 返回后连接仍然要保持。
// 如果把读写写在 handler 里同步执行，这个 HTTP 请求就永远不结束了。
func (c *Client) Start() {
	c.hub.Register(c)
	go c.writePump()
	go c.readPump()
}

// PushJSON 把任意结构体序列化后放进发送队列。
// 供业务层给某个客户端回错误信息或确认消息。
//
// 这里是“序列化 + 入队”，不直接写 conn：
// 直接写会和其他 goroutine 并发写同一个连接，gorilla/websocket 会 panic。
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
//
// select + default 是 Go 里“尝试发送、不等”的固定写法：
//   - case c.send <- data：队列有空位，放进去；
//   - default：队列已满，立即走这里，绝不阻塞。
//
// 为什么不能阻塞：Hub.SendToUser 是持有读锁遍历所有连接调用 push 的，
// 一旦这里卡住，整个 Hub 的推送都会被拖慢，甚至阻塞注册 / 注销。
// 对聊天场景来说，丢一条实时推送可以接受——消息已经入库，
// 用户重新拉取历史记录就能补齐（最终一致）。
func (c *Client) push(data []byte) {
	select {
	case c.send <- data:
	default:
		log.Printf("[chat] 发送队列已满，丢弃消息 user=%s", c.Username)
	}
}

// readPump 持续读取客户端发来的消息。
//
// 【为什么只能有一个读 goroutine】
// websocket.Conn 的读接口同样不是并发安全的；而且 TCP 流需要按顺序消费，
// 多个 goroutine 同时 ReadMessage 会打乱消息边界。所以整个连接只在这里读。
//
// 【超时与心跳】
//   - SetReadLimit：限制单帧大小，防止恶意客户端发超大包耗尽内存；
//   - SetReadDeadline：设置“读超时”。如果 60 秒内既没有业务消息、
//     也没有 Pong，ReadMessage 会返回错误，循环退出，连接被清理；
//   - SetPongHandler：浏览器收到服务端 Ping 后会自动回 Pong，
//     这里在收到 Pong 时把截止时间往后推，形成“心跳续命”。
//   - 收到任何业务数据也会刷新截止时间（说明连接还活着）。
//
// 【退出清理顺序】
// defer 是后进先出：先 Unregister（从 Hub 摘除）-> close(send) 通知 writer 退出
// -> conn.Close()。close(send) 只在这里做一次，避免重复关闭 channel 导致 panic。
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
//
// 【为什么所有写都集中在这里】
// gorilla/websocket 明确要求：同一个 Conn 上不能有并发的 writer。
// 如果多个 HTTP 请求同时调用 WriteMessage，轻则消息交错，重则直接 panic。
// 这里用单 goroutine + channel 把并发写“串行化”：
//
//	业务 goroutine --push--> send channel --> writePump --> conn.WriteMessage
//
// 【select 的两个分支】
//   - c.send 有数据：先设置 10 秒写超时，再写一帧 TextMessage；
//     如果 channel 已被 close（ok == false），说明读侧已经退出，
//     发送一个 CloseMessage 关闭帧后结束；
//   - ticker.C 到点：发送 WebSocket 协议层的 Ping 控制帧（不是业务 JSON），
//     浏览器会自动回 Pong，从而触发 readPump 里的 PongHandler 续期。
//
// 任意一次写失败都直接 return：连接已经不可用，
// conn.Close() 会让阻塞中的 ReadMessage 返回错误，readPump 随之完成清理。
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
