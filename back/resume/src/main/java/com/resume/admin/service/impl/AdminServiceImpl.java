package com.resume.admin.service.impl;

import com.github.pagehelper.PageHelper;
import com.github.pagehelper.PageInfo;
import com.resume.admin.mapper.AdminMapper;
import com.resume.admin.service.AdminService;
import com.resume.company.service.SearchService;
import com.resume.model.dto.BatchOperationResultDTO;
import com.resume.model.dto.LanguageDTO;
import com.resume.model.dto.StudentDetailDTO;
import com.resume.model.dto.StudentQueryDTO;
import com.resume.model.entity.Company;
import com.resume.model.entity.Student;
import com.resume.model.entity.User;
import com.resume.model.result.Result;
import com.resume.user.mapper.UserMapper;
import com.resume.utils.PasswordUtil;
import org.apache.poi.ss.usermodel.*;
import org.apache.poi.xssf.usermodel.XSSFWorkbook;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.multipart.MultipartFile;

import java.io.InputStream;
import java.util.*;

/**
 * 管理员服务实现类
 */
@Service
public class AdminServiceImpl implements AdminService {

    @Autowired
    private AdminMapper adminMapper;

    @Autowired
    private UserMapper userMapper;

    @Autowired
    private SearchService searchService;


    @Override
    public boolean updateStuPassword(int stu_id, String new_Password) {
        int Num = adminMapper.updateStuPassword(stu_id, new_Password);
        return Num == 1;
    }


    @Override
    @Transactional(rollbackFor = Exception.class)
    public BatchOperationResultDTO importStudents(MultipartFile file) {
        List<Student> students = new ArrayList<>();
        List<User> users = new ArrayList<>();
        List<BatchOperationResultDTO.ErrorDetail> errors = new ArrayList<>();
        int successCount = 0;
        int failureCount = 0;
        
        try (InputStream is = file.getInputStream();
             Workbook workbook = new XSSFWorkbook(is)) {
            
            Sheet sheet = workbook.getSheetAt(0);
            int rowNum = 0;
            
            for (Row row : sheet) {
                rowNum++;
                // 跳过表头
                if (rowNum == 1) {
                    continue;
                }
                
                try {
                    Student student = parseStudentFromRow(row);
                    
                    // 检查学号是否已存在
                    Student existing = adminMapper.selectStudentByStuId(student.getStuId());
                    if (existing != null) {
                        errors.add(new BatchOperationResultDTO.ErrorDetail(
                            rowNum, "学号" + student.getStuId() + "已存在"));
                        failureCount++;
                        continue;
                    }
                    
                    // 创建User对象，username使用stuId
                    User user = new User();
                    user.setUsername(String.valueOf(student.getStuId()));
                    user.setPassword(PasswordUtil.encode(String.valueOf(student.getStuId()))); // 默认密码为学号
                    user.setUserType("student");
                    
                    users.add(user);
                    students.add(student);
                } catch (Exception e) {
                    errors.add(new BatchOperationResultDTO.ErrorDetail(
                        rowNum, "数据解析失败: " + e.getMessage()));
                    failureCount++;
                }
            }
            
            // 批量插入
            if (!users.isEmpty() && !students.isEmpty()) {
                // 先插入user表，获取自增ID
                userMapper.batchInsertUsers(users);
                
                // 将user_id设置到student中
                for (int i = 0; i < students.size(); i++) {
                    students.get(i).setUserId(users.get(i).getUserId());
                }
                
                // 再插入student表
                successCount = adminMapper.batchInsertStudents(students);
            }
            
        } catch (Exception e) {
            throw new RuntimeException("Excel文件解析失败: " + e.getMessage());
        }
        
        return new BatchOperationResultDTO(successCount, failureCount, errors);
    }
    
