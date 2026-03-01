package com.resume.company.service.impl;

import com.github.pagehelper.PageHelper;
import com.github.pagehelper.PageInfo;
import com.resume.company.mapper.SearchMapper;
import com.resume.company.mapper.StudentDetailMapper;
import com.resume.company.service.SearchService;
import com.resume.model.dto.*;
import com.resume.model.entity.Student;
import com.resume.model.dto.StudentQueryDTO;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.stream.Collectors;

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
        
        // 执行查询 - 获取学生基本详细信息
        List<StudentDetailDTO> students = studentDetailMapper.selectStudentDetailsByCondition(queryDTO);
        
        if (students.isEmpty()) {
            return new PageInfo<>(students);
        }
        
        // 提取学生ID列表
        List<Integer> studentIds = students.stream()
            .map(StudentDetailDTO::getId)
            .collect(Collectors.toList());
        
        // 根据学生ID批量查询关联信息
        List<LanguageDTO> languagesList = studentDetailMapper.selectLanguagesByStudentIds(studentIds);
        List<CertificateDTO> certificatesList = studentDetailMapper.selectCertificatesByStudentIds(studentIds);
        List<InternshipDTO> internshipsList = studentDetailMapper.selectInternshipsByStudentIds(studentIds);
        List<JobDTO> jobsList = studentDetailMapper.selectJobsByStudentIds(studentIds);
        List<SpecialityDTO> specialitiesList = studentDetailMapper.selectSpecialitiesByStudentIds(studentIds);
        
        // 构建学生ID与关联信息的映射
        Map<Integer, List<LanguageDTO>> languagesMap = languagesList.stream()
            .collect(Collectors.groupingBy(LanguageDTO::getStuId));
        Map<Integer, List<CertificateDTO>> certificatesMap = certificatesList.stream()
            .collect(Collectors.groupingBy(CertificateDTO::getStuId));
        Map<Integer, List<InternshipDTO>> internshipsMap = internshipsList.stream()
            .collect(Collectors.groupingBy(InternshipDTO::getStuId));
        Map<Integer, List<JobDTO>> jobsMap = jobsList.stream()
            .collect(Collectors.groupingBy(JobDTO::getStuId));
        Map<Integer, List<SpecialityDTO>> specialitiesMap = specialitiesList.stream()
            .collect(Collectors.groupingBy(SpecialityDTO::getStuId));
        
        // 将关联信息填充到学生详细信息中
        for (StudentDetailDTO student : students) {
            student.setLanguages(languagesMap.getOrDefault(student.getId(), new ArrayList<>()));
            student.setCertificates(certificatesMap.getOrDefault(student.getId(), new ArrayList<>()));
            student.setInternships(internshipsMap.getOrDefault(student.getId(), new ArrayList<>()));
            student.setJobs(jobsMap.getOrDefault(student.getId(), new ArrayList<>()));
            student.setSpecialities(specialitiesMap.getOrDefault(student.getId(), new ArrayList<>()));
        }
        
        // 返回分页结果
        return new PageInfo<>(students);
    }
}