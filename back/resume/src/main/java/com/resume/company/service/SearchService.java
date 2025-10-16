package com.resume.company.service;

import com.github.pagehelper.PageInfo;
import com.resume.model.dto.StudentDetailDTO;
import com.resume.model.entity.Student;
import com.resume.model.dto.StudentQueryDTO;

public interface SearchService {
    
    /**
     * 根据条件查询学生列表
     * @param queryDTO 查询条件
     * @return 学生列表
     */
    PageInfo<Student> getStudentsByCondition(StudentQueryDTO queryDTO);
    
    /**
     * 根据条件查询学生详细信息列表
     * @param queryDTO 查询条件
     * @return 学生详细信息列表
     */
    PageInfo<StudentDetailDTO> getStudentDetailsByCondition(StudentQueryDTO queryDTO);
}