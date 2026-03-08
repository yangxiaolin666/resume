package com.resume.student.service;

import com.resume.model.dto.StudentCompleteInfoDTO;
import com.resume.model.dto.StudentDetailDTO;
import com.resume.model.dto.ResumeUploadVO;
import com.resume.model.entity.StudentStatisticsVO;
import org.springframework.web.multipart.MultipartFile;

/**
 * 学生信息服务接口
 */
public interface StudentInfoService {
    
    /**
     * 完善学生信息
     * @param userId 用户ID
     * @param infoDTO 完善信息DTO
     * @return 更新后的学生详细信息
     */
    StudentDetailDTO completeInfo(Integer userId, StudentCompleteInfoDTO infoDTO);
    
    /**
     * 上传简历
     * @param userId 用户ID
     * @param file 简历文件
     * @return 简历上传结果
     */
    ResumeUploadVO uploadResume(Integer userId, MultipartFile file);

    //获取所有学生统计信息
    StudentStatisticsVO getAllStudentStatistics();

    //根据原密码修改密码
    void updatePwd(Integer studentId, String newPwd);

    //查询学生密码
    String queryPwd(Integer userId);
}
