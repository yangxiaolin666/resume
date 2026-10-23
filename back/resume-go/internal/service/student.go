package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"resume-backend/internal/auth"
	"resume-backend/internal/model"
)

const maxResumeSize = 10 * 1024 * 1024

var allowedResumeExt = map[string]bool{".pdf": true, ".doc": true, ".docx": true}

func (s *Service) findStudent(userID int) (*model.Student, error) {
	user, err := s.DB.FindUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, model.NewAppError(404, "用户不存在")
	}
	// 优先按 user_id 关联学生，兼容非学号形式的登录名（例如 x1）。
	student, err := s.DB.FindStudentByUserID(userID)
	if err != nil {
		return nil, err
	}
	if student != nil {
		return student, nil
	}
	stuID, err := strconv.Atoi(strings.TrimSpace(user.Username))
	if err != nil {
		return nil, model.NewAppError(400, "用户名格式错误，无法转换为学号")
	}
	student, err = s.DB.FindStudentByStuID(stuID)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, model.NewAppError(404, "学生信息不存在")
	}
	return student, nil
}

func (s *Service) CompleteInfo(userID int, info model.StudentCompleteInfo) (*model.StudentDetailDTO, error) {
	student, err := s.findStudent(userID)
	if err != nil {
		return nil, err
	}

	if info.Age != nil && (*info.Age < 16 || *info.Age > 60) {
		return nil, model.NewAppError(42201, "年龄超出合理范围(16-60岁)")
	}
	if info.GraduationYear != nil {
		year := time.Now().Year()
		if *info.GraduationYear < year-50 || *info.GraduationYear > year+10 {
			return nil, model.NewAppError(42202,
				fmt.Sprintf("毕业年份不合法，应在%d-%d范围内", year-50, year+10))
		}
	}

	if info.Age != nil {
		student.Age = info.Age
	}
	if info.Gender != nil {
		student.Gender = *info.Gender
	}
	if info.Nation != nil {
		student.Nation = *info.Nation
	}
	if info.Major != nil {
		student.Major = *info.Major
	}
	if info.Degree != nil {
		student.Degree = *info.Degree
	}
	if info.GraduationYear != nil {
		student.GraduationYear = info.GraduationYear
	}
	if info.Internship != nil {
		student.Internship = info.Internship
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := s.DB.UpdateStudentBasicTx(tx, student); err != nil {
		return nil, err
	}

	studentID := student.ID

	if len(info.Languages) > 0 {
		for _, lang := range info.Languages {
			if strings.TrimSpace(lang.Language) == "" {
				return nil, model.NewAppError(400, "语言名称不能为空")
			}
			if strings.TrimSpace(lang.Level) == "" {
				return nil, model.NewAppError(400, "语言等级不能为空")
			}
		}
		if err := s.DB.DeleteLanguagesTx(tx, studentID); err != nil {
			return nil, err
		}
		if err := s.DB.InsertLanguagesTx(tx, studentID, info.Languages); err != nil {
			return nil, err
		}
	}

	regions := []model.RegionDTO{}
	for _, city := range info.Cities {
		if strings.TrimSpace(city) == "" {
			return nil, model.NewAppError(400, "城市名称不能为空")
		}
		c := city
		regions = append(regions, model.RegionDTO{City: &c})
	}
	for _, province := range info.Provinces {
		if strings.TrimSpace(province) == "" {
			return nil, model.NewAppError(400, "省份名称不能为空")
		}
		p := province
		regions = append(regions, model.RegionDTO{Province: &p})
	}
	if len(regions) > 0 {
		if err := s.DB.DeleteRegionsTx(tx, studentID); err != nil {
			return nil, err
		}
		if err := s.DB.InsertRegionsTx(tx, studentID, regions); err != nil {
			return nil, err
		}
	}

	if len(info.Certificates) > 0 {
		for _, cert := range info.Certificates {
			if strings.TrimSpace(cert.CertName) == "" {
				return nil, model.NewAppError(400, "证书名称不能为空")
			}
		}
		if err := s.DB.DeleteCertificatesTx(tx, studentID); err != nil {
			return nil, err
		}
		if err := s.DB.InsertCertificatesTx(tx, studentID, info.Certificates); err != nil {
			return nil, err
		}
	}

	if len(info.Internships) > 0 {
		for _, intern := range info.Internships {
			if strings.TrimSpace(intern.Company) == "" {
				return nil, model.NewAppError(400, "实习公司名称不能为空")
			}
		}
		if err := s.DB.DeleteInternshipsTx(tx, studentID); err != nil {
			return nil, err
		}
		if err := s.DB.InsertInternshipsTx(tx, studentID, info.Internships); err != nil {
			return nil, err
		}
	}

	if len(info.Specialities) > 0 {
		for _, speciality := range info.Specialities {
			if strings.TrimSpace(speciality) == "" {
				return nil, model.NewAppError(400, "专业技能名称不能为空")
			}
		}
		if err := s.DB.DeleteSpecialitiesTx(tx, studentID); err != nil {
			return nil, err
		}
		if err := s.DB.InsertSpecialitiesTx(tx, studentID, info.Specialities); err != nil {
			return nil, err
		}
	}

	if len(info.Jobs) > 0 {
		for _, job := range info.Jobs {
			if strings.TrimSpace(job) == "" {
				return nil, model.NewAppError(400, "求职岗位名称不能为空")
			}
		}
		if err := s.DB.DeleteJobsTx(tx, studentID); err != nil {
			return nil, err
		}
		if err := s.DB.InsertJobsTx(tx, studentID, info.Jobs); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.buildStudentDetail(student)
}

func (s *Service) buildStudentDetail(student *model.Student) (*model.StudentDetailDTO, error) {
	detail := &model.StudentDetailDTO{
		ID:             student.ID,
		StuID:          student.StuID,
		UserID:         student.UserID,
		Name:           student.Name,
		Age:            student.Age,
		Gender:         student.Gender,
		Nation:         student.Nation,
		Major:          student.Major,
		Degree:         student.Degree,
		GraduationYear: student.GraduationYear,
		Internship:     student.Internship,
	}

	languages, err := s.DB.FindLanguagesByStudentID(student.ID)
	if err != nil {
		return nil, err
	}
	if len(languages) > 0 {
		detail.Languages = languages
	}

	certificates, err := s.DB.FindCertificatesByStudentID(student.ID)
	if err != nil {
		return nil, err
	}
	if len(certificates) > 0 {
		detail.Certificates = certificates
	}

	internships, err := s.DB.FindInternshipsByStudentID(student.ID)
	if err != nil {
		return nil, err
	}
	if len(internships) > 0 {
		detail.Internships = internships
	}

	jobs, err := s.DB.FindJobsByStudentID(student.ID)
	if err != nil {
		return nil, err
	}
	if len(jobs) > 0 {
		detail.Jobs = jobs
	}

	specialities, err := s.DB.FindSpecialitiesByStudentID(student.ID)
	if err != nil {
		return nil, err
	}
	if len(specialities) > 0 {
		detail.Specialities = specialities
	}

	regions, err := s.DB.FindRegionsByStudentID(student.ID)
	if err != nil {
		return nil, err
	}
	if len(regions) > 0 {
		provinces := make([]string, 0, len(regions))
		cities := make([]string, 0, len(regions))
		for _, r := range regions {
			if r.Province != nil && *r.Province != "" {
				provinces = append(provinces, *r.Province)
			}
			if r.City != nil && *r.City != "" {
				cities = append(cities, *r.City)
			}
		}
		region := &model.RegionDTO{}
		if joined := joinDistinct(provinces); joined != "" {
			region.Province = &joined
		}
		if joined := joinDistinct(cities); joined != "" {
			region.City = &joined
		}
		detail.Region = region
	}

	resume, err := s.DB.FindResumeByStudentID(student.ID)
	if err != nil {
		return nil, err
	}
	if resume != "" {
		detail.Resume = &model.ResumeDTO{Resume: resume}
	}
	return detail, nil
}

func joinDistinct(values []string) string {
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))
	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		unique = append(unique, v)
	}
	return strings.Join(unique, ",")
}

