package com.resume.model.dto;

import com.resume.model.entity.Company;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

/**
 * 批量更新企业DTO
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class UpdateCompaniesDTO {
    /**
     * 要更新的企业列表
     */
    private List<Company> companies;
}
