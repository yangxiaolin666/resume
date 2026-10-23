package service

import (
	"resume-backend/internal/auth"
	"resume-backend/internal/chat"
	"resume-backend/internal/data"
	"resume-backend/internal/store"
)

type Service struct {
	DB        *data.DB
	Auth      *auth.Manager
	Favorites *store.FavoriteStore
	ChatHub   *chat.Hub
	UploadDir string
}

func New(db *data.DB, manager *auth.Manager, favorites *store.FavoriteStore,
	chatHub *chat.Hub, uploadDir string) *Service {
	return &Service{
		DB:        db,
		Auth:      manager,
		Favorites: favorites,
		ChatHub:   chatHub,
		UploadDir: uploadDir,
	}
}
