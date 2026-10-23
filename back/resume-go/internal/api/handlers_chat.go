package api

import (
	"errors"
	"net/http"

	"github.com/gorilla/websocket"

	"resume-backend/internal/chat"
	"resume-backend/internal/model"
	"resume-backend/internal/service"
)

// chatUpgrader 把普通 HTTP 请求升级成 WebSocket。
// CheckOrigin 返回 true 表示允许任意来源（本地开发方便）。
// 生产环境应该只允许自己的前端域名，防止 CSWSH（跨站 WebSocket 劫持）。
var chatUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// chatPeers GET /chat/peers?keyword=
// 返回“我可以聊天的人”：学生看到企业，企业看到学生。
func (a *API) chatPeers(w http.ResponseWriter, r *http.Request) {
	peers, err := a.Service.ChatPeers(userType(r), queryString(r, "keyword"))
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, map[string]any{"list": peers})
}

// chatConversations GET /chat/conversations
// 返回会话列表：对方、最后一条消息、未读数、在线状态。
func (a *API) chatConversations(w http.ResponseWriter, r *http.Request) {
	conversations, err := a.Service.ChatConversations(userID(r), userType(r))
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, map[string]any{"list": conversations})
}

// chatHistory GET /chat/messages?peerId=&peerType=&beforeId=&limit=
func (a *API) chatHistory(w http.ResponseWriter, r *http.Request) {
	peerID := queryInt(r, "peerId", 0)
	peerType := queryString(r, "peerType")
	beforeID := int64(queryInt(r, "beforeId", 0))
	limit := queryInt(r, "limit", 30)

	messages, hasMore, err := a.Service.ChatHistory(userID(r), userType(r), peerID, peerType, beforeID, limit)
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, map[string]any{"messages": messages, "hasMore": hasMore})
}

// chatSend POST /chat/messages
// REST 发送接口。WebSocket 断开时前端可以降级用它发送。
func (a *API) chatSend(w http.ResponseWriter, r *http.Request) {
	var req model.SendChatRequest
	if err := decodeJSON(r, &req); err != nil {
		a.fail(w, err)
		return
	}
	message, err := a.Service.SendChatMessage(userID(r), userType(r), username(r),
		req.ReceiverID, req.ReceiverType, req.Content, req.MsgType)
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, message)
}

// chatUploadImage POST /chat/uploadImage
// 聊天图片上传，表单字段名 image，最大 5MB。
func (a *API) chatUploadImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(6 << 20); err != nil {
		respond(w, model.Error(500, "服务器内部错误: "+err.Error()))
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		respond(w, model.Error(400, "请选择要上传的图片"))
		return
	}
	defer file.Close()

	vo, svcErr := a.Service.UploadChatImage(userID(r), header.Filename, header.Size, file)
	if svcErr != nil {
		a.fail(w, svcErr)
		return
	}
	ok(w, vo)
}

// chatPeerProfile GET /chat/profile?peerId=&peerType=
// 点击聊天头像时预览对方资料：学生返回完整简历，企业返回简介与在招岗位。
func (a *API) chatPeerProfile(w http.ResponseWriter, r *http.Request) {
	peerID := queryInt(r, "peerId", 0)
	peerType := queryString(r, "peerType")
	if peerID <= 0 || (peerType != "student" && peerType != "company") {
		respond(w, model.Error(400, "聊天对象不正确"))
		return
	}
	profile, err := a.Service.ChatPeerProfile(peerID, peerType)
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, profile)
}

// chatRead POST /chat/read
func (a *API) chatRead(w http.ResponseWriter, r *http.Request) {
	var req model.MarkReadRequest
	if err := decodeJSON(r, &req); err != nil {
		a.fail(w, err)
		return
	}
	if err := a.Service.MarkChatRead(userID(r), userType(r), req.PeerID, req.PeerType); err != nil {
		a.fail(w, err)
		return
	}
	ok(w, nil)
}

// chatWebSocket GET /chat/ws?token=xxx
//
// 为什么 token 放在 query 里？
// 浏览器原生 WebSocket API 不支持自定义请求头，无法像 axios 那样带 Authorization。
// 所以这里用 ?token=... 传递 JWT，并在升级连接前完成校验。
func (a *API) chatWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "缺少 token", http.StatusUnauthorized)
		return
	}
	claims, err := a.Service.Auth.Parse(token)
	if err != nil {
		http.Error(w, "token 无效或已过期", http.StatusUnauthorized)
		return
	}
	if claims.UserType != "student" && claims.UserType != "company" {
		http.Error(w, "只有学生和企业可以使用聊天", http.StatusForbidden)
		return
	}

	conn, err := chatUpgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade 内部已经写过 HTTP 错误响应，这里直接返回
		return
	}

	client := chat.NewClient(a.Service.ChatHub, conn, claims.UserID, claims.UserType, claims.Username)
	client.Start()
	client.PushJSON(map[string]any{
		"type":     "connected",
		"userId":   claims.UserID,
		"userType": claims.UserType,
	})
}

// registerChatWSHandler 把“收到消息 -> 业务处理”的回调注入 Hub。
// 放在 api 包是为了让 chat 包保持纯粹（只负责连接管理），避免循环依赖。
func registerChatWSHandler(svc *service.Service) {
	svc.ChatHub.OnMessage = func(client *chat.Client, msg chat.IncomingMessage) {
		_, err := svc.SendChatMessage(client.UserID, client.UserType, client.Username,
			msg.ReceiverID, msg.ReceiverType, msg.Content, msg.MsgType)
		if err == nil {
			return
		}
		text := err.Error()
		var appErr *model.AppError
		if errors.As(err, &appErr) {
			text = appErr.Msg
		}
		client.PushJSON(map[string]any{"type": "error", "message": text})
	}
}
