package service

import (
	"fmt"
	"strings"

	"resume-backend/internal/data"
	"resume-backend/internal/model"
)

func (s *Service) QueryComID(userID int) (*int, error) {
	return s.DB.QueryComIDByUserID(userID)
}

func (s *Service) AddLikes(userID int, items []model.LikeItem) (int, error) {
	pairs := make([]data.LikePair, 0, len(items))
	for _, item := range items {
		if item.StuID == nil || item.ComID == nil {
			continue
		}
		existsStudent, err := s.DB.ExistsStudent(*item.StuID)
		if err != nil {
			return 0, err
		}
		if !existsStudent {
			return 0, model.NewAppError(500, fmt.Sprintf("学生ID不存在: %d", *item.StuID))
		}
		existsCompany, err := s.DB.ExistsCompany(*item.ComID)
		if err != nil {
			return 0, err
		}
		if !existsCompany {
			return 0, model.NewAppError(500, fmt.Sprintf("企业ID不存在: %d", *item.ComID))
		}
		pairs = append(pairs, data.LikePair{ComID: *item.ComID, StuID: *item.StuID})
	}
	if len(pairs) == 0 {
		return 0, nil
	}

	existing, err := s.DB.ExistingLikes(pairs)
	if err != nil {
		return 0, err
	}
	toInsert := make([]data.LikePair, 0, len(pairs))
	for _, pair := range pairs {
		duplicated := false
		for _, had := range existing {
			if had.ComID == pair.ComID && had.StuID == pair.StuID {
				duplicated = true
				break
			}
		}
		if !duplicated {
			toInsert = append(toInsert, pair)
		}
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if err := s.DB.InsertLikesTx(tx, toInsert); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	for _, pair := range pairs {
		s.Favorites.Add(userID, pair.StuID)
	}
	return len(existing), nil
}

func (s *Service) DeleteLikes(userID int, items []model.LikeItem) (int, error) {
	pairs := make([]data.LikePair, 0, len(items))
	for _, item := range items {
		if item.StuID == nil || item.ComID == nil {
			continue
		}
		pairs = append(pairs, data.LikePair{ComID: *item.ComID, StuID: *item.StuID})
	}
	existing, err := s.DB.ExistingLikes(pairs)
	if err != nil {
		return 0, err
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if err := s.DB.DeleteLikesTx(tx, existing); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	for _, pair := range pairs {
		s.Favorites.Remove(userID, pair.StuID)
	}
	return len(pairs) - len(existing), nil
}

func (s *Service) FavoriteStudents(dto model.LikeQueryDTO) (*model.PageInfo, error) {
	dto.Normalize()
	if dto.ComID == nil {
		return nil, model.NewAppError(500, "企业ID不能为空")
	}
	list, total, err := s.DB.FavoriteStudents(*dto.ComID, dto.PageNum, dto.PageSize)
	if err != nil {
		return nil, err
	}
	page := model.NewPageInfo(list, total, dto.PageNum, dto.PageSize, len(list))
	return &page, nil
}

func (s *Service) IsLiked(comID, stuID int) (bool, error) {
	existing, err := s.DB.ExistingLikes([]data.LikePair{{ComID: comID, StuID: stuID}})
	if err != nil {
		return false, err
	}
	return len(existing) > 0, nil
}

func (s *Service) BatchIsFavorite(userID int, raw string) map[int]bool {
	ids := []int{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		var id int
		if _, err := fmt.Sscanf(part, "%d", &id); err == nil {
			ids = append(ids, id)
		}
	}
	return s.Favorites.BatchIsFavorite(userID, ids)
}
