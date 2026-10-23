package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"resume-backend/internal/model"
)

// ==================================================================
// 本文件职责：管理端（/admin 分组）的 HTTP 接口。
// ==================================================================
// 这些路由在 api.go 里统一挂了两层中间件：
//   a.authMiddleware()  —— 必须登录，解析出当前登录用户（身份来自 token，前端无法伪造）；
//   a.adminMiddleware() —— 必须是 admin 角色，否则直接拒绝。
// 因此 handler 内部不再重复校验权限。
//
// 路由一览：
//   PUT    /admin/updateStuPwd/:stu_id?pwd=   重置学生密码
//   POST   /admin/addStu                       Excel 批量导入学生
//   POST   /admin/addCom                       Excel 批量导入企业
//   DELETE /admin/delStu                       批量删除学生
//   DELETE /admin/delCom                       批量删除企业
//   PUT    /admin/updateStu                    批量修改学生
//   PUT    /admin/updateCom                    批量修改企业
//   GET    /admin/searchStu                    学生分页搜索
//   GET    /admin/searchCom                    企业分页搜索
//   POST   /admin/searchStuDetail              学生详情分页搜索
//   GET    /admin/getAllLanguage               语言字典查询
//   POST   /admin/addLanguage                  新增语言（表单）
//   POST   /admin/delLanguage                  删除语言（表单）
//   GET    /admin/getAllCertificate            证书字典查询
//   POST   /admin/addCertificate               新增证书（表单）
//   POST   /admin/delCertificate               删除证书（表单）
//   POST   /admin/addSkillAll                  批量添加技能（原始文本）
//   GET    /admin/getAllSkill                  技能字典查询
//   DELETE /admin/deleteSkill?skills=          批量删除技能
//
// 【本文件的返回特点】
//   - 单个操作用统一响应 {code,message,data}；
//   - 批量操作返回 model.BatchOperationResult，并在有失败时仍用 HTTP 200 +
//     特定业务 code（如 40402）表达“部分成功”，让前端能展示每一行的失败原因。
// ==================================================================

// updateStuPwd 管理员重置某个学生的登录密码。
//
//	PUT /admin/updateStuPwd/:stu_id?pwd=新密码
//
// 请求参数：路径 :stu_id（学号，Atoi 失败返回 400）；查询串 pwd（不能为空）。
// 注意密码走 URL 查询串是历史接口约定，实际生产更推荐放在请求体里。
// 调用：Service.UpdateStudentPassword(stuID, pwd)
//
//	service 会用 bcrypt 生成哈希，只有当恰好更新 1 行时才算成功。
//
// 响应：成功 data="修改密码成功"；未命中学生时 code=50001、message="密码错误"（沿用旧接口文案）。
func (a *API) updateStuPwd(c *gin.Context) {
	stuID, err := strconv.Atoi(c.Param("stu_id"))
	if err != nil {
		respond(c, model.Error(400, "学号格式不正确"))
		return
	}
	pwd := c.Query("pwd")
	if pwd == "" {
		respond(c, model.Error(400, "密码不能为空"))
		return
	}
	success, svcErr := a.Service.UpdateStudentPassword(stuID, pwd)
	if svcErr != nil {
		a.fail(c, svcErr)
		return
	}
	if success {
		ok(c, "修改密码成功")
		return
	}
	respond(c, model.Error(50001, "密码错误"))
}

// importStudents Excel 批量导入学生。
//
//	POST /admin/addStu
//	Content-Type: multipart/form-data，文件字段名 stuFile
//
// 只是把「文件字段名」和「导入函数」传给公共方法 importExcel，
// 真正的解析、校验、入库都在 service.ImportStudents 里。
func (a *API) importStudents(c *gin.Context) {
	a.importExcel(c, "stuFile", a.Service.ImportStudents)
}

// importCompanies Excel 批量导入企业。
//
//	POST /admin/addCom
//	Content-Type: multipart/form-data，文件字段名 comFile
func (a *API) importCompanies(c *gin.Context) {
	a.importExcel(c, "comFile", a.Service.ImportCompanies)
}

