package com.resume.student.service.impl;

import com.resume.model.dto.*;
import com.resume.model.entity.*;
import com.resume.student.mapper.StudentInfoMapper;
import com.resume.student.service.StudentInfoService;
import com.resume.user.mapper.UserMapper;
import com.resume.user.service.impl.UserServiceImpl;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.multipart.MultipartFile;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

/**
 * 学生信息服务实现类
 */
@Service
public class StudentInfoServiceImpl implements StudentInfoService {
    
    @Autowired
    private StudentInfoMapper studentInfoMapper;
    
    @Autowired
    private UserMapper userMapper;
    
    // 简历文件存储路径（可在application.yml中配置）
    @Value("${file.upload.resume-path:uploads/resume/}")
    private String resumeUploadPath;
    
    // 简历文件最大大小（10MB）
    private static final long MAX_FILE_SIZE = 10 * 1024 * 1024;
    
    // 支持的文件格式
    private static final String[] ALLOWED_EXTENSIONS = {".pdf", ".doc", ".docx"};

    @Override
    @Transactional(rollbackFor = Exception.class)
    public StudentDetailDTO completeInfo(Integer userId, StudentCompleteInfoDTO infoDTO) {
        try {
            // 1. 从 token 中获取 userId，查询 user 表
            User user = userMapper.findByUserId(userId);
            if (user == null) {
                throw new RuntimeException("404:用户不存在");
            }
            
            // 2. 获取 username，将其作为学号查询 student 表
            Integer stuId;
            try {
                stuId = Integer.parseInt(user.getUsername());
            } catch (NumberFormatException e) {
                throw new RuntimeException("400:用户名格式错误，无法转换为学号");
            }
            
            Student student = studentInfoMapper.findByStuId(stuId);
            if (student == null) {
                throw new RuntimeException("404:学生信息不存在");
            }
            
            // 验证年龄合理性
            if (infoDTO.getAge() != null && (infoDTO.getAge() < 16 || infoDTO.getAge() > 60)) {
                throw new RuntimeException("42201:年龄超出合理范围(16-60岁)");
            }
            
            // 验证毕业年份
            if (infoDTO.getGraduationYear() != null) {
                int currentYear = LocalDateTime.now().getYear();
                if (infoDTO.getGraduationYear() < currentYear - 50 || infoDTO.getGraduationYear() > currentYear + 10) {
                    throw new RuntimeException("42202:毕业年份不合法，应在" + (currentYear - 50) + "-" + (currentYear + 10) + "范围内");
                }
            }
            
            // 更新学生基本信息（不允许修改学号和姓名）
            if (infoDTO.getAge() != null) student.setAge(infoDTO.getAge());
            if (infoDTO.getGender() != null) student.setGender(infoDTO.getGender());
            if (infoDTO.getNation() != null) student.setNation(infoDTO.getNation());
            if (infoDTO.getMajor() != null) student.setMajor(infoDTO.getMajor());
            if (infoDTO.getDegree() != null) student.setDegree(infoDTO.getDegree());
            if (infoDTO.getGraduationYear() != null) student.setGraduationYear(infoDTO.getGraduationYear());
            if (infoDTO.getInternship() != null) student.setInternship(infoDTO.getInternship());
            
            // 更新学生基本信息
            int updateCount = studentInfoMapper.updateStudentInfo(student);
            if (updateCount == 0) {
                throw new RuntimeException("500:更新学生基本信息失败");
            }
            
            // 使用 student 的 id（主键）作为关联表的 stu_id
            Integer studentId = student.getId();
            
            // 更新语言信息
            if (infoDTO.getLanguages() != null && !infoDTO.getLanguages().isEmpty()) {
                // 先删除旧数据
                studentInfoMapper.deleteLanguagesByStudentId(studentId);
                
                // 构建新数据
                List<StuLanguage> languages = new ArrayList<>();
                for (StudentCompleteInfoDTO.LanguageInfo langInfo : infoDTO.getLanguages()) {
                    if (langInfo.getLanguage() == null || langInfo.getLanguage().trim().isEmpty()) {
                        throw new RuntimeException("400:语言名称不能为空");
                    }
                    if (langInfo.getLevel() == null || langInfo.getLevel().trim().isEmpty()) {
                        throw new RuntimeException("400:语言等级不能为空");
                    }
                    StuLanguage language = new StuLanguage();
                    language.setStuId(studentId.longValue());
                    language.setLanguage(langInfo.getLanguage());
                    language.setLevel(langInfo.getLevel());
                    languages.add(language);
                }
                
                // 批量插入新数据
                if (!languages.isEmpty()) {
                    studentInfoMapper.batchInsertLanguages(studentId, languages);
                }
            }
            
            // 更新地区信息
            List<StuRegion> regions = new ArrayList<>();
            if (infoDTO.getCities() != null && !infoDTO.getCities().isEmpty()) {
                for (String city : infoDTO.getCities()) {
                    if (city == null || city.trim().isEmpty()) {
                        throw new RuntimeException("400:城市名称不能为空");
                    }
                    StuRegion region = new StuRegion();
                    region.setStuId(studentId);
                    region.setCity(city);
                    regions.add(region);
                }
            }
            if (infoDTO.getProvinces() != null && !infoDTO.getProvinces().isEmpty()) {
                for (String province : infoDTO.getProvinces()) {
                    if (province == null || province.trim().isEmpty()) {
                        throw new RuntimeException("400:省份名称不能为空");
                    }
                    StuRegion region = new StuRegion();
                    region.setStuId(studentId);
                    region.setProvince(province);
                    regions.add(region);
                }
            }
            if (!regions.isEmpty()) {
                studentInfoMapper.deleteRegionsByStudentId(studentId);
                studentInfoMapper.batchInsertRegions(studentId, regions);
            }
            
            // 更新证书信息
            if (infoDTO.getCertificates() != null && !infoDTO.getCertificates().isEmpty()) {
                studentInfoMapper.deleteCertificatesByStudentId(studentId);
                List<StuCertificate> certificates = new ArrayList<>();
                for (StudentCompleteInfoDTO.CertificateInfo certInfo : infoDTO.getCertificates()) {
                    if (certInfo.getCertName() == null || certInfo.getCertName().trim().isEmpty()) {
                        throw new RuntimeException("400:证书名称不能为空");
                    }
                    StuCertificate certificate = new StuCertificate();
                    certificate.setStuId(studentId.longValue());
                    certificate.setCertName(certInfo.getCertName());
                    certificate.setLevel(certInfo.getLevel());
                    certificates.add(certificate);
                }
                if (!certificates.isEmpty()) {
                    studentInfoMapper.batchInsertCertificates(studentId, certificates);
                }
            }
            
            // 更新实习信息
            if (infoDTO.getInternships() != null && !infoDTO.getInternships().isEmpty()) {
                studentInfoMapper.deleteInternshipsByStudentId(studentId);
                List<StuInternship> internships = new ArrayList<>();
                for (StudentCompleteInfoDTO.InternshipInfo internInfo : infoDTO.getInternships()) {
                    if (internInfo.getCompany() == null || internInfo.getCompany().trim().isEmpty()) {
                        throw new RuntimeException("400:实习公司名称不能为空");
                    }
                    StuInternship internship = new StuInternship();
                    internship.setStuId(studentId);
                    internship.setCompany(internInfo.getCompany());
                    internship.setExperience(internInfo.getExperience());
                    internships.add(internship);
                }
                if (!internships.isEmpty()) {
                    studentInfoMapper.batchInsertInternships(studentId, internships);
                }
            }
            
            // 更新专业技能信息
            if (infoDTO.getSpecialities() != null && !infoDTO.getSpecialities().isEmpty()) {
                studentInfoMapper.deleteSpecialitiesByStudentId(studentId);
                List<StuSpeciality> specialities = new ArrayList<>();
                for (String speciality : infoDTO.getSpecialities()) {
                    if (speciality == null || speciality.trim().isEmpty()) {
                        throw new RuntimeException("400:专业技能名称不能为空");
                    }
                    StuSpeciality spec = new StuSpeciality();
                    spec.setStuId(studentId);
                    spec.setSpeciality(speciality);
                    specialities.add(spec);
                }
                if (!specialities.isEmpty()) {
                    studentInfoMapper.batchInsertSpecialities(studentId, specialities);
                }
            }
            
            // 更新求职意向信息
            if (infoDTO.getJobs() != null && !infoDTO.getJobs().isEmpty()) {
                studentInfoMapper.deleteJobsByStudentId(studentId);
                List<StuJob> jobs = new ArrayList<>();
                for (String job : infoDTO.getJobs()) {
                    if (job == null || job.trim().isEmpty()) {
                        throw new RuntimeException("400:求职岗位名称不能为空");
                    }
                    StuJob stuJob = new StuJob();
                    stuJob.setStuId(studentId);
                    stuJob.setPost(job);
                    jobs.add(stuJob);
                }
                if (!jobs.isEmpty()) {
                    studentInfoMapper.batchInsertJobs(studentId, jobs);
                }
            }
            
            // 事务成功，查询并返回学生详细信息
            StudentDetailDTO studentDetail = buildStudentDetail(student, studentId);
            return studentDetail;
            
        } catch (RuntimeException e) {
            // 运行时异常直接抛出，触发事务回滚
            throw e;
        } catch (Exception e) {
            // 其他异常包装后抛出，触发事务回滚
            throw new RuntimeException("500:完善学生信息失败，" + e.getMessage(), e);
        }
    }
    
