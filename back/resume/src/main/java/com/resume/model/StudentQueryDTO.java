package com.resume.model;

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
     * 最小年龄
     */
    private Integer minAge;
    
    /**
     * 最大年龄
     */
    private Integer maxAge;
    
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
    private Integer minGraduationYear;
    
    /**
     * 最大毕业年份
     */
    private Integer maxGraduationYear;
    
    /**
     * 最小实习次数
     */
    private Integer minInternship;
    
    /**
     * 最大实习次数
     */
    private Integer maxInternship;
    
    /**
     * 语言要求
     */
    private String language;
    
    /**
     * 语言等级
     */
    private String languageLevel;
    
    /**
     * 城市
     */
    private String city;
    
    /**
     * 省份
     */
    private String province;
    
    /**
     * 证书名称
     */
    private String certName;
    
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
    private String speciality;
    
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