// importExcel 是两种导入接口的公共流程：取文件 -> 校验 -> 交给 importer -> 组装“部分成功”响应。
//
// 参数 importer 是一个函数值（签名 func(io.Reader) (*model.BatchOperationResult, error)），
// 这样学生导入和企业导入可以复用同一套 HTTP 处理，只需换一个业务函数，
// 属于最朴素的“策略注入”，不必为两个接口写两遍代码。
//
// 【文件校验顺序（每一步失败都返回不同的业务码）】
//  1. ParseMultipartForm(12 << 20)：最多用 12MB 内存缓存文件，超出部分自动落临时文件，
//     避免管理员上传超大 Excel 时把服务内存打满；临时文件由 net/http 在请求结束后清理。
//  2. c.FormFile(field) 取不到文件 -> 40009“请选择要导入的 Excel 文件”；
//  3. header.Size == 0 -> 40009“Excel文件内容为空”；
//  4. 扩展名不是 .xlsx -> 41501“文件格式不支持，仅支持.xlsx格式”
//     （这里是区分大小写的后缀判断，所以 .XLSX 会被拒绝）；
//  5. header.Open() 打开失败 -> 50001（磁盘 / 临时文件问题）。
//
// 【失败计数与部分成功】
// service 会逐行解析：能读的行入库并 SuccessCount++；
// 单行数据有问题（学号为空、重复等）则 FailureCount++ 并往 Errors 里追加
// {identifier: 行号, reason: 原因}，不会因为一行错误就整批失败。
// 因此这里判断 FailureCount > 0 时，返回 code=200、message="部分数据导入失败"、data=result，
// 前端可以据此展示“成功 N 条、失败 M 条”以及每一行的失败原因。
// 只有 service 返回 error（如整批事务失败、文件根本不是 xlsx）才返回 50001。
func (a *API) importExcel(c *gin.Context, field string,
	importer func(io.Reader) (*model.BatchOperationResult, error)) {
	// 限制 multipart 解析占用的内存，超出部分落到临时文件。
	if err := c.Request.ParseMultipartForm(12 << 20); err != nil {
		respond(c, model.Error(50001, "数据保存失败，请稍后重试"))
		return
	}
	header, err := c.FormFile(field)
	if err != nil {
		respond(c, model.Error(40009, "请选择要导入的 Excel 文件"))
		return
	}
	if header.Size == 0 {
		respond(c, model.Error(40009, "Excel文件内容为空"))
		return
	}
	if !strings.HasSuffix(header.Filename, ".xlsx") {
		respond(c, model.Error(41501, "文件格式不支持，仅支持.xlsx格式"))
		return
	}
	file, err := header.Open()
	if err != nil {
		respond(c, model.Error(50001, "数据保存失败，请稍后重试"))
		return
	}
	defer file.Close()

	result, svcErr := importer(file)
	if svcErr != nil {
		respond(c, model.Error(50001, "数据保存失败，请稍后重试"))
		return
	}
	if result.FailureCount > 0 {
		c.JSON(http.StatusOK, model.Result{Code: 200, Message: "部分数据导入失败", Data: result})
		return
	}
	ok(c, result)
}

// deleteStudents 批量删除学生。
//
//	DELETE /admin/delStu
//
// 请求体：{"stuIds":[1,2,3]}（不是查询串）。
// 空列表直接返回 400，避免误触发全表删除。
// 调用：Service.DeleteStudents(stuIDs)
// 响应：
//   - 全部存在并删除：code=200，data=BatchOperationResult；
//   - 有 ID 不存在：code=40402，message="部分学生不存在"，data 中仍带成功/失败明细。
func (a *API) deleteStudents(c *gin.Context) {
	var dto model.DeleteStudentsDTO
	if err := decodeJSON(c, &dto); err != nil {
		a.fail(c, err)
		return
	}
	if len(dto.StuIDs) == 0 {
		respond(c, model.Error(400, "学生ID列表不能为空"))
		return
	}
	result, err := a.Service.DeleteStudents(dto.StuIDs)
	if err != nil {
		respond(c, model.Error(500, "删除失败，请稍后重试"))
		return
	}
	if result.FailureCount > 0 {
		c.JSON(http.StatusOK, model.Result{Code: 40402, Message: "部分学生不存在", Data: result})
		return
	}
	ok(c, result)
}

