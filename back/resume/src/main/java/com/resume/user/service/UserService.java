package com.resume.user.service;

import com.resume.model.dto.LoginDTO;
import com.resume.model.dto.LoginVO;

/**
 * 用户服务接口
 */
public interface UserService {
    
    /**
     * 用户登录
     * @param loginDTO 登录信息
     * @return 登录结果（包含用户信息和token）
     */
    LoginVO login(LoginDTO loginDTO);
    
    /**
     * 刷新Token
     * @param refreshToken 刷新Token
     * @return 新的Access Token
     */
    String refreshToken(String refreshToken);
}
