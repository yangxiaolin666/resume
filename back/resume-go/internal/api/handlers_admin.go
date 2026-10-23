package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"resume-backend/internal/model"
)

func (a *API) updateStuPwd(w http.ResponseWriter, r *http.Request) {
	stuID, err := strconv.Atoi(r.PathValue("stu_id"))
	if err != nil {
		respond(w, model.Error(500, "服务器内部错误: 学号格式错误"))
		return
	}
	pwd := r.URL.Query().Get("pwd")
	if pwd == "" {
		respond(w, model.Error(400, "密码不能为空"))
		return
	}
	success, svcErr := a.Service.UpdateStudentPassword(stuID, pwd)
	if svcErr != nil {
		a.fail(w, svcErr)
		return
	}
	if success {
		ok(w, "修改密码成功")
		return
	}
	respond(w, model.Error(50001, "密码错误"))
}

func (a *API) importStudents(w http.ResponseWriter, r *http.Request) {
	a.importExcel(w, r, "stuFile", a.Service.ImportStudents)
}

func (a *API) importCompanies(w http.ResponseWriter, r *http.Request) {
	a.importExcel(w, r, "comFile", a.Service.ImportCompanies)
}

func (a *API) importExcel(w http.ResponseWriter, r *http.Request, field string,
	importer func(io.Reader) (*model.BatchOperationResult, error)) {
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		respond(w, model.Error(50001, "数据库写入失败: "+err.Error()))
		return
	}
	file, header, err := r.FormFile(field)
	if err != nil {
		respond(w, model.Error(50001, "数据库写入失败: Required part '"+field+"' is not present"))
		return
	}
	defer file.Close()
	if header.Size == 0 {
		respond(w, model.Error(40009, "Excel文件内容为空"))
		return
	}
	if !strings.HasSuffix(header.Filename, ".xlsx") {
		respond(w, model.Error(41501, "文件格式不支持，仅支持.xlsx格式"))
		return
	}
	result, svcErr := importer(file)
	if svcErr != nil {
		respond(w, model.Error(50001, "数据库写入失败: "+svcErr.Error()))
		return
	}
	if result.FailureCount > 0 {
		writeJSON(w, http.StatusOK, model.Result{Code: 200, Message: "部分数据导入失败", Data: result})
		return
	}
	ok(w, result)
}

func (a *API) deleteStudents(w http.ResponseWriter, r *http.Request) {
	var dto model.DeleteStudentsDTO
	if err := decodeJSON(r, &dto); err != nil {
		a.fail(w, err)
		return
	}
	if len(dto.StuIDs) == 0 {
		respond(w, model.Error(400, "学生ID列表不能为空"))
		return
	}
	result, err := a.Service.DeleteStudents(dto.StuIDs)
	if err != nil {
		respond(w, model.Error(500, "删除失败: "+err.Error()))
		return
	}
	if result.FailureCount > 0 {
		writeJSON(w, http.StatusOK, model.Result{Code: 40402, Message: "部分学生不存在", Data: result})
		return
	}
	ok(w, result)
}

func (a *API) deleteCompanies(w http.ResponseWriter, r *http.Request) {
	var dto model.DeleteCompaniesDTO
	if err := decodeJSON(r, &dto); err != nil {
		a.fail(w, err)
		return
	}
	if len(dto.ComIDs) == 0 {
		respond(w, model.Error(400, "企业ID列表不能为空"))
		return
	}
	result, err := a.Service.DeleteCompanies(dto.ComIDs)
	if err != nil {
		respond(w, model.Error(500, "删除失败: "+err.Error()))
		return
	}
	if result.FailureCount > 0 {
		writeJSON(w, http.StatusOK, model.Result{Code: 40403, Message: "部分企业不存在", Data: result})
		return
	}
	ok(w, result)
}

func (a *API) updateStudents(w http.ResponseWriter, r *http.Request) {
	var dto model.UpdateStudentsDTO
	if err := decodeJSON(r, &dto); err != nil {
		a.fail(w, err)
		return
	}
	if len(dto.Students) == 0 {
		respond(w, model.Error(400, "学生列表不能为空"))
		return
	}
	result, err := a.Service.UpdateStudents(dto.Students)
	if err != nil {
		respond(w, model.Error(500, "更新失败: "+err.Error()))
		return
	}
	if result.FailureCount > 0 {
		writeJSON(w, http.StatusOK, model.Result{Code: 42205, Message: "学生数据验证失败", Data: result})
		return
	}
	ok(w, result)
}

func (a *API) updateCompanies(w http.ResponseWriter, r *http.Request) {
	var dto model.UpdateCompaniesDTO
	if err := decodeJSON(r, &dto); err != nil {
		a.fail(w, err)
		return
	}
	if len(dto.Companies) == 0 {
		respond(w, model.Error(400, "企业列表不能为空"))
		return
	}
	result, err := a.Service.UpdateCompanies(dto.Companies)
	if err != nil {
		respond(w, model.Error(500, "更新失败: "+err.Error()))
		return
	}
	if result.FailureCount > 0 {
		writeJSON(w, http.StatusOK, model.Result{Code: 42206, Message: "企业数据验证失败", Data: result})
		return
	}
	ok(w, result)
}

