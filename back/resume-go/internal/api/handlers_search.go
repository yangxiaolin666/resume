package api

// ==================================================================
// 本文件职责：企业端「搜索学生」的两个接口。
// ==================================================================
//   POST /com/search        学生列表搜索（结果只有列表字段，适合卡片 / 表格）
//   POST /com/detailSearch  学生详情搜索（每条结果带上语言、证书等完整信息）
//
// 两者都用 POST + JSON 请求体而不是 GET 查询串，是因为筛选条件多且包含数组
// （languages、cities、specialities...），拼到 URL 里既冗长又容易超长。
//
// 请求体统一解析成 model.StudentQuery。这个 DTO 做了大量兼容处理：
// 驼峰 / 下划线字段都认、数字既接受 11 也接受 "11"、列表既接受数组也接受单值。
// 因此 handler 里看起来只有一行 decodeJSON，真正的“脏活”都在 model 层。
//
// 响应：data 为 model.PageInfo；成功 {code:200}，解析失败由 a.fail 统一转成 400/500。
// ==================================================================

import (
	"github.com/gin-gonic/gin"

	"resume-backend/internal/model"
)

// searchStudents 按条件分页搜索学生（列表级信息）。
//
//	POST /com/search
//
// 请求体：JSON -> model.StudentQuery。
// 调用：Service.SearchStudents(query)（service 内部会先 query.Normalize()）。
// 响应：data 为 model.PageInfo，list 为 []model.Student。
func (a *API) searchStudents(c *gin.Context) {
	var query model.StudentQuery
	if err := decodeJSON(c, &query); err != nil {
		a.fail(c, err)
		return
	}
	page, err := a.Service.SearchStudents(query)
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, page)
}

// searchStudentDetails 按条件分页搜索学生，并返回完整详情。
//
//	POST /com/detailSearch
//
// 请求体：同 /com/search。
// 调用：Service.SearchStudentDetails(query)
// 响应：data 为 model.PageInfo，list 为 []model.StudentDetailDTO
// （每条记录聚合了语言、证书、实习、特长等子表数据，字段更多、查询也更重）。
func (a *API) searchStudentDetails(c *gin.Context) {
	var query model.StudentQuery
	if err := decodeJSON(c, &query); err != nil {
		a.fail(c, err)
		return
	}
	page, err := a.Service.SearchStudentDetails(query)
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, page)
}
