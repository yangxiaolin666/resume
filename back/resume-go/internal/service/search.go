package service

import "resume-backend/internal/model"

func (s *Service) SearchStudents(q model.StudentQuery) (*model.PageInfo, error) {
	q.Normalize()
	list, total, err := s.DB.SearchStudents(q)
	if err != nil {
		return nil, err
	}
	page := model.NewPageInfo(list, total, q.PageNum, q.PageSize, len(list))
	return &page, nil
}

func (s *Service) SearchStudentDetails(q model.StudentQuery) (*model.PageInfo, error) {
	q.Normalize()
	list, total, err := s.DB.SearchStudentDetails(q)
	if err != nil {
		return nil, err
	}
	page := model.NewPageInfo(list, total, q.PageNum, q.PageSize, len(list))
	return &page, nil
}
