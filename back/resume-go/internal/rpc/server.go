// Package rpc 提供内部 gRPC 服务的服务端与客户端封装。
//
// 【调用链示例】
//
//	学生发消息 -> 写入 MySQL -> 投递 RabbitMQ -> 消费者取出事件
//	    -> 通过 gRPC 调用 GetUserBrief 获取发送者昵称
//	    -> 写入离线通知表
//
// 这样通知消费者不需要知道 student / company 表的结构，
// 只依赖一个稳定的接口契约（proto），符合「服务通信与部署」的简历描述。
package rpc

import (
	"context"
	"errors"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"resume-backend/internal/data"
	"resume-backend/internal/rpc/recruitpb"
)

// Server 实现 proto 中定义的 RecruitInternal 服务。
// 当前它和 HTTP 服务在同一个进程里启动（端口不同），
// 这样学习时只跑一个程序即可；真实微服务拆分时，把这个包单独部署即可。
type Server struct {
	recruitpb.UnimplementedRecruitInternalServer
	db *data.DB
}

func NewServer(db *data.DB) *Server {
	return &Server{db: db}
}

// GetUserBrief 查询学生/企业的显示名称。
func (s *Server) GetUserBrief(ctx context.Context, req *recruitpb.GetUserBriefRequest) (*recruitpb.UserBrief, error) {
	if req.GetUserType() != "student" && req.GetUserType() != "company" {
		return nil, status.Error(codes.InvalidArgument, "user_type 只能是 student 或 company")
	}
	userID := int(req.GetUserId())

	exists, err := s.db.ChatUserExists(userID, req.GetUserType())
	if err != nil {
		return nil, status.Error(codes.Internal, "查询用户失败")
	}
	if !exists {
		return &recruitpb.UserBrief{UserId: req.GetUserId(), UserType: req.GetUserType(), Exists: false}, nil
	}
	name, err := s.db.ChatUserName(userID, req.GetUserType())
	if err != nil {
		return nil, status.Error(codes.Internal, "查询用户昵称失败")
	}
	return &recruitpb.UserBrief{
		UserId:      req.GetUserId(),
		UserType:    req.GetUserType(),
		DisplayName: name,
		Exists:      true,
	}, nil
}

// GetJobBrief 查询职位摘要（职位标题 + 企业名称）。
func (s *Server) GetJobBrief(ctx context.Context, req *recruitpb.GetJobBriefRequest) (*recruitpb.JobBrief, error) {
	job, err := s.db.FindJobPostingByID(int(req.GetJobId()))
	if err != nil {
		return nil, status.Error(codes.Internal, "查询职位失败")
	}
	if job == nil {
		return &recruitpb.JobBrief{JobId: req.GetJobId(), Exists: false}, nil
	}
	return &recruitpb.JobBrief{
		JobId:       int32(job.ID),
		Title:       job.Title,
		CompanyName: job.ComName,
		Status:      job.Status,
		Exists:      true,
	}, nil
}

// Serve 启动 gRPC 服务，直到 ctx 被取消。
// 注意：graceful stop 会等待正在处理的请求结束，避免直接断开连接。
func Serve(ctx context.Context, addr string, db *data.DB) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		// 生产环境可在这里加拦截器：日志、鉴权、限流、链路追踪
		grpc.UnaryInterceptor(loggingInterceptor),
	)
	recruitpb.RegisterRecruitInternalServer(grpcServer, NewServer(db))

	go func() {
		<-ctx.Done()
		log.Printf("[grpc] 正在停止 gRPC 服务...")
		done := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			grpcServer.Stop() // 超时强制停止
		}
	}()

	log.Printf("[grpc] 内部服务已启动: %s", addr)
	if err := grpcServer.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return err
	}
	return nil
}

// loggingInterceptor 是最简单的 gRPC 一元拦截器示例：
// 记录方法名和耗时。理解它之后，可以按同样方式加鉴权、Prometheus 指标等。
func loggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	log.Printf("[grpc] %s 耗时=%s err=%v", info.FullMethod, time.Since(start), err)
	return resp, err
}
