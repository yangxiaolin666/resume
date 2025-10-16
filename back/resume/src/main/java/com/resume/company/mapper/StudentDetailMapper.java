package com.resume.company.mapper;

import com.resume.model.dto.StudentDetailDTO;
import com.resume.model.dto.StudentQueryDTO;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import java.util.List;

@Mapper
public interface StudentDetailMapper {
    
    /**
     * 根据条件查询学生详细信息列表
     * @param queryDTO 查询条件
     * @return 学生详细信息列表
     */
    List<StudentDetailDTO> selectStudentDetailsByCondition(@Param("query") StudentQueryDTO queryDTO);
}