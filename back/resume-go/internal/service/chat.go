package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"resume-backend/internal/model"
)

// ==================================================================
// 聊天业务层
// ==================================================================
// 这里集中写“聊天规则”，例如：
//   - 只有学生和企业可以聊天，管理员不参与；
//   - 学生只能找企业，企业只能找学生；
//   - 消息不能为空，最长 800 个字符；
//   - 接收者必须真实存在，防止前端伪造 receiverId。
// ==================================================================

const (
	maxChatContentLength = 800
	maxChatImageSize     = 5 * 1024 * 1024
)

var allowedChatImageExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

func isChatRole(userType string) bool {
	return userType == "student" || userType == "company"
}

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
func (s *Service) ChatPeers(meType, keyword string) ([]model.ChatPeer, error) {
	if !isChatRole(meType) {
		return nil, model.NewAppError(403, "只有学生和企业可以使用聊天")
	}
	peers, err := s.DB.ChatPeers(meType, keyword)
	if err != nil {
		return nil, err
	}
	for i := range peers {
		peers[i].Online = s.ChatHub.IsOnline(peers[i].UserID, peers[i].UserType)
	}
	return peers, nil
}

// ChatConversations 获取会话列表（按最后一条消息倒序）。
func (s *Service) ChatConversations(meID int, meType string) ([]model.Conversation, error) {
	if !isChatRole(meType) {
		return nil, model.NewAppError(403, "只有学生和企业可以使用聊天")
	}
	conversations, err := s.DB.ListConversations(meID, meType)
	if err != nil {
		return nil, err
	}
	for i := range conversations {
		conversations[i].Online = s.ChatHub.IsOnline(conversations[i].PeerID, conversations[i].PeerType)
	}
	return conversations, nil
}

// ChatHistory 获取与某个联系人的历史消息。
// beforeID 用于“向上加载更多”，返回的 hasMore 表示是否可能还有更早的消息。
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

// SendChatMessage 发送一条消息：校验 -> 入库 -> WebSocket 推送。
// 注意：即使对方不在线，消息也会入库，对方下次打开聊天页能看到。
// msgType：text（文本，content 为文字）/ image（图片，content 为图片地址）。
func (s *Service) SendChatMessage(senderID int, senderType, senderUsername string,
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
		msgType = "text"
	}
	if msgType != "text" && msgType != "image" {
		return nil, model.NewAppError(400, "不支持的消息类型")
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, model.NewAppError(400, "消息内容不能为空")
	}
	if msgType == "image" {
		// 图片必须是本站上传接口返回的地址，避免伪造外链
		if !strings.HasPrefix(content, "/uploads/chat/") {
			return nil, model.NewAppError(400, "图片地址不合法")
		}
	} else if utf8.RuneCountInString(content) > maxChatContentLength {
		return nil, model.NewAppError(400, "消息内容不能超过800个字符")
	}

	exists, err := s.DB.ChatUserExists(receiverID, receiverType)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, model.NewAppError(404, "接收者不存在")
	}

	// 查询一次发送者显示名，用于消息推送给对方时展示
	senderName, err := s.DB.ChatUserName(senderID, senderType)
	if err != nil {
		return nil, err
	}
	if senderName == "" {
		senderName = senderUsername
	}

	id, err := s.DB.InsertChatMessage(senderID, senderType, receiverID, receiverType, msgType, content)
	if err != nil {
		return nil, err
	}
	message, err := s.DB.FindChatMessageByID(id)
	if err != nil {
		return nil, err
	}

	// 同一条消息推给双方：
	//   - 接收者：实时看到新消息
	//   - 发送者：多标签页/多设备同步（前端按消息 id 去重）
	payload := map[string]any{"type": "message", "data": message}
	s.ChatHub.SendToUser(receiverID, receiverType, payload)
	s.ChatHub.SendToUser(senderID, senderType, payload)

	return message, nil
}

// MarkChatRead 进入会话时，把对方发给我的消息标记为已读。
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

// UploadChatImage 保存聊天图片，返回可直接访问的相对地址。
// 图片统一放在 uploads/chat 目录，静态路由 /uploads/ 会直接映射到 uploads 根目录。
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
