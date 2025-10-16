package com.resume.company.mapper;

import com.resume.model.entity.Student;
import com.resume.model.StudentQueryDTO;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface SearchMapper {
    
    /**
     * 根据条件查询学生列表
     * @param queryDTO 查询条件
     * @return 学生列表
     */
    List<Student> selectStudentsByCondition(@Param("query") StudentQueryDTO queryDTO);
}