// deleteCompanies 批量删除企业。
//
//	DELETE /admin/delCom
//
// 请求体：{"comIds":[1,2,3]}。
// 调用：Service.DeleteCompanies(comIDs)
// 响应：全部成功 code=200；有 ID 不存在时 code=40403、message="部分企业不存在"。
func (a *API) deleteCompanies(c *gin.Context) {
	var dto model.DeleteCompaniesDTO
	if err := decodeJSON(c, &dto); err != nil {
		a.fail(c, err)
		return
	}
	if len(dto.ComIDs) == 0 {
		respond(c, model.Error(400, "企业ID列表不能为空"))
		return
	}
	result, err := a.Service.DeleteCompanies(dto.ComIDs)
	if err != nil {
		respond(c, model.Error(500, "删除失败，请稍后重试"))
		return
	}
	if result.FailureCount > 0 {
		c.JSON(http.StatusOK, model.Result{Code: 40403, Message: "部分企业不存在", Data: result})
		return
	}
	ok(c, result)
}

// updateStudents 批量修改学生信息。
//
//	PUT /admin/updateStu
//
// 请求体：{"students":[{...}, ...]}，每项为 model.StudentInput
// （字段支持 stuId/stu_id、graduationYear/graduation_year）。
// 空数组返回 400；service 内部会逐条 Normalize 并统计成功 / 失败。
// 响应：全部成功 code=200；有失败时 code=42205、message="学生数据验证失败"，data 带明细。
func (a *API) updateStudents(c *gin.Context) {
	var dto model.UpdateStudentsDTO
	if err := decodeJSON(c, &dto); err != nil {
		a.fail(c, err)
		return
	}
	if len(dto.Students) == 0 {
		respond(c, model.Error(400, "学生列表不能为空"))
		return
	}
	result, err := a.Service.UpdateStudents(dto.Students)
	if err != nil {
		respond(c, model.Error(500, "更新失败，请稍后重试"))
		return
	}
	if result.FailureCount > 0 {
		c.JSON(http.StatusOK, model.Result{Code: 42205, Message: "学生数据验证失败", Data: result})
		return
	}
	ok(c, result)
}

// updateCompanies 批量修改企业信息。
//
//	PUT /admin/updateCom
//
// 请求体：{"companies":[{...}, ...]}，每项为 model.CompanyInput
// （comName/com_name、comRegion/com_region 都接受）。
// 调用：Service.UpdateCompanies(companies)
// 响应：全部成功 code=200；有失败时 code=42206、message="企业数据验证失败"。
func (a *API) updateCompanies(c *gin.Context) {
	var dto model.UpdateCompaniesDTO
	if err := decodeJSON(c, &dto); err != nil {
		a.fail(c, err)
		return
	}
	if len(dto.Companies) == 0 {
		respond(c, model.Error(400, "企业列表不能为空"))
		return
	}
	result, err := a.Service.UpdateCompanies(dto.Companies)
	if err != nil {
		respond(c, model.Error(500, "更新失败，请稍后重试"))
		return
	}
	if result.FailureCount > 0 {
		c.JSON(http.StatusOK, model.Result{Code: 42206, Message: "企业数据验证失败", Data: result})
		return
	}
	ok(c, result)
}

// adminSearchStudents 管理端按姓名分页搜索学生。
//
//	GET /admin/searchStu?username=张&page=1&pageSize=10
//
// 请求参数：查询串 username（模糊匹配）、page、pageSize（缺省 1 / 10）。
// 注意分页参数名是 page（不是 pageNum），与用户端接口略有差异。
// 调用：Service.AdminSearchStudents(name, page, pageSize)
// 响应：data 为分页结果；任何错误统一返回 500“查询失败，请稍后重试”。
func (a *API) adminSearchStudents(c *gin.Context) {
	name := queryString(c, "username")
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "pageSize", 10)
	result, err := a.Service.AdminSearchStudents(name, page, pageSize)
	if err != nil {
		respond(c, model.Error(500, "查询失败，请稍后重试"))
		return
	}
	ok(c, result)
}

