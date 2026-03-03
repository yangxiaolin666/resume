package com.resume.student.mapper;

import com.resume.model.entity.*;
import org.apache.ibatis.annotations.*;

import java.util.List;

/**
 * 学生信息Mapper
 */
@Mapper
public interface StudentInfoMapper {
    
    /**
     * 根据用户ID查询学生信息
     */
    @Select("SELECT * FROM student WHERE user_id = #{userId}")
    Student findByUserId(Integer userId);
    
    /**
     * 根据学号查询学生信息
     */
    @Select("SELECT * FROM student WHERE stu_id = #{stuId}")
    Student findByStuId(Integer stuId);
    
    /**
     * 更新学生基本信息
     */
    @Update("UPDATE student SET age = #{age}, gender = #{gender}, nation = #{nation}, " +
            "major = #{major}, degree = #{degree}, graduation_year = #{graduationYear}, " +
            "internship = #{internship} WHERE id = #{id}")
    int updateStudentInfo(Student student);
    
    /**
     * 根据学生ID删除语言信息
     */
    @Delete("DELETE FROM stu_language WHERE stu_id = #{stuId}")
    void deleteLanguagesByStudentId(Integer stuId);
    
    /**
     * 批量插入语言信息
     */
    @Insert("<script>" +
            "INSERT INTO stu_language (stu_id, language, level) VALUES " +
            "<foreach collection='list' item='item' separator=','>" +
            "(#{stuId}, #{item.language}, #{item.level})" +
            "</foreach>" +
            "</script>")
    void batchInsertLanguages(@Param("stuId") Integer stuId, @Param("list") List<StuLanguage> languages);
    
    /**
     * 根据学生ID删除地区信息
     */
    @Delete("DELETE FROM stu_region WHERE stu_id = #{stuId}")
    void deleteRegionsByStudentId(Integer stuId);
    
    /**
     * 批量插入地区信息
     */
    @Insert("<script>" +
            "INSERT INTO stu_region (stu_id, province, city) VALUES " +
            "<foreach collection='list' item='item' separator=','>" +
            "(#{stuId}, #{item.province}, #{item.city})" +
            "</foreach>" +
            "</script>")
    void batchInsertRegions(@Param("stuId") Integer stuId, @Param("list") List<StuRegion> regions);
    
    /**
     * 根据学生ID删除证书信息
     */
    @Delete("DELETE FROM stu_certificate WHERE stu_id = #{stuId}")
    void deleteCertificatesByStudentId(Integer stuId);
    
    /**
     * 批量插入证书信息
     */
    @Insert("<script>" +
            "INSERT INTO stu_certificate (stu_id, cert_name, level) VALUES " +
            "<foreach collection='list' item='item' separator=','>" +
            "(#{stuId}, #{item.certName}, #{item.level})" +
            "</foreach>" +
            "</script>")
    void batchInsertCertificates(@Param("stuId") Integer stuId, @Param("list") List<StuCertificate> certificates);
    
    /**
     * 根据学生ID删除实习信息
     */
    @Delete("DELETE FROM stu_internship WHERE stu_id = #{stuId}")
    void deleteInternshipsByStudentId(Integer stuId);
    
    /**
     * 批量插入实习信息
     */
    @Insert("<script>" +
            "INSERT INTO stu_internship (stu_id, company, experience) VALUES " +
            "<foreach collection='list' item='item' separator=','>" +
            "(#{stuId}, #{item.company}, #{item.experience})" +
            "</foreach>" +
            "</script>")
    @Options(useGeneratedKeys = true, keyProperty = "list.id")
    void batchInsertInternships(@Param("stuId") Integer stuId, @Param("list") List<StuInternship> internships);
    
    /**
     * 根据学生ID删除专业技能信息
     */
    @Delete("DELETE FROM stu_speciality WHERE stu_id = #{stuId}")
    void deleteSpecialitiesByStudentId(Integer stuId);
    
    /**
     * 批量插入专业技能信息
     */
    @Insert("<script>" +
            "INSERT INTO stu_speciality (stu_id, speciality) VALUES " +
            "<foreach collection='list' item='item' separator=','>" +
            "(#{stuId}, #{item.speciality})" +
            "</foreach>" +
            "</script>")
    @Options(useGeneratedKeys = true, keyProperty = "list.id")
    void batchInsertSpecialities(@Param("stuId") Integer stuId, @Param("list") List<StuSpeciality> specialities);
    