    @Override
    @Transactional(rollbackFor = Exception.class)
    public BatchOperationResultDTO importCompanies(MultipartFile file) {
        List<Company> companies = new ArrayList<>();
        List<User> users = new ArrayList<>();
        List<BatchOperationResultDTO.ErrorDetail> errors = new ArrayList<>();
        int successCount = 0;
        int failureCount = 0;
        
        try (InputStream is = file.getInputStream();
             Workbook workbook = new XSSFWorkbook(is)) {
            
            Sheet sheet = workbook.getSheetAt(0);
            int rowNum = 0;
            
            for (Row row : sheet) {
                rowNum++;
                // 跳过表头
                if (rowNum == 1) {
                    continue;
                }
                
                try {
                    Company company = parseCompanyFromRow(row);
                    
                    // 检查企业名称是否已存在
                    Company existing = adminMapper.selectCompanyByName(company.getComName());
                    if (existing != null) {
                        errors.add(new BatchOperationResultDTO.ErrorDetail(
                            rowNum, "企业" + company.getComName() + "已存在"));
                        failureCount++;
                        continue;
                    }
                    
                    // 创建User对象，username使用企业名称
                    User user = new User();
                    user.setUsername(company.getComName());
                    user.setPassword(PasswordUtil.encode(company.getComName())); // 默认密码为企业名称
                    user.setUserType("company");
                    
                    users.add(user);
                    companies.add(company);
                } catch (Exception e) {
                    errors.add(new BatchOperationResultDTO.ErrorDetail(
                        rowNum, "数据解析失败: " + e.getMessage()));
                    failureCount++;
                }
            }
            
            // 批量插入
            if (!users.isEmpty() && !companies.isEmpty()) {
                // 先插入user表，获取自增ID
                userMapper.batchInsertUsers(users);
                
                // 将user_id设置到company中
                for (int i = 0; i < companies.size(); i++) {
                    companies.get(i).setUserId(users.get(i).getUserId());
                }
                
                // 再插入company表
                successCount = adminMapper.batchInsertCompanies(companies);
            }
            
        } catch (Exception e) {
            throw new RuntimeException("Excel文件解析失败: " + e.getMessage());
        }
        
        return new BatchOperationResultDTO(successCount, failureCount, errors);
    }
    
    @Override
    @Transactional(rollbackFor = Exception.class)
    public BatchOperationResultDTO deleteStudents(List<Integer> stuIds) {
        List<BatchOperationResultDTO.ErrorDetail> errors = new ArrayList<>();
        int successCount = 0;
        int failureCount = 0;
        
        List<Integer> validIds = new ArrayList<>();
        List<Integer> userIds = new ArrayList<>();
        
        for (Integer id : stuIds) {
            Student student = adminMapper.selectStudentById(id);
            if (student == null) {
                errors.add(new BatchOperationResultDTO.ErrorDetail(
                    id, "学生不存在"));
                failureCount++;
            } else {
                validIds.add(id);
                if (student.getUserId() != null) {
                    userIds.add(student.getUserId());
                }
            }
        }
        
        if (!validIds.isEmpty()) {
            // 先删除student表数据
            successCount = adminMapper.batchDeleteStudents(validIds);
            
            // 再删除user表数据
            if (!userIds.isEmpty()) {
                userMapper.batchDeleteUsers(userIds);
            }
        }
        
        return new BatchOperationResultDTO(successCount, failureCount, errors);
    }
    
    @Override
    @Transactional(rollbackFor = Exception.class)
    public BatchOperationResultDTO deleteCompanies(List<Integer> comIds) {
        List<BatchOperationResultDTO.ErrorDetail> errors = new ArrayList<>();
        int successCount = 0;
        int failureCount = 0;
        
        List<Integer> validIds = new ArrayList<>();
        List<Integer> userIds = new ArrayList<>();
        
        for (Integer id : comIds) {
            Company company = adminMapper.selectCompanyById(id);
            if (company == null) {
                errors.add(new BatchOperationResultDTO.ErrorDetail(
                    id, "企业不存在"));
                failureCount++;
            } else {
                validIds.add(id);
                if (company.getUserId() != null) {
                    userIds.add(company.getUserId());
                }
            }
        }
        
        if (!validIds.isEmpty()) {
            // 先删除company表数据
            successCount = adminMapper.batchDeleteCompanies(validIds);
            
            // 再删除user表数据
            if (!userIds.isEmpty()) {
                userMapper.batchDeleteUsers(userIds);
            }
        }
        
        return new BatchOperationResultDTO(successCount, failureCount, errors);
    }
    