func (s *Service) UploadResume(userID int, filename string, size int64, reader io.Reader) (*model.ResumeUploadVO, error) {
	student, err := s.findStudent(userID)
	if err != nil {
		return nil, err
	}
	if filename == "" {
		return nil, model.NewAppError(400, "文件名无效")
	}
	if size > maxResumeSize {
		return nil, model.NewAppError(41301, "文件大小超限，最大支持10MB")
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedResumeExt[ext] {
		return nil, model.NewAppError(41501, "文件格式不支持，仅支持PDF、DOC、DOCX格式")
	}

	random := make([]byte, 4)
	if _, err := rand.Read(random); err != nil {
		return nil, err
	}
	uniqueName := fmt.Sprintf("%d_%s_%s%s", student.StuID,
		time.Now().Format("20060102150405"), hex.EncodeToString(random), ext)

	if err := os.MkdirAll(s.UploadDir, 0o755); err != nil {
		return nil, model.NewAppError(500, "创建上传目录失败")
	}
	filePath := filepath.Join(s.UploadDir, uniqueName)
	file, err := os.Create(filePath)
	if err != nil {
		return nil, model.NewAppError(500, "文件保存失败，"+err.Error())
	}
	if _, err := io.Copy(file, reader); err != nil {
		file.Close()
		return nil, model.NewAppError(500, "文件保存失败，"+err.Error())
	}
	file.Close()

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	resumeID, err := s.DB.InsertResumeTx(tx, student.ID, filePath)
	if err != nil {
		return nil, model.NewAppError(500, "保存简历记录失败")
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &model.ResumeUploadVO{
		ResumeID:   resumeID,
		ResumeName: uniqueName,
		ResumeURL:  "/api/resume/download/" + uniqueName,
	}, nil
}

func (s *Service) StudentStatistics() (*model.StudentStatistics, error) {
	return s.DB.StudentStatistics()
}

func (s *Service) UpdatePwd(userID int, oldPwd, newPwd, rePwd string) error {
	stored, err := s.DB.FindUserByID(userID)
	if err != nil {
		return err
	}
	hash := ""
	if stored != nil {
		hash = stored.Password
	}
	if hash == "" || !auth.CheckPassword(oldPwd, hash) {
		return model.NewAppError(400, "原密码不正确")
	}
	if auth.CheckPassword(newPwd, hash) {
		return model.NewAppError(400, "新密码不能与原密码一样")
	}
	if rePwd != newPwd {
		return model.NewAppError(400, "两次填写的新密码不一样")
	}
	newHash, err := auth.HashPassword(newPwd)
	if err != nil {
		return err
	}
	return s.DB.UpdateUserPassword(userID, newHash)
}
