package com.resume.company.controller;

import com.github.pagehelper.PageInfo;
import com.resume.company.service.SearchService;
import com.resume.model.StudentDetailDTO;
import com.resume.model.entity.Student;
import com.resume.model.StudentQueryDTO;
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
    public PageInfo<Student> searchStudents(@RequestBody StudentQueryDTO queryDTO) {
        return searchService.getStudentsByCondition(queryDTO);
    }
    
    /**
     * 根据条件查询学生详细信息列表
     * @param queryDTO 查询条件
     * @return 学生详细信息列表
     */
    @PostMapping("/detailSearch")
    public PageInfo<StudentDetailDTO> searchStudentDetails(@RequestBody StudentQueryDTO queryDTO) {
        return searchService.getStudentDetailsByCondition(queryDTO);
    }
}