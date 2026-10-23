package model

// ==================================================================
// 简历投递（学生 -> 职位）数据模型
// ==================================================================
// 一条记录表示「某个学生把简历投给了某个职位」。
//
// 【表设计要点】
//   - job_id + student_id 建唯一索引，从数据库层面防止重复投递；
//   - com_id 是冗余字段：企业查看投递列表时直接按 com_id 过滤，
//     不用每次都 JOIN job_postings 再判断归属，查询更快、权限校验也更简单；
//   - resume_url 是「投递时的简历快照」：学生之后修改/删除简历，
//     企业看到的仍然是当初投递的那一份，符合招聘业务的实际语义。
// ==================================================================

// 投递状态机（面试常问：为什么用字符串而不是数字枚举）
//
//	pending   已投递，企业还没查看
//	viewed    企业已查看
//	interview 邀请面试
//	offer     已录用
//	rejected  不合适
//
// 用字符串是为了排查问题时一眼能看懂，且新增状态不需要改表结构；
// 代价是占用空间略大、写入前必须做白名单校验（见 service/application.go）。
const (
	ApplicationPending   = "pending"
	ApplicationViewed    = "viewed"
	ApplicationInterview = "interview"
	ApplicationOffer     = "offer"
	ApplicationRejected  = "rejected"
)

// JobApplication 投递记录（同时用于学生端“我的投递”和企业端“收到的简历”）。
// 除了表里的原始字段，还通过 JOIN 带出职位标题、企业名称、学生姓名等展示字段。
type JobApplication struct {
	ID            int64  `json:"id"`
	JobID         int    `json:"jobId"`
	JobTitle      string `json:"jobTitle"`
	ComID         int    `json:"comId"`
	ComName       string `json:"comName"`
	StudentID     int    `json:"studentId"`
	StudentUserID int    `json:"studentUserId"`
	StudentName   string `json:"studentName"`
	Major         string `json:"major"`
	Degree        string `json:"degree"`
	ResumeURL     string `json:"resumeUrl"`
	Status        string `json:"status"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// ApplyJobRequest 学生投递职位的请求体。
// ResumeURL 为空时，后端会自动使用该学生最新上传的简历。
type ApplyJobRequest struct {
	JobID     int    `json:"jobId"`
	ResumeURL string `json:"resumeUrl"`
}

// UpdateApplicationStatusRequest 企业修改投递状态的请求体。
type UpdateApplicationStatusRequest struct {
	Status string `json:"status"`
}
