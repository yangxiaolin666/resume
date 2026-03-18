package com.resume.common.util;

import com.resume.admin.mapper.CompanyMapper;
import com.resume.utils.JwtUtil;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

/**
 * 当前用户工具类
 */
@Component
public class CurrentUserUtil {

    @Autowired
    private CompanyMapper companyMapper;

    /**
     * 从请求中获取当前企业ID
     * @param request 请求对象
     * @return 企业ID
     */
    public Integer getComIdFromRequest(HttpServletRequest request) {
        String token = getTokenFromRequest(request);
        if (token == null) {
            return null;
        }

        try {
            // 验证token
            JwtUtil.verifyToken(token);

            // 获取用户ID
            Integer userId = JwtUtil.getUserId(token);

            // 根据userId查询企业ID
            return companyMapper.selectCompanyById(userId).getId();
        } catch (Exception e) {
            return null;
        }
    }

    /**
     * 从请求中获取当前用户ID
     * @param request 请求对象
     * @return 用户ID
     */
    public Integer getUserIdFromRequest(HttpServletRequest request) {
        String token = getTokenFromRequest(request);
        if (token == null) {
            return null;
        }

        try {
            // 验证token
            JwtUtil.verifyToken(token);

            // 获取用户ID
            return JwtUtil.getUserId(token);
        } catch (Exception e) {
            return null;
        }
    }

    /**
     * 从请求中获取Token
     */
    private String getTokenFromRequest(HttpServletRequest request) {
        String authHeader = request.getHeader("Authorization");
        if (authHeader != null && authHeader.startsWith("Bearer ")) {
            return authHeader.substring(7);
        }
        return null;
    }
}