func (a *API) adminSearchStudents(w http.ResponseWriter, r *http.Request) {
	name := queryString(r, "username")
	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "pageSize", 10)
	result, err := a.Service.AdminSearchStudents(name, page, pageSize)
	if err != nil {
		respond(w, model.Error(500, "查询失败: "+err.Error()))
		return
	}
	ok(w, result)
}

func (a *API) adminSearchCompanies(w http.ResponseWriter, r *http.Request) {
	name := queryString(r, "companyName")
	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "pageSize", 10)
	result, err := a.Service.AdminSearchCompanies(name, page, pageSize)
	if err != nil {
		respond(w, model.Error(500, "查询失败: "+err.Error()))
		return
	}
	ok(w, result)
}

func (a *API) adminSearchStudentDetails(w http.ResponseWriter, r *http.Request) {
	var query model.StudentQuery
	if err := decodeJSON(r, &query); err != nil {
		a.fail(w, err)
		return
	}
	page, err := a.Service.SearchStudentDetails(query)
	if err != nil {
		respond(w, model.Error(500, "查询失败: "+err.Error()))
		return
	}
	ok(w, page)
}

func (a *API) getAllLanguages(w http.ResponseWriter, r *http.Request) {
	languages, err := a.Service.GetAllLanguages()
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, languages)
}

func (a *API) addLanguage(w http.ResponseWriter, r *http.Request) {
	language := strings.TrimSpace(r.FormValue("language"))
	if language == "" {
		respond(w, model.Error(400, "语言不能为空"))
		return
	}
	existing, err := a.Service.GetAllLanguages()
	if err != nil {
		respond(w, model.Error(500, "添加失败: "+err.Error()))
		return
	}
	for _, item := range existing {
		if item == language {
			respond(w, model.Error(400, "语言已存在"))
			return
		}
	}
	if err := a.Service.AddLanguage(language); err != nil {
		respond(w, model.Error(500, "添加失败: "+err.Error()))
		return
	}
	ok(w, true)
}

func (a *API) delLanguage(w http.ResponseWriter, r *http.Request) {
	language := strings.TrimSpace(r.FormValue("language"))
	if language == "" {
		respond(w, model.Error(400, "语言不能为空"))
		return
	}
	existing, err := a.Service.GetAllLanguages()
	if err != nil {
		respond(w, model.Error(500, "删除失败: "+err.Error()))
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
		respond(w, model.Error(400, "语言不存在"))
		return
	}
	if err := a.Service.DeleteLanguage(language); err != nil {
		respond(w, model.Error(500, "删除失败: "+err.Error()))
		return
	}
	ok(w, true)
}

func (a *API) getAllCertificates(w http.ResponseWriter, r *http.Request) {
	certificates, err := a.Service.GetAllCertificates()
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, certificates)
}

func (a *API) addCertificate(w http.ResponseWriter, r *http.Request) {
	certificate := strings.TrimSpace(r.FormValue("certificate"))
	if certificate == "" {
		respond(w, model.Error(400, "证书不能为空"))
		return
	}
	existing, err := a.Service.GetAllCertificates()
	if err != nil {
		respond(w, model.Error(500, "添加失败: "+err.Error()))
		return
	}
	for _, item := range existing {
		if item == certificate {
			respond(w, model.Error(400, "证书已存在"))
			return
		}
	}
	if err := a.Service.AddCertificate(certificate); err != nil {
		respond(w, model.Error(500, "添加失败: "+err.Error()))
		return
	}
	ok(w, true)
}

func (a *API) delCertificate(w http.ResponseWriter, r *http.Request) {
	certificate := strings.TrimSpace(r.FormValue("certificate"))
	if certificate == "" {
		respond(w, model.Error(400, "证书不能为空"))
		return
	}
	existing, err := a.Service.GetAllCertificates()
	if err != nil {
		respond(w, model.Error(500, "删除失败: "+err.Error()))
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
		respond(w, model.Error(400, "证书不存在"))
		return
	}
	if err := a.Service.DeleteCertificate(certificate); err != nil {
		respond(w, model.Error(500, "删除失败: "+err.Error()))
		return
	}
	ok(w, true)
}

func (a *API) addSkillAll(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		a.fail(w, err)
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
		a.fail(w, svcErr)
		return
	}
	ok(w, fmt.Sprintf("添加成功%d个学生技能信息", count))
}

func (a *API) deleteSkill(w http.ResponseWriter, r *http.Request) {
	skills := r.URL.Query().Get("skills")
	if strings.TrimSpace(skills) == "" {
		respond(w, model.Error(500, "服务器内部错误: Required request parameter 'skills' is not present"))
		return
	}
	count, err := a.Service.DeleteSkills(skills)
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, fmt.Sprintf("删除%d个技能选项成功", count))
}