    @Override
    @Transactional(rollbackFor = Exception.class)
    public ResumeUploadVO uploadResume(Integer userId, MultipartFile file) {
        try {
            // 1. 从 token 中获取 userId，查询 user 表
            User user = userMapper.findByUserId(userId);
            if (user == null) {
                throw new RuntimeException("404:用户不存在");
            }
            
            // 2. 获取 username，将其作为学号查询 student 表
            Integer stuId;
            try {
                stuId = Integer.parseInt(user.getUsername());
            } catch (NumberFormatException e) {
                throw new RuntimeException("400:用户名格式错误，无法转换为学号");
            }
            
            Student student = studentInfoMapper.findByStuId(stuId);
            if (student == null) {
                throw new RuntimeException("404:学生信息不存在");
            }
            
            // 校验文件
            if (file == null || file.isEmpty()) {
                throw new RuntimeException("400:请选择要上传的文件");
            }
            
            // 检查文件大小
            if (file.getSize() > MAX_FILE_SIZE) {
                throw new RuntimeException("41301:文件大小超限，最大支持10MB");
            }
            
            // 检查文件格式
            String originalFilename = file.getOriginalFilename();
            if (originalFilename == null || originalFilename.trim().isEmpty()) {
                throw new RuntimeException("400:文件名无效");
            }
            
            String extension = getFileExtension(originalFilename);
            if (!isAllowedExtension(extension)) {
                throw new RuntimeException("41501:文件格式不支持，仅支持PDF、DOC、DOCX格式");
            }
            
            // 生成唯一文件名
            String timestamp = LocalDateTime.now().format(DateTimeFormatter.ofPattern("yyyyMMddHHmmss"));
            String uniqueFileName = student.getStuId() + "_" + timestamp + "_" + UUID.randomUUID().toString().substring(0, 8) + extension;
            
            // 确保上传目录存在
            File uploadDir = new File(resumeUploadPath);
            if (!uploadDir.exists()) {
                boolean created = uploadDir.mkdirs();
                if (!created) {
                    throw new RuntimeException("500:创建上传目录失败");
                }
            }
            
            // 保存文件到本地
            Path filePath = Paths.get(resumeUploadPath, uniqueFileName);
            try {
                Files.copy(file.getInputStream(), filePath);
            } catch (IOException e) {
                throw new RuntimeException("500:文件保存失败，" + e.getMessage(), e);
            }
            
            // 保存简历记录到数据库
            StuResume stuResume = new StuResume();
            stuResume.setStuId(student.getId());
            stuResume.setResume(filePath.toString());
            
            int insertCount = studentInfoMapper.insertResume(stuResume);
            if (insertCount == 0) {
                throw new RuntimeException("500:保存简历记录失败");
            }
            
            // 构造返回结果
            ResumeUploadVO vo = new ResumeUploadVO();
            vo.setResumeId(stuResume.getId());
            vo.setResumeName(uniqueFileName);
            vo.setResumeUrl("/api/resume/download/" + uniqueFileName);
            
            // 事务成功，返回结果
            return vo;
            
        } catch (RuntimeException e) {
            // 运行时异常直接抛出，触发事务回滚
            throw e;
        } catch (Exception e) {
            // 其他异常包装后抛出，触发事务回滚
            throw new RuntimeException("500:上传简历失败，" + e.getMessage(), e);
        }
    }
    
