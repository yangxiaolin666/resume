package com.resume.model.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

/**
 * 删除请求DTO - 学生
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class DeleteStudentsDTO {
    /**
     * 要删除的学生ID列表
     */
    private List<Integer> stuIds;
}
