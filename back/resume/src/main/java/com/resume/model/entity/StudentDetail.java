package com.resume.model.entity;

import com.resume.model.dto.*;
import lombok.Data;
import java.util.List;

@Data
public class StudentDetail extends Student {
    /**
     * 语言证书列表
     */
    private List<LanguageDTO> languages;
    
    /**
     * 专业证书列表
     */
    private List<CertificateDTO> certificates;
    
    /**
     * 实习经历列表
     */
    private List<InternshipDTO> internships;
    
    /**
     * 工作岗位列表
     */
    private List<JobDTO> jobs;
    
    /**
     * 地区信息
     */
    private RegionDTO region;
    
    /**
     * 简历信息
     */
    private ResumeDTO resume;
    
    /**
     * 专业技能列表
     */
    private List<SpecialityDTO> specialities;
}