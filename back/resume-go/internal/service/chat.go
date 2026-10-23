package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"resume-backend/internal/chat"
	"resume-backend/internal/model"
	"resume-backend/internal/mq"
)

// ==================================================================
// 聊天业务层
// ==================================================================
// 这里集中写“聊天规则”，例如：
//   - 只有学生和企业可以聊天，管理员不参与；
//   - 学生只能找企业，企业只能找学生；
//   - 消息不能为空，最长 800 个字符；
//   - 接收者必须真实存在，防止前端伪造 receiverId。
//
// 【与其它层的关系】
//   - api：负责把 HTTP/JSON 参数绑定成方法入参，并把返回的 AppError 转成响应；
//   - data：消息、会话、通知的持久化（MySQL 是唯一可靠数据源）；
//   - chat.Hub：本进程内的 WebSocket 连接表，负责实时推送；
//   - cache：在线状态（多实例可见）与 MQ 消费幂等键；
//   - mq：把“离线通知”异步化，消费者最终回调本文件的 HandleChatNotify；
//   - rpc：消费者生成通知时通过内部 gRPC 取发送者昵称，失败可降级查库。
//
// 【对应简历能力】WebSocket 实时聊天、RabbitMQ 异步通知与幂等消费、Redis 在线状态。
// ==================================================================

const (
	maxChatContentLength = 800
	maxChatImageSize     = 5 * 1024 * 1024
)

var allowedChatImageExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

// 聊天只允许「学生 <-> 企业」这一对角色。
// 管理员虽然也能登录管理后台，但不参与业务聊天，
// 把判断收口成一个函数，避免每个方法各写一遍导致规则漂移。
func isChatRole(userType string) bool {
	return userType == "student" || userType == "company"
}

// oppositeRole 返回聊天对象的角色：学生找企业，企业找学生。
// 传入未知角色返回空串，调用方用「peerType != oppositeRole(meType)」即可一并拦住非法值。
func oppositeRole(userType string) string {
	if userType == "student" {
		return "company"
	}
	if userType == "company" {
		return "student"
	}
	return ""
}

// ChatPeers 获取当前用户可聊天的联系人列表。
//
// 角色不合法直接返回 403：这是权限问题，不是参数格式问题。
// 列表本身由 data 层按角色查询（学生看企业、企业看学生），
// 查完再逐个补充在线状态——在线状态在 Redis 里，不适合写进 SQL。
func (s *Service) ChatPeers(ctx context.Context, meType, keyword string) ([]model.ChatPeer, error) {
	if !isChatRole(meType) {
		return nil, model.NewAppError(403, "只有学生和企业可以使用聊天")
	}
	peers, err := s.DB.ChatPeers(meType, keyword)
	if err != nil {
		return nil, err
	}
	for i := range peers {
		peers[i].Online = s.isOnline(ctx, peers[i].UserID, peers[i].UserType)
	}
	return peers, nil
}

// ChatConversations 获取会话列表（按最后一条消息倒序）。
// 排序由 SQL 完成（按最后消息时间），这里同样只补在线状态。
func (s *Service) ChatConversations(ctx context.Context, meID int, meType string) ([]model.Conversation, error) {
	if !isChatRole(meType) {
		return nil, model.NewAppError(403, "只有学生和企业可以使用聊天")
	}
	conversations, err := s.DB.ListConversations(meID, meType)
	if err != nil {
		return nil, err
	}
	for i := range conversations {
		conversations[i].Online = s.isOnline(ctx, conversations[i].PeerID, conversations[i].PeerType)
	}
	return conversations, nil
}

// isOnline 判断用户是否在线。
// 优先查 Redis：即使服务部署了多个副本、WebSocket 连在别的实例上，也能得到正确结果；
// Redis 不可用时退回本进程 Hub 的连接表。
func (s *Service) isOnline(ctx context.Context, userID int, userType string) bool {
	if s.Cache.IsOnline(ctx, userID, userType) {
		return true
	}
	return s.ChatHub.IsOnline(userID, userType)
}

