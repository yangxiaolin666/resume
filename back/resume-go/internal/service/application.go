package service

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"resume-backend/internal/data"
	"resume-backend/internal/model"
)

// ==================================================================
// 简历投递业务层
// ==================================================================
// 对应简历中的「简历编辑与投递」。
//
// 核心流程：
//  1. 学生（必须已建档）选择职位投递；
//  2. 后端校验职位存在且处于 active 状态；
//  3. 简历地址优先使用前端传入的快照，否则取该学生最新上传的简历；
//  4. 写入 job_application，唯一索引负责防重复投递；
//  5. 企业按 com_id 查看自己收到的简历，并推进状态。
//
// 【为什么简历地址要做白名单校验】
// 如果不校验，学生可以传一个外部链接甚至 javascript: 伪协议，
// 企业点击时就可能被钓鱼或触发 XSS。这里只允许本站上传/下载地址。
// ==================================================================

// applicationStatuses 是投递状态白名单。
// 状态只能在这 5 个值之间变化，防止前端传入奇怪字符串污染数据。
var applicationStatuses = map[string]bool{
	model.ApplicationPending:   true,
	model.ApplicationViewed:    true,
	model.ApplicationInterview: true,
	model.ApplicationOffer:     true,
	model.ApplicationRejected:  true,
}

// ApplyJob 学生投递职位。
func (s *Service) ApplyJob(ctx context.Context, userID int, req model.ApplyJobRequest) (*model.JobApplication, error) {
	if req.JobID <= 0 {
		return nil, model.NewAppError(400, "请选择要投递的职位")
	}
	student, err := s.findStudent(userID)
	if err != nil {
		return nil, err
	}

	job, err := s.DB.FindJobPostingByID(req.JobID)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, model.NewAppError(404, "职位不存在")
	}
	if job.Status != "active" {
		return nil, model.NewAppError(400, "该职位已停止招聘")
	}

	resumeURL := strings.TrimSpace(req.ResumeURL)
	if resumeURL == "" {
		// 没有传快照时，自动使用最新上传的简历
		stored, err := s.DB.FindResumeByStudentID(student.ID)
		if err != nil {
			return nil, err
		}
		resumeURL = publicResumeURL(stored)
	}
	if resumeURL == "" {
		return nil, model.NewAppError(400, "请先上传简历后再投递")
	}
	if !strings.HasPrefix(resumeURL, "/uploads/") && !strings.HasPrefix(resumeURL, "/api/resume/download/") {
		return nil, model.NewAppError(400, "简历地址不合法")
	}

	id, err := s.DB.InsertApplication(job.ID, job.ComID, student.ID, resumeURL)
	if errors.Is(err, data.ErrDuplicateApplication) {
		return nil, model.NewAppError(409, "你已经投递过该职位，请勿重复投递")
	}
	if err != nil {
		return nil, err
	}
	return s.DB.FindApplicationByID(id)
}

// MyApplications 学生端「我的投递」列表。
func (s *Service) MyApplications(userID int) ([]model.JobApplication, error) {
	student, err := s.findStudent(userID)
	if err != nil {
		return nil, err
	}
	return s.DB.ListApplicationsByStudent(student.ID)
}

// WithdrawApplication 学生撤回投递。
func (s *Service) WithdrawApplication(userID int, id int64) error {
	student, err := s.findStudent(userID)
	if err != nil {
		return err
	}
	affected, err := s.DB.DeleteApplication(student.ID, id)
	if err != nil {
		return err
	}
	if affected == 0 {
		return model.NewAppError(404, "投递记录不存在或无权操作")
	}
	return nil
}

// CompanyApplications 企业端查看收到的简历，可按职位和状态筛选。
func (s *Service) CompanyApplications(userID, jobID int, status string) ([]model.JobApplication, error) {
	company, err := s.jobCompany(userID)
	if err != nil {
		return nil, err
	}
	if status != "" && !applicationStatuses[status] {
		return nil, model.NewAppError(400, "投递状态不正确")
	}
	return s.DB.ListApplicationsByCompany(company.ID, jobID, status)
}

// UpdateApplicationStatus 企业推进投递状态（已查看/面试/录用/不合适）。
func (s *Service) UpdateApplicationStatus(ctx context.Context, userID int, id int64, status string) (*model.JobApplication, error) {
	if !applicationStatuses[status] {
		return nil, model.NewAppError(400, "投递状态不正确")
	}
	company, err := s.jobCompany(userID)
	if err != nil {
		return nil, err
	}
	affected, err := s.DB.UpdateApplicationStatus(company.ID, id, status)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		// 注意：MySQL 在“新状态与旧状态相同”时也可能返回 0，
		// 所以这里再查一次记录是否存在，避免误报。
		item, err := s.DB.FindApplicationByID(id)
		if err != nil {
			return nil, err
		}
		if item == nil || item.ComID != company.ID {
			return nil, model.NewAppError(404, "投递记录不存在或无权操作")
		}
		return item, nil
	}
	return s.DB.FindApplicationByID(id)
}

// publicResumeURL 把数据库里保存的磁盘路径转换成静态资源 URL。
//
// 上传时数据库存的是 uploads/resume/xxx.pdf（磁盘路径），
// 不能直接返回给前端，否则会暴露服务器目录结构。
// 静态路由把 /uploads 映射到 uploads 目录，所以这里拼成 /uploads/resume/文件名。
func publicResumeURL(stored string) string {
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return ""
	}
	// 已经是 URL 形式（历史数据）就直接返回
	if strings.HasPrefix(stored, "/uploads/") || strings.HasPrefix(stored, "/api/resume/download/") {
		return stored
	}
	return "/uploads/resume/" + filepath.Base(stored)
}
