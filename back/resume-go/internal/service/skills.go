package service

import (
	"strings"

	"resume-backend/internal/model"
)

func splitSkills(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func (s *Service) AddSkillsAll(raw string) (int, error) {
	return s.DB.AddSkills(splitSkills(raw))
}

func (s *Service) GetAllSkills() (string, error) {
	skills, err := s.DB.GetSkills()
	if err != nil {
		return "", err
	}
	return strings.Join(skills, ","), nil
}

func (s *Service) DeleteSkills(raw string) (int64, error) {
	skills := splitSkills(raw)
	if len(skills) == 0 {
		return 0, model.NewAppError(400, "技能不能为空")
	}
	return s.DB.DeleteSkills(skills)
}
