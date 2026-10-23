package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ---------- 统一响应 ----------

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func Success(data any) Result {
	return Result{Code: 200, Message: "操作成功", Data: data}
}

func Error(code int, message string) Result {
	return Result{Code: code, Message: message, Data: nil}
}

// ---------- 分页（兼容 PageHelper 的 PageInfo 结构） ----------

type PageInfo struct {
	Total             int   `json:"total"`
	List              any   `json:"list"`
	PageNum           int   `json:"pageNum"`
	PageSize          int   `json:"pageSize"`
	Size              int   `json:"size"`
	StartRow          int   `json:"startRow"`
	EndRow            int   `json:"endRow"`
	Pages             int   `json:"pages"`
	PrePage           int   `json:"prePage"`
	NextPage          int   `json:"nextPage"`
	IsFirstPage       bool  `json:"isFirstPage"`
	IsLastPage        bool  `json:"isLastPage"`
	HasPreviousPage   bool  `json:"hasPreviousPage"`
	HasNextPage       bool  `json:"hasNextPage"`
	NavigatePages     int   `json:"navigatePages"`
	NavigatePageNums  []int `json:"navigatepageNums"`
	NavigateFirstPage int   `json:"navigateFirstPage"`
	NavigateLastPage  int   `json:"navigateLastPage"`
	Count             int   `json:"count"`
}

func NewPageInfo(list any, total, pageNum, pageSize, size int) PageInfo {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	pages := 0
	if total > 0 {
		pages = (total + pageSize - 1) / pageSize
	}
	startRow := 0
	endRow := 0
	if size > 0 {
		startRow = (pageNum-1)*pageSize + 1
		endRow = startRow + size - 1
	}
	prePage := 0
	if pageNum > 1 {
		prePage = pageNum - 1
	}
	nextPage := 0
	if pageNum < pages {
		nextPage = pageNum + 1
	}
	navNums := make([]int, 0, 8)
	navFirst := 0
	navLast := 0
	if pages > 0 {
		begin := pageNum - 3
		if begin < 1 {
			begin = 1
		}
		end := begin + 7
		if end > pages {
			end = pages
		}
		for i := begin; i <= end; i++ {
			navNums = append(navNums, i)
		}
		navFirst = begin
		navLast = end
	}
	return PageInfo{
		Total:             total,
		List:              list,
		PageNum:           pageNum,
		PageSize:          pageSize,
		Size:              size,
		StartRow:          startRow,
		EndRow:            endRow,
		Pages:             pages,
		PrePage:           prePage,
		NextPage:          nextPage,
		IsFirstPage:       pageNum == 1,
		IsLastPage:        pageNum == pages || pages == 0,
		HasPreviousPage:   pageNum > 1,
		HasNextPage:       pageNum < pages,
		NavigatePages:     8,
		NavigatePageNums:  navNums,
		NavigateFirstPage: navFirst,
		NavigateLastPage:  navLast,
		Count:             total,
	}
}

// ---------- 实体 ----------

type User struct {
	UserID     int        `json:"userId"`
	Username   string     `json:"username"`
	Password   string     `json:"-"`
	UserType   string     `json:"userType"`
	CreateTime *time.Time `json:"createTime"`
	UpdateTime *time.Time `json:"updateTime"`
}

type Student struct {
	ID             int    `json:"id"`
	StuID          int    `json:"stuId"`
	Name           string `json:"name"`
	Age            *int   `json:"age"`
	Gender         string `json:"gender"`
	Nation         string `json:"nation"`
	Major          string `json:"major"`
	Degree         string `json:"degree"`
	GraduationYear *int   `json:"graduationYear"`
	Internship     *int   `json:"internship"`
	UserID         *int   `json:"userId"`
}

type Company struct {
	ID        int    `json:"id"`
	ComName   string `json:"comName"`
	ComRegion string `json:"comRegion"`
	UserID    *int   `json:"userId"`
}

// JobPosting 招聘信息。返回时同时输出驼峰与下划线两套字段名，兼容现有前端。
type JobPosting struct {
	ID           int
	ComID        int
	ComUserID    int
	ComName      string
	Title        string
	Description  string
	ExternalLink string
	Status       string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

func fmtDateTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.Format("2006-01-02 15:04:05")
}

func (j JobPosting) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"id":            j.ID,
		"comId":         j.ComID,
		"com_id":        j.ComID,
		"comUserId":     j.ComUserID,
		"com_user_id":   j.ComUserID,
		"comName":       j.ComName,
		"com_name":      j.ComName,
		"title":         j.Title,
		"description":   j.Description,
		"externalLink":  j.ExternalLink,
		"external_link": j.ExternalLink,
		"status":        j.Status,
		"createdAt":     fmtDateTime(j.CreatedAt),
		"created_at":    fmtDateTime(j.CreatedAt),
		"updatedAt":     fmtDateTime(j.UpdatedAt),
		"updated_at":    fmtDateTime(j.UpdatedAt),
	})
}