// ChatHistory 获取与某个联系人的历史消息。
// beforeID 用于“向上加载更多”，返回的 hasMore 表示是否可能还有更早的消息。
//
// 规则与取舍：
//   - 先校验双方角色必须互补、peerID 不能是自己，避免通过改参数读到别人的会话；
//   - limit 做上下限收敛（默认 30、最多 100），防止前端传 100000 把数据库拖垮；
//   - hasMore 用「本页是否查满 limit」来近似判断，省掉一次 COUNT 查询；
//     代价是当剩余消息数恰好等于 limit 时会多请求一次空页，可以接受。
func (s *Service) ChatHistory(meID int, meType string, peerID int, peerType string, beforeID int64, limit int) ([]model.ChatMessage, bool, error) {
	if !isChatRole(meType) {
		return nil, false, model.NewAppError(403, "只有学生和企业可以使用聊天")
	}
	if peerType != oppositeRole(meType) {
		return nil, false, model.NewAppError(400, "聊天对象类型不正确")
	}
	if peerID <= 0 || peerID == meID {
		return nil, false, model.NewAppError(400, "聊天对象不正确")
	}
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}

	messages, err := s.DB.ListChatMessages(meID, meType, peerID, peerType, beforeID, limit)
	if err != nil {
		return nil, false, err
	}
	return messages, len(messages) == limit, nil
}

