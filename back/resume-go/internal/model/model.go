// Package model 定义整个后端的「数据形状」：实体、请求 DTO、响应 VO 以及分页结构。
//
// 学习重点（按文件内的小节顺序阅读）：
//  1. 统一响应 Result            —— 前后端约定好的 {code, message, data} 三件套；
//  2. 分页 PageInfo              —— 刻意模仿 Java PageHelper 的字段，前端分页组件不用改；
//  3. 实体（User / Student / JobPosting ...）—— 与数据库表一一对应；
//  4. 各类 DTO                   —— 接收请求参数，并做「驼峰 / 下划线」双写兼容；
//  5. 业务错误 AppError          —— 让 HTTP 层能拿到业务错误码。
//
// 本文件只放结构体和与序列化相关的方法，不写数据库访问逻辑（那属于 repository/service）。
package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ---------- 统一响应 ----------

// Result 是所有接口的统一响应外壳，序列化后形如：
//
//	{"code":200,"message":"操作成功","data":{...}}
//
// 【为什么要统一】
//   - 前端只需要写一套判断逻辑：code == 200 视为成功，其余按 message 弹提示；
//   - HTTP 状态码始终是 200（见 api.respond），业务成败看 code，
//     这样浏览器 / axios 不会因为 4xx、5xx 走到 catch 分支，前端处理更简单；
//   - data 用 any 承载任意结构：对象、数组、分页对象都可以，不需要为每种响应写一个类型。
//
// 【和前端怎么配合】
//   - 成功：code=200，data 是业务数据（可能是 nil，例如“删除成功”）；
//   - 失败：code 是业务错误码（如 400 参数错误、401 未登录、500 服务器错误），data 为 null。
type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// Success 构造成功响应，固定 code=200。data 传 nil 表示“没有返回数据”。
func Success(data any) Result {
	return Result{Code: 200, Message: "操作成功", Data: data}
}

// Error 构造失败响应：code/message 由调用方决定，data 固定为 nil。
func Error(code int, message string) Result {
	return Result{Code: code, Message: message, Data: nil}
}

// ---------- 分页（兼容 PageHelper 的 PageInfo 结构） ----------

// PageInfo 统一的分页响应结构。
//
// 【为什么字段这么多、还要模仿 PageHelper】
// 这个项目是从 Java（MyBatis + PageHelper）迁移过来的，前端分页组件早就按
// PageHelper 的 PageInfo 字段写死了（total、pageNum、pages、isFirstPage...）。
// Go 端如果只返回 {total,list}，前端就得改动；所以这里完整复刻一套同名同义的字段，
// 让接口契约保持不变，前端零改动。
//
// 【真正有用的字段】
//   - Total / List：总条数与当前页数据，最常用；
//   - PageNum / PageSize / Pages：当前页码、每页条数、总页数；
//   - HasNextPage / HasPreviousPage：翻页按钮是否可用；
//   - NavigatePageNums：页码条上要渲染的页码数组（如 [1 2 3 4 5 6 7 8]）。
//
// 其余如 StartRow / EndRow / PrePage 属于 PageHelper 的历史字段，保留只为兼容。
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

// NewPageInfo 根据查询结果组装分页对象。
//
// 参数说明：
//   - list：当前页的数据切片（这里用 any，学生 / 企业 / 职位都能复用）；
//   - total：满足条件的总记录数（由 COUNT 查询得到）；
//   - pageNum / pageSize：前端请求的页码与每页条数；
//   - size：本页实际查到的条数（最后一页通常小于 pageSize）。
//
// 计算要点：
//   - 总页数用「向上取整」：(total + pageSize - 1) / pageSize；
//   - 页码条从「当前页 - 3」开始（不足 1 时从 1 开始），最多连续展示 8 个页码；
//   - 入参非法时兜底成第 1 页、每页 10 条，避免除零和负数。
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
//
// 这里的结构体对应数据库表，字段的 json tag 就是返回给前端的字段名。
// 注意指针类型（*int、*time.Time）的含义：允许为 NULL。
// 例如 Student.Age 用 *int，是因为数据库里年龄可以为空，
// 若用 int 会把 NULL 变成 0，前端就无法区分“0 岁”和“未填写”。

