package api

import (
	"fmt"
	"net/http"
	"strconv"

	"resume-backend/internal/model"
)

func (a *API) queryComID(w http.ResponseWriter, r *http.Request) {
	comID, err := a.Service.QueryComID(userID(r))
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, comID)
}

func (a *API) likeAdd(w http.ResponseWriter, r *http.Request) {
	var items []model.LikeItem
	if err := decodeJSON(r, &items); err != nil {
		a.fail(w, err)
		return
	}
	failCount, err := a.Service.AddLikes(userID(r), items)
	if err != nil {
		a.fail(w, err)
		return
	}
	if failCount == 0 {
		ok(w, nil)
		return
	}
	respond(w, model.Error(500, fmt.Sprintf("添加%d个学生失败，原因：学生已存在", failCount)))
}

func (a *API) likeQuery(w http.ResponseWriter, r *http.Request) {
	var dto model.LikeQueryDTO
	if err := decodeJSON(r, &dto); err != nil {
		a.fail(w, err)
		return
	}
	page, err := a.Service.FavoriteStudents(dto)
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, page)
}

func (a *API) likeDelete(w http.ResponseWriter, r *http.Request) {
	var items []model.LikeItem
	if err := decodeJSON(r, &items); err != nil {
		a.fail(w, err)
		return
	}
	failCount, err := a.Service.DeleteLikes(userID(r), items)
	if err != nil {
		a.fail(w, err)
		return
	}
	if failCount == 0 {
		ok(w, nil)
		return
	}
	respond(w, model.Error(500, fmt.Sprintf("删除%d个学生失败，原因：学生不存在", failCount)))
}

func (a *API) likeIf(w http.ResponseWriter, r *http.Request) {
	comID, err1 := strconv.Atoi(queryString(r, "com_id"))
	stuID, err2 := strconv.Atoi(queryString(r, "stu_id"))
	if err1 != nil || err2 != nil {
		respond(w, model.Error(500, "服务器内部错误: Required request parameter is not present"))
		return
	}
	liked, err := a.Service.IsLiked(comID, stuID)
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, liked)
}

func (a *API) likeIfByRedis(w http.ResponseWriter, r *http.Request) {
	raw := queryString(r, "stuIdList")
	result := a.Service.BatchIsFavorite(userID(r), raw)
	ok(w, result)
}
