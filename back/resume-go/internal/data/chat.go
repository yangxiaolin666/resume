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

// 本文件负责 chat_message 表：学生 <-> 企业的私聊消息，以及由它派生出来的
// 「可聊天对象列表」「会话列表」「已读标记」。
//
// 业务定位：简历项目里的「站内即时通讯」模块，前端是类微信的聊天窗口；
// 调用方是 internal/service/chat.go。消息发送成功后会再投递一条 MQ 事件，
// 由消费者异步写 chat_notification（离线通知），这部分见 notification.go。
//
// 表上的两个索引决定了下面每条 SQL 的写法（schema.sql）：
//   - idx_chat_pair(sender_id, sender_type, receiver_id, receiver_type, id)
//     支撑「拉取我和某人的历史消息」：SQL 的 WHERE 条件必须按这个顺序给出，
//     ORDER BY m.id 也能顺着索引读，不需要额外排序；
//   - idx_chat_inbox(receiver_id, receiver_type, is_read, id)
//     支撑「我的未读消息」和「收件箱」方向的查询。
// 注意 idx_chat_pair 是「发送方视角」建的，所以查询两个人之间的消息时
// 要写两个方向的 OR 条件（我发给他 + 他发给我），否则只能用到一半索引。

// chatMessageSelect 统一的消息查询语句。
// 这里通过 LEFT JOIN 学生表和企业表，把发送者/接收者的显示名称一起查出来，
// 避免前端再发一次请求去查名字。
//
// 为什么是 4 个 LEFT JOIN：
// 一条消息有「发送者」和「接收者」两个身份，每个身份都可能是学生或企业，
// 所以发送者要 join 一次 student、一次 company，接收者同理，共 4 个别名。
// 每个 JOIN 的 ON 条件里都带上 `m.sender_type = 'student'` 这样的类型判断：
// 只有类型匹配时才真正关联，否则 join 结果为空——
// 这样可以避免“user_id 恰好和某个学生 id 相同”造成的错配。
//
// 为什么用 LEFT JOIN 而不是 INNER JOIN：
// 消息必须能查出来，哪怕对应的学生/企业资料已被删除（名字就为空）。
//
// COALESCE(ss.name, cc.com_name, ”) 的作用：
// 依次取第一个非 NULL 的值，最后兜底成空字符串，
// 保证 Go 侧 Scan 到的一定是 string 而不是 NULL，
// 前端也不用处理 null 名称。
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

// scanChatMessage 把一行结果映射成 model.ChatMessage。
//
// is_read 在 MySQL 里是 tinyint（0/1），Go 没有内建的“tinyint -> bool”转换，
// 所以先 Scan 进 int 再显式比较 == 1，语义更直观，也避免依赖驱动的隐式转换。
//
// created_at 用 time.Time 接收后格式化成 "2006-01-02 15:04:05" 字符串：
// 聊天界面只按这个格式展示，统一在这里转好，前端就不用各自处理时区格式。
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
//
// 这里多写了一个 LIMIT 1，属于防御性写法：
// user_id 在 student/company 上通常没有唯一约束，一旦出现重复资料，
// 查询会返回多行。QueryRow 只会取第一行、静默丢弃其余行，
// 但“取到的是哪一行”是不确定的。显式 LIMIT 1 让意图清楚，也避免数据库多算。
//
// 返回 ("", nil) 表示查不到（或类型不认识），调用方可以退回用用户 ID 展示。
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
//
// 查询条件是两个字段的等值组合，属于「认证 + 授权」的前置检查：
// 光有 user_id 不够，还要类型匹配，才能确定对方确实是可以对话的学生/企业。
// 这里只返回 bool，不返回具体用户，避免把不必要的资料暴露给调用方。
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
//
// 实现细节：
//
//  1. 先 JOIN user 再 JOIN 资料表，是为了确保“账号和资料都存在”：
//     只有 user 但没填 company 资料的账号不应该出现在聊天对象里。
//     JOIN 走 c.user_id / s.user_id，依赖这两列上的索引（否则每次都要全表扫）。
//
//  2. 关键技巧是 `(? = ” OR c.com_name LIKE ?)`：
//     用「关键字为空 或 名称匹配」来表达“关键字可选”。
//     这样一条 SQL 同时覆盖“不搜索”和“搜索”两种情况，
//     不需要在 Go 里拼两种 SQL，也不会因为拼字符串引入注入风险。
//     代价是同一个 keyword 要传两次（一次用于判空、一次用于 LIKE），
//     而且 OPTIONAL 条件可能让优化器放弃索引、走全表扫描（数据量小无所谓）。
//
//  3. like := "%" + keyword + "%" 是在 Go 里拼的通配符，但它是作为参数值绑定的，
//     不是拼进 SQL 文本，所以安全；LIKE ? 里的 ? 由驱动负责转义。
//
//  4. 企业端用 CONCAT_WS(' / ', NULLIF(...), NULLIF(...)) 生成“专业 / 学历”的描述：
//     CONCAT_WS 会自动跳过 NULL 参数，NULLIF(x, ”) 把空串转成 NULL，
//     于是没填学历时不会出现 "计算机 / " 这样的多余分隔符。
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
//
// 入参含义：
//   - senderID / receiverID：双方 user_id（不是 student.id / company.id）；
//   - senderType / receiverType：student 或 company，和 ID 一起才能唯一确定一个人；
//   - msgType：text / image；
//   - content：文本内容，或图片地址。
//
// is_read 显式写 0（未读），created_at 用 NOW() 取数据库时间：
// 消息顺序完全依赖自增 id，时间只用于展示，所以不需要应用层参与。
//
// 六个值全部用 ? 绑定：聊天内容直接来自用户输入，是最典型的注入高风险点，
// 绝对不能拼字符串。返回的 id 会作为 MQ 事件的 messageId，用于离线通知关联。
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
// 用于发送成功后的回显、以及 MQ 消费者按 messageId 反查消息内容。
// 主键等值查询，走 chat_message 的聚簇索引，最多一行。
func (d *DB) FindChatMessageByID(id int64) (*model.ChatMessage, error) {
	return scanChatMessage(d.QueryRow(chatMessageSelect+` WHERE m.id = ?`, id))
}

