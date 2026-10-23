package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"resume-backend/internal/api"
	"resume-backend/internal/auth"
	"resume-backend/internal/cache"
	"resume-backend/internal/chat"
	"resume-backend/internal/config"
	"resume-backend/internal/data"
	"resume-backend/internal/mq"
	"resume-backend/internal/rpc"
	"resume-backend/internal/service"
)

// main 是整个后端的「组装车间」。
//
// 【启动顺序】
//  1. 读取配置；
//  2. 连接 MySQL（强依赖，失败直接退出）；
//  3. 连接 Redis / RabbitMQ（弱依赖，失败自动降级，不阻塞启动）；
//  4. 启动 gRPC 内部服务 + RabbitMQ 消费者；
//  5. 启动 Gin HTTP 服务；
//  6. 监听 Ctrl+C / SIGTERM，优雅关闭。
//
// 这种「强依赖必须成功、弱依赖允许降级」的设计，
// 可以避免因为一个缓存或消息队列没起来，整个网站都打不开。
func main() {
	cfg := config.Load()

	// ---------- 1. MySQL：核心数据源 ----------
	db, err := data.Open(cfg.DSN())
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}

	// ---------- 2. Redis：缓存 / 在线状态 / 幂等 ----------
	cacheStore := cache.New(cache.Config{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
		Enabled:  cfg.RedisEnabled,
	})
	defer cacheStore.Close()

	// ---------- 3. RabbitMQ：异步通知 ----------
	publisher := mq.NewPublisher(cfg.RabbitMQURL, cfg.RabbitMQEnabled)
	defer publisher.Close()

	// ---------- 4. gRPC 客户端：内部服务调用 ----------
	rpcClient := rpc.NewClient(cfg.GRPCInternal, cfg.GRPCEnabled)
	defer rpcClient.Close()

	// ---------- 5. 业务层 ----------
	chatHub := chat.NewHub()
	svc := service.New(db, auth.NewManager(cfg.JWTSecret), cacheStore, chatHub,
		cfg.UploadDir, publisher, rpcClient)

	// WebSocket 上线/下线时，把状态同步到 Redis
	chatHub.OnRegister = svc.HandleUserOnline
	chatHub.OnUnregister = svc.HandleUserOffline

	// 用 context 统一管理所有后台 goroutine 的生命周期：
	// 收到退出信号后，ctx.Done() 会关闭，gRPC 服务、MQ 消费者都会随之退出。
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ---------- 6. gRPC 服务端 ----------
	if cfg.GRPCEnabled {
		go func() {
			if err := rpc.Serve(ctx, ":"+cfg.GRPCPort, db); err != nil {
				log.Printf("[grpc] 服务退出: %v", err)
			}
		}()
	}

	// ---------- 7. RabbitMQ 消费者 ----------
	consumer := mq.NewConsumer(cfg.RabbitMQURL, cfg.RabbitMQEnabled, svc.HandleChatNotify)
	consumer.Start(ctx)

	// ---------- 8. HTTP 服务 ----------
	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           api.NewRouter(cfg, svc),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// 优雅关闭：给正在处理的请求 5 秒时间完成，避免直接断开连接
	go func() {
		<-ctx.Done()
		log.Printf("收到退出信号，正在关闭 HTTP 服务...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("resume-go 启动成功，HTTP :%s，gRPC :%s", cfg.ServerPort, cfg.GRPCPort)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("服务启动失败: %v", err)
	}
}
