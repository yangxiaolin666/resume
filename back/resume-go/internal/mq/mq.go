// Package mq 封装 RabbitMQ 的「发布 / 消费」流程。
//
// 【简历对应点】
// 「通过 RabbitMQ 解耦消息通知与异步任务，针对重复消费设计幂等处理机制」。
//
// 【为什么要异步】
// 学生发一条聊天消息，核心链路只需要：校验 -> 写入 MySQL -> WebSocket 推送。
// 而「给对方记录一条离线通知」并不是用户马上要看到的结果，
// 放在主链路里会拖慢接口响应，也会因为通知服务抖动导致发消息失败。
// 所以主链路只负责把事件丢进队列，通知逻辑由消费者慢慢处理。
//
// 【幂等设计（面试高频）】
// RabbitMQ 保证的是「至少一次投递」，网络抖动、消费者重启都可能导致同一条消息被消费多次。
// 本项目的做法是三层防护：
//  1. 事件自带唯一的 EventID（由消息 ID + 事件类型拼成，天然唯一）；
//  2. 消费者先用 Redis SETNX 做快速去重；
//  3. 再向 MySQL 的 mq_consumed_event 表插入唯一键，数据库唯一索引做最终兜底。
//
// 【没有 RabbitMQ 时】
// 与 cache 一样做降级：Publish 变成打印日志的空操作，业务不受影响，
// 方便你在本地只启动 MySQL 就能跑通整个项目。
package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// QueueChatNotify 是聊天消息通知队列名。
const QueueChatNotify = "resume.chat.notify"

// EventTypeChatMessage 表示「一条聊天消息已创建」。
const EventTypeChatMessage = "chat.message.created"

// Event 是投递到队列里的消息体。
// 字段全部使用 JSON 可序列化的基础类型，避免消费者与生产者共享 Go 结构体版本。
type Event struct {
	EventID      string    `json:"eventId"`      // 全局唯一事件 ID，用于幂等
	Type         string    `json:"type"`         // 事件类型
	MessageID    int64     `json:"messageId"`    // chat_message.id
	SenderID     int       `json:"senderId"`     // 发送者 user_id
	SenderType   string    `json:"senderType"`   // student / company
	ReceiverID   int       `json:"receiverId"`   // 接收者 user_id
	ReceiverType string    `json:"receiverType"` // student / company
	Content      string    `json:"content"`      // 消息内容（文本或图片地址）
	MsgType      string    `json:"msgType"`      // text / image
	CreatedAt    time.Time `json:"createdAt"`
}

// Publisher 负责把事件写入队列。
// 内部维护一条长连接 + 一个 channel，写操作加锁，避免并发写 channel 触发协议错误。
type Publisher struct {
	url     string
	enabled bool

	mu   sync.Mutex
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewPublisher(url string, enabled bool) *Publisher {
	p := &Publisher{url: url, enabled: enabled}
	if !enabled {
		log.Printf("[mq] 已通过配置关闭 RabbitMQ，消息通知将直接跳过")
		return p
	}
	if err := p.connect(); err != nil {
		log.Printf("[mq] RabbitMQ 连接失败(%v)，消息通知降级为跳过", err)
	}
	return p
}

// connect 建立连接并声明队列。
// 队列声明是幂等的：不存在就创建，存在则校验参数是否一致。
func (p *Publisher) connect() error {
	conn, err := amqp.Dial(p.url)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}
	if _, err := ch.QueueDeclare(
		QueueChatNotify,
		true,  // durable：RabbitMQ 重启后队列还在
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return err
	}
	p.conn = conn
	p.ch = ch
	log.Printf("[mq] RabbitMQ 连接成功，队列=%s", QueueChatNotify)
	return nil
}

// Publish 发送事件。
// 业务方只需要打日志记录失败，不需要中断主流程——消息已经落库，
// 最坏情况只是少一条离线通知，不会丢聊天记录。
func (p *Publisher) Publish(ctx context.Context, event Event) error {
	if p == nil || !p.enabled {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ch == nil || p.conn == nil || p.conn.IsClosed() {
		if err := p.connect(); err != nil {
			return fmt.Errorf("RabbitMQ 不可用: %w", err)
		}
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	err = p.ch.PublishWithContext(ctx, "", QueueChatNotify, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent, // 持久化消息，配合 durable 队列防止重启丢消息
		MessageId:    event.EventID,   // 方便在 RabbitMQ 控制台按事件 ID 排查
		Timestamp:    time.Now(),
		Body:         body,
	})
	if err != nil {
		// 连接可能已经断了，下次调用时重连
		p.ch = nil
		return err
	}
	return nil
}

// Close 释放连接。
func (p *Publisher) Close() {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ch != nil {
		_ = p.ch.Close()
		p.ch = nil
	}
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
	}
}

// Handler 是消费者的业务回调。
// 返回 error 表示处理失败，消息会被丢弃（本项目的通知允许丢失，避免毒消息无限重试）。
// 如果你做主业务队列，建议配合死信队列做有限次重试。
type Handler func(ctx context.Context, event Event) error

// Consumer 负责从队列拉取事件并执行 Handler。
type Consumer struct {
	url     string
	enabled bool
	handler Handler
}

func NewConsumer(url string, enabled bool, handler Handler) *Consumer {
	return &Consumer{url: url, enabled: enabled, handler: handler}
}

// Start 启动消费循环。
// 这里用「外层 for + 断线重连」的经典写法：
// RabbitMQ 连接断开后，Consume 的 deliveries channel 会关闭，
// 我们等待几秒再重新建连，而不是直接退出进程。
func (c *Consumer) Start(ctx context.Context) {
	if c == nil || !c.enabled || c.handler == nil {
		log.Printf("[mq] 消费者未启用，跳过启动")
		return
	}
	go func() {
		for {
			if ctx.Err() != nil {
				return
			}
			if err := c.consume(ctx); err != nil {
				log.Printf("[mq] 消费连接中断: %v，5 秒后重连", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
		}
	}()
}

func (c *Consumer) consume(ctx context.Context) error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if _, err := ch.QueueDeclare(QueueChatNotify, true, false, false, false, nil); err != nil {
		return err
	}
	// 公平分发：每个消费者最多同时持有 20 条未 ACK 的消息，
	// 防止某个实例被一次性压入过多任务。
	if err := ch.Qos(20, 0, false); err != nil {
		return err
	}
	deliveries, err := ch.Consume(
		QueueChatNotify,
		"",    // consumer tag 留空由服务端生成
		false, // autoAck=false：处理成功后才手动 ACK
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		return err
	}
	log.Printf("[mq] 消费者已启动，监听队列=%s", QueueChatNotify)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("deliveries channel 已关闭")
			}

			var event Event
			if err := json.Unmarshal(delivery.Body, &event); err != nil {
				log.Printf("[mq] 消息格式错误，直接丢弃: %v", err)
				_ = delivery.Nack(false, false)
				continue
			}
			if err := c.handler(ctx, event); err != nil {
				log.Printf("[mq] 事件处理失败 event=%s err=%v", event.EventID, err)
				// requeue=false：避免坏消息反复入队形成死循环
				_ = delivery.Nack(false, false)
				continue
			}
			_ = delivery.Ack(false)
		}
	}
}
