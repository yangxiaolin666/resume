package com.resume.admin.controller;

import com.github.pagehelper.PageInfo;
import com.resume.admin.service.AdminService;
import com.resume.company.service.SearchService;
import com.resume.model.dto.*;
import com.resume.model.result.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.util.List;
import java.util.Map;

/**
 * 管理端控制器
 */
@Tag(name = "管理端接口")
@RestController
@RequestMapping("/admin")
public class AdminController {

    @Autowired
    private AdminService adminService;

    @Autowired
    private SearchService searchService;
    
    /**
     * 1. 导入学生数据
     */
    @Operation(summary = "导入学生数据")
    @PostMapping("/addStu")
    public Result<? > importStudents(@RequestParam("stuFile") MultipartFile stuFile) {
        try {
            // 验证文件
            if (stuFile.isEmpty()) {
                return Result.error(40009, "Excel文件内容为空");
            }
            
            String filename = stuFile.getOriginalFilename();
            if (filename == null || !filename.endsWith(".xlsx")) {
                return Result.error(41501, "文件格式不支持，仅支持.xlsx格式");
            }
            
            BatchOperationResultDTO result = adminService.importStudents(stuFile);
            
            if (result.getFailureCount() > 0) {
                return new Result<>(200, "部分数据导入失败", result);
            }
            
            return Result.success(result);
        } catch (Exception e) {
            return Result.error(50001, "数据库写入失败: " + e.getMessage());
        }
    }
    
    /**
     * 2. 导入企业数据
     */
    @Operation(summary = "导入企业数据")
    @PostMapping("/addCom")
    public Result<?> importCompanies(@RequestParam("comFile") MultipartFile comFile) {
        try {
            // 验证文件
            if (comFile.isEmpty()) {
                return Result.error(40009, "Excel文件内容为空");
            }
            
            String filename = comFile.getOriginalFilename();
            if (filename == null || !filename.endsWith(".xlsx")) {
                return Result.error(41501, "文件格式不支持，仅支持.xlsx格式");
            }
            
            BatchOperationResultDTO result = adminService.importCompanies(comFile);
            
            if (result.getFailureCount() > 0) {
                return new Result<>(200, "部分数据导入失败", result);
            }
            
            return Result.success(result);
        } catch (Exception e) {
            return Result.error(50001, "数据库写入失败: " + e.getMessage());
        }
    }
    
    /**
     * 3. 删除学生数据
     */
    @Operation(summary = "删除学生数据")
    @DeleteMapping("/delStu")
    public Result<?> deleteStudents(@RequestBody DeleteStudentsDTO dto) {
        try {
            if (dto.getStuIds() == null || dto.getStuIds().isEmpty()) {
                return Result.error(400, "学生ID列表不能为空");
            }
            
            BatchOperationResultDTO result = adminService.deleteStudents(dto.getStuIds());
            
            if (result.getFailureCount() > 0) {
                return new Result<>(40402, "部分学生不存在", result);
            }
            
            return Result.success(result);
        } catch (Exception e) {
            return Result.error(500, "删除失败: " + e.getMessage());
        }
    }
    
    /**
     * 4. 删除企业数据
     */
    @Operation(summary = "删除企业数据")
    @DeleteMapping("/delCom")
    public Result<?> deleteCompanies(@RequestBody DeleteCompaniesDTO dto) {
        try {
            if (dto.getComIds() == null || dto.getComIds().isEmpty()) {
                return Result.error(400, "企业ID列表不能为空");
            }
            
            BatchOperationResultDTO result = adminService.deleteCompanies(dto.getComIds());
            
            if (result.getFailureCount() > 0) {
                return new Result<>(40403, "部分企业不存在", result);
            }
            
            return Result.success(result);
        } catch (Exception e) {
            return Result.error(500, "删除失败: " + e.getMessage());
        }
    }
    
