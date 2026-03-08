package com.resume.model.dto;

import lombok.Data;

@Data
public class LanguageQueryDTO {
    /**
     * 语言名称
     */
    private String language;
    
    /**
     * 语言等级
     */
    private String level;
}