// SendChatMessage 发送一条消息：校验 -> 入库 -> WebSocket 推送 -> 投递 MQ 事件。
// 注意：即使对方不在线，消息也会入库，对方下次打开聊天页能看到。
// msgType：text（文本，content 为文字）/ image（图片，content 为图片地址）。
//
// 【完整链路】
//  1. 参数校验：
//     - 发送者必须是学生/企业；
//     - 接收者角色必须与发送者互补，且不能是自己；
//     - 消息类型只能是 text/image；
//     - 内容去空格后不能为空；文本按「字符数」限制 800，
//     图片必须是本站上传接口返回的 /uploads/chat/ 地址。
//  2. 接收者存在性校验：查库确认 receiverId 真实存在，
//     防止前端伪造一个不存在的接收者，产生永远没人能看到的“孤儿消息”。
//  3. 入库：先 INSERT 拿到自增 ID，再按 ID 回查完整行。
//     回查是为了拿到数据库生成的 create_time 等字段，保证推送和返回给前端的数据一致。
//  4. WebSocket 双端推送：同一条消息同时推给接收者和发送者。
//     推给接收者是为了实时到达；推给发送者是为了多标签页/多设备同步，
//     前端按消息 id 去重即可，不会重复显示。
//  5. RabbitMQ 投递：交给消费者判断是否需要生成离线通知。
//
// 【为什么 MQ 放在最后】
// 前面的步骤都是「必须成功」的核心链路；投递通知属于「尽力而为」，
// 即使 RabbitMQ 暂时不可用，也不能让用户发不出消息。
//
// 失败时：校验类错误返回 4xx AppError；数据库错误原样返回（api 层按 500 处理）。
func (s *Service) SendChatMessage(ctx context.Context, senderID int, senderType, senderUsername string,
	receiverID int, receiverType, content, msgType string) (*model.ChatMessage, error) {
	if !isChatRole(senderType) {
		return nil, model.NewAppError(403, "只有学生和企业可以发送消息")
	}
	if receiverType != oppositeRole(senderType) {
		return nil, model.NewAppError(400, "只能和学生或企业聊天")
	}
	if receiverID <= 0 || receiverID == senderID {
		return nil, model.NewAppError(400, "接收者不正确")
	}

	msgType = strings.TrimSpace(msgType)
	if msgType == "" {
		// 前端老版本可能不传 msgType，默认按文本处理，保持向后兼容
		msgType = "text"
	}
	if msgType != "text" && msgType != "image" && msgType != "resume" {
		return nil, model.NewAppError(400, "不支持的消息类型")
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, model.NewAppError(400, "消息内容不能为空")
	}
	switch msgType {
	case "image":
		// 图片必须是本站上传接口返回的地址，避免伪造外链
		// 只信任自己生成的 /uploads/chat/ 前缀，可以挡住「把任意 URL 当图片发」的滥用。
		if !strings.HasPrefix(content, "/uploads/chat/") {
			return nil, model.NewAppError(400, "图片地址不合法")
		}
	case "resume":
		// 简历只有学生能发；地址同样只允许本站上传/下载地址，
		// 防止前端塞入外部链接或 javascript: 伪协议。
		if senderType != "student" {
			return nil, model.NewAppError(403, "只有学生可以发送简历")
		}
		if !isLocalResumeURL(content) {
			return nil, model.NewAppError(400, "简历地址不合法")
		}
		if utf8.RuneCountInString(content) > 500 {
			return nil, model.NewAppError(400, "简历地址过长")
		}
		// 频控：对方回复之前只能发一次简历。
		// 后端必须再校验一次，不能只依赖前端置灰按钮，
		// 否则用户直接调接口就能绕过限制。
		sent, err := s.DB.HasUnansweredResume(senderID, senderType, receiverID, receiverType)
		if err != nil {
			return nil, err
		}
		if sent {
			return nil, model.NewAppError(409, "对方回复前只能发送一次简历，请等待对方回复")
		}
	default: // text
		// 必须按 rune（字符）而不是 len（字节）计数：
		// 一个中文字符占 3 字节，用 len 会把 267 个汉字就判成超长。
		if utf8.RuneCountInString(content) > maxChatContentLength {
			return nil, model.NewAppError(400, "消息内容不能超过800个字符")
		}
	}

	// 校验接收者真实存在：JWT 只能证明「发送者是谁」，不能证明 receiverId 合法，
	// 所以必须回库确认，否则会写入指向不存在用户的消息。
	exists, err := s.DB.ChatUserExists(receiverID, receiverType)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, model.NewAppError(404, "接收者不存在")
	}

	// 查询一次发送者显示名，用于消息推送给对方时展示
	// 显示名可能为空（例如资料还没填），此时退回登录账号名，保证通知里一定有称呼。
	senderName, err := s.DB.ChatUserName(senderID, senderType)
	if err != nil {
		return nil, err
	}
	if senderName == "" {
		senderName = senderUsername
	}

	// 先写入数据库再推送：数据库是唯一可靠数据源。
	// 如果反过来先推送、入库失败，对方会看到一条刷新后就消失的“幽灵消息”。
	id, err := s.DB.InsertChatMessage(senderID, senderType, receiverID, receiverType, msgType, content)
	if err != nil {
		return nil, err
	}
	// 再按主键回查，拿到数据库生成的 create_time 等完整字段
	message, err := s.DB.FindChatMessageByID(id)
	if err != nil {
		return nil, err
	}

	// 同一条消息推给双方：
	//   - 接收者：实时看到新消息
	//   - 发送者：多标签页/多设备同步（前端按消息 id 去重）
	// SendToUser 只遍历本进程的连接表：跨实例的实时推送不在本方案范围内。
	// 但消息已经可靠入库，对方重新进入会话/刷新历史时仍能看到，不会丢消息。
	payload := map[string]any{"type": "message", "data": message}
	s.ChatHub.SendToUser(receiverID, receiverType, payload)
	s.ChatHub.SendToUser(senderID, senderType, payload)

	// 异步通知：把事件交给 RabbitMQ，由消费者判断是否需要生成离线通知。
	// EventID 由「事件类型 + 消息 ID」组成，天然唯一，消费者据此做幂等。
	// 为什么不用随机 UUID：重复投递同一条消息时 EventID 必须保持一致，去重才有意义。
	event := mq.Event{
		EventID:      fmt.Sprintf("%s:%d", mq.EventTypeChatMessage, message.ID),
		Type:         mq.EventTypeChatMessage,
		MessageID:    message.ID,
		SenderID:     senderID,
		SenderType:   senderType,
		ReceiverID:   receiverID,
		ReceiverType: receiverType,
		Content:      content,
		MsgType:      msgType,
		CreatedAt:    time.Now(),
	}
	if err := s.MQ.Publish(ctx, event); err != nil {
		// 通知失败不影响聊天主流程，只记录日志方便排查
		log.Printf("[chat] 消息通知投递失败 messageId=%d: %v", message.ID, err)
	}

	return message, nil
}

