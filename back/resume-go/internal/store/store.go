package store

import "sync"

// FavoriteStore 用进程内集合替代原 Java 版的 Redis Set（user:favorite:{userId}）。
// 接口行为一致，重启后清空；本地运行无需额外安装 Redis。
type FavoriteStore struct {
	mu   sync.RWMutex
	data map[int]map[int]struct{}
}

func NewFavoriteStore() *FavoriteStore {
	return &FavoriteStore{data: make(map[int]map[int]struct{})}
}

func (s *FavoriteStore) Add(userID, contentID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	set, ok := s.data[userID]
	if !ok {
		set = make(map[int]struct{})
		s.data[userID] = set
	}
	set[contentID] = struct{}{}
}

func (s *FavoriteStore) Remove(userID, contentID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if set, ok := s.data[userID]; ok {
		delete(set, contentID)
	}
}

func (s *FavoriteStore) BatchIsFavorite(userID int, contentIDs []int) map[int]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[int]bool, len(contentIDs))
	set := s.data[userID]
	for _, id := range contentIDs {
		_, ok := set[id]
		result[id] = ok
	}
	return result
}
