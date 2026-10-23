package service

import (
	"context"
	"fmt"
	"strings"

	"resume-backend/internal/data"
	"resume-backend/internal/model"
)

// ==================================================================
// 企业收藏学生业务层
// ==================================================================
// 对应简历能力：「Redis 维护收藏集合（LikeIfByRedis）+ MySQL 持久化」。
//
// 【数据存放在哪】
//   - MySQL 的 com_like_stu 是唯一可靠数据源，收藏关系必须落库；
//   - Redis 的 Set 只是「查询加速」：列表页要一次性判断几十个学生是否被收藏，
//     走 Redis 的 SMISMEMBER 一次网络往返即可，比逐条查库快得多。
//   因此写操作永远是「先落库、成功后再更新 Redis」，Redis 写失败只影响加速，不影响正确性。
//
// 【与其它层的关系】
//   - api：传入批量 LikeItem（stuId + comId）和当前企业用户；
//   - data：存在性查询、去重查询、事务内批量插入/删除；
//   - cache：AddFavorite / RemoveFavorite / BatchIsFavorite。
// ==================================================================

// QueryComID 根据登录用户 ID 找到对应的企业 ID。
// 企业相关接口只信任 JWT 里的 userID，comId 必须由后端反查，
// 不能直接采用前端传入的企业 ID，否则可以冒充其他企业操作数据。
func (s *Service) QueryComID(userID int) (*int, error) {
	return s.DB.QueryComIDByUserID(userID)
}

// AddLikes 批量收藏学生。
//
// 【步骤与原因】
//  1. 先跳过 stuId/comId 为空的不完整项（不报错，直接忽略），
//     再确认学生、企业真实存在；不存在时直接返回错误，并在文案里带上具体 ID，
//     方便调用方定位是哪条数据有问题。
//  2. 用「一次查询」把请求里已经存在的收藏全部捞出来，再在内存里做差集。
//     为什么不直接逐条 INSERT？
//     - 若表上有 (stu_id, com_id) 唯一索引，重复插入会让整批事务报错回滚，
//     一条重复导致全部失败，体验很差；
//     - 若没有唯一索引，会插入重复数据，后续统计和取消收藏都会出错。
//     所以先查后插，既避免报错，也避免产生脏数据。
//  3. 去重后统一在「一个事务」里插入：批量写要么全成功、要么全失败，
//     不会出现“收藏了一半”的中间状态。
//  4. 事务提交成功后再更新 Redis 收藏集合；顺序不能反，
//     否则数据库回滚了缓存却已经标记为收藏，出现缓存与库不一致。
//
// 返回值：本次请求中「数据库里原本就存在」的条数，接口层可用它提示重复收藏。
// 失败时：参数/存在性问题返回 4xx/5xx AppError；SQL 错误原样返回。
func (s *Service) AddLikes(ctx context.Context, userID int, items []model.LikeItem) (int, error) {
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
	// 在内存里做差集：把「库里已有」的 pair 过滤掉，只留下真正需要插入的。
	// existing 通常很小（最多等于本次请求条数），所以双重循环足够，不需要额外建 map。
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
	// defer Rollback 是固定写法：Commit 成功后 Rollback 是空操作，
	// 而中途任何 return 都能保证连接被释放、未提交的写入被撤销。
	defer tx.Rollback()
	if err := s.DB.InsertLikesTx(tx, toInsert); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	for _, pair := range pairs {
		// Redis 只是加速层，写失败不影响主流程（cache 内部只记日志），
		// 所以这里不需要收集错误；最坏情况是列表页多查一次库。
		s.Cache.AddFavorite(ctx, userID, pair.StuID)
	}
	return len(existing), nil
}

// DeleteLikes 批量取消收藏。
//
// 与新增对称：先查出「真正存在」的记录，再在一个事务里按 (stu_id, com_id) 批量删除。
// 只删除存在的记录，避免对不存在的行做无用写入；删除后再同步 Redis。
// Redis 里删多了也没关系（SREM 对不存在的成员是幂等的）。
//
// 返回值：本次传入但数据库里原本就不存在的条数（即没删到的数量）。
func (s *Service) DeleteLikes(ctx context.Context, userID int, items []model.LikeItem) (int, error) {
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
		s.Cache.RemoveFavorite(ctx, userID, pair.StuID)
	}
	return len(pairs) - len(existing), nil
}

// FavoriteStudents 分页查询某企业收藏的学生列表。
// Normalize 会补默认分页，并兼容驼峰/下划线两种入参写法；
// comId 为空说明调用方没传企业，返回业务错误而不是查全表。
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

// IsLiked 判断某个企业是否收藏了某个学生（单条场景，直接查库）。
// 复用批量查询 ExistingLikes 传一个元素，避免为单条判断再加一个 data 方法。
func (s *Service) IsLiked(comID, stuID int) (bool, error) {
	existing, err := s.DB.ExistingLikes([]data.LikePair{{ComID: comID, StuID: stuID}})
	if err != nil {
		return false, err
	}
	return len(existing) > 0, nil
}

// BatchIsFavorite 批量判断「这些学生是否被我收藏」，供学生列表页渲染收藏按钮。
//
// 入参 raw 是逗号分隔的 id 字符串（前端 query 参数），这里做容错解析：
// 空串、多余空格、非数字都会被跳过，而不是让整个请求失败——
// 列表页的辅助信息不值得因为一个脏 id 就 400。
// 合法的 id 交给 Redis 一次 SMISMEMBER 批量判断；未命中/Redis 不可用时结果为空 map，
// 前端会显示为「未收藏」，属于可接受的降级。
func (s *Service) BatchIsFavorite(ctx context.Context, userID int, raw string) map[int]bool {
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
	return s.Cache.BatchIsFavorite(ctx, userID, ids)
}