// MarkChatRead 进入会话时，把对方发给我的消息标记为已读。
// 只按「对方发给我的」方向更新，不会误改我发给对方的消息；
// peerID <= 0 直接拒绝，避免 WHERE 条件被拼成非法值。
func (s *Service) MarkChatRead(meID int, meType string, peerID int, peerType string) error {
	if !isChatRole(meType) {
		return model.NewAppError(403, "只有学生和企业可以使用聊天")
	}
	if peerType != oppositeRole(meType) || peerID <= 0 {
		return model.NewAppError(400, "聊天对象不正确")
	}
	return s.DB.MarkChatRead(meID, meType, peerID, peerType)
}

// ChatPeerProfile 聊天窗口中点击头像查看对方资料。
// 学生返回完整简历信息（与 /studentDetail 页一致），企业返回简介 + 在招岗位。
// 复用 buildStudentDetail 和岗位查询，保证聊天页看到的资料与其他页面一致。
// 对方资料不存在时返回 404；类型既不是 student 也不是 company 时返回 400。
func (s *Service) ChatPeerProfile(peerID int, peerType string) (map[string]any, error) {
	switch peerType {
	case "student":
		student, err := s.DB.FindStudentByUserID(peerID)
		if err != nil {
			return nil, err
		}
		if student == nil {
			return nil, model.NewAppError(404, "学生信息不存在")
		}
		detail, err := s.buildStudentDetail(student)
		if err != nil {
			return nil, err
		}
		return map[string]any{"userType": "student", "student": detail}, nil

	case "company":
		company, err := s.DB.FindCompanyByUserID(peerID)
		if err != nil {
			return nil, err
		}
		if company == nil {
			return nil, model.NewAppError(404, "企业信息不存在")
		}
		jobs, err := s.DB.ListJobPostingsByComID("active", company.ID)
		if err != nil {
			return nil, err
		}
		return map[string]any{"userType": "company", "company": company, "jobs": jobs}, nil

	default:
		return nil, model.NewAppError(400, "聊天对象类型不正确")
	}
}

// ---------- 在线状态钩子 ----------
// chat.Hub 在连接建立/断开时回调这两个方法，把状态同步到 Redis。
// 这样多实例部署时，A 实例也能知道用户是不是连在 B 实例上。
//
// 为什么用 Redis 而不是只查本进程 Hub：
// WebSocket 连接是有状态的，只会落在某一台实例上；
// 如果只用本地连接表，负载均衡把请求打到另一台实例时就会误判「不在线」。
// 回调里使用 context.Background()，因为连接生命周期不绑定某一次 HTTP 请求。
func (s *Service) HandleUserOnline(client *chat.Client) {
	s.Cache.SetOnline(context.Background(), client.UserID, client.UserType)
}

func (s *Service) HandleUserOffline(client *chat.Client) {
	s.Cache.SetOffline(context.Background(), client.UserID, client.UserType)
}

// ---------- 离线通知 ----------

