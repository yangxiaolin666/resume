package com.resume.model.entity;

import lombok.Data;
import lombok.Getter;
import lombok.Setter;


import java.io.Serializable;

/**
 * (Student)实体类
 *
 * @author makejava
 * @since 2025-10-16 15:35:32
 */

@Data
public class Student implements Serializable {
    private static final long serialVersionUID = -29015782958147942L;

    private Integer id;

    private Integer stuId;

    private String name;

    private Integer age;

    private String gender;

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
     * 关联用户ID
     */
    private Integer userId;


}