    @Override
    @Transactional(rollbackFor = Exception.class)
    public BatchOperationResultDTO updateStudents(List<Student> students) {
        List<BatchOperationResultDTO.ErrorDetail> errors = new ArrayList<>();
        int successCount = 0;
        int failureCount = 0;
        int index = 0;
        
        for (Student student : students) {
            index++;
            try {
                if (student.getId() == null) {
                    errors.add(new BatchOperationResultDTO.ErrorDetail(
                        index, "学生ID不能为空"));
                    failureCount++;
                    continue;
                }
                
                Student existing = adminMapper.selectStudentById(student.getId());
                if (existing == null) {
                    errors.add(new BatchOperationResultDTO.ErrorDetail(
                        index, "学生不存在"));
                    failureCount++;
                    continue;
                }
                
                int result = adminMapper.updateStudent(student);
                if (result > 0) {
                    successCount++;
                } else {
                    errors.add(new BatchOperationResultDTO.ErrorDetail(
                        index, "更新失败"));
                    failureCount++;
                }
            } catch (Exception e) {
                errors.add(new BatchOperationResultDTO.ErrorDetail(
                    index, "更新失败: " + e.getMessage()));
                failureCount++;
            }
        }
        
        return new BatchOperationResultDTO(successCount, failureCount, errors);
    }
    
    @Override
    @Transactional(rollbackFor = Exception.class)
    public BatchOperationResultDTO updateCompanies(List<Company> companies) {
        List<BatchOperationResultDTO.ErrorDetail> errors = new ArrayList<>();
        int successCount = 0;
        int failureCount = 0;
        int index = 0;
        
        for (Company company : companies) {
            index++;
            try {
                if (company.getId() == null) {
                    errors.add(new BatchOperationResultDTO.ErrorDetail(
                        index, "企业ID不能为空"));
                    failureCount++;
                    continue;
                }
                
                Company existing = adminMapper.selectCompanyById(company.getId());
                if (existing == null) {
                    errors.add(new BatchOperationResultDTO.ErrorDetail(
                        index, "企业不存在"));
                    failureCount++;
                    continue;
                }
                
                int result = adminMapper.updateCompany(company);
                if (result > 0) {
                    successCount++;
                } else {
                    errors.add(new BatchOperationResultDTO.ErrorDetail(
                        index, "更新失败"));
                    failureCount++;
                }
            } catch (Exception e) {
                errors.add(new BatchOperationResultDTO.ErrorDetail(
                    index, "更新失败: " + e.getMessage()));
                failureCount++;
            }
        }
        
        return new BatchOperationResultDTO(successCount, failureCount, errors);
    }
    
    @Override
    public Map<String, Object> searchStudents(String name, Integer page, Integer pageSize) {
        // 设置默认值
        if (page == null || page < 1) {
            page = 1;
        }
        if (pageSize == null || pageSize < 1) {
            pageSize = 10;
        }
        
        // 分页查询
        PageHelper.startPage(page, pageSize);
        List<Student> students = adminMapper.selectStudentsByPage(name);
        int total = adminMapper.countStudents(name);
        
        // 封装结果
        Map<String, Object> result = new HashMap<>();
        result.put("list", students);
        
        Map<String, Object> pagination = new HashMap<>();
        pagination.put("total", total);
        pagination.put("page", page);
        pagination.put("pageSize", pageSize);
        result.put("pagination", pagination);
        
        return result;
    }
    
    @Override
    public Map<String, Object> searchCompanies(String comName, Integer page, Integer pageSize) {
        // 设置默认值
        if (page == null || page < 1) {
            page = 1;
        }
        if (pageSize == null || pageSize < 1) {
            pageSize = 10;
        }

        // 分页查询
        PageHelper.startPage(page, pageSize);
        List<Company> companies = adminMapper.selectCompaniesByPage(comName);
        int total = adminMapper.countCompanies(comName);

        // 封装结果
        Map<String, Object> result = new HashMap<>();
        result.put("list", companies);

        Map<String, Object> pagination = new HashMap<>();
        pagination.put("total", total);
        pagination.put("page", page);
        pagination.put("pageSize", pageSize);
        result.put("pagination", pagination);

        return result;
    }

