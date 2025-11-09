package com.resume.model.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * 简历上传响应VO
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class ResumeUploadVO {
    
    /**
     * 系统生成的简历唯一ID
     */
    private Integer resumeId;
    
    /**
     * 简历文件存储的名称
     */
    private String resumeName;
    
    /**
     * 简历文件的在线访问或下载地址
     */
    private String resumeUrl;
}