// User 登录账号表（学生、企业、管理员共用一个账号表，靠 UserType 区分）。
type User struct {
	UserID     int        `json:"userId"`
	Username   string     `json:"username"`
	Password   string     `json:"-"` // json:"-" 表示永远不参与序列化，避免密码泄露到接口响应
	UserType   string     `json:"userType"`
	CreateTime *time.Time `json:"createTime"`
	UpdateTime *time.Time `json:"updateTime"`
}

// Student 学生表的基本信息（不含语言、证书等子表）。
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

// Company 企业表的基本信息。
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

// fmtDateTime 把可空的 time.Time 转成 "2006-01-02 15:04:05" 字符串。
// 数据库里时间为 NULL 时返回 nil，序列化后就是 JSON 的 null。
// 前端统一按字符串展示，不需要再处理时区与 RFC3339 的 "T"/"Z"。
func fmtDateTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.Format("2006-01-02 15:04:05")
}

// MarshalJSON 自定义 JobPosting 的 JSON 序列化。
//
// 【为什么同一个值要输出两遍（comId 和 com_id）】
// 该项目前端经历了从 Java 版（下划线命名，如 com_id）到新前端（驼峰命名，如 comId）的演进，
// 两个版本都还在用。与其让前端改字段，不如后端一次返回两套 key：
//   - 新前端读 comId / createdAt；
//   - 老页面仍读 com_id / created_at。
//
// 代价是响应体略大，但兼容成本最低。
//
// 注意：结构体字段本身没有 json tag，所以默认序列化不会生效，
// 一旦实现了 MarshalJSON，encoding/json 就完全以这里返回的 map 为准。
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
//
// 字段全部用指针类型，是为了区分「没传这个字段」和「传了空值」：
//   - 指针为 nil：请求里没出现该字段，更新时保持数据库原值不变；
//   - 指针非 nil：请求里明确传了值（哪怕是 ""），更新时就要覆盖。
//
// 如果直接用 string，没传和传空串都会变成 ""，无法做局部更新。
type JobPostingInput struct {
	ComID        *int
	Title        *string
	Description  *string
	ExternalLink *string
	Status       *string
}

// UnmarshalJSON 自定义反序列化，把 comId / com_id 这类别名统一到一个字段上。
//
// 【技巧拆解】
//  1. 先把整个 JSON 解析成 map[string]json.RawMessage。
//     RawMessage 是「原样的字节」，可以延后再解析，避免一开始就因类型不匹配报错。
//  2. pickInt / pickStr 按候选 key 顺序查找，找到第一个存在的就用。
//     例如 pickInt("comId", "com_id")：新前端传 comId，老前端传 com_id，都能读到。
//  3. 判断 string(v) != "null"：显式传 null 等同于“没有传”，保持 nil。
//  4. 单个 key 解析失败不影响其他字段：ComID 传错类型时只是它为 nil，
//     真正的校验放在 service 层，报错信息更友好。
//
// 这就是「后端做兼容层」的典型写法：脏活留在边界，业务代码只面对统一字段。
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
//
// 学生的信息分散在多张子表（语言、证书、实习、岗位、地区、简历、特长），
// 查询详情时由 service 分别查出来再拼成 StudentDetailDTO 一次返回，
// 前端就不必发七八个请求。这就是常说的「接口聚合」。

// LanguageDTO 学生语言能力（对应 language 子表的一行）。
type LanguageDTO struct {
	StuID     int    `json:"stuId"`
	Language  string `json:"language"`
	Level     string `json:"level"`
	CertNo    string `json:"certNo"`
	IssueDate any    `json:"issueDate"` // 可能是字符串或 null，用 any 直接透传
}

// CertificateDTO 学生证书。
type CertificateDTO struct {
	StuID     int    `json:"stuId"`
	CertName  string `json:"certName"`
	Level     string `json:"level"`
	CertNo    string `json:"certNo"`
	IssueDate any    `json:"issueDate"`
}

// InternshipDTO 实习经历。
type InternshipDTO struct {
	StuID      int    `json:"stuId"`
	Company    string `json:"company"`
	Experience string `json:"experience"`
}

// JobDTO 意向岗位 / 历史岗位。
type JobDTO struct {
	StuID int    `json:"stuId"`
	Post  string `json:"post"`
}

// SpecialityDTO 专业特长。
type SpecialityDTO struct {
	StuID      int    `json:"stuId"`
	Speciality string `json:"speciality"`
}