    /**
     * 根据学生ID删除求职意向信息
     */
    @Delete("DELETE FROM stu_job WHERE stu_id = #{stuId}")
    void deleteJobsByStudentId(Integer stuId);
    
    /**
     * 批量插入求职意向信息
     */
    @Insert("<script>" +
            "INSERT INTO stu_job (stu_id, post) VALUES " +
            "<foreach collection='list' item='item' separator=','>" +
            "(#{stuId}, #{item.post})" +
            "</foreach>" +
            "</script>")
    void batchInsertJobs(@Param("stuId") Integer stuId, @Param("list") List<StuJob> jobs);
    
    /**
     * 插入简历记录
     */
    @Insert("INSERT INTO stu_resume (stu_id, resume) VALUES (#{stuId}, #{resume})")
    @Options(useGeneratedKeys = true, keyProperty = "id")
    int insertResume(StuResume stuResume);
    
    /**
     * 根据学生ID查询简历路径
     */
    @Select("SELECT resume FROM stu_resume WHERE stu_id = #{stuId} ORDER BY id DESC LIMIT 1")
    String findResumeByStudentId(Integer stuId);
    
    /**
     * 根据学生ID查询语言信息
     */
    @Select("SELECT language, level, cert_no AS certNo, issue_date AS issueDate FROM stu_language WHERE stu_id = #{stuId}")
    List<StuLanguage> findLanguagesByStudentId(@Param("stuId") Integer stuId);
    
    /**
     * 根据学生ID查询证书信息
     */
    @Select("SELECT cert_name AS certName, level, cert_no AS certNo, issue_date AS issueDate FROM stu_certificate WHERE stu_id = #{stuId}")
    List<StuCertificate> findCertificatesByStudentId(@Param("stuId") Integer stuId);
    
    /**
     * 根据学生ID查询实习信息
     */
    @Select("SELECT company, experience FROM stu_internship WHERE stu_id = #{stuId}")
    List<StuInternship> findInternshipsByStudentId(@Param("stuId") Integer stuId);
    
    /**
     * 根据学生ID查询求职意向
     */
    @Select("SELECT post FROM stu_job WHERE stu_id = #{stuId}")
    List<StuJob> findJobsByStudentId(@Param("stuId") Integer stuId);
    
    /**
     * 根据学生ID查询专业技能
     */
    @Select("SELECT speciality FROM stu_speciality WHERE stu_id = #{stuId}")
    List<StuSpeciality> findSpecialitiesByStudentId(@Param("stuId") Integer stuId);
    
    /**
     * 根据学生ID查询地区信息
     */
    @Select("SELECT province, city FROM stu_region WHERE stu_id = #{stuId}")
    List<StuRegion> findRegionsByStudentId(@Param("stuId") Integer stuId);


    // 语言统计
    @Select("SELECT COUNT(*) as value, language, level " +
            "FROM stu_language " +
            "GROUP BY language, level " +
            "ORDER BY value DESC ")
    List<StudentStatisticsVO.LanguageStat> selectLanguageStats();

    // 城市统计
    @Select("SELECT COUNT(*) as value, city as name " +
            "FROM stu_region " +
            "WHERE city IS NOT NULL " +
            "GROUP BY city " +
            "ORDER BY value DESC ")
    List<StudentStatisticsVO.CityStat> selectCityStats();

    // 实习统计
    @Select("SELECT COUNT(*) as value, company, experience " +
            "FROM stu_internship " +
            "GROUP BY company, experience " +
            "ORDER BY value DESC ")
    List<StudentStatisticsVO.InternshipStat> selectInternshipStats();

    // 技能统计
    @Select("SELECT COUNT(*) as value, speciality as name " +
            "FROM stu_speciality " +
            "GROUP BY speciality " +
            "ORDER BY value DESC ")
    List<StudentStatisticsVO.SpecialityStat> selectSpecialityStats();

    // 职位统计
    @Select("SELECT COUNT(*) as value, post as name " +
            "FROM stu_job " +
            "GROUP BY post " +
            "ORDER BY value DESC ")
    List<StudentStatisticsVO.JobStat> selectJobStats();

    //查询密码
    @Select("SELECT password FROM user WHERE user_id = #{id}")
    String findPwdByStudentId(Integer id);

    //修改密码
    @Update("UPDATE user SET password = #{password} WHERE user_id = #{id}")
    boolean updatePwd(Integer id, String password);

}
