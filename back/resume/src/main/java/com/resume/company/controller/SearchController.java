package com.resume.company.controller;

import com.github.pagehelper.PageInfo;
import com.resume.company.service.SearchService;
import com.resume.model.dto.StudentDetailDTO;
import com.resume.model.entity.Student;
import com.resume.model.dto.StudentQueryDTO;
import com.resume.model.result.Result;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/com")
public class SearchController {
    
    @Autowired
    private SearchService searchService;
    
    /**
     * 根据条件查询学生列表
     * @param queryDTO 查询条件
     * @return 学生列表
     */
    @PostMapping("/search")
    public Result<PageInfo<Student>> searchStudents(@RequestBody StudentQueryDTO queryDTO) {
        return Result.success(searchService.getStudentsByCondition(queryDTO));
}
    
    /**
     * 根据条件查询学生详细信息列表
     * @param queryDTO 查询条件
     * @return 学生详细信息列表
     */
    @PostMapping("/detailSearch")
    public Result<PageInfo<StudentDetailDTO>> searchStudentDetails(@RequestBody StudentQueryDTO queryDTO) {
        return Result.success(searchService.getStudentDetailsByCondition(queryDTO));
    }
}