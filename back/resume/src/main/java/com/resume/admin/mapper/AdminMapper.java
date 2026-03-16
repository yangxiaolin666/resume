package com.resume.admin.mapper;

import com.resume.model.dto.LanguageDTO;
import com.resume.model.entity.Company;
import com.resume.model.entity.Student;
import com.resume.model.result.Result;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

/**
 * 管理员Mapper接口
 */
@Mapper
public interface AdminMapper {

    // ============ 学生相关操作 ============
    
    /**
     * 批量插入学生
     */
    int batchInsertStudents(@Param("students") List<Student> students);
    
    /**
     * 根据学号查询学生
     */
    Student selectStudentByStuId(@Param("stuId") Integer stuId);
    
    /**
     * 根据ID查询学生
     */
    Student selectStudentById(@Param("id") Integer id);
    
    /**
     * 批量删除学生
     */
    int batchDeleteStudents(@Param("ids") List<Integer> ids);
    
    /**
     * 更新学生信息
     */
    int updateStudent(Student student);
    
    /**
     * 分页查询学生（支持姓名模糊查询）
     */
    List<Student> selectStudentsByPage(@Param("name") String name);
    
    /**
     * 统计学生总数（支持姓名模糊查询）
     */
    int countStudents(@Param("name") String name);
    
    // ============ 企业相关操作 ============
    
    /**
     * 批量插入企业
     */
    int batchInsertCompanies(@Param("companies") List<Company> companies);
    
    /**
     * 根据企业名称查询企业
     */
    Company selectCompanyByName(@Param("comName") String comName);
    
    /**
     * 根据ID查询企业
     */
    Company selectCompanyById(@Param("id") Integer id);
    
    /**
     * 批量删除企业
     */
    int batchDeleteCompanies(@Param("ids") List<Integer> ids);
    
    /**
     * 更新企业信息
     */
    int updateCompany(Company company);
    
    /**
     * 分页查询企业（支持名称模糊查询）
     */
    List<Company> selectCompaniesByPage(@Param("comName") String comName);
    
    /**
     * 统计企业总数（支持名称模糊查询）
     */
    int countCompanies(@Param("comName") String comName);

    /**
     * 查询所有语言并返回
     */
    @Select("SELECT DISTINCT language_name FROM language")
    List<String> getAllLanguages();

    /**
     * 添加语言
     */
    @Select("INSERT INTO language (language_name) VALUES (#{language})")
    Result<?> addLanguage(String language);

    /**
     * 删除语言
     */
    @Select("DELETE FROM language WHERE language_name = #{language}")
    Result<?> deleteLanguage(String language);
}
