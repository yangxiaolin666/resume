package model

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
type Conversation struct {
	PeerID      int    `json:"peerId"`
	PeerType    string `json:"peerType"`
	PeerName    string `json:"peerName"`
	LastMessage string `json:"lastMessage"`
	LastMsgType string `json:"lastMsgType"`
	LastTime    string `json:"lastTime"`
	LastFromMe  bool   `json:"lastFromMe"`
	Unread      int    `json:"unread"`
	Online      bool   `json:"online"`
}

// ChatPeer 可发起聊天的联系人
type ChatPeer struct {
	UserID      int    `json:"userId"`
	UserType    string `json:"userType"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Online      bool   `json:"online"`
}

// SendChatRequest REST 发送消息的请求体
type SendChatRequest struct {
	ReceiverID   int    `json:"receiverId"`
	ReceiverType string `json:"receiverType"`
	Content      string `json:"content"`
	MsgType      string `json:"msgType"`
}

// ChatImageVO 图片上传结果
type ChatImageVO struct {
	URL string `json:"url"`
}

// MarkReadRequest 标记会话已读的请求体
type MarkReadRequest struct {
	PeerID   int    `json:"peerId"`
	PeerType string `json:"peerType"`
}
