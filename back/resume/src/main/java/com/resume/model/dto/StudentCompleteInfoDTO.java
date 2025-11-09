package com.resume.model.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

/**
 * 学生完善信息DTO
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class StudentCompleteInfoDTO {
    
    /**
     * 年龄
     */
    private Integer age;
    
    /**
     * 性别
     */
    private String gender;
    
    /**
     * 民族
     */
    private String nation;
    
    /**
     * 所属专业
     */
    private String major;
    
    /**
     * 学位
     */
    private String degree;
    
    /**
     * 毕业年份
     */
    private Integer graduationYear;
    
    /**
     * 实习次数
     */
    private Integer internship;
    
    /**
     * 掌握的语言列表
     */
    private List<LanguageInfo> languages;
    
    /**
     * 意向工作城市
     */
    private List<String> cities;
    
    /**
     * 意向工作省份
     */
    private List<String> provinces;
    
    /**
     * 持有的证书名称列表
     */
    private List<CertificateInfo> certificates;
    
    /**
     * 实习公司信息列表
     */
    private List<InternshipInfo> internships;
    
    /**
     * 专业技能标签列表
     */
    private List<String> specialities;
    
    /**
     * 求职意向列表
     */
    private List<String> jobs;
    
    /**
     * 语言信息内部类
     */
    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class LanguageInfo {
        private String language;
        private String level;
    }
    
    /**
     * 证书信息内部类
     */
    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class CertificateInfo {
        private String certName;
        private String level;
    }
    
    /**
     * 实习信息内部类
     */
    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class InternshipInfo {
        private String company;
        private String experience;
    }
}
