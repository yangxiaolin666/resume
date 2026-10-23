package service

import (
	"fmt"
	"strings"

	"resume-backend/internal/data"
	"resume-backend/internal/model"
)

func (s *Service) jobCompany(userID int) (*model.Company, error) {
	company, err := s.DB.FindCompanyByUserID(userID)
	if err != nil {
		return nil, err
	}
	if company == nil {
		return nil, model.NewAppError(404, "企业信息不存在")
	}
	return company, nil
}

func (s *Service) PublishJob(userID int, input model.JobPostingInput) (*model.JobPosting, error) {
	company, err := s.DB.FindCompanyByUserID(userID)
	if err != nil {
		return nil, err
	}
	if company == nil {
		return nil, model.NewAppError(404, fmt.Sprintf("企业信息不存在（userId: %d）", userID))
	}
	if input.Title == nil || strings.TrimSpace(*input.Title) == "" {
		return nil, model.NewAppError(400, "招聘标题不能为空")
	}
	status := "active"
	if input.Status != nil {
		status = *input.Status
	}
	description, link := "", ""
	if input.Description != nil {
		description = *input.Description
	}
	if input.ExternalLink != nil {
		link = *input.ExternalLink
	}
	id, err := s.DB.InsertJobPosting(company.ID, *input.Title, description, link, status)
	if err != nil {
		return nil, err
	}
	return s.DB.FindJobPostingByID(id)
}

func (s *Service) UpdateJob(userID, id int, input model.JobPostingInput) (*model.JobPosting, error) {
	company, err := s.jobCompany(userID)
	if err != nil {
		return nil, err
	}
	existing, err := s.DB.FindJobPostingByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, model.NewAppError(500, "服务器内部错误: 招聘信息不存在")
	}
	if existing.ComID != company.ID {
		return nil, model.NewAppError(403, "无权限修改该招聘信息")
	}
	if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
		return nil, model.NewAppError(400, "招聘标题不能为空")
	}
	if err := s.DB.UpdateJobPosting(id, input); err != nil {
		return nil, err
	}
	return s.DB.FindJobPostingByID(id)
}

func (s *Service) DeleteJob(userID, id int) (bool, error) {
	company, err := s.jobCompany(userID)
	if err != nil {
		return false, err
	}
	existing, err := s.DB.FindJobPostingByID(id)
	if err != nil {
		return false, err
	}
	if existing == nil {
		return false, model.NewAppError(500, "服务器内部错误: 招聘信息不存在")
	}
	if existing.ComID != company.ID {
		return false, model.NewAppError(403, "无权限删除该招聘信息")
	}
	rows, err := s.DB.DeleteJobPosting(id)
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (s *Service) JobPostings(status, keyword string, comID *int, pageNum, pageSize int) (*model.PageInfo, error) {
	filter := data.JobFilter{ComID: comID}
	if status != "" {
		filter.Status = &status
	}
	if keyword != "" {
		filter.Keyword = &keyword
	}
	list, total, err := s.DB.ListJobPostings(filter, pageNum, pageSize)
	if err != nil {
		return nil, err
	}
	page := model.NewPageInfo(list, total, pageNum, pageSize, len(list))
	return &page, nil
}

func (s *Service) JobDetail(id int) (*model.JobPosting, error) {
	job, err := s.DB.FindJobPostingByID(id)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, model.NewAppError(500, "服务器内部错误: 招聘信息不存在")
	}
	return job, nil
}

func (s *Service) MyJobPostings(userID int, status string) ([]model.JobPosting, error) {
	company, err := s.jobCompany(userID)
	if err != nil {
		return nil, err
	}
	return s.DB.ListJobPostingsByComID(status, company.ID)
}

func (s *Service) UpdateJobStatus(userID, id int, status string) (*model.JobPosting, error) {
	company, err := s.jobCompany(userID)
	if err != nil {
		return nil, err
	}
	existing, err := s.DB.FindJobPostingByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, model.NewAppError(500, "服务器内部错误: 招聘信息不存在")
	}
	if existing.ComID != company.ID {
		return nil, model.NewAppError(403, "无权限修改该招聘信息")
	}
	if status != "active" && status != "inactive" {
		return nil, model.NewAppError(500, "服务器内部错误: 无效的状态值，只能是active或inactive")
	}
	if err := s.DB.UpdateJobPostingStatus(id, status); err != nil {
		return nil, err
	}
	return s.DB.FindJobPostingByID(id)
}