// RegionDTO 期望工作地区，省 / 市都允许为空。
type RegionDTO struct {
	Province *string `json:"province"`
	City     *string `json:"city"`
}

// ResumeDTO 简历文件信息（实际只返回访问地址）。
type ResumeDTO struct {
	Resume string `json:"resume"`
}

// StudentDetailDTO 学生完整详情，聚合了上面所有子表的数据。
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

// StuFavorite 企业「收藏的学生」列表中的一行（已把各子表拼成扁平字符串，方便表格展示）。
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
//
// 管理端经常一次处理多行数据（批量导入 / 批量删除 / 批量修改）。
// 这类接口不适合「遇到第一个错误就整体回滚」：用户会希望知道哪些成功了、哪些失败、为什么。
// 因此统一返回 BatchOperationResult，让前端能逐行展示失败原因。

// ErrorDetail 单条失败记录：Identifier 是被操作对象的标识（学号 / 名称等），Reason 是失败原因。
type ErrorDetail struct {
	Identifier any    `json:"identifier"`
	Reason     string `json:"reason"`
}

// BatchOperationResult 批量操作结果汇总。
// 只要 FailureCount > 0，接口就会返回“部分成功”的提示，而不是简单地报 500。
type BatchOperationResult struct {
	SuccessCount int           `json:"successCount"`
	FailureCount int           `json:"failureCount"`
	Errors       []ErrorDetail `json:"errors"`
}

// NewBatchResult 创建空结果。
// 注意 Errors 初始化为空切片而不是 nil：序列化后是 [] 而不是 null，
// 前端可以直接 .length / .map，不用额外判空。
func NewBatchResult() *BatchOperationResult {
	return &BatchOperationResult{Errors: []ErrorDetail{}}
}

// ---------- 学生统计 ----------
//
// 首页 / 管理端仪表盘用的聚合数据：按语言、城市、实习公司、专业、岗位分组计数。
// 每个 Stat 都带 Value（数量）和对应的名称字段，前端直接喂给图表组件。

// LanguageStat 按语言（可细分到等级）统计的人数。
type LanguageStat struct {
	Value    int    `json:"value"`
	Language string `json:"language"`
	Level    string `json:"level"`
}

// NameStat 通用「名称 + 数量」统计项（城市、专业、岗位都用它）。
type NameStat struct {
	Value int    `json:"value"`
	Name  string `json:"name"`
}

// InternshipStat 按实习公司统计。
type InternshipStat struct {
	Value      int    `json:"value"`
	Company    string `json:"company"`
	Experience string `json:"experience"`
}

// StudentStatistics 学生统计总览，一次性返回五组图表数据。
type StudentStatistics struct {
	Languages    []LanguageStat   `json:"languages"`
	Cities       []NameStat       `json:"cities"`
	Internships  []InternshipStat `json:"internships"`
	Specialities []NameStat       `json:"specialities"`
	Jobs         []NameStat       `json:"jobs"`
}

// ---------- 请求 DTO ----------
//
// 这一节是「入口数据」：HTTP 层把请求体解析成这些结构，再交给 service。
// 命名习惯：XxxDTO 表示请求体，XxxVO 表示返回给前端的展示对象。

// LoginDTO 登录请求体。
//
// UserType 与 UserTypeSnake 是同一个含义的两种字段名：
// 新前端传 userType，老前端传 user_type，二者都保留，取值时用 Type() 统一。
type LoginDTO struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	UserType      string `json:"userType"`
	UserTypeSnake string `json:"user_type"`
}

// Type 返回用户类型，优先取驼峰字段，找不到再取蛇形字段。
// 这样调用方只关心“最终类型是什么”，不用自己写 if 判断两遍。
func (l LoginDTO) Type() string {
	if l.UserType != "" {
		return l.UserType
	}
	return l.UserTypeSnake
}

// LoginVO 登录成功后返回给前端的用户信息 + 双 Token。
//
// Token 用于访问业务接口（放在 Authorization 头里）；
// RefreshToken 专门用来换取新 Token，有效期更长，避免频繁要求用户重新登录。
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

// LanguageInfo 语言 + 等级（提交表单时用）。
type LanguageInfo struct {
	Language string `json:"language"`
	Level    string `json:"level"`
}

// CertificateInfo 证书名 + 等级（提交表单时用）。
type CertificateInfo struct {
	CertName string `json:"certName"`
	Level    string `json:"level"`
}