    @Override
    public PageInfo<StudentDetailDTO> searchStudentDetails(StudentQueryDTO queryDTO) {
        // 调用企业端接口查询学生详细信息
        return searchService.getStudentDetailsByCondition(queryDTO);
    }


    /**
     * 从Excel行解析学生数据
     * 列顺序: 学号 | 姓名 | 年龄 | 性别 | 民族 | 专业 | 学位 | 毕业年份 | 实习次数
     */
    private Student parseStudentFromRow(Row row) {
        Student student = new Student();
        
        Cell cell0 = row.getCell(0);
        if (cell0 != null) {
            student.setStuId(getCellIntValue(cell0));
        }
        
        Cell cell1 = row.getCell(1);
        if (cell1 != null) {
            student.setName(getCellStringValue(cell1));
        }
        
        Cell cell2 = row.getCell(2);
        if (cell2 != null) {
            student.setAge(getCellIntValue(cell2));
        }
        
        Cell cell3 = row.getCell(3);
        if (cell3 != null) {
            student.setGender(getCellStringValue(cell3));
        }
        
        Cell cell4 = row.getCell(4);
        if (cell4 != null) {
            student.setNation(getCellStringValue(cell4));
        }
        
        Cell cell5 = row.getCell(5);
        if (cell5 != null) {
            student.setMajor(getCellStringValue(cell5));
        }
        
        Cell cell6 = row.getCell(6);
        if (cell6 != null) {
            student.setDegree(getCellStringValue(cell6));
        }
        
        Cell cell7 = row.getCell(7);
        if (cell7 != null) {
            student.setGraduationYear(getCellIntValue(cell7));
        }
        
        Cell cell8 = row.getCell(8);
        if (cell8 != null) {
            student.setInternship(getCellIntValue(cell8));
        }
        
        return student;
    }
    
    /**
     * 从Excel行解析企业数据
     * 列顺序: 企业名称 | 所在地区
     */
    private Company parseCompanyFromRow(Row row) {
        Company company = new Company();
        
        Cell cell0 = row.getCell(0);
        if (cell0 != null) {
            company.setComName(getCellStringValue(cell0));
        }
        
        Cell cell1 = row.getCell(1);
        if (cell1 != null) {
            company.setComRegion(getCellStringValue(cell1));
        }
        
        return company;
    }
    
    /**
     * 获取单元格字符串值
     */
    private String getCellStringValue(Cell cell) {
        if (cell == null) {
            return null;
        }
        
        switch (cell.getCellType()) {
            case STRING:
                return cell.getStringCellValue().trim();
            case NUMERIC:
                return String.valueOf((int) cell.getNumericCellValue());
            case BOOLEAN:
                return String.valueOf(cell.getBooleanCellValue());
            default:
                return null;
        }
    }
    
    /**
     * 获取单元格整数值
     */
    private Integer getCellIntValue(Cell cell) {
        if (cell == null) {
            return null;
        }
        
        switch (cell.getCellType()) {
            case NUMERIC:
                return (int) cell.getNumericCellValue();
            case STRING:
                String value = cell.getStringCellValue().trim();
                return value.isEmpty() ? null : Integer.parseInt(value);
            default:
                return null;
        }
    }

    /**
     * 获取所有语言
     */
    @Override
    public List<String> getAllLanguages() {
        return adminMapper.getAllLanguages();
    }

    /**
     * 添加语言
     */
    @Override
    public Result<?> addLanguage(String language){
        return adminMapper.addLanguage(language);
    }

    /**
     * 删除语言
     */
    @Override
    public Result<?> delLanguage(String language){
        return adminMapper.deleteLanguage(language);
    }

    /**
     * 获取所有证书
     */
    @Override
    public List<String> getAllCertificates() {
        return adminMapper.getAllCertificates();
    }

    /**
     * 添加证书
     */
    @Override
    public Result<?> addCertificate(String certificate){
        return adminMapper.addCertificate(certificate);
    }
    /**
     * 删除证书
     */
    @Override
    public Result<?> delCertificate(String certificate){
        return adminMapper.deleteCertificate(certificate);
    }
}
