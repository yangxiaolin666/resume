package rpc

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"resume-backend/internal/rpc/recruitpb"
)

// Client 是 gRPC 客户端封装。
// 采用「懒连接」：grpc.NewClient 不会立刻建立 TCP 连接，
// 第一次调用时才会真正连接，因此启动顺序不影响程序运行。
type Client struct {
	conn *grpc.ClientConn
	cli  recruitpb.RecruitInternalClient
}

// NewClient 创建客户端。enabled=false 时返回一个「空客户端」，
// 调用方通过 Enabled() 判断是否可用，从而实现无 gRPC 环境下的降级。
func NewClient(addr string, enabled bool) *Client {
	if !enabled {
		log.Printf("[grpc] 已通过配置关闭 gRPC 客户端")
		return &Client{}
	}
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("[grpc] 创建客户端失败: %v", err)
		return &Client{}
	}
	return &Client{conn: conn, cli: recruitpb.NewRecruitInternalClient(conn)}
}

func (c *Client) Enabled() bool { return c != nil && c.cli != nil }

// UserDisplayName 获取用户昵称；调用失败时返回 ok=false，由调用方兜底。
// 每次都设置 2 秒超时：内部调用不能无限等待，否则会拖垮消费者。
func (c *Client) UserDisplayName(ctx context.Context, userID int, userType string) (string, bool) {
	if !c.Enabled() {
		return "", false
	}
	callCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := c.cli.GetUserBrief(callCtx, &recruitpb.GetUserBriefRequest{
		UserId:   int32(userID),
		UserType: userType,
	})
	if err != nil {
		log.Printf("[grpc] GetUserBrief 调用失败: %v", err)
		return "", false
	}
	return resp.GetDisplayName(), resp.GetExists()
}

// JobBrief 查询职位摘要。
func (c *Client) JobBrief(ctx context.Context, jobID int) (*recruitpb.JobBrief, bool) {
	if !c.Enabled() {
		return nil, false
	}
	callCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := c.cli.GetJobBrief(callCtx, &recruitpb.GetJobBriefRequest{JobId: int32(jobID)})
	if err != nil {
		log.Printf("[grpc] GetJobBrief 调用失败: %v", err)
		return nil, false
	}
	return resp, resp.GetExists()
}

// Close 关闭连接。
func (c *Client) Close() {
	if c != nil && c.conn != nil {
		_ = c.conn.Close()
	}
}
