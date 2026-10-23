package data

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"resume-backend/internal/model"
)

// ==================================================================
// 聊天数据访问层
// ==================================================================
// 这一层只负责“怎么读写 MySQL”，不处理业务规则。
// 业务规则（谁可以给谁发、内容长度校验等）放在 service 层。
// ==================================================================

// chatMessageSelect 统一的消息查询语句。
// 这里通过 LEFT JOIN 学生表和企业表，把发送者/接收者的显示名称一起查出来，
// 避免前端再发一次请求去查名字。
const chatMessageSelect = `SELECT
	m.id, m.sender_id, m.sender_type, m.receiver_id, m.receiver_type,
	m.content, m.is_read, m.msg_type, m.created_at,
	COALESCE(ss.name, cc.com_name, '') AS sender_name,
	COALESCE(rs.name, rc.com_name, '') AS receiver_name
	FROM chat_message m
	LEFT JOIN student ss ON m.sender_type = 'student' AND ss.user_id = m.sender_id
	LEFT JOIN company cc ON m.sender_type = 'company' AND cc.user_id = m.sender_id
	LEFT JOIN student rs ON m.receiver_type = 'student' AND rs.user_id = m.receiver_id
	LEFT JOIN company rc ON m.receiver_type = 'company' AND rc.user_id = m.receiver_id`

func scanChatMessage(sc rowScanner) (*model.ChatMessage, error) {
	var msg model.ChatMessage
	var isRead int
	var createdAt time.Time
	err := sc.Scan(&msg.ID, &msg.SenderID, &msg.SenderType, &msg.ReceiverID, &msg.ReceiverType,
		&msg.Content, &isRead, &msg.MsgType, &createdAt, &msg.SenderName, &msg.ReceiverName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	msg.IsRead = isRead == 1
	msg.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
	return &msg, nil
}

// ChatUserName 查询某个聊天用户的显示名称。
// 学生取 student.name，企业取 company.com_name。
func (d *DB) ChatUserName(userID int, userType string) (string, error) {
	var name sql.NullString
	var err error
	switch userType {
	case "student":
		err = d.QueryRow(`SELECT name FROM student WHERE user_id = ? LIMIT 1`, userID).Scan(&name)
	case "company":
		err = d.QueryRow(`SELECT com_name FROM company WHERE user_id = ? LIMIT 1`, userID).Scan(&name)
	default:
		return "", nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return str(name), nil
}

// ChatUserExists 校验 user_id + user_type 是否真实存在。
// 发消息前必须校验，防止伪造 receiverId。
func (d *DB) ChatUserExists(userID int, userType string) (bool, error) {
	var count int
	err := d.QueryRow(`SELECT COUNT(*) FROM user WHERE user_id = ? AND user_type = ?`, userID, userType).Scan(&count)
	return count > 0, err
}

// ChatPeers 查询当前用户可聊天的对象：
//   - 学生 -> 所有企业
//   - 企业 -> 所有学生
//
// keyword 支持按名称模糊搜索。
func (d *DB) ChatPeers(userType, keyword string) ([]model.ChatPeer, error) {
	peers := []model.ChatPeer{}
	like := "%" + strings.TrimSpace(keyword) + "%"

	switch userType {
	case "student":
		// 学生只能找企业聊天
		rows, err := d.Query(`SELECT u.user_id, c.com_name, COALESCE(c.com_region, '')
			FROM user u JOIN company c ON c.user_id = u.user_id
			WHERE u.user_type = 'company' AND (? = '' OR c.com_name LIKE ?)
			ORDER BY c.com_name`, strings.TrimSpace(keyword), like)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var peer model.ChatPeer
			var name, description sql.NullString
			if err := rows.Scan(&peer.UserID, &name, &description); err != nil {
				return nil, err
			}
			peer.UserType = "company"
			peer.Name = str(name)
			peer.Description = str(description)
			peers = append(peers, peer)
		}
		return peers, rows.Err()

	case "company":
		// 企业只能找学生聊天
		rows, err := d.Query(`SELECT u.user_id, s.name,
			CONCAT_WS(' / ', NULLIF(s.major, ''), NULLIF(s.degree, ''))
			FROM user u JOIN student s ON s.user_id = u.user_id
			WHERE u.user_type = 'student' AND (? = '' OR s.name LIKE ?)
			ORDER BY s.name`, strings.TrimSpace(keyword), like)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var peer model.ChatPeer
			var name, description sql.NullString
			if err := rows.Scan(&peer.UserID, &name, &description); err != nil {
				return nil, err
			}
			peer.UserType = "student"
			peer.Name = str(name)
			peer.Description = str(description)
			peers = append(peers, peer)
		}
		return peers, rows.Err()
	}
	return peers, nil
}

