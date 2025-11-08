package com.resume.user.controller;

import com.resume.model.dto.LoginDTO;
import com.resume.model.dto.LoginVO;
import com.resume.model.result.Result;
import com.resume.user.service.UserService;
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
     * 刷新Token
     * @param refreshToken Refresh Token
     * @return 新的Access Token
     */
    @PostMapping("/refresh-token")
    @Operation(summary = "刷新Token")
    public Result<Map<String, String>> refreshToken(@RequestHeader("Refresh-Token") String refreshToken) {
        String newAccessToken = userService.refreshToken(refreshToken);
        Map<String, String> data = new HashMap<>();
        data.put("token", newAccessToken);
        return Result.success(data);
    }
}
