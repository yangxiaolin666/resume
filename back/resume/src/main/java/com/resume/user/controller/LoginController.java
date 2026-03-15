package com.resume.user.controller;

import com.resume.model.dto.LoginDTO;
import com.resume.model.dto.LoginVO;
import com.resume.model.dto.TokenValidateDTO;
import com.resume.model.result.Result;
import com.resume.user.service.UserService;
import com.resume.utils.JwtUtil;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.Map;

/**
 * 用户登录控制器
 */
@RestController
@Tag(name = "公共模块")
public class LoginController {
    
    @Autowired
    private UserService userService;
    
    /**
     * 用户登录
     * @param loginDTO 登录信息
     * @return 登录结果
     */
    @PostMapping("/login")
    @Operation(summary = "用户登录")
    public Result<LoginVO> login(@Validated @RequestBody LoginDTO loginDTO) {
        LoginVO loginVO = userService.login(loginDTO);
        return Result.success(loginVO);
    }
    
    /**
     * 刷新 Token
     * @param refreshToken Refresh Token
     * @return 新的 Access Token
     */
    @PostMapping("/refresh-token")
    @Operation(summary = "刷新 Token")
    public Result<Map<String, String>> refreshToken(@RequestHeader("Refresh-Token") String refreshToken) {
        String newAccessToken = userService.refreshToken(refreshToken);
        Map<String, String> data = new HashMap<>();
        data.put("token", newAccessToken);
        return Result.success(data);
    }
    
    /**
     * 验证 Token 是否有效且未过期
     * @param dto 请求参数（可选，token 也可以从请求头获取）
     * @return 验证结果
     */
    @PostMapping("/validateToken")
    @Operation(summary = "验证 Token")
    public Result<Map<String, Object>> validateToken(@RequestBody(required = false) TokenValidateDTO dto) {
        try {
            // 优先从请求体获取 token，如果没有则从请求头获取
            String token = null;
            if (dto != null && dto.getToken() != null) {
                token = dto.getToken();
            }
            
            // 如果请求体和请求头都没有 token，返回错误
            if (token == null || token.isEmpty()) {
                Map<String, Object> result = new HashMap<>();
                result.put("valid", false);
                result.put("message", "Token 不能为空");
                return Result.error(401, "Token 不能为空");
            }
            
            // 验证 token
            boolean isValid = false;
            try {
                JwtUtil.verifyToken(token);
                isValid = true;
            } catch (Exception e) {
                // Token 无效或过期
            }
            
            Map<String, Object> result = new HashMap<>();
            result.put("valid", isValid);
            
            if (isValid) {
                // 如果 token 有效，可以返回一些额外信息
                Integer userId = JwtUtil.getUserId(token);
                String userType = JwtUtil.getUserType(token);
                
                result.put("message", "Token 有效且未过期");
                result.put("userId", userId);
                result.put("userType", userType);
                
                // 检查是否快过期了（剩余时间少于 30 分钟）
                boolean needRefresh = JwtUtil.isTokenExpiringSoon(token);
                result.put("needRefresh", needRefresh);
            } else {
                result.put("message", "Token 无效或已过期");
            }
            
            return Result.success(result);
            
        } catch (Exception e) {
            Map<String, Object> result = new HashMap<>();
            result.put("valid", false);
            result.put("message", "Token 验证失败：" + e.getMessage());
            return Result.error(500, "Token 验证失败");
        }
    }
}