// InternshipInfo 实习公司 + 经历描述。
type InternshipInfo struct {
	Company    string `json:"company"`
	Experience string `json:"experience"`
}

// StudentCompleteInfo 学生完善资料页一次性提交的全部信息。
//
// 字段用指针：nil 表示“这次没有提交该字段”，service 可以据此决定跳过更新；
// 切片字段则整体替换（先删旧数据再插新数据）。
// 注意这些字段没有 json tag，默认按字段名匹配（大小写不敏感），
// JSON 里的 age、gender 等都能正常解析。
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

// UnmarshalJSON 兼容 specialities 的两种提交格式：
// ["Java"] 与 [{"speciality":"Java"}]（旧前端提交的是对象数组）。
//
// 【type alias 技巧】
// 如果直接在方法里 json.Unmarshal(data, s)，会再次调用本方法，无限递归。
// 这里先定义 type alias StudentCompleteInfo（新类型，不继承方法集），
// 把 *alias 嵌进临时结构体，让其余字段走默认解析逻辑；
// 只把 specialities 单独用 RawMessage 接住，事后再手工拆成 []string。
//
// 解析时逐项尝试：先按字符串解析，成功就收下；
// 失败再按 {speciality} / {name} 对象解析，兼容旧数据结构。
// 空字符串会被过滤，避免前端提交空行时写入脏数据。
func (s *StudentCompleteInfo) UnmarshalJSON(data []byte) error {
	type alias StudentCompleteInfo
	aux := struct {
		*alias
		Specialities json.RawMessage `json:"specialities"`
	}{alias: (*alias)(s)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if len(aux.Specialities) == 0 {
		return nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(aux.Specialities, &items); err != nil {
		return err
	}
	specialities := make([]string, 0, len(items))
	for _, item := range items {
		var text string
		if err := json.Unmarshal(item, &text); err == nil {
			if strings.TrimSpace(text) != "" {
				specialities = append(specialities, text)
			}
			continue
		}
		var obj struct {
			Speciality string `json:"speciality"`
			Name       string `json:"name"`
		}
		if err := json.Unmarshal(item, &obj); err != nil {
			return err
		}
		if obj.Speciality != "" {
			specialities = append(specialities, obj.Speciality)
		} else if obj.Name != "" {
			specialities = append(specialities, obj.Name)
		}
	}
	s.Specialities = specialities
	return nil
}

// LikeItem 企业收藏学生的一条记录。
// 请求里两个 ID 都可能用到：企业端自己收藏时由后端补 ComID，删除时前端会带上两个 ID。
type LikeItem struct {
	ComID *int
	StuID *int
}

// UnmarshalJSON 同样兼容 comId / com_id、stuId / stu_id 两种写法。
// 这是收藏接口被新老前端同时调用时最容易踩的坑，因此在边界统一掉。
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

// LikeQueryDTO 收藏列表的查询条件。
//
// ComID 与 ComIDSnake 一个接下划线、一个接驼峰；Normalize 负责合并成 ComID，
// 同时给分页参数兜底。调用 service 前先 Normalize，业务层就只认 ComID 一个字段。
type LikeQueryDTO struct {
	ComID      *int   `json:"com_id"`
	ComIDSnake *int   `json:"comId"`
	ComName    string `json:"comName"`
	PageNum    int    `json:"pageNum"`
	PageSize   int    `json:"pageSize"`
}

// Normalize 把两种命名的字段合并，并补齐默认分页值（第 1 页、每页 10 条）。
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

// StudentQuery 学生搜索条件，是本文件里最“宽松”的入参结构。
//
// 【为什么同一个字段出现两次（带 Snake 后缀）】
// 前端新页面用驼峰（stuId、graduationYear、pageNum），
// 旧页面和老接口文档用下划线（stu_id、graduation_year、page_num）。
// 两套字段都声明出来，由 UnmarshalJSON + Normalize 合并，service 只读合并后的那个。
//
// 【为什么全部用指针】
// 搜索是「有哪些条件就过滤哪些」：nil 表示用户没填，不能当成 0 参与 SQL 条件。
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

// UnmarshalJSON 手工解析招生搜索条件，处理三类“宽容”场景：
//  1. 驼峰 / 下划线双写：pageNum 与 page_num 都接受；
//  2. 数字既可能是 11 也可能是 "11"（URL query 里全是字符串）；
//  3. 列表字段既可能传数组，也可能只传一个字符串。
//
// 这里不直接依赖 struct tag，而是先转成 map[string]json.RawMessage，
// 再用下面的 rawXxxField 工具函数逐个读取，把兼容逻辑集中在一处。
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

// rawIntField 从原始 JSON map 中按候选 key 读整数，支持 11 和 "11" 两种写法。
// 返回 nil 表示这些 key 都没出现（或值是 null / 无法解析）。
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

// rawStringField 读取字符串字段，不存在 / null 时返回 nil。
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

// rawStringList 读取字符串数组；若前端传的是单个字符串，则包成长度为 1 的切片。
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

// rawLanguageList 读取语言条件，兼容两种格式：
//   - [{"language":"英语","level":"CET6"}]：标准对象数组；
//   - ["英语","日语"]：只传语言名，等级留空。
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

// Normalize 合并驼峰 / 下划线两套字段，并给分页兜底。
// 顺序很重要：先「用 Snake 补空」再「做默认值」，
// 否则默认值会先把空字段填成 1，Snake 传来的真实页码就永远覆盖不上。
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

// UpdateStudentsDTO 管理端批量修改学生：外层包一个数组，便于一次提交多行。
type UpdateStudentsDTO struct {
	Students []StudentInput `json:"students"`
}

// StudentInput 批量修改时单个学生的字段集合（只包含可编辑列）。
// 同样保留 Snake 别名，Normalize 后再交给 service。
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

// Normalize 合并 StudentInput 的两种字段写法。
func (s *StudentInput) Normalize() {
	if s.StuID == nil {
		s.StuID = s.StuIDSnake
	}
	if s.GraduationYear == nil {
		s.GraduationYear = s.GraduationYearSnake
	}
}

// UpdateCompaniesDTO 管理端批量修改企业。
type UpdateCompaniesDTO struct {
	Companies []CompanyInput `json:"companies"`
}

// CompanyInput 批量修改时单个企业的字段集合。
type CompanyInput struct {
	ID             *int    `json:"id"`
	ComName        *string `json:"comName"`
	ComNameSnake   *string `json:"com_name"`
	ComRegion      *string `json:"comRegion"`
	ComRegionSnake *string `json:"com_region"`
}

// Normalize 合并 CompanyInput 的两种字段写法。
func (c *CompanyInput) Normalize() {
	if c.ComName == nil {
		c.ComName = c.ComNameSnake
	}
	if c.ComRegion == nil {
		c.ComRegion = c.ComRegionSnake
	}
}

// DeleteStudentsDTO 批量删除学生，只需要学号列表。
type DeleteStudentsDTO struct {
	StuIDs []int `json:"stuIds"`
}

// DeleteCompaniesDTO 批量删除企业，只需要企业 ID 列表。
type DeleteCompaniesDTO struct {
	ComIDs []int `json:"comIds"`
}

// ResumeUploadVO 简历上传成功后返回的信息：
// ResumeID 用于后续关联，ResumeURL 是可直接访问 / 下载的静态地址。
type ResumeUploadVO struct {
	ResumeID   int    `json:"resumeId"`
	ResumeName string `json:"resumeName"`
	ResumeURL  string `json:"resumeUrl"`
}

// ResumeInfo 是「当前学生最新简历」的精简信息，供聊天发送简历使用。
// URL 是可直接访问的静态资源地址，Name 是文件名。
type ResumeInfo struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// AppError 携带业务错误码，Error() 返回 "code:message" 以保持与 Java 异常相同的解析方式。
//
// 【为什么不用普通 error】
// service 层需要告诉 HTTP 层「用哪个业务码返回」（如 401 未登录、40009 文件为空）。
// 普通 error 只有文本，HTTP 层无法区分；自定义类型则可以用 errors.As 取出 Code，
// 然后通过 model.Error(code, msg) 输出统一响应。
// 详见 api 包的 fail 方法。
type AppError struct {
	Code int
	Msg  string
}

// Error 实现 error 接口，文本格式为 "code:message"，便于日志里直接看到错误码。
func (e *AppError) Error() string {
	return fmt.Sprintf("%d:%s", e.Code, e.Msg)
}

// NewAppError 构造业务错误。
func NewAppError(code int, msg string) *AppError {
	return &AppError{Code: code, Msg: msg}
}
