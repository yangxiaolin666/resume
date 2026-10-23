package service

import (
	"strings"

	"resume-backend/internal/model"
)

// ==================================================================
// 技能字典业务层
// ==================================================================
// 技能既是「字典」（管理端维护可选项），也是学生档案里的「专业技能」来源。
// 前端习惯用逗号分隔的字符串传参/展示，所以这一层负责字符串与切片之间的转换。
//
// 【与其它层的关系】
//   - api：接收 skills=Java,Go,MySQL 这样的字符串；
//   - data：批量写入/删除 skill 表，返回受影响行数；
//   - cache/mq/rpc：不涉及，字典数据量小，直接查库即可。
//
// 【对应简历能力】字典维护与批量写入、前后端数据格式适配。
// ==================================================================

// splitSkills 把逗号分隔的字符串切成去空格、去空项的切片。
// 例如 "Java, Go, ,MySQL " -> ["Java","Go","MySQL"]。
// 这里不做去重：skill 表没有唯一约束，重复项会被原样插入，
// 清洗只负责格式，避免在多个地方重复实现同一套规则。
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

// AddSkillsAll 批量新增技能。
// 返回受影响的行数；空字符串会被 splitSkills 清洗成空切片，
// 最终一次都不会插入，不会产生空技能记录。
func (s *Service) AddSkillsAll(raw string) (int, error) {
	return s.DB.AddSkills(splitSkills(raw))
}

// GetAllSkills 返回全部技能，拼成逗号分隔的字符串。
// 这是为了兼容前端已有接口的数据格式（直接展示/放进输入框），
// 而不是返回 JSON 数组；查询失败时返回空串和错误。
func (s *Service) GetAllSkills() (string, error) {
	skills, err := s.DB.GetSkills()
	if err != nil {
		return "", err
	}
	return strings.Join(skills, ","), nil
}

// DeleteSkills 批量删除技能。
// 清洗后为空说明没有任何有效名称，直接返回 400，
// 避免执行一条无意义的 DELETE（也防止把空条件误当成“删除全部”）。
// 返回实际删除的行数，前端可据此提示是否有名称不存在。
func (s *Service) DeleteSkills(raw string) (int64, error) {
	skills := splitSkills(raw)
	if len(skills) == 0 {
		return 0, model.NewAppError(400, "技能不能为空")
	}
	return s.DB.DeleteSkills(skills)
}
