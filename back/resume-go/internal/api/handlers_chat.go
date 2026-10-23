package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
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
func (a *API) chatPeers(c *gin.Context) {
	peers, err := a.Service.ChatPeers(c.Request.Context(), userType(c), queryString(c, "keyword"))
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, map[string]any{"list": peers})
}

// chatConversations GET /chat/conversations
// 返回会话列表：对方、最后一条消息、未读数、在线状态。
func (a *API) chatConversations(c *gin.Context) {
	conversations, err := a.Service.ChatConversations(c.Request.Context(), userID(c), userType(c))
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, map[string]any{"list": conversations})
}

// chatHistory GET /chat/messages?peerId=&peerType=&beforeId=&limit=
func (a *API) chatHistory(c *gin.Context) {
	peerID := queryInt(c, "peerId", 0)
	peerType := queryString(c, "peerType")
	beforeID := int64(queryInt(c, "beforeId", 0))
	limit := queryInt(c, "limit", 30)

	messages, hasMore, err := a.Service.ChatHistory(userID(c), userType(c), peerID, peerType, beforeID, limit)
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, map[string]any{"messages": messages, "hasMore": hasMore})
}

// chatSend POST /chat/messages
// REST 发送接口。WebSocket 断开时前端可以降级用它发送。
func (a *API) chatSend(c *gin.Context) {
	var req model.SendChatRequest
	if err := decodeJSON(c, &req); err != nil {
		a.fail(c, err)
		return
	}
	message, err := a.Service.SendChatMessage(c.Request.Context(), userID(c), userType(c), username(c),
		req.ReceiverID, req.ReceiverType, req.Content, req.MsgType)
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, message)
}

// chatUploadImage POST /chat/uploadImage
// 聊天图片上传，表单字段名 image，最大 5MB。
func (a *API) chatUploadImage(c *gin.Context) {
	header, err := c.FormFile("image")
	if err != nil {
		respond(c, model.Error(400, "请选择要上传的图片"))
		return
	}
	file, err := header.Open()
	if err != nil {
		respond(c, model.Error(500, "图片上传失败，请稍后重试"))
		return
	}
	defer file.Close()

	vo, svcErr := a.Service.UploadChatImage(userID(c), header.Filename, header.Size, file)
	if svcErr != nil {
		a.fail(c, svcErr)
		return
	}
	ok(c, vo)
}

// chatPeerProfile GET /chat/profile?peerId=&peerType=
// 点击聊天头像时预览对方资料：学生返回完整简历，企业返回简介与在招岗位。
func (a *API) chatPeerProfile(c *gin.Context) {
	peerID := queryInt(c, "peerId", 0)
	peerType := queryString(c, "peerType")
	if peerID <= 0 || (peerType != "student" && peerType != "company") {
		respond(c, model.Error(400, "聊天对象不正确"))
		return
	}
	profile, err := a.Service.ChatPeerProfile(peerID, peerType)
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, profile)
}

// chatRead POST /chat/read
func (a *API) chatRead(c *gin.Context) {
	var req model.MarkReadRequest
	if err := decodeJSON(c, &req); err != nil {
		a.fail(c, err)
		return
	}
	if err := a.Service.MarkChatRead(userID(c), userType(c), req.PeerID, req.PeerType); err != nil {
		a.fail(c, err)
		return
	}
	ok(c, nil)
}

// ---------- 离线通知（由 RabbitMQ 异步生成） ----------

// chatResumePermission GET /chat/resume-permission?peerId=&peerType=
// 查询当前学生是否还能给这个企业发简历（对方回复前只能发一次）。
func (a *API) chatResumePermission(c *gin.Context) {
	permission, err := a.Service.ResumePermission(userID(c), userType(c),
		queryInt(c, "peerId", 0), queryString(c, "peerType"))
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, permission)
}

// chatNotifications GET /chat/notifications?limit=
func (a *API) chatNotifications(c *gin.Context) {
	list, err := a.Service.ChatNotifications(userID(c), userType(c), queryInt(c, "limit", 20))
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, map[string]any{"list": list})
}

func (a *API) chatNotificationsRead(c *gin.Context) {
	if err := a.Service.ReadChatNotifications(userID(c), userType(c)); err != nil {
		a.fail(c, err)
		return
	}
	ok(c, nil)
}

func (a *API) chatUnreadCount(c *gin.Context) {
	count, err := a.Service.UnreadNotificationCount(userID(c), userType(c))
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, map[string]any{"count": count})
}

// chatWebSocket GET /chat/ws?token=xxx
//
// 为什么 token 放在 query 里？
// 浏览器原生 WebSocket API 不支持自定义请求头，无法像 axios 那样带 Authorization。
// 所以这里用 ?token=... 传递 JWT，并在升级连接前完成校验。
func (a *API) chatWebSocket(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.String(http.StatusUnauthorized, "缺少 token")
		return
	}
	claims, err := a.Service.Auth.Parse(token)
	if err != nil {
		c.String(http.StatusUnauthorized, "token 无效或已过期")
		return
	}
	if claims.UserType != "student" && claims.UserType != "company" {
		c.String(http.StatusForbidden, "只有学生和企业可以使用聊天")
		return
	}

	conn, err := chatUpgrader.Upgrade(c.Writer, c.Request, nil)
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
		// WebSocket 没有 HTTP 请求上下文，这里用 Background 作为任务根上下文
		_, err := svc.SendChatMessage(context.Background(), client.UserID, client.UserType, client.Username,
			msg.ReceiverID, msg.ReceiverType, msg.Content, msg.MsgType)
		if err == nil {
			return
		}
		text := "消息发送失败，请稍后重试"
		var appErr *model.AppError
		if errors.As(err, &appErr) {
			text = appErr.Msg
		}
		client.PushJSON(map[string]any{"type": "error", "message": text})
	}
}