// adminSearchCompanies 管理端按企业名称分页搜索企业。
//
//	GET /admin/searchCom?companyName=科技&page=1&pageSize=10
//
// 请求参数：查询串 companyName、page、pageSize。
// 调用：Service.AdminSearchCompanies(name, page, pageSize)
// 响应：data 为分页结果。
func (a *API) adminSearchCompanies(c *gin.Context) {
	name := queryString(c, "companyName")
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "pageSize", 10)
	result, err := a.Service.AdminSearchCompanies(name, page, pageSize)
	if err != nil {
		respond(c, model.Error(500, "查询失败，请稍后重试"))
		return
	}
	ok(c, result)
}

// adminSearchStudentDetails 管理端按条件分页搜索学生完整详情。
//
//	POST /admin/searchStuDetail
//
// 请求体：JSON -> model.StudentQuery（与学生端搜索共用同一套宽松解析）。
// 调用：Service.SearchStudentDetails(query)
// 响应：data 为 model.PageInfo，list 为 []model.StudentDetailDTO；
// 出错时统一返回 500“查询失败，请稍后重试”（不暴露具体 SQL 错误）。
func (a *API) adminSearchStudentDetails(c *gin.Context) {
	var query model.StudentQuery
	if err := decodeJSON(c, &query); err != nil {
		a.fail(c, err)
		return
	}
	page, err := a.Service.SearchStudentDetails(query)
	if err != nil {
		respond(c, model.Error(500, "查询失败，请稍后重试"))
		return
	}
	ok(c, page)
}

// getAllLanguages 获取语言字典（下拉选项）。
//
//	GET /admin/getAllLanguage
//
// 请求参数：无。
// 调用：Service.GetAllLanguages()
// 响应：data 为 []string。
func (a *API) getAllLanguages(c *gin.Context) {
	languages, err := a.Service.GetAllLanguages()
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, languages)
}

// addLanguage 新增语言选项。
//
//	POST /admin/addLanguage
//	Content-Type: application/x-www-form-urlencoded，字段 language
//
// 请求参数：c.PostForm("language")，去掉首尾空格后不能为空。
// 唯一性校验放在 handler：先查全部语言再遍历比较，存在就返回 400“语言已存在”。
// 调用：Service.AddLanguage(language)
// 响应：成功 data=true。
func (a *API) addLanguage(c *gin.Context) {
	language := strings.TrimSpace(c.PostForm("language"))
	if language == "" {
		respond(c, model.Error(400, "语言不能为空"))
		return
	}
	existing, err := a.Service.GetAllLanguages()
	if err != nil {
		respond(c, model.Error(500, "添加失败，请稍后重试"))
		return
	}
	for _, item := range existing {
		if item == language {
			respond(c, model.Error(400, "语言已存在"))
			return
		}
	}
	if err := a.Service.AddLanguage(language); err != nil {
		respond(c, model.Error(500, "添加失败，请稍后重试"))
		return
	}
	ok(c, true)
}

// delLanguage 删除语言选项。
//
//	POST /admin/delLanguage
//	Content-Type: application/x-www-form-urlencoded，字段 language
//
// 删除前先查一遍字典，不存在时返回 400“语言不存在”，避免“删了但没反馈”。
// 调用：Service.DeleteLanguage(language)
// 响应：成功 data=true。
func (a *API) delLanguage(c *gin.Context) {
	language := strings.TrimSpace(c.PostForm("language"))
	if language == "" {
		respond(c, model.Error(400, "语言不能为空"))
		return
	}
	existing, err := a.Service.GetAllLanguages()
	if err != nil {
		respond(c, model.Error(500, "删除失败，请稍后重试"))
		return
	}
	found := false
	for _, item := range existing {
		if item == language {
			found = true
			break
		}
	}
	if !found {
		respond(c, model.Error(400, "语言不存在"))
		return
	}
	if err := a.Service.DeleteLanguage(language); err != nil {
		respond(c, model.Error(500, "删除失败，请稍后重试"))
		return
	}
	ok(c, true)
}

// getAllCertificates 获取证书字典（下拉选项）。
//
//	GET /admin/getAllCertificate
//
// 请求参数：无。
// 调用：Service.GetAllCertificates()
// 响应：data 为 []string。
func (a *API) getAllCertificates(c *gin.Context) {
	certificates, err := a.Service.GetAllCertificates()
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, certificates)
}

