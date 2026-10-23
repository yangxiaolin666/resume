package data

import (
	"database/sql"
	"errors"
	"time"

	"resume-backend/internal/model"
)

// ==================================================================
// 通知 & MQ 幂等的数据访问
// ==================================================================
// 这里对应简历里「RabbitMQ 解耦消息通知」「重复消费幂等」的持久化部分。
// ==================================================================

// 本文件涉及两张表：
//   - mq_consumed_event：MQ 消费幂等记录，event_id 是主键；
//   - chat_notification：离线消息通知，用于聊天列表的红点和离线提醒。
//
// 业务链路：
//
//	发送消息（chat.go 写 chat_message）-> 投递 MQ 事件
//	-> 消费者先 MarkEventConsumed 抢占事件（幂等）
//	-> 抢到了再 InsertNotification 写通知，抢不到直接 ack 跳过。
//
// 调用方是 internal/mq（消费者）和 internal/service/chat.go（查询/已读）。
// 设计上通知是「最终一致」的：即使通知写失败，聊天记录本身已经落库，
// 不影响核心功能，所以消费者可以做有限次重试。

// MarkEventConsumed 尝试把一个事件 ID 记录为「已消费」。
//
// 实现要点：
//   - mq_consumed_event.event_id 是主键；
//   - 使用 INSERT IGNORE，重复插入不会报错，只会影响 0 行；
//   - RowsAffected == 1 说明这是第一次消费，可以继续执行业务；
//   - RowsAffected == 0 说明之前已经消费过，直接跳过。
//
// 这是「数据库唯一索引兜底」的经典幂等方案：
// 即使 Redis 去重失效、多个消费者同时处理，也只有一个能插入成功。
//
// 为什么不用「先 SELECT 查有没有，没有就 INSERT」：
//
//	这是典型的 check-then-act 竞态——两个消费者可能同时查到“不存在”，
//	然后都去 INSERT，最终产生两条记录或一个报错。
//	加唯一索引后，判断和写入由数据库在一次原子操作里完成，
//	无论并发多高都只会有一个赢家。这是幂等实现里最可靠的一招。
//
// 为什么不能只看 error 判断是否重复：
//
//	INSERT IGNORE 会把“主键冲突”和“其他可忽略的警告/错误”（如数据截断）
//	一起吞掉，都表现为影响 0 行或 1 行。所以这里用 RowsAffected 作为判断依据，
//	语义是“这次真的插进去了 -> 首次消费”。若业务对错误敏感，
//	也可以改用 INSERT ... ON DUPLICATE KEY UPDATE 或其他方式区分。
//
// 附加说明：这条记录只增不删，表会随时间增长，
// 生产环境通常按 consumed_at 定期归档（比如只保留 7~30 天）。
func (d *DB) MarkEventConsumed(eventID string) (bool, error) {
	res, err := d.Exec(`INSERT IGNORE INTO mq_consumed_event (event_id, consumed_at) VALUES (?, NOW())`, eventID)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

// InsertNotification 写入一条离线通知。
// 同样使用 INSERT IGNORE + 唯一键 (message_id, receiver_id, receiver_type)，
// 保证同一条消息不会给同一个接收者生成多条通知。
//
// 唯一键 uk_notification_message 覆盖「一条消息 + 一个接收者」这个业务主键：
// 消息重投、消费者重启后重复处理，第二次插入会被唯一索引拦下，
// 影响 0 行但不报错，于是消费逻辑可以安全地重复执行。
// 这也是为什么这个函数只返回 error、不返回 id —— 调用方只关心“有没有写进去”，
// 不关心具体是哪一行。
//
// 注意 is_read 固定写 0，created_at 用 NOW()，都由数据库决定，
// 不接收调用方传时间，避免不同消费者机器时钟不一致。
func (d *DB) InsertNotification(n model.ChatNotification) error {
	_, err := d.Exec(`INSERT IGNORE INTO chat_notification
		(message_id, receiver_id, receiver_type, sender_id, sender_type, title, content, is_read, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 0, NOW())`,
		n.MessageID, n.ReceiverID, n.ReceiverType, n.SenderID, n.SenderType, n.Title, n.Content)
	return err
}

// ListNotifications 查询某个用户最近的离线通知（用于聊天列表的红点提醒）。
//
// 索引配合：WHERE (receiver_id, receiver_type) + ORDER BY id DESC
// 正好命中 idx_notification_inbox(receiver_id, receiver_type, is_read, id)
// 的前两列，id 在该前缀内有序，所以排序可以顺索引读取，不必额外 filesort。
//
// limit 做了白名单式收敛：
//   - 传 0、负数或超过 100 时一律回落到 20，
//     既防止调用方误传 0 导致“查不到数据”，
//     也防止有人传个巨大值把整张通知表拉出来（接口安全/性能兜底）。
//
// 只取最近若干条而不是全部，是因为通知只是“提醒”，
// 历史数据价值低，前端红点也只需要少量条目。
func (d *DB) ListNotifications(userID int, userType string, limit int) ([]model.ChatNotification, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := d.Query(`SELECT id, message_id, receiver_id, sender_id, sender_type, title, content, is_read, created_at
		FROM chat_notification
		WHERE receiver_id = ? AND receiver_type = ?
		ORDER BY id DESC LIMIT ?`, userID, userType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.ChatNotification{}
	for rows.Next() {
		var item model.ChatNotification
		var isRead int
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.MessageID, &item.ReceiverID, &item.SenderID,
			&item.SenderType, &item.Title, &item.Content, &isRead, &createdAt); err != nil {
			return nil, err
		}
		item.IsRead = isRead == 1
		item.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		list = append(list, item)
	}
	return list, rows.Err()
}

// MarkNotificationsRead 把某个用户的通知全部标记为已读。
//
// 条件只按接收者过滤，不区分是哪条消息：用户点开通知中心即视为全部查看。
// is_read = 0 让已读行不被重复更新，减少行锁和无谓写入。
// 与 chat_notification 的收件箱索引前缀一致，可走索引定位。
func (d *DB) MarkNotificationsRead(userID int, userType string) error {
	_, err := d.Exec(`UPDATE chat_notification SET is_read = 1
		WHERE receiver_id = ? AND receiver_type = ? AND is_read = 0`, userID, userType)
	return err
}

// CountUnreadNotifications 未读通知数量（前端红点）。
//
// COUNT(*) 配合 (receiver_id, receiver_type, is_read) 这组前缀，
// 可以只读索引完成统计（覆盖索引），不必回表取行数据。
//
// 这里的 ErrNoRows 分支其实是防御性的：
// COUNT(*) 永远会返回一行（没有匹配时返回 0），不会出现 ErrNoRows。
// 保留它是为了和其他以 QueryRow 取单值的函数保持一致的写法，
// 阅读时不要误以为“计数可能查不到”。
func (d *DB) CountUnreadNotifications(userID int, userType string) (int, error) {
	var count int
	err := d.QueryRow(`SELECT COUNT(*) FROM chat_notification
		WHERE receiver_id = ? AND receiver_type = ? AND is_read = 0`, userID, userType).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return count, err
}
