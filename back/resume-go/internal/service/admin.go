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

// ==================================================================
// 管理端业务层：Excel 批量导入、批量更新、批量删除、字典维护
// ==================================================================
// 【与其它层的关系】
//   - api：接收 multipart 上传的 Excel 文件流，或 JSON 数组形式的批量数据；
//   - data：单条/批量 SQL，导入和删除使用事务；
//   - auth：为新建账号生成 bcrypt 初始密码；
//   - cache/mq/rpc：管理端操作不涉及，保证以数据库结果为准。
//
// 【本文件的核心取舍：部分成功】
// 批量导入/更新时，一行数据有问题不应该让整个文件都失败——
// 管理员更希望看到「成功 98 条、失败 2 条，失败原因是第 12 行学号重复」，
// 修正后只重传失败的部分即可。
// 因此这里采用「逐条处理 + 失败计数 + 收集错误明细」的模式：
//   - 行级问题（必填缺失、重复、记录不存在）：跳过该行，FailureCount++，继续下一行；
//   - 数据库级错误（连接断开、SQL 失败）：无法保证剩余数据正确，直接返回 error，
//     由 defer Rollback 撤销整批，避免留下不可预期的半成品。
//
// 【对应简历能力】Excel 批量导入/批量更新、事务与失败明细设计。
// ==================================================================

// ImportStudents 从 Excel 批量导入学生账号与档案。
//
// 处理规则：
//   - 只读第一个工作表，第一行视为表头跳过；
//   - 整行为空则跳过（Excel 末尾常带空行，不能算失败）；
//   - 学号必填且必须能解析成整数，否则记一条失败明细；
//   - 学号已存在则跳过并记录原因，不覆盖已有档案（导入不应变成“静默改数据”）；
//   - 初始密码就是学号，经 bcrypt 哈希后写入；
//   - user 与 student 两条记录在同一个事务里插入，保证不会出现“有账号没档案”。
//
// 行号使用 Excel 的真实行号（i+1），管理员能直接定位到文件中的哪一行。
// 注意：重复检查走的是普通连接而非 tx，判断的是「导入开始前」库里的状态；
// 同一文件内部的重复这里没有额外去重，依赖数据库约束或后续数据检查兜底。
//
// 返回值：SuccessCount/FailureCount/Errors；只有数据库级错误才返回 error。
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
		// 空行直接跳过：Excel 的已用区域经常包含格式残留的空行，
		// 把空行算成失败会让导入结果看起来“错了很多”。
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

// ImportCompanies 从 Excel 批量导入企业账号与档案。
//
// 与企业手工创建保持一致：企业名称必填且唯一，初始密码为企业名称（bcrypt 哈希）。
// 其余列（地区）缺失时留空，不阻断导入。
// 失败处理策略与 ImportStudents 相同：行级问题跳过并计数，数据库错误整批回滚。
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

// DeleteStudents 批量删除学生。
//
// 两阶段处理：
//  1. 先逐个确认学生存在，收集有效学生 ID 及其关联的 userID；
//     不存在的 ID 不算严重错误，计入失败明细后继续，保证批量操作尽量完成；
//  2. 对有效 ID 在一个事务里删除学生档案和对应账号。
//
// 为什么要同时删 user：只删 student 会留下能登录却没有档案的“幽灵账号”。
// 若一个有效 ID 都没有，直接返回结果，避免开启无意义的事务。
// SuccessCount 取实际删除的学生行数。
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

// DeleteCompanies 批量删除企业。
// 逻辑与 DeleteStudents 对称：先筛选出真实存在的企业，再在同一事务里
// 删除企业档案和登录账号，避免留下孤儿账号。
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