// HandleChatNotify 是 RabbitMQ 消费者的业务处理函数。
//
// 处理步骤：
//  1. 幂等：数据库唯一索引 + Redis SETNX 双重去重；
//  2. 接收者在线则跳过（实时消息已经推送过了，不需要离线提醒）；
//  3. 通过内部 gRPC 获取发送者昵称（拿不到就用兜底文案）；
//  4. 写入 chat_notification 表。
//
// 【为什么要两层幂等】
// RabbitMQ 只能保证「至少一次」投递，网络抖动、消费者重启、ACK 丢失都可能让同一条消息被消费多次。
//   - 第一层 MySQL：mq_consumed_event.event_id 是主键，INSERT IGNORE 只有第一次能写成功，
//     这一层最可靠，即使 Redis 整个挂掉、多个消费者并发处理，也只有一个能拿到 first=true；
//   - 第二层 Redis SetNX：key 只存在 24 小时，作用是在多实例高并发时挡住大部分重复请求，
//     减少数据库压力。它只是「快」，不是「准」，所以不能单独依赖它。
//
// 顺序上先写数据库再占 Redis：
// 数据库是最终事实来源；若反过来，Redis 占位成功但数据库写入失败，
// 第二次重试会被 Redis 误判为重复而永久丢失通知。
//
// 失败时返回 error，由 MQ 消费框架决定是否重新入队/重试；
// 重复事件返回 nil，表示「已处理过，可以正常 ACK」。
func (s *Service) HandleChatNotify(ctx context.Context, event mq.Event) error {
	// 第一步：数据库唯一键兜底。这是最可靠的去重，不依赖 Redis 是否可用。
	first, err := s.DB.MarkEventConsumed(event.EventID)
	if err != nil {
		return err
	}
	if !first {
		log.Printf("[notify] 重复事件，跳过 event=%s", event.EventID)
		return nil
	}
	// 第二步：Redis 快速去重（多实例并发时减少数据库压力）。
	// 如果 Redis 不可用，MarkOnce 会返回 true，让数据库唯一索引继续兜底。
	if !s.Cache.MarkOnce(ctx, "resume:mq:consumed:"+event.EventID, 24*time.Hour) {
		return nil
	}

	if s.isOnline(ctx, event.ReceiverID, event.ReceiverType) {
		// 在线说明 WebSocket 已经实时推送过了，再写离线通知会让用户看到重复提醒。
		// 这里存在极小的竞态（查询后用户刚好下线），属于可接受的取舍。
		return nil
	}

	// 昵称获取按「gRPC -> 本地数据库 -> 固定文案」逐级降级：
	// 通知只是辅助功能，不能因为某个下游不可用就让整条消息消费失败、无限重试。
	senderName := "对方"
	if name, ok := s.RPC.UserDisplayName(ctx, event.SenderID, event.SenderType); ok && name != "" {
		senderName = name
	} else if name, err := s.DB.ChatUserName(event.SenderID, event.SenderType); err == nil && name != "" {
		// gRPC 不可用时的本地兜底，保证功能可用
		senderName = name
	}

	// 通知里只放摘要：图片统一显示 [图片]，文本截断到 60 个字符，
	// 避免私聊全文出现在通知列表里，也控制列表查询的数据量。
	preview := event.Content
	switch event.MsgType {
	case "image":
		preview = "[图片]"
	case "resume":
		preview = "[简历]"
	}
	preview = truncateRunes(preview, 60)

	notification := model.ChatNotification{
		MessageID:    event.MessageID,
		ReceiverID:   event.ReceiverID,
		ReceiverType: event.ReceiverType,
		SenderID:     event.SenderID,
		SenderType:   event.SenderType,
		Title:        senderName + " 发来一条新消息",
		Content:      preview,
	}
	if err := s.DB.InsertNotification(notification); err != nil {
		return err
	}
	// 兜底说明：InsertNotification 内部是 INSERT IGNORE + (message_id, receiver_id, receiver_type)
	// 唯一键，即使前面两层幂等都失效，同一条消息也只会生成一条通知。
	log.Printf("[notify] 已生成离线通知 messageId=%d receiver=%s:%d",
		event.MessageID, event.ReceiverType, event.ReceiverID)
	return nil
}

// ResumePermission 查询当前学生能否给指定企业发送简历。
// 企业端没有简历，直接返回不可发送；前端据此置灰按钮。
func (s *Service) ResumePermission(meID int, meType string, peerID int, peerType string) (*model.ResumePermission, error) {
	if !isChatRole(meType) {
		return nil, model.NewAppError(403, "只有学生和企业可以使用聊天")
	}
	if meType != "student" {
		return &model.ResumePermission{CanSend: false}, nil
	}
	if peerType != oppositeRole(meType) || peerID <= 0 {
		return nil, model.NewAppError(400, "聊天对象不正确")
	}
	blocked, err := s.DB.HasUnansweredResume(meID, meType, peerID, peerType)
	if err != nil {
		return nil, err
	}
	return &model.ResumePermission{CanSend: !blocked, WaitingReply: blocked}, nil
}

