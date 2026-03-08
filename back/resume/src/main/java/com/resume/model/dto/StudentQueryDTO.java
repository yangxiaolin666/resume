package com.resume.model.dto;

import lombok.Data;
import java.util.List;

@Data
public class StudentQueryDTO {
    /**
     * 学生ID
     */
    private Integer id;
    
    /**
     * 学号
     */
    private Integer stuId;
    
    /**
     * 姓名（模糊查询）
     */
    private String name;
    
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
     * 所属专业（模糊查询）
     */
    private String major;
    
    /**
     * 学位
     */
    private String degree;
    
    /**
     * 最小毕业年份
     */
    private Integer graduationYear;

    /**
     * 实习次数
     */
    private Integer internship;
    
    /**
     * 语言要求（对象数组）
     */
    private List<LanguageQueryDTO> languages;
    
    /**
     * 城市
     */
    private List<String> cities;
    
    /**
     * 省份
     */
    private List<String> provinces;
    
    /**
     * 证书名称
     */
    private List<String> certNames;
    
    /**
     * 证书等级
     */
    private String certLevel;
    
    /**
     * 实习公司
     */
    private String internshipCompany;
    
    /**
     * 专业技能
     */
    private List<String> specialities;
    
    /**
     * 求职岗位
     */
    private List<String> jobs;
    
    /**
     * 当前页码
     */
    private Integer pageNum = 1;
    
    /**
     * 每页大小
     */
    private Integer pageSize = 10;
    
    /**
     * 偏移量
     */
    public Integer getOffset() {
        if (pageNum != null && pageSize != null) {
            return (pageNum - 1) * pageSize;
        }
        return 0;
    }
}