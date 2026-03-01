package com.resume.company.mapper;

import com.resume.model.dto.*;
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
    
    /**
     * 根据学生ID列表查询语言信息
     * @param studentIds 学生ID列表
     * @return 语言信息列表
     */
    List<LanguageDTO> selectLanguagesByStudentIds(@Param("studentIds") List<Integer> studentIds);
    
    /**
     * 根据学生ID列表查询证书信息
     * @param studentIds 学生ID列表
     * @return 证书信息列表
     */
    List<CertificateDTO> selectCertificatesByStudentIds(@Param("studentIds") List<Integer> studentIds);
    
    /**
     * 根据学生ID列表查询实习信息
     * @param studentIds 学生ID列表
     * @return 实习信息列表
     */
    List<InternshipDTO> selectInternshipsByStudentIds(@Param("studentIds") List<Integer> studentIds);
    
    /**
     * 根据学生ID列表查询求职信息
     * @param studentIds 学生ID列表
     * @return 求职信息列表
     */
    List<JobDTO> selectJobsByStudentIds(@Param("studentIds") List<Integer> studentIds);
    
    /**
     * 根据学生ID列表查询专业技能信息
     * @param studentIds 学生ID列表
     * @return 专业技能信息列表
     */
    List<SpecialityDTO> selectSpecialitiesByStudentIds(@Param("studentIds") List<Integer> studentIds);
}