// ListChatMessages 查询当前用户与某个联系人的历史消息。
// beforeID > 0 时表示“加载更早的消息”（按 id 倒序取一页，再翻转为正序）。
//
// WHERE 里为什么要写两段 OR：
//
//	聊天记录是双向的——我发给他的、他发给我的都属于同一个会话。
//	而 idx_chat_pair 是按 (sender_id, sender_type, receiver_id, receiver_type, id)
//	建的「发送方视角」索引，所以必须拆成两个方向分别匹配：
//	第一段覆盖“我发给他”（能用上索引的最左前缀），
//	第二段覆盖“他发给我”（在索引里是另一端的有序区间）。
//	MySQL 对 OR 可能做 index merge，或拆成两次索引范围扫描再合并。
//
// 为什么这里用 id 游标而不是 OFFSET 分页：
//
//	聊天是「上滑加载更多」的无限列表，如果用 LIMIT/OFFSET，
//	每次都要扫描并丢弃前面所有消息，越往前翻越慢；
//	更重要的是，新消息不断插入会让 OFFSET 的窗口整体偏移，
//	导致翻页时出现重复或漏读。用 `m.id < beforeID` 做游标（keyset pagination），
//	锚点是“上一页最旧那条的 id”，新消息插入完全不影响，且每次都能走索引定位，
//	耗时与已加载的消息数量无关。这是深分页优化在聊天场景的典型应用。
//
// 先按 id DESC 取最近 limit 条，再在 Go 里反转成升序：
//
//	这样一次查询就能拿到“离 now 最近的 N 条”，而不是从头开始取；
//	反转后前端按“旧 -> 新”渲染，和微信的上滑加载体验一致。
func (d *DB) ListChatMessages(meID int, meType string, peerID int, peerType string, beforeID int64, limit int) ([]model.ChatMessage, error) {
	// 8 个参数对应两段 OR 里的 4 个 ?：先是“我->对方”，再是“对方->我”。
	// 顺序必须与 SQL 中 ? 出现的顺序严格一致，否则条件会串位。
	query := chatMessageSelect + `
		WHERE ((m.sender_id = ? AND m.sender_type = ? AND m.receiver_id = ? AND m.receiver_type = ?)
			OR (m.sender_id = ? AND m.sender_type = ? AND m.receiver_id = ? AND m.receiver_type = ?))`
	args := []any{meID, meType, peerID, peerType, peerID, peerType, meID, meType}
	// 游标条件：只要比上一页最旧那条更早（id 更小）的消息，天然去重、无窗口偏移问题。
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
//
// 再展开几个细节，这段 SQL 是「先聚合再回表」的典型写法：
//
//	· CASE 的作用是「把双向的消息归一成单向」：
//	  对我而言，只要 sender 是我，对话对象就是 receiver；否则对话对象就是 sender。
//	  这样同一个会话的两个方向才会被分到同一组里。
//
//	· GROUP BY peer_id, peer_type 直接用了 SELECT 里的别名。
//	  MySQL 允许 GROUP BY 引用 select 别名，但在标准 SQL 里不合法，
//	  写成 GROUP BY 完整表达式会更可移植。
//	  分组键必须是 (peer_id, peer_type) 两个字段：不同角色的 user_id 可能重复，
//	  只按 peer_id 分组会把“学生 5 号”和“企业 5 号”混成一个会话。
//
//	· MAX(id) 而不是 MAX(created_at)：id 自增、唯一且单调，顺序最可靠；
//	  created_at 只精确到秒，同一秒的多条消息无法区分先后。
//
//	· SUM(CASE WHEN ... THEN 1 ELSE 0 END) 是条件计数的经典写法
//	  （等价于 COUNT(CASE WHEN ... THEN 1 END)，因为 COUNT 会跳过 NULL）。
//	  这里统计的是「所有历史未读」，而不是“当前页未读”，所以必须覆盖整个会话。
//
//	· 外层 JOIN chat_message m ON m.id = t.max_id：
//	  子查询只能给出 max_id，拿不到最后一条的内容/类型/时间，需要再按主键回表取一次。
//	  不回表直接在子查询里取 content 是不行的——聚合查询里非聚合列的值是不确定的
//	  （除非用 MySQL 的 loose group by 特性，但那不可靠）。
//
//	· 排序用 t.max_id DESC：max_id 越大说明会话越新，
//	  且子查询里已经算好了这个值，外层不用再按时间排序。
//
// 索引方面：
//
//	WHERE 的 (sender_id, sender_type) 和 (receiver_id, receiver_type) 两个方向
//	分别对应 idx_chat_pair 和 idx_chat_inbox 的最左前缀；
//	SUM 里的 is_read = 0 正是 idx_chat_inbox 的第三列，
//	所以“我的未读消息”这部分可以只读索引、不回表。
//	不过分组和聚合仍然要处理我参与的全部消息，消息量很大时，
//	通常会把会话摘要单独冗余成一张表（或用 Redis 维护），而不是每次都现算。
//
// 参数说明：子查询里 peer_id/peer_type 两个 CASE、SUM 里一个 CASE、
// WHERE 里两个方向，一共 5 组 (meID, meType)，也就是 10 个占位符。
// 这是这段 SQL 看着参数多的原因，顺序必须与 ? 出现的位置完全一致。
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
		// 最后一条是不是我自己发的：前端用它决定是否显示“我：xxx”前缀。
		// 必须同时比较 ID 和类型，因为学生和企业可能撞号。
		item.LastFromMe = lastSenderID == meID && lastSenderType == meType
		conversations = append(conversations, item)
	}
	return conversations, rows.Err()
}