// UpdateStudents 批量更新学生基本信息。
//
// 逐条处理的细节：
//   - Normalize 兼容驼峰/下划线两种字段写法；
//   - ID 缺失或学生不存在：记失败明细并跳过，不影响其他行；
//   - RowsAffected 为 0 也记为失败：说明没有匹配到行或值完全没变化，
//     管理员需要知道这一条“没有生效”；
//   - 所有成功行共用一个事务，最后统一 Commit。
//
// 这是「部分成功」的典型实现：单行业务失败不回滚其他行，
// 但数据库级错误仍然整批回滚，保证不会只更新一半字段。
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

// UpdateCompanies 批量更新企业信息。
// 与 UpdateStudents 使用同一套「逐条校验 + 失败计数 + 统一事务提交」策略，
// 便于管理端复用同一份结果展示组件。
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

// AdminSearchStudents 按姓名分页查询学生（管理端）。
// page/pageSize 非法时回落到 1/10，避免 OFFSET 为负或一次拉取过多数据；
// 返回 map 是为了与管理端前端约定的 { list, pagination } 结构保持一致。
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

// AdminSearchCompanies 按名称分页查询企业（管理端）。
// 分页兜底与返回结构同 AdminSearchStudents。
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

// UpdateStudentPassword 管理员重置学生密码。
//
// 与用户自助改密不同：这里不需要原密码，属于管理权限操作。
// 新密码同样必须经 bcrypt 哈希后再入库，绝不能存明文。
// rows == 1 才视为成功：0 表示学号不存在（或没有对应账号），
// 返回 bool 让 api 层可以给出明确提示。
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

// GetAllLanguages 返回语言字典，供学生填写语言能力时下拉选择。
func (s *Service) GetAllLanguages() ([]string, error) {
	return s.DB.GetAllLanguages()
}

// AddLanguage 新增语言字典项。
// 唯一性由数据库约束保证；重复添加会返回底层错误，由 api 层统一处理。
func (s *Service) AddLanguage(language string) error {
	return s.DB.AddLanguage(language)
}

// DeleteLanguage 删除语言字典项。
// 只删字典，不影响学生档案里已经填写的语言记录。
func (s *Service) DeleteLanguage(language string) error {
	return s.DB.DeleteLanguage(language)
}

// GetAllCertificates 返回证书字典，供学生填写证书时下拉选择。
func (s *Service) GetAllCertificates() ([]string, error) {
	return s.DB.GetAllCertificates()
}

// AddCertificate 新增证书字典项。
func (s *Service) AddCertificate(certificate string) error {
	return s.DB.AddCertificate(certificate)
}

// DeleteCertificate 删除证书字典项。
func (s *Service) DeleteCertificate(certificate string) error {
	return s.DB.DeleteCertificate(certificate)
}

// ---------- Excel 辅助 ----------

// readExcelRows 把上传的文件流解析成二维字符串表。
// 只取第一个工作表，和模板约定一致；
// 所有错误都补上「Excel文件解析失败」前缀，方便前端直接展示给管理员。
// defer file.Close() 及时释放 excelize 内部资源，避免大文件占用内存。
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

// rowEmpty 判断整行是否全为空白（含只填空格的行）。
func rowEmpty(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

// cellString 安全读取某一列并去掉首尾空格。
// excelize 返回的行长度不固定，短行直接越界访问会 panic，
// 所以统一做边界检查，缺列按空字符串处理。
func cellString(row []string, index int) string {
	if index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

// cellInt 读取整数列，返回指针以区分「没填」和「填了 0」：
//   - 空单元格返回 (nil, nil)，表示调用方没提供这个字段；
//   - 形如 "2024.0" 的值会先去掉 ".0"——Excel 常把整数存成浮点文本；
//   - 确实不是数字时返回 error，由调用方决定记为失败还是忽略。
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

// mustInt 是 cellInt 的「宽松版」：解析失败时返回 nil 而不是 error。
// 用于年龄、毕业年份这类非关键列——填错就当作没填，不因此让整行导入失败；
// 真正必填的列（如学号）必须用 cellInt 并显式判断 error。
func mustInt(row []string, index int) *int {
	n, _ := cellInt(row, index)
	return n
}
