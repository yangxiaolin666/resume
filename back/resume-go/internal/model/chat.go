package model

// ==================================================================
// 本文件职责：聊天业务的数据模型（学生 <-> 企业 的实时沟通）
// ==================================================================
// 这里只定义「消息长什么样」，不包含任何收发逻辑：
//   - 收发动作在 internal/api/handlers_chat.go（REST）与 internal/chat（WebSocket）；
//   - 持久化在 internal/data；
//   - 离线通知经 RabbitMQ 异步投递。
// 本文件的类型同时被 REST 响应和 WebSocket 推送复用，
// 保证「历史消息」和「实时消息」的字段结构完全一致，前端只写一套渲染逻辑。
// ==================================================================

// ==================================================================
// 聊天模块数据模型
// ==================================================================
// 设计说明（学习重点）：
//  1. 聊天双方用 user_id + user_type 共同标识。
//     user_id 是 user 表主键；user_type 只能是 student / company。
//     这样学生和企业可以复用同一套消息表，不需要建两张表。
//  2. 一条消息只存一行：发送者 -> 接收者。
//     查询“某个会话”时用双向条件：
//       (A 发给 B) OR (B 发给 A)
//  3. is_read 只针对“接收者”有意义：接收者读取会话时把它置为 1。
// ==================================================================

// ChatMessage 一条聊天消息（同时用于 REST 响应和 WebSocket 推送）
//
// 字段分组理解：
//   - ID / CreatedAt：消息身份与时间，前端用来排序、去重；
//   - Sender* / Receiver*：双方身份，Receiver 决定这条消息推给谁；
//   - Content / MsgType：正文与类型（text 文本、image 图片地址）；
//   - IsRead：接收者是否已读，只对接收方有意义；
//   - SenderName / ReceiverName：冗余的展示名，避免前端为了显示昵称再查一次用户。
type ChatMessage struct {
	ID           int64  `json:"id"`
	SenderID     int    `json:"senderId"`
	SenderType   string `json:"senderType"`
	ReceiverID   int    `json:"receiverId"`
	ReceiverType string `json:"receiverType"`
	Content      string `json:"content"`
	MsgType      string `json:"msgType"`
	IsRead       bool   `json:"isRead"`
	CreatedAt    string `json:"createdAt"`
	SenderName   string `json:"senderName"`
	ReceiverName string `json:"receiverName"`
}

// Conversation 会话列表中的一项（左侧聊天列表）
//
// 一个会话并不是数据库表，而是把「双方之间最后一条消息 + 未读数」聚合出来的视图对象，
// 由 service 查询时拼装。Peer* 表示“对方”，LastFromMe 决定最后一条消息是否显示“我：”前缀。
type Conversation struct {
	PeerID      int    `json:"peerId"`      // 对方的 user_id
	PeerType    string `json:"peerType"`    // 对方身份：student / company
	PeerName    string `json:"peerName"`    // 对方展示名
	LastMessage string `json:"lastMessage"` // 最后一条消息内容（图片消息可能是地址）
	LastMsgType string `json:"lastMsgType"` // 最后一条消息类型
	LastTime    string `json:"lastTime"`    // 最后一条消息时间，用于排序
	LastFromMe  bool   `json:"lastFromMe"`  // 最后一条是否由我发出
	Unread      int    `json:"unread"`      // 我未读的条数（红点）
	Online      bool   `json:"online"`      // 对方当前是否有 WebSocket 连接
}

// ChatPeer 可发起聊天的联系人
//
// 与 Conversation 的区别：这是“可以聊的人”（如企业收藏的学生、投递过的学生），
// 还没有任何消息；Conversation 是“已经聊过的会话”。
type ChatPeer struct {
	UserID      int    `json:"userId"`
	UserType    string `json:"userType"`
	Name        string `json:"name"`
	Description string `json:"description"` // 副标题，如专业 / 公司地区
	Online      bool   `json:"online"`
}

// SendChatRequest REST 发送消息的请求体。
// MsgType 为空时 service 会按 text 处理；图片消息的 Content 存的是上传后的 URL。
type SendChatRequest struct {
	ReceiverID   int    `json:"receiverId"`
	ReceiverType string `json:"receiverType"`
	Content      string `json:"content"`
	MsgType      string `json:"msgType"`
}

// ChatImageVO 图片上传结果，前端拿到 URL 后作为 content 再发一条 image 消息。
type ChatImageVO struct {
	URL string `json:"url"`
}

// ResumePermission 描述「当前学生能否给这个企业发简历」。
//
// 【业务规则】
// 对方回复之前，学生只能发一次简历；企业回复后额度恢复，可以再发一次。
// 这样既避免刷屏，也保证企业每次沟通都能拿到最新简历。
type ResumePermission struct {
	CanSend      bool `json:"canSend"`
	WaitingReply bool `json:"waitingReply"`
}

// MarkReadRequest 标记会话已读的请求体：Peer 指“对方”，
// 服务端会把「对方发给我的、还没读的」消息全部置为已读。
type MarkReadRequest struct {
	PeerID   int    `json:"peerId"`
	PeerType string `json:"peerType"`
}

// ChatNotification 离线消息通知。
//
// 【它和 ChatMessage 的区别】
// ChatMessage 是聊天记录本身，必须实时、完整；
// ChatNotification 只是「对方不在线时留下的一条提醒」，由 RabbitMQ 异步生成。
// 即使通知丢了，用户重新打开聊天页仍然能从 chat_message 里看到完整记录，
// 所以通知链路允许最终一致，不会影响核心业务。
type ChatNotification struct {
	ID           int64  `json:"id"`
	MessageID    int64  `json:"messageId"`
	ReceiverID   int    `json:"receiverId"`
	ReceiverType string `json:"receiverType"`
	SenderID     int    `json:"senderId"`
	SenderType   string `json:"senderType"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	IsRead       bool   `json:"isRead"`
	CreatedAt    string `json:"createdAt"`
}