// InsertChatMessage 插入一条消息，返回自增主键。
func (d *DB) InsertChatMessage(senderID int, senderType string, receiverID int, receiverType, msgType, content string) (int64, error) {
	res, err := d.Exec(`INSERT INTO chat_message
		(sender_id, sender_type, receiver_id, receiver_type, content, msg_type, is_read, created_at)
		VALUES (?, ?, ?, ?, ?, ?, 0, NOW())`,
		senderID, senderType, receiverID, receiverType, content, msgType)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// FindChatMessageByID 查询单条消息（带双方名称）。
func (d *DB) FindChatMessageByID(id int64) (*model.ChatMessage, error) {
	return scanChatMessage(d.QueryRow(chatMessageSelect+` WHERE m.id = ?`, id))
}

// ListChatMessages 查询当前用户与某个联系人的历史消息。
// beforeID > 0 时表示“加载更早的消息”（按 id 倒序取一页，再翻转为正序）。
func (d *DB) ListChatMessages(meID int, meType string, peerID int, peerType string, beforeID int64, limit int) ([]model.ChatMessage, error) {
	query := chatMessageSelect + `
		WHERE ((m.sender_id = ? AND m.sender_type = ? AND m.receiver_id = ? AND m.receiver_type = ?)
			OR (m.sender_id = ? AND m.sender_type = ? AND m.receiver_id = ? AND m.receiver_type = ?))`
	args := []any{meID, meType, peerID, peerType, peerID, peerType, meID, meType}
	if beforeID > 0 {
		query += ` AND m.id < ?`
		args = append(args, beforeID)
	}
	query += ` ORDER BY m.id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := d.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 数据库按 id 倒序查（方便取最近 N 条），这里反转为“旧 -> 新”，
	// 前端渲染时就是从下往上追加，和微信一致。
	messages := []model.ChatMessage{}
	for rows.Next() {
		msg, err := scanChatMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, *msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

// ListConversations 查询会话列表。
// 思路：
//  1. 先用 CASE 把“我发给别人”和“别人发给我”的消息统一成 peer_id / peer_type；
//  2. 按 peer 分组，取 MAX(id) 作为该会话最后一条消息；
//  3. 同时 SUM(is_read=0 且接收者是我) 得到未读数；
//  4. 再 JOIN 回消息表拿到最后一条消息的内容和时间。
func (d *DB) ListConversations(meID int, meType string) ([]model.Conversation, error) {
	query := `SELECT t.peer_id, t.peer_type, t.unread,
		COALESCE(ss.name, cc.com_name, '') AS peer_name,
		m.content, m.msg_type, m.created_at, m.sender_id, m.sender_type
		FROM (
			SELECT
				CASE WHEN sender_id = ? AND sender_type = ? THEN receiver_id ELSE sender_id END AS peer_id,
				CASE WHEN sender_id = ? AND sender_type = ? THEN receiver_type ELSE sender_type END AS peer_type,
				MAX(id) AS max_id,
				SUM(CASE WHEN receiver_id = ? AND receiver_type = ? AND is_read = 0 THEN 1 ELSE 0 END) AS unread
			FROM chat_message
			WHERE (sender_id = ? AND sender_type = ?) OR (receiver_id = ? AND receiver_type = ?)
			GROUP BY peer_id, peer_type
		) t
		JOIN chat_message m ON m.id = t.max_id
		LEFT JOIN student ss ON t.peer_type = 'student' AND ss.user_id = t.peer_id
		LEFT JOIN company cc ON t.peer_type = 'company' AND cc.user_id = t.peer_id
		ORDER BY t.max_id DESC`
	rows, err := d.Query(query,
		meID, meType, meID, meType, meID, meType, meID, meType, meID, meType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations := []model.Conversation{}
	for rows.Next() {
		var item model.Conversation
		var peerName, lastMessage, lastMsgType sql.NullString
		var createdAt time.Time
		var lastSenderID int
		var lastSenderType string
		if err := rows.Scan(&item.PeerID, &item.PeerType, &item.Unread, &peerName,
			&lastMessage, &lastMsgType, &createdAt, &lastSenderID, &lastSenderType); err != nil {
			return nil, err
		}
		item.PeerName = str(peerName)
		item.LastMessage = str(lastMessage)
		item.LastMsgType = str(lastMsgType)
		item.LastTime = createdAt.Format("2006-01-02 15:04:05")
		item.LastFromMe = lastSenderID == meID && lastSenderType == meType
		conversations = append(conversations, item)
	}
	return conversations, rows.Err()
}

// MarkChatRead 把“某个联系人发给我的未读消息”全部标记为已读。
func (d *DB) MarkChatRead(meID int, meType string, peerID int, peerType string) error {
	_, err := d.Exec(`UPDATE chat_message SET is_read = 1
		WHERE sender_id = ? AND sender_type = ? AND receiver_id = ? AND receiver_type = ? AND is_read = 0`,
		peerID, peerType, meID, meType)
	return err
}