    /**
     * 5. 修改学生数据
     */
    @Operation(summary = "修改学生数据")
    @PutMapping("/updateStu")
    public Result<?> updateStudents(@RequestBody UpdateStudentsDTO dto) {
        try {
            if (dto.getStudents() == null || dto.getStudents().isEmpty()) {
                return Result.error(400, "学生列表不能为空");
            }
            
            BatchOperationResultDTO result = adminService.updateStudents(dto.getStudents());
            
            if (result.getFailureCount() > 0) {
                return new Result<>(42205, "学生数据验证失败", result);
            }
            
            return Result.success(result);
        } catch (Exception e) {
            return Result.error(500, "更新失败: " + e.getMessage());
        }
    }
    
    /**
     * 6. 修改企业数据
     */
    @Operation(summary = "修改企业数据")
    @PutMapping("/updateCom")
    public Result<?> updateCompanies(@RequestBody UpdateCompaniesDTO dto) {
        try {
            if (dto.getCompanies() == null || dto.getCompanies().isEmpty()) {
                return Result.error(400, "企业列表不能为空");
            }
            
            BatchOperationResultDTO result = adminService.updateCompanies(dto.getCompanies());
            
            if (result.getFailureCount() > 0) {
                return new Result<>(42206, "企业数据验证失败", result);
            }
            
            return Result.success(result);
        } catch (Exception e) {
            return Result.error(500, "更新失败: " + e.getMessage());
        }
    }
    
    /**
     * 7. 查找学生数据
     */
    @Operation(summary = "查找学生数据")
    @GetMapping("/searchStu")
    public Result<?> searchStudents(
            @RequestParam(required = false) String username,
            @RequestParam(required = false, defaultValue = "1") Integer page,
            @RequestParam(required = false, defaultValue = "10") Integer pageSize) {
        try {
            Map<String, Object> result = adminService.searchStudents(username, page, pageSize);
            return Result.success(result);
        } catch (Exception e) {
            return Result.error(500, "查询失败: " + e.getMessage());
        }
    }
    
    /**
     * 8. 查找企业数据
     */
    @Operation(summary = "查找企业数据")
    @GetMapping("/searchCom")
    public Result<?> searchCompanies(
            @RequestParam(required = false) String companyName,
            @RequestParam(required = false, defaultValue = "1") Integer page,
            @RequestParam(required = false, defaultValue = "10") Integer pageSize) {
        try {
            Map<String, Object> result = adminService.searchCompanies(companyName, page, pageSize);
            return Result.success(result);
        } catch (Exception e) {
            return Result.error(500, "查询失败: " + e.getMessage());
        }
    }

    /**
     * 9. 查找学生详细信息（调用企业端接口）
     */
    @Operation(summary = "查找学生详细信息")
    @PostMapping("/searchStuDetail")
    public Result<PageInfo<StudentDetailDTO>> searchStudentDetails(@RequestBody StudentQueryDTO queryDTO) {
        try {
            PageInfo<StudentDetailDTO> result = adminService.searchStudentDetails(queryDTO);
            return Result.success(result);
        } catch (Exception e) {
            return Result.error(500, "查询失败: " + e.getMessage());
        }
    }

    /**
     * 10.  获取所有语言
     */
    @GetMapping("/getAllLanguage")
    public Result<List<String>> getAllLanguage(){
        return Result.success(adminService.getAllLanguages());
    }

    /**
    * 11.  增加语言
     */
    @PostMapping("/addLanguage")
    public Result<?> addLanguage(@RequestParam String language){
        try {
            if (language == null || language.isEmpty()) {
                return Result.error(400, "语言不能为空");
            }
            if (adminService.getAllLanguages().contains(language)) {
                return Result.error(400, "语言已存在");
            }
            return Result.success(adminService.addLanguage(language));
        } catch (Exception e) {
            return Result.error(500, "添加失败: " + e.getMessage());
        }
    }

    /**
     * 12.删除语言
     */
    @PostMapping("/delLanguage")
    public Result<?> delLanguage(@RequestParam String language){
        try {
            if (language == null || language.isEmpty()) {
                return Result.error(400, "语言不能为空");
            }
            if (!adminService.getAllLanguages().contains(language)) {
                return Result.error(400, "语言不存在");
            }
            return Result.success(adminService.delLanguage(language));
        } catch (Exception e) {
            return Result.error(500, "删除失败: " + e.getMessage());
        }
    }
}
