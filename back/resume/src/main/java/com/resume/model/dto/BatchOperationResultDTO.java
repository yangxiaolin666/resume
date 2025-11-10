package com.resume.model.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

/**
 * 批量操作结果DTO
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class BatchOperationResultDTO {
    
    /**
     * 成功数量
     */
    private Integer successCount;
    
    /**
     * 失败数量
     */
    private Integer failureCount;
    
    /**
     * 错误信息列表
     */
    private List<ErrorDetail> errors;
    
    /**
     * 错误详情
     */
    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class ErrorDetail {
        /**
         * 行号或ID
         */
        private Object identifier;
        
        /**
         * 失败原因
         */
        private String reason;
    }
}
