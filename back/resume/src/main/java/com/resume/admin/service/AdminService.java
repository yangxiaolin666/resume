package com.resume.admin.service;

import com.github.pagehelper.PageInfo;
import com.resume.model.dto.BatchOperationResultDTO;
import com.resume.model.dto.LanguageDTO;
import com.resume.model.dto.StudentDetailDTO;
import com.resume.model.dto.StudentQueryDTO;
import com.resume.model.entity.Company;
import com.resume.model.entity.Student;
import com.resume.model.result.Result;
import org.springframework.web.multipart.MultipartFile;

import java.util.List;
import java.util.Map;

/**
 * 管理员服务接口
 */
public interface AdminService {



    /**
     * 通过Excel导入学生数据
     */
    BatchOperationResultDTO importStudents(MultipartFile file);

    /**
     * 通过Excel导入企业数据
     */
    BatchOperationResultDTO importCompanies(MultipartFile file);

    /**
     * 批量删除学生
     */
    BatchOperationResultDTO deleteStudents(List<Integer> stuIds);

    /**
     * 批量删除企业
     */
    BatchOperationResultDTO deleteCompanies(List<Integer> comIds);

    /**
     * 批量更新学生
     */
    BatchOperationResultDTO updateStudents(List<Student> students);

    /**
     * 批量更新企业
     */
    BatchOperationResultDTO updateCompanies(List<Company> companies);

    /**
     * 分页查询学生
     */
    Map<String, Object> searchStudents(String name, Integer page, Integer pageSize);

    /**
     * 分页查询企业
     */
    Map<String, Object> searchCompanies(String comName, Integer page, Integer pageSize);

    /**
     * 查询学生详细信息（调用企业端接口）
     * @param queryDTO 查询条件
     * @return 学生详细信息分页列表
     */
    PageInfo<StudentDetailDTO> searchStudentDetails(StudentQueryDTO queryDTO);

    /**
    * 获取所有语言
     */
    List<String> getAllLanguages();

    Result<?> addLanguage(String language);

    Result<?> delLanguage(String language);

    List<String> getAllCertificates();

    Object addCertificate(String certificate);

    Object delCertificate(String certificate);
}