// JobPostingInput 招聘信息请求体，同时接受驼峰与下划线字段。
type JobPostingInput struct {
	ComID        *int
	Title        *string
	Description  *string
	ExternalLink *string
	Status       *string
}

func (j *JobPostingInput) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	pickInt := func(keys ...string) *int {
		for _, k := range keys {
			if v, ok := raw[k]; ok && string(v) != "null" {
				var n int
				if err := json.Unmarshal(v, &n); err == nil {
					return &n
				}
			}
		}
		return nil
	}
	pickStr := func(keys ...string) *string {
		for _, k := range keys {
			if v, ok := raw[k]; ok && string(v) != "null" {
				var s string
				if err := json.Unmarshal(v, &s); err == nil {
					return &s
				}
			}
		}
		return nil
	}
	j.ComID = pickInt("comId", "com_id")
	j.Title = pickStr("title")
	j.Description = pickStr("description")
	j.ExternalLink = pickStr("externalLink", "external_link")
	j.Status = pickStr("status")
	return nil
}

// ---------- 学生详情相关 DTO ----------

type LanguageDTO struct {
	StuID     int    `json:"stuId"`
	Language  string `json:"language"`
	Level     string `json:"level"`
	CertNo    string `json:"certNo"`
	IssueDate any    `json:"issueDate"`
}

type CertificateDTO struct {
	StuID     int    `json:"stuId"`
	CertName  string `json:"certName"`
	Level     string `json:"level"`
	CertNo    string `json:"certNo"`
	IssueDate any    `json:"issueDate"`
}

type InternshipDTO struct {
	StuID      int    `json:"stuId"`
	Company    string `json:"company"`
	Experience string `json:"experience"`
}

type JobDTO struct {
	StuID int    `json:"stuId"`
	Post  string `json:"post"`
}

type SpecialityDTO struct {
	StuID      int    `json:"stuId"`
	Speciality string `json:"speciality"`
}

type RegionDTO struct {
	Province *string `json:"province"`
	City     *string `json:"city"`
}

type ResumeDTO struct {
	Resume string `json:"resume"`
}

type StudentDetailDTO struct {
	ID             int              `json:"id"`
	StuID          int              `json:"stuId"`
	UserID         *int             `json:"userId"`
	Name           string           `json:"name"`
	Age            *int             `json:"age"`
	Gender         string           `json:"gender"`
	Nation         string           `json:"nation"`
	Major          string           `json:"major"`
	Degree         string           `json:"degree"`
	GraduationYear *int             `json:"graduationYear"`
	Internship     *int             `json:"internship"`
	Languages      []LanguageDTO    `json:"languages"`
	Certificates   []CertificateDTO `json:"certificates"`
	Internships    []InternshipDTO  `json:"internships"`
	Jobs           []JobDTO         `json:"jobs"`
	Region         *RegionDTO       `json:"region"`
	Resume         *ResumeDTO       `json:"resume"`
	Specialities   []SpecialityDTO  `json:"specialities"`
}

type StuFavorite struct {
	StuID          int    `json:"stuId"`
	UserID         int    `json:"userId"`
	StudentName    string `json:"studentName"`
	Major          string `json:"major"`
	Gender         string `json:"gender"`
	Degree         string `json:"degree"`
	GraduationYear string `json:"graduationYear"`
	Languages      string `json:"languages"`
	Level          string `json:"level"`
	Specialities   string `json:"specialities"`
	Post           string `json:"post"`
}

// ---------- 批量操作 ----------

type ErrorDetail struct {
	Identifier any    `json:"identifier"`
	Reason     string `json:"reason"`
}

type BatchOperationResult struct {
	SuccessCount int           `json:"successCount"`
	FailureCount int           `json:"failureCount"`
	Errors       []ErrorDetail `json:"errors"`
}

func NewBatchResult() *BatchOperationResult {
	return &BatchOperationResult{Errors: []ErrorDetail{}}
}

// ---------- 学生统计 ----------

type LanguageStat struct {
	Value    int    `json:"value"`
	Language string `json:"language"`
	Level    string `json:"level"`
}

type NameStat struct {
	Value int    `json:"value"`
	Name  string `json:"name"`
}

type InternshipStat struct {
	Value      int    `json:"value"`
	Company    string `json:"company"`
	Experience string `json:"experience"`
}

type StudentStatistics struct {
	Languages    []LanguageStat   `json:"languages"`
	Cities       []NameStat       `json:"cities"`
	Internships  []InternshipStat `json:"internships"`
	Specialities []NameStat       `json:"specialities"`
	Jobs         []NameStat       `json:"jobs"`
}

// ---------- 请求 DTO ----------

