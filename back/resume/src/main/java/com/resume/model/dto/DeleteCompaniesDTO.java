package com.resume.model.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

/**
 * 删除请求DTO - 企业
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class DeleteCompaniesDTO {
    /**
     * 要删除的企业ID列表
     */
    private List<Integer> comIds;
}