    /**
     * 获取文件扩展名
     */
    private String getFileExtension(String filename) {
        int lastDotIndex = filename.lastIndexOf(".");
        if (lastDotIndex == -1) {
            return "";
        }
        return filename.substring(lastDotIndex).toLowerCase();
    }
    
    /**
     * 检查文件扩展名是否允许
     */
    private boolean isAllowedExtension(String extension) {
        for (String allowed : ALLOWED_EXTENSIONS) {
            if (allowed.equalsIgnoreCase(extension)) {
                return true;
            }
        }
        return false;
    }
    
    /**
     * 构建学生详细信息
     */
    private StudentDetailDTO buildStudentDetail(Student student, Integer studentId) {
        StudentDetailDTO detail = new StudentDetailDTO();
        
        // 基本信息
        detail.setId(student.getId());
        detail.setStuId(student.getStuId());
        detail.setName(student.getName());
        detail.setAge(student.getAge());
        detail.setGender(student.getGender());
        detail.setNation(student.getNation());
        detail.setMajor(student.getMajor());
        detail.setDegree(student.getDegree());
        detail.setGraduationYear(student.getGraduationYear());
        detail.setInternship(student.getInternship());
        
        // 查询并转换语言信息
        List<StuLanguage> languages = studentInfoMapper.findLanguagesByStudentId(studentId);
        if (languages != null && !languages.isEmpty()) {
            detail.setLanguages(languages.stream().map(lang -> {
                LanguageDTO dto = new LanguageDTO();
                dto.setLanguage(lang.getLanguage());
                dto.setLevel(lang.getLevel());
                dto.setCertNo(lang.getCertNo());
                dto.setIssueDate(lang.getIssueDate());
                return dto;
            }).collect(Collectors.toList()));
        }
        
        // 查询并转换证书信息
        List<StuCertificate> certificates = studentInfoMapper.findCertificatesByStudentId(studentId);
        if (certificates != null && !certificates.isEmpty()) {
            detail.setCertificates(certificates.stream().map(cert -> {
                CertificateDTO dto = new CertificateDTO();
                dto.setCertName(cert.getCertName());
                dto.setLevel(cert.getLevel());
                dto.setCertNo(cert.getCertNo());
                dto.setIssueDate(cert.getIssueDate());
                return dto;
            }).collect(Collectors.toList()));
        }
        
        // 查询并转换实习信息
        List<StuInternship> internships = studentInfoMapper.findInternshipsByStudentId(studentId);
        if (internships != null && !internships.isEmpty()) {
            detail.setInternships(internships.stream().map(intern -> {
                InternshipDTO dto = new InternshipDTO();
                dto.setCompany(intern.getCompany());
                dto.setExperience(intern.getExperience());
                return dto;
            }).collect(Collectors.toList()));
        }
        
        // 查询并转换求职意向
        List<StuJob> jobs = studentInfoMapper.findJobsByStudentId(studentId);
        if (jobs != null && !jobs.isEmpty()) {
            detail.setJobs(jobs.stream().map(job -> {
                JobDTO dto = new JobDTO();
                dto.setPost(job.getPost());
                return dto;
            }).collect(Collectors.toList()));
        }
        
        // 查询并转换专业技能
        List<StuSpeciality> specialities = studentInfoMapper.findSpecialitiesByStudentId(studentId);
        if (specialities != null && !specialities.isEmpty()) {
            detail.setSpecialities(specialities.stream().map(spec -> {
                SpecialityDTO dto = new SpecialityDTO();
                dto.setSpeciality(spec.getSpeciality());
                return dto;
            }).collect(Collectors.toList()));
        }
        
        // 查询地区信息（取第一条或合并）
        List<StuRegion> regions = studentInfoMapper.findRegionsByStudentId(studentId);
        if (regions != null && !regions.isEmpty()) {
            RegionDTO regionDTO = new RegionDTO();
            // 合并所有省份和城市
            String provinces = regions.stream()
                .map(StuRegion::getProvince)
                .filter(p -> p != null && !p.isEmpty())
                .distinct()
                .collect(Collectors.joining(","));
            String cities = regions.stream()
                .map(StuRegion::getCity)
                .filter(c -> c != null && !c.isEmpty())
                .distinct()
                .collect(Collectors.joining(","));
            regionDTO.setProvince(provinces.isEmpty() ? null : provinces);
            regionDTO.setCity(cities.isEmpty() ? null : cities);
            detail.setRegion(regionDTO);
        }
        
        // 查询简历信息
        String resumePath = studentInfoMapper.findResumeByStudentId(studentId);
        if (resumePath != null && !resumePath.isEmpty()) {
            ResumeDTO resumeDTO = new ResumeDTO();
            resumeDTO.setResume(resumePath);
            detail.setResume(resumeDTO);
        }
        
        return detail;
    }
}
