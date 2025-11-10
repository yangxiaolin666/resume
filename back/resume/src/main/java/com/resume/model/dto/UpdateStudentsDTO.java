package com.resume.model.dto;

import com.resume.model.entity.Student;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

/**
 * 批量更新学生DTO
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class UpdateStudentsDTO {
    /**
     * 要更新的学生列表
     */
    private List<Student> students;
}
