package api

// ==================================================================
// 本文件职责：企业「收藏学生」相关的 HTTP 接口。
// ==================================================================
// 业务背景：企业浏览学生搜索结果时，可以把感兴趣的学生加入收藏，
// 之后在收藏列表里集中查看、批量取消。
//
// 路由（均需登录）：
//   GET    /com/QueryComIdByUserId       用当前登录用户查出企业 ID
//   POST   /com/LikeAdd                  批量收藏
//   POST   /com/LikeQuery                分页查询收藏的学生
//   DELETE /com/LikeDelete               批量取消收藏
//   GET    /com/LikeIf?com_id=&stu_id=   判断某个学生是否已被收藏
//   GET    /com/LikeIfByRedis?stuIdList= 批量判断（走 Redis 缓存）
//
// 响应依旧统一为 {code,message,data}；收藏类接口的 data 常常是 nil 或布尔值。
// ==================================================================

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"resume-backend/internal/model"
)

// queryComID 用当前登录用户查出其所属的企业 ID。
//
//	GET /com/QueryComIdByUserId
//
// 请求参数：无（userID 来自 token）。
// 调用：Service.QueryComID(userID)
// 响应：data 为企业 ID（整数）。前端拿到后作为后续收藏接口的 comId 使用。
func (a *API) queryComID(c *gin.Context) {
	comID, err := a.Service.QueryComID(userID(c))
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, comID)
}

// likeAdd 批量收藏学生。
//
//	POST /com/LikeAdd
//
// 请求体：JSON 数组 [{"comId":1,"stuId":2}, ...]，字段也接受 com_id / stu_id。
// 两列 ID 都必须传：service 会跳过缺任一 ID 的项，并校验学生 / 企业是否真实存在。
// 调用：Service.AddLikes(ctx, userID, items) 返回失败条数。
// 响应：
//   - 全部成功：code=200，data=null；
//   - 有失败：code=500，message 形如“添加N个学生失败，原因：学生已存在”。
func (a *API) likeAdd(c *gin.Context) {
	var items []model.LikeItem
	if err := decodeJSON(c, &items); err != nil {
		a.fail(c, err)
		return
	}
	failCount, err := a.Service.AddLikes(c.Request.Context(), userID(c), items)
	if err != nil {
		a.fail(c, err)
		return
	}
	if failCount == 0 {
		ok(c, nil)
		return
	}
	respond(c, model.Error(500, fmt.Sprintf("添加%d个学生失败，原因：学生已存在", failCount)))
}

// likeQuery 分页查询企业收藏的学生。
//
//	POST /com/LikeQuery
//
// 请求体：JSON -> model.LikeQueryDTO（comId/com_id、comName、pageNum、pageSize）。
// service 内部会先 Normalize 合并两种字段名并补齐分页默认值。
// 调用：Service.FavoriteStudents(dto)
// 响应：data 为 model.PageInfo，list 中是 model.StuFavorite（已把语言、特长等拼成字符串）。
func (a *API) likeQuery(c *gin.Context) {
	var dto model.LikeQueryDTO
	if err := decodeJSON(c, &dto); err != nil {
		a.fail(c, err)
		return
	}
	page, err := a.Service.FavoriteStudents(dto)
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, page)
}

// likeDelete 批量取消收藏。
//
//	DELETE /com/LikeDelete
//
// 请求体：与收藏相同，JSON 数组 [{"comId":1,"stuId":2}, ...]。
// 调用：Service.DeleteLikes(ctx, userID, items) 返回失败条数。
// 响应：全部成功 data=null；有失败时 code=500，message 为“删除N个学生失败，原因：学生不存在”。
func (a *API) likeDelete(c *gin.Context) {
	var items []model.LikeItem
	if err := decodeJSON(c, &items); err != nil {
		a.fail(c, err)
		return
	}
	failCount, err := a.Service.DeleteLikes(c.Request.Context(), userID(c), items)
	if err != nil {
		a.fail(c, err)
		return
	}
	if failCount == 0 {
		ok(c, nil)
		return
	}
	respond(c, model.Error(500, fmt.Sprintf("删除%d个学生失败，原因：学生不存在", failCount)))
}

// likeIf 判断“该企业是否已收藏该学生”（学生详情页的收藏按钮状态）。
//
//	GET /com/LikeIf?com_id=1&stu_id=2
//
// 请求参数：查询串 com_id、stu_id（这里固定用下划线命名，Atoi 失败返回 400）。
// 调用：Service.IsLiked(comID, stuID)
// 响应：data 为 bool（true=已收藏）。
func (a *API) likeIf(c *gin.Context) {
	comID, err1 := strconv.Atoi(queryString(c, "com_id"))
	stuID, err2 := strconv.Atoi(queryString(c, "stu_id"))
	if err1 != nil || err2 != nil {
		respond(c, model.Error(400, "请求参数不完整，请刷新页面后重试"))
		return
	}
	liked, err := a.Service.IsLiked(comID, stuID)
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, liked)
}

// likeIfByRedis 批量判断“哪些学生已被收藏”，用于搜索结果列表一次性点亮收藏图标。
//
//	GET /com/LikeIfByRedis?stuIdList=1,2,3
//
// 请求参数：查询串 stuIdList，逗号分隔的学生 ID（service 负责拆分并容错）。
// 调用：Service.BatchIsFavorite(ctx, userID, raw) —— 实际读 Redis 缓存，避免逐个查库。
// 响应：data 为 map[int]bool，键是学生 ID，值表示是否已收藏；
// 注意 JSON 序列化后对象的 key 会变成字符串，前端按字符串下标读取即可。
func (a *API) likeIfByRedis(c *gin.Context) {
	raw := queryString(c, "stuIdList")
	result := a.Service.BatchIsFavorite(c.Request.Context(), userID(c), raw)
	ok(c, result)
}
