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

// ==================================================================
// 学生档案业务层
// ==================================================================
// 负责学生「基本资料 + 多张子表（语言/地区/证书/实习/技能/求职意向）+ 简历文件 + 改密」。
//
// 【与其它层的关系】
//   - api：接收完整的 StudentCompleteInfo（字段都是指针，支持局部更新）；
//   - data：student 主表与 stu_* 子表的读写，以及事务对象 *sql.Tx；
//   - cache/mq/rpc：本文件基本不用，档案读写要求强一致，直接以 MySQL 为准；
//   - auth：改密时用 bcrypt 重新哈希。
//
// 【对应简历能力】复杂表单的多表写入与事务回滚、简历文件上传。
// ==================================================================

const maxResumeSize = 10 * 1024 * 1024

var allowedResumeExt = map[string]bool{".pdf": true, ".doc": true, ".docx": true}

// findStudent 把「登录用户」定位到「学生档案」。
//
// 查找顺序：
//  1. 先确认用户本身存在，否则 404；
//  2. 优先按 user_id 关联学生——这是新数据的标准关系，也兼容登录名不是学号的情况；
//  3. 老数据可能只按学号关联，于是退回「把用户名当学号」再查一次。
//
// 用户名不是纯数字时无法转学号，返回 400 并说明原因；
// 两种方式都找不到则返回 404「学生信息不存在」。
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

// CompleteInfo 保存学生完整信息。
//
// 【局部更新语义】
// 入参字段全部是指针：nil 表示「这次没传、不要动」，非 nil 才覆盖。
// 这样同一个接口既能支持「只改年龄」，也能支持「整页保存」，
// 避免前端漏传某个字段就把它清空。
//
// 【校验规则（为什么在写库前做）】
//   - 年龄 16-60：明显超出常理的数据多半是误填或攻击，提前拒绝，别污染档案；
//   - 毕业年份 [今年-50, 今年+10]：给往届和应届预留合理区间，
//     既能挡住 1900/2999 这类脏值，也不会误伤正常用户。
//
// 先校验再开事务，可以避免「开了事务才发现参数错」的无谓开销。
//
// 【多表写入与事务回滚】
// 一次保存涉及 1 张主表 + 6 张子表。所有写操作共用同一个 tx：
//   - 任何一步失败都直接 return，defer tx.Rollback() 会把前面已执行的
//     UPDATE/DELETE/INSERT 全部撤销，数据库不会留下“主表改了、子表没改”的半成品；
//   - 只有全部成功才 Commit，保证前端看到的是「要么全保存、要么全没保存」。
//
// 子表采用「先删后插」的替换策略：整块数据一起换，避免逐条 diff 的复杂逻辑。
// 注意外层有 len(...) > 0 判断：空数组代表「本次没提交这一块」，
// 不能理解成「清空」，否则前端只改基本信息就会误删所有证书/实习。
//
// 子表字段的非空校验放在各自写入之前；即使校验失败，事务也会因 rollback 而不生效。
// 全部提交后重新聚合查询一遍，返回与数据库完全一致的详情，避免返回内存里的旧值。
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
	// 统一用 defer 兜底回滚：只要没走到 Commit，中途任何 return 都不会留下部分写入。
	defer tx.Rollback()

	if _, err := s.DB.UpdateStudentBasicTx(tx, student); err != nil {
		return nil, err
	}

	studentID := student.ID

	// 以下每个子表都是同一套模式：校验 -> DELETE 旧数据 -> INSERT 新数据。
	// 它们共享上面的事务，任何一步失败都会整体回滚。
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

// buildStudentDetail 把主表 + 各子表聚合成前端需要的完整详情。
//
// 这里固定执行 7 次查询（主表数据已由调用方传入），而不是写一条巨型 JOIN：
// JOIN 多张一对多表会产生笛卡尔积、需要去重，代码和 SQL 都更难维护；
// 单条学生详情属于低频操作，固定次数的简单查询完全够用。
// 任一查询出错都向上返回，由 api 层转成 5xx；
// 某类子数据为空时保持 nil，前端可据此隐藏对应板块。
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

// joinDistinct 去重后用逗号拼接。
// 数据库里地区是「一行一个省份/城市」，前端 DTO 只接受一个字符串，
// 所以这里做去重合并，保证同一地区不会因为多行重复出现。
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

// LatestResume 返回当前学生最新上传的简历，供「聊天发送简历」使用。
//
// 数据库里存的是磁盘路径，这里统一转换成 /uploads/resume/xxx 的静态地址，
// 避免把服务器目录结构暴露给前端。
func (s *Service) LatestResume(userID int) (*model.ResumeInfo, error) {
	student, err := s.findStudent(userID)
	if err != nil {
		return nil, err
	}
	stored, err := s.DB.FindResumeByStudentID(student.ID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(stored) == "" {
		return nil, model.NewAppError(404, "你还没有上传简历，请先到个人中心上传")
	}
	return &model.ResumeInfo{
		URL:  publicResumeURL(stored),
		Name: filepath.Base(stored),
	}, nil
}

// UploadResume 保存简历文件并记录到数据库。
//
// 规则：
//   - 必须是已建档的学生；
//   - 文件名不能为空、大小不超过 10MB、扩展名只允许 PDF/DOC/DOCX；
//   - 文件名由后端重新生成（学号 + 时间 + 随机数 + 扩展名），
//     不信任原始文件名，既避免重名覆盖，也防止路径穿越。
//
// 顺序是「先落盘、再写库」：文件写成功但数据库失败时，会留下一个没有被引用的孤儿文件，
// 这比「数据库有记录、文件却不存在」要好——后者会让用户下载到 404。
// 孤儿文件可以由定时任务清理，属于可接受的取舍。
// 返回的 ResumeURL 指向下载接口，而不是磁盘路径，避免暴露服务器目录结构。
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

// StudentStatistics 返回学生维度的统计数据，供首页/大屏展示。
// 纯读取，直接透传 data 层结果。
func (s *Service) StudentStatistics() (*model.StudentStatistics, error) {
	return s.DB.StudentStatistics()
}

// UpdatePwd 修改当前用户密码（学生、企业、管理员共用的账号能力）。
//
// 校验顺序与原因：
//  1. 先取当前哈希；用户不存在或哈希为空时，同样提示「原密码不正确」，
//     不透露账号是否存在；
//  2. 原密码必须匹配，否则任何人都能在已登录状态下直接改密；
//  3. 新密码不能与原密码相同，避免“改了等于没改”；
//  4. 两次输入必须一致，防止用户手误设置成自己不知道的密码。
//
// 通过后重新用 bcrypt 生成哈希（自动换新盐）再写库。
// 注意：本项目改密后不会让已签发的 JWT 立即失效，
// 因为 JWT 是无状态的；若要强制下线，需要额外维护 token 版本号/黑名单。
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
