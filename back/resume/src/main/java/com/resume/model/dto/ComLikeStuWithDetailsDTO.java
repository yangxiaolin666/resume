package com.resume.model.dto;

import com.resume.model.entity.Student;
import lombok.Data;

import java.util.List;

// ComLikeStuWithDetailsDTO.java
@Data
public class ComLikeStuWithDetailsDTO {
    // com_like_stu 基础字段
    private Long id;
    private Long comId;
    private Long stuId;
    private java.time.LocalDateTime createTime;

    // 学生基本信息
    private Student student;

    // 语言列表（前端希望的List形式）
    private List<String> languages;

    // 特长列表（前端希望的List形式）
    private List<String> specialities;

    // getters/setters
}
