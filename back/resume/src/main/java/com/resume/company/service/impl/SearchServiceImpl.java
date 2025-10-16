package com.resume.company.service.impl;

import com.github.pagehelper.PageHelper;
import com.github.pagehelper.PageInfo;
import com.resume.company.mapper.SearchMapper;
import com.resume.company.mapper.StudentDetailMapper;
import com.resume.company.service.SearchService;
import com.resume.model.dto.StudentDetailDTO;
import com.resume.model.entity.Student;
import com.resume.model.dto.StudentQueryDTO;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class SearchServiceImpl implements SearchService {
    
    @Autowired
    private SearchMapper searchMapper;
    
    @Autowired
    private StudentDetailMapper studentDetailMapper;
    
    @Override
    public PageInfo<Student> getStudentsByCondition(StudentQueryDTO queryDTO) {
        // 设置分页参数
        PageHelper.startPage(queryDTO.getPageNum(), queryDTO.getPageSize());
        
        // 执行查询
        List<Student> students = searchMapper.selectStudentsByCondition(queryDTO);
        
        // 返回分页结果
        return new PageInfo<>(students);
    }
    
    @Override
    public PageInfo<StudentDetailDTO> getStudentDetailsByCondition(StudentQueryDTO queryDTO) {
        // 设置分页参数
        PageHelper.startPage(queryDTO.getPageNum(), queryDTO.getPageSize());
        
        // 执行查询
        List<StudentDetailDTO> students = studentDetailMapper.selectStudentDetailsByCondition(queryDTO);
        
        // 返回分页结果
        return new PageInfo<>(students);
    }
}