type LoginDTO struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	UserType      string `json:"userType"`
	UserTypeSnake string `json:"user_type"`
}

func (l LoginDTO) Type() string {
	if l.UserType != "" {
		return l.UserType
	}
	return l.UserTypeSnake
}

type LoginVO struct {
	UserID       int    `json:"userId"`
	Username     string `json:"username"`
	UserType     string `json:"userType"`
	DisplayName  string `json:"displayName"`
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

// UserProfile 当前登录用户的展示信息（账号名 + 昵称）
type UserProfile struct {
	UserID      int    `json:"userId"`
	Username    string `json:"username"`
	UserType    string `json:"userType"`
	DisplayName string `json:"displayName"`
}

type LanguageInfo struct {
	Language string `json:"language"`
	Level    string `json:"level"`
}

type CertificateInfo struct {
	CertName string `json:"certName"`
	Level    string `json:"level"`
}

type InternshipInfo struct {
	Company    string `json:"company"`
	Experience string `json:"experience"`
}

type StudentCompleteInfo struct {
	Age            *int
	Gender         *string
	Nation         *string
	Major          *string
	Degree         *string
	GraduationYear *int
	Internship     *int
	Languages      []LanguageInfo
	Cities         []string
	Provinces      []string
	Certificates   []CertificateInfo
	Internships    []InternshipInfo
	Specialities   []string
	Jobs           []string
}

type LikeItem struct {
	ComID *int
	StuID *int
}

func (l *LikeItem) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	pick := func(keys ...string) *int {
		for _, k := range keys {
			if v, ok := raw[k]; ok && string(v) != "null" {
				var n int
				if err := json.Unmarshal(v, &n); err == nil {
					return &n
				}
			}
		}
		return nil
	}
	l.ComID = pick("comId", "com_id")
	l.StuID = pick("stuId", "stu_id")
	return nil
}

type LikeQueryDTO struct {
	ComID      *int   `json:"com_id"`
	ComIDSnake *int   `json:"comId"`
	ComName    string `json:"comName"`
	PageNum    int    `json:"pageNum"`
	PageSize   int    `json:"pageSize"`
}

func (q *LikeQueryDTO) Normalize() {
	if q.ComID == nil {
		q.ComID = q.ComIDSnake
	}
	if q.PageNum < 1 {
		q.PageNum = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 10
	}
}

// StudentQuery 支持驼峰与下划线两种入参写法。
type StudentQuery struct {
	ID                  *int           `json:"id"`
	StuID               *int           `json:"stuId"`
	StuIDSnake          *int           `json:"stu_id"`
	Name                *string        `json:"name"`
	Age                 *int           `json:"age"`
	Gender              *string        `json:"gender"`
	Nation              *string        `json:"nation"`
	Major               *string        `json:"major"`
	Degree              *string        `json:"degree"`
	GraduationYear      *int           `json:"graduationYear"`
	GraduationYearSnake *int           `json:"graduation_year"`
	Internship          *int           `json:"internship"`
	Languages           []LanguageInfo `json:"languages"`
	Cities              []string       `json:"cities"`
	Provinces           []string       `json:"provinces"`
	CertNames           []string       `json:"certNames"`
	CertLevel           *string        `json:"certLevel"`
	InternshipCompany   *string        `json:"internshipCompany"`
	Specialities        []string       `json:"specialities"`
	Jobs                []string       `json:"jobs"`
	PageNum             int            `json:"pageNum"`
	PageNumSnake        int            `json:"page_num"`
	PageSize            int            `json:"pageSize"`
	PageSizeSnake       int            `json:"page_size"`
}

func (q *StudentQuery) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	// 数字字段同时接受 11 和 "11" 两种写法：
	// 前端从 URL query 里拿到的 id 默认是字符串，Java 版会自动转换，这里保持同样的宽松行为。
	q.ID = rawIntField(raw, "id")
	q.StuID = rawIntField(raw, "stuId")
	q.StuIDSnake = rawIntField(raw, "stu_id")
	q.Age = rawIntField(raw, "age")
	q.GraduationYear = rawIntField(raw, "graduationYear")
	q.GraduationYearSnake = rawIntField(raw, "graduation_year")
	q.Internship = rawIntField(raw, "internship")

	q.Name = rawStringField(raw, "name")
	q.Gender = rawStringField(raw, "gender")
	q.Nation = rawStringField(raw, "nation")
	q.Major = rawStringField(raw, "major")
	q.Degree = rawStringField(raw, "degree")
	q.CertLevel = rawStringField(raw, "certLevel")
	q.InternshipCompany = rawStringField(raw, "internshipCompany")

	q.Languages = rawLanguageList(raw, "languages")
	q.Cities = rawStringList(raw, "cities")
	q.Provinces = rawStringList(raw, "provinces")
	q.CertNames = rawStringList(raw, "certNames")
	q.Specialities = rawStringList(raw, "specialities")
	q.Jobs = rawStringList(raw, "jobs")

	if value := rawIntField(raw, "pageNum", "page_num"); value != nil {
		q.PageNum = *value
	}
	if value := rawIntField(raw, "pageSize", "page_size"); value != nil {
		q.PageSize = *value
	}
	return nil
}

