package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"resume-backend/internal/api"
	"resume-backend/internal/auth"
	"resume-backend/internal/chat"
	"resume-backend/internal/config"
	"resume-backend/internal/data"
	"resume-backend/internal/service"
	"resume-backend/internal/store"
)

func main() {
	cfg := config.Load()

	db, err := data.Open(cfg.DSN())
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}

	favorites := store.NewFavoriteStore()
	chatHub := chat.NewHub()
	svc := service.New(db, auth.NewManager(cfg.JWTSecret), favorites, chatHub, cfg.UploadDir)

	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           api.NewRouter(cfg, svc),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("resume-go 启动成功，监听 :%s", cfg.ServerPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务启动失败: %v", err)
	}
}