// addCertificate 新增证书选项。
//
//	POST /admin/addCertificate
//	Content-Type: application/x-www-form-urlencoded，字段 certificate
//
// 校验逻辑与语言一致：非空 + 不允许重复。
// 调用：Service.AddCertificate(certificate)
// 响应：成功 data=true。
func (a *API) addCertificate(c *gin.Context) {
	certificate := strings.TrimSpace(c.PostForm("certificate"))
	if certificate == "" {
		respond(c, model.Error(400, "证书不能为空"))
		return
	}
	existing, err := a.Service.GetAllCertificates()
	if err != nil {
		respond(c, model.Error(500, "添加失败，请稍后重试"))
		return
	}
	for _, item := range existing {
		if item == certificate {
			respond(c, model.Error(400, "证书已存在"))
			return
		}
	}
	if err := a.Service.AddCertificate(certificate); err != nil {
		respond(c, model.Error(500, "添加失败，请稍后重试"))
		return
	}
	ok(c, true)
}

// delCertificate 删除证书选项。
//
//	POST /admin/delCertificate
//	Content-Type: application/x-www-form-urlencoded，字段 certificate
//
// 先确认证书存在再删除；不存在返回 400“证书不存在”。
// 调用：Service.DeleteCertificate(certificate)
// 响应：成功 data=true。
func (a *API) delCertificate(c *gin.Context) {
	certificate := strings.TrimSpace(c.PostForm("certificate"))
	if certificate == "" {
		respond(c, model.Error(400, "证书不能为空"))
		return
	}
	existing, err := a.Service.GetAllCertificates()
	if err != nil {
		respond(c, model.Error(500, "删除失败，请稍后重试"))
		return
	}
	found := false
	for _, item := range existing {
		if item == certificate {
			found = true
			break
		}
	}
	if !found {
		respond(c, model.Error(400, "证书不存在"))
		return
	}
	if err := a.Service.DeleteCertificate(certificate); err != nil {
		respond(c, model.Error(500, "删除失败，请稍后重试"))
		return
	}
	ok(c, true)
}

// addSkillAll 批量添加技能选项。
//
//	POST /admin/addSkillAll
//
// 请求体：原始文本（不是 JSON 对象），技能之间用英文逗号分隔；
// 也兼容带引号的 JSON 字符串（"Java,Go"）。
//
// 因为旧前端直接把 textarea 的内容当 body 发过来，所以这里用 GetRawData 读原文：
//  1. 先 TrimSpace；
//  2. 如果内容首尾都是双引号，再按 JSON 字符串解析一次，剥掉引号与转义；
//  3. 交给 Service.AddSkillsAll，由 service 按逗号拆分、过滤空项后逐条入库。
//
// 调用：Service.AddSkillsAll(raw) 返回实际新增条数。
// 响应：data 为字符串“添加成功N个学生技能信息”。
func (a *API) addSkillAll(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		a.fail(c, err)
		return
	}
	raw := strings.TrimSpace(string(body))
	// Java @RequestBody String 接收原始文本；同时兼容 JSON 字符串（带引号）。
	if strings.HasPrefix(raw, "\"") && strings.HasSuffix(raw, "\"") {
		var value string
		if err := json.Unmarshal(body, &value); err == nil {
			raw = value
		}
	}
	count, svcErr := a.Service.AddSkillsAll(raw)
	if svcErr != nil {
		a.fail(c, svcErr)
		return
	}
	ok(c, fmt.Sprintf("添加成功%d个学生技能信息", count))
}

// deleteSkill 批量删除技能选项。
//
//	DELETE /admin/deleteSkill?skills=Java,Go
//
// 请求参数：查询串 skills，英文逗号分隔；为空返回 400“请选择要删除的技能”。
// 调用：Service.DeleteSkills(skills) 返回受影响行数。
// 响应：data 为字符串“删除N个技能选项成功”。
func (a *API) deleteSkill(c *gin.Context) {
	skills := c.Query("skills")
	if strings.TrimSpace(skills) == "" {
		respond(c, model.Error(400, "请选择要删除的技能"))
		return
	}
	count, err := a.Service.DeleteSkills(skills)
	if err != nil {
		a.fail(c, err)
		return
	}
	ok(c, fmt.Sprintf("删除%d个技能选项成功", count))
}