// HasUnansweredResume 判断「发送者在对方最后一次回复之后，是否已经发过简历」。
//
// 业务规则：对方回复前只能发一次简历。
// 判断方式不是简单看历史上有没有简历，而是看「对方最后一条消息之后」是否已有简历：
//   - 企业回复后，MAX(id) 变大，之前的简历不再计入，额度恢复；
//   - 企业一直没回复，则之前发过的简历仍然拦住第二次发送。
//
// SQL 说明：
//   - 子查询取「接收者发给发送者」的最大消息 id（COALESCE 处理从未回复的情况）；
//   - 外层统计 id 大于该值的简历消息数量；
//   - 两个方向都能命中 idx_chat_pair(sender_id, sender_type, receiver_id, receiver_type, id)。
func (d *DB) HasUnansweredResume(senderID int, senderType string, receiverID int, receiverType string) (bool, error) {
	var count int
	err := d.QueryRow(`SELECT COUNT(*) FROM chat_message
		WHERE sender_id = ? AND sender_type = ? AND receiver_id = ? AND receiver_type = ?
		  AND msg_type = 'resume'
		  AND id > COALESCE((
			SELECT MAX(id) FROM chat_message
			WHERE sender_id = ? AND sender_type = ? AND receiver_id = ? AND receiver_type = ?
		  ), 0)`,
		senderID, senderType, receiverID, receiverType,
		receiverID, receiverType, senderID, senderType).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// MarkChatRead 把“某个联系人发给我的未读消息”全部标记为已读。
//
// 条件方向很重要：sender 是对方、receiver 是我，
// 即只更新「我收到的」消息，不能把「我发出去的」也算进来。
// 末尾的 is_read = 0 有两个作用：
//   - 语义上只处理未读，已读的不必重复写；
//   - 减少实际更新的行数，降低锁范围和无谓的 binlog 写入。
//
// 这条 UPDATE 的 WHERE 前缀 (sender_id, sender_type, receiver_id, receiver_type)
// 与 idx_chat_pair 的最左前缀一致，可以走索引定位而不是全表扫描。
// 打开会话时调用一次，属于高频但轻量的写操作。
func (d *DB) MarkChatRead(meID int, meType string, peerID int, peerType string) error {
	_, err := d.Exec(`UPDATE chat_message SET is_read = 1
		WHERE sender_id = ? AND sender_type = ? AND receiver_id = ? AND receiver_type = ? AND is_read = 0`,
		peerID, peerType, meID, meType)
	return err
}