// ChatNotifications 查询离线通知列表。
// limit 的具体收敛在 data 层完成（默认 20、最多 100）；
// 这里只校验角色，防止管理员或非法角色读到业务通知。
func (s *Service) ChatNotifications(userID int, userType string, limit int) ([]model.ChatNotification, error) {
	if !isChatRole(userType) {
		return nil, model.NewAppError(403, "只有学生和企业可以使用聊天")
	}
	return s.DB.ListNotifications(userID, userType, limit)
}

// ReadChatNotifications 把当前用户的通知全部标记为已读。
// 是幂等操作：重复调用不会报错，也不影响已读数据。
func (s *Service) ReadChatNotifications(userID int, userType string) error {
	if !isChatRole(userType) {
		return model.NewAppError(403, "只有学生和企业可以使用聊天")
	}
	return s.DB.MarkNotificationsRead(userID, userType)
}

// UnreadNotificationCount 返回未读通知数，供前端红点使用。
// 只查计数、不取明细，避免为了一个数字拉回整页数据。
func (s *Service) UnreadNotificationCount(userID int, userType string) (int, error) {
	if !isChatRole(userType) {
		return 0, model.NewAppError(403, "只有学生和企业可以使用聊天")
	}
	return s.DB.CountUnreadNotifications(userID, userType)
}

// truncateRunes 按「字符」截断，而不是按字节。
// 中文一个字符占 3 个字节，直接按字节截断会得到乱码。
// 超长时补 "..."，让用户能看出内容被省略了。
func truncateRunes(text string, max int) string {
	text = strings.TrimSpace(text)
	if utf8.RuneCountInString(text) <= max {
		return text
	}
	runes := []rune(text)
	return string(runes[:max]) + "..."
}

// isLocalResumeURL 判断简历地址是否来自本站。
// 允许两种形式：
//   - /uploads/resume/xxx.pdf      静态文件地址（当前使用）
//   - /api/resume/download/xxx.pdf 旧版下载接口地址（历史数据兼容）
func isLocalResumeURL(url string) bool {
	return strings.HasPrefix(url, "/uploads/resume/") ||
		strings.HasPrefix(url, "/api/resume/download/")
}

// UploadChatImage 保存聊天图片，返回可直接访问的相对地址。
// 图片统一放在 uploads/chat 目录，静态路由 /uploads/ 会直接映射到 uploads 根目录。
//
// 安全与工程要点：
//   - 先校验大小和扩展名，避免把可执行文件当图片存进来；
//   - 不直接使用用户上传的原始文件名，而是用「用户 ID + 时间 + 随机数 + 扩展名」重新生成，
//     这样既不会重名覆盖，也能防止 ../../ 之类的路径穿越攻击；
//   - size <= 0 说明拿不到有效长度，同样拒绝，避免写入空文件；
//   - 失败时返回带业务码的 AppError，由 api 层转换成对应状态码。
func (s *Service) UploadChatImage(userID int, filename string, size int64, reader io.Reader) (*model.ChatImageVO, error) {
	if strings.TrimSpace(filename) == "" {
		return nil, model.NewAppError(400, "文件名无效")
	}
	if size <= 0 || size > maxChatImageSize {
		return nil, model.NewAppError(41301, "图片大小超限，最大支持5MB")
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedChatImageExt[ext] {
		return nil, model.NewAppError(41501, "图片格式不支持，仅支持 JPG、PNG、GIF、WEBP")
	}

	random := make([]byte, 4)
	if _, err := rand.Read(random); err != nil {
		return nil, err
	}
	uniqueName := fmt.Sprintf("%d_%s_%s%s", userID,
		time.Now().Format("20060102150405"), hex.EncodeToString(random), ext)

	dir := filepath.Join(filepath.Dir(s.UploadDir), "chat")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, model.NewAppError(500, "创建上传目录失败")
	}
	filePath := filepath.Join(dir, uniqueName)
	file, err := os.Create(filePath)
	if err != nil {
		return nil, model.NewAppError(500, "图片保存失败，"+err.Error())
	}
	if _, err := io.Copy(file, reader); err != nil {
		file.Close()
		return nil, model.NewAppError(500, "图片保存失败，"+err.Error())
	}
	file.Close()

	return &model.ChatImageVO{URL: "/uploads/chat/" + uniqueName}, nil
}