func rawIntField(raw map[string]json.RawMessage, keys ...string) *int {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok || string(value) == "null" {
			continue
		}
		var number int
		if err := json.Unmarshal(value, &number); err == nil {
			return &number
		}
		var text string
		if err := json.Unmarshal(value, &text); err == nil {
			if number, err := strconv.Atoi(strings.TrimSpace(text)); err == nil {
				return &number
			}
		}
	}
	return nil
}

func rawStringField(raw map[string]json.RawMessage, keys ...string) *string {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok || string(value) == "null" {
			continue
		}
		var text string
		if err := json.Unmarshal(value, &text); err == nil {
			return &text
		}
	}
	return nil
}

func rawStringList(raw map[string]json.RawMessage, key string) []string {
	value, ok := raw[key]
	if !ok || string(value) == "null" {
		return nil
	}
	var list []string
	if err := json.Unmarshal(value, &list); err == nil {
		return list
	}
	// 兼容单值写法："cities": "北京"
	var single string
	if err := json.Unmarshal(value, &single); err == nil && single != "" {
		return []string{single}
	}
	return nil
}

func rawLanguageList(raw map[string]json.RawMessage, key string) []LanguageInfo {
	value, ok := raw[key]
	if !ok || string(value) == "null" {
		return nil
	}
	var list []LanguageInfo
	if err := json.Unmarshal(value, &list); err == nil {
		return list
	}
	// 兼容 ["英语", "日语"] 这种只传语言名的写法
	var names []string
	if err := json.Unmarshal(value, &names); err == nil {
		result := make([]LanguageInfo, 0, len(names))
		for _, name := range names {
			result = append(result, LanguageInfo{Language: name})
		}
		return result
	}
	return nil
}

func (q *StudentQuery) Normalize() {
	if q.StuID == nil {
		q.StuID = q.StuIDSnake
	}
	if q.GraduationYear == nil {
		q.GraduationYear = q.GraduationYearSnake
	}
	if q.PageNum < 1 {
		q.PageNum = q.PageNumSnake
	}
	if q.PageSize < 1 {
		q.PageSize = q.PageSizeSnake
	}
	if q.PageNum < 1 {
		q.PageNum = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 10
	}
}

type UpdateStudentsDTO struct {
	Students []StudentInput `json:"students"`
}

type StudentInput struct {
	ID                  *int    `json:"id"`
	StuID               *int    `json:"stuId"`
	StuIDSnake          *int    `json:"stu_id"`
	Name                *string `json:"name"`
	Age                 *int    `json:"age"`
	Gender              *string `json:"gender"`
	Nation              *string `json:"nation"`
	Major               *string `json:"major"`
	Degree              *string `json:"degree"`
	GraduationYear      *int    `json:"graduationYear"`
	GraduationYearSnake *int    `json:"graduation_year"`
	Internship          *int    `json:"internship"`
}

func (s *StudentInput) Normalize() {
	if s.StuID == nil {
		s.StuID = s.StuIDSnake
	}
	if s.GraduationYear == nil {
		s.GraduationYear = s.GraduationYearSnake
	}
}

type UpdateCompaniesDTO struct {
	Companies []CompanyInput `json:"companies"`
}

type CompanyInput struct {
	ID             *int    `json:"id"`
	ComName        *string `json:"comName"`
	ComNameSnake   *string `json:"com_name"`
	ComRegion      *string `json:"comRegion"`
	ComRegionSnake *string `json:"com_region"`
}

func (c *CompanyInput) Normalize() {
	if c.ComName == nil {
		c.ComName = c.ComNameSnake
	}
	if c.ComRegion == nil {
		c.ComRegion = c.ComRegionSnake
	}
}

type DeleteStudentsDTO struct {
	StuIDs []int `json:"stuIds"`
}

type DeleteCompaniesDTO struct {
	ComIDs []int `json:"comIds"`
}

type ResumeUploadVO struct {
	ResumeID   int    `json:"resumeId"`
	ResumeName string `json:"resumeName"`
	ResumeURL  string `json:"resumeUrl"`
}

// AppError 携带业务错误码，Error() 返回 "code:message" 以保持与 Java 异常相同的解析方式。
type AppError struct {
	Code int
	Msg  string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%d:%s", e.Code, e.Msg)
}

func NewAppError(code int, msg string) *AppError {
	return &AppError{Code: code, Msg: msg}
}
