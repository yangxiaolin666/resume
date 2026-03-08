package com.resume.model.entity;


import lombok.Data;

@Data
public class StuFavorite {
    private Integer stuId;        // 对应列别名 stu_id 或 stuId（开启驼峰后自动映射）
    private String studentName;  // 对应列别名 student_name
    private String major;
    private String gender;
    private String degree;
    private String graduationYear;
    private String languages;    // 对应列别名 languages
    private String specialities; // 对应列别名 specialities
}
