package service

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"

	"resume-backend/internal/auth"
	"resume-backend/internal/model"
)

func (s *Service) ImportStudents(reader io.Reader) (*model.BatchOperationResult, error) {
	rows, err := readExcelRows(reader)
	if err != nil {
		return nil, err
	}
	result := model.NewBatchResult()

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for i := 1; i < len(rows); i++ {
		rowNum := i + 1
		row := rows[i]
		if rowEmpty(row) {
			continue
		}
		stuID, err := cellInt(row, 0)
		if err != nil || stuID == nil {
			result.FailureCount++
			result.Errors = append(result.Errors, model.ErrorDetail{Identifier: rowNum, Reason: "数据解析失败: 学号不能为空"})
			continue
		}
		student := model.Student{
			StuID:          *stuID,
			Name:           cellString(row, 1),
			Age:            mustInt(row, 2),
			Gender:         cellString(row, 3),
			Nation:         cellString(row, 4),
			Major:          cellString(row, 5),
			Degree:         cellString(row, 6),
			GraduationYear: mustInt(row, 7),
			Internship:     mustInt(row, 8),
		}

		existing, err := s.DB.FindStudentByStuID(student.StuID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			result.FailureCount++
			result.Errors = append(result.Errors, model.ErrorDetail{
				Identifier: rowNum, Reason: fmt.Sprintf("学号%d已存在", student.StuID),
			})
			continue
		}

		hash, err := auth.HashPassword(strconv.Itoa(student.StuID))
		if err != nil {
			return nil, err
		}
		userID, err := s.DB.InsertUserTx(tx, strconv.Itoa(student.StuID), hash, "student")
		if err != nil {
			return nil, err
		}
		student.UserID = &userID
		if err := s.DB.InsertStudentTx(tx, student); err != nil {
			return nil, err
		}
		result.SuccessCount++
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) ImportCompanies(reader io.Reader) (*model.BatchOperationResult, error) {
	rows, err := readExcelRows(reader)
	if err != nil {
		return nil, err
	}
	result := model.NewBatchResult()

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for i := 1; i < len(rows); i++ {
		rowNum := i + 1
		row := rows[i]
		if rowEmpty(row) {
			continue
		}
		comName := strings.TrimSpace(cellString(row, 0))
		if comName == "" {
			result.FailureCount++
			result.Errors = append(result.Errors, model.ErrorDetail{Identifier: rowNum, Reason: "数据解析失败: 企业名称不能为空"})
			continue
		}
		company := model.Company{ComName: comName, ComRegion: cellString(row, 1)}

		existing, err := s.DB.FindCompanyByName(company.ComName)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			result.FailureCount++
			result.Errors = append(result.Errors, model.ErrorDetail{
				Identifier: rowNum, Reason: fmt.Sprintf("企业%s已存在", company.ComName),
			})
			continue
		}

		hash, err := auth.HashPassword(company.ComName)
		if err != nil {
			return nil, err
		}
		userID, err := s.DB.InsertUserTx(tx, company.ComName, hash, "company")
		if err != nil {
			return nil, err
		}
		company.UserID = &userID
		if err := s.DB.InsertCompanyTx(tx, company); err != nil {
			return nil, err
		}
		result.SuccessCount++
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) DeleteStudents(ids []int) (*model.BatchOperationResult, error) {
	result := model.NewBatchResult()
	validIDs := []int{}
	userIDs := []int{}
	for _, id := range ids {
		student, err := s.DB.FindStudentByID(id)
		if err != nil {
			return nil, err
		}
		if student == nil {
			result.FailureCount++
			result.Errors = append(result.Errors, model.ErrorDetail{Identifier: id, Reason: "学生不存在"})
			continue
		}
		validIDs = append(validIDs, id)
		if student.UserID != nil {
			userIDs = append(userIDs, *student.UserID)
		}
	}
	if len(validIDs) == 0 {
		return result, nil
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	deleted, err := s.DB.DeleteStudentsTx(tx, validIDs)
	if err != nil {
		return nil, err
	}
	if err := s.DB.DeleteUsersTx(tx, userIDs); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	result.SuccessCount = int(deleted)
	return result, nil
}

func (s *Service) DeleteCompanies(ids []int) (*model.BatchOperationResult, error) {
	result := model.NewBatchResult()
	validIDs := []int{}
	userIDs := []int{}
	for _, id := range ids {
		company, err := s.DB.FindCompanyByID(id)
		if err != nil {
			return nil, err
		}
		if company == nil {
			result.FailureCount++
			result.Errors = append(result.Errors, model.ErrorDetail{Identifier: id, Reason: "企业不存在"})
			continue
		}
		validIDs = append(validIDs, id)
		if company.UserID != nil {
			userIDs = append(userIDs, *company.UserID)
		}
	}
	if len(validIDs) == 0 {
		return result, nil
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	deleted, err := s.DB.DeleteCompaniesTx(tx, validIDs)
	if err != nil {
		return nil, err
	}
	if err := s.DB.DeleteUsersTx(tx, userIDs); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	result.SuccessCount = int(deleted)
	return result, nil
}

func (s *Service) UpdateStudents(students []model.StudentInput) (*model.BatchOperationResult, error) {
	result := model.NewBatchResult()
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for i, student := range students {
		index := i + 1
		student.Normalize()
		if student.ID == nil {
			result.FailureCount++
			result.Errors = append(result.Errors, model.ErrorDetail{Identifier: index, Reason: "学生ID不能为空"})
			continue
		}
		existing, err := s.DB.FindStudentByID(*student.ID)
		if err != nil {
			return nil, err
		}
		if existing == nil {
			result.FailureCount++
			result.Errors = append(result.Errors, model.ErrorDetail{Identifier: index, Reason: "学生不存在"})
			continue
		}
		rows, err := s.DB.UpdateStudentTx(tx, student)
		if err != nil {
			return nil, err
		}
		if rows > 0 {
			result.SuccessCount++
		} else {
			result.FailureCount++
			result.Errors = append(result.Errors, model.ErrorDetail{Identifier: index, Reason: "更新失败"})
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) UpdateCompanies(companies []model.CompanyInput) (*model.BatchOperationResult, error) {
	result := model.NewBatchResult()
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for i, company := range companies {
		index := i + 1
		company.Normalize()
		if company.ID == nil {
			result.FailureCount++
			result.Errors = append(result.Errors, model.ErrorDetail{Identifier: index, Reason: "企业ID不能为空"})
			continue
		}
		existing, err := s.DB.FindCompanyByID(*company.ID)
		if err != nil {
			return nil, err
		}
		if existing == nil {
			result.FailureCount++
			result.Errors = append(result.Errors, model.ErrorDetail{Identifier: index, Reason: "企业不存在"})
			continue
		}
		rows, err := s.DB.UpdateCompanyTx(tx, company)
		if err != nil {
			return nil, err
		}
		if rows > 0 {
			result.SuccessCount++
		} else {
			result.FailureCount++
			result.Errors = append(result.Errors, model.ErrorDetail{Identifier: index, Reason: "更新失败"})
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) AdminSearchStudents(name string, page, pageSize int) (map[string]any, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	list, total, err := s.DB.AdminStudents(name, page, pageSize)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"list": list,
		"pagination": map[string]any{
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	}, nil
}

func (s *Service) AdminSearchCompanies(name string, page, pageSize int) (map[string]any, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	list, total, err := s.DB.AdminCompanies(name, page, pageSize)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"list": list,
		"pagination": map[string]any{
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	}, nil
}

func (s *Service) UpdateStudentPassword(stuID int, pwd string) (bool, error) {
	hash, err := auth.HashPassword(pwd)
	if err != nil {
		return false, err
	}
	rows, err := s.DB.UpdateStudentPasswordByStuID(stuID, hash)
	if err != nil {
		return false, err
	}
	return rows == 1, nil
}

func (s *Service) GetAllLanguages() ([]string, error) {
	return s.DB.GetAllLanguages()
}

func (s *Service) AddLanguage(language string) error {
	return s.DB.AddLanguage(language)
}

func (s *Service) DeleteLanguage(language string) error {
	return s.DB.DeleteLanguage(language)
}

func (s *Service) GetAllCertificates() ([]string, error) {
	return s.DB.GetAllCertificates()
}

func (s *Service) AddCertificate(certificate string) error {
	return s.DB.AddCertificate(certificate)
}

func (s *Service) DeleteCertificate(certificate string) error {
	return s.DB.DeleteCertificate(certificate)
}

// ---------- Excel 辅助 ----------

func readExcelRows(reader io.Reader) ([][]string, error) {
	file, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("Excel文件解析失败: %w", err)
	}
	defer file.Close()
	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel文件解析失败: 没有工作表")
	}
	rows, err := file.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("Excel文件解析失败: %w", err)
	}
	return rows, nil
}

func rowEmpty(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func cellString(row []string, index int) string {
	if index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func cellInt(row []string, index int) (*int, error) {
	value := cellString(row, index)
	if value == "" {
		return nil, nil
	}
	value = strings.TrimSuffix(value, ".0")
	n, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func mustInt(row []string, index int) *int {
	n, _ := cellInt(row, index)
	return n
}
