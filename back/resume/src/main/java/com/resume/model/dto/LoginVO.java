package com.resume.model.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * 登录响应VO
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class LoginVO {
    
    /**
     * 用户ID
     */
    private Integer userId;
    
    /**
     * 用户名
     */
    private String username;
    
    /**
     * 用户类型
     */
    private String userType;
    
    /**
     * 访问Token（短期）
     */
    private String token;
    
    /**
     * 刷新Token（长期）
     */
    private String refreshToken;
}
