package com.resume.model.dto;

import lombok.Data;

@Data
public class TokenValidateDTO {
    /**
     * Token（可选，如果不传则从请求头获取）
     */
    private String token;
}
