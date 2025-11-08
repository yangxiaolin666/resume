package com.resume.common.interceptor;

import com.resume.utils.JwtUtil;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.HandlerInterceptor;

/**
 * Token刷新拦截器（第二层）
 * 检查Access Token是否即将过期，如果即将过期则在响应头中返回新Token
 */
@Component
public class TokenRefreshInterceptor implements HandlerInterceptor {
    
    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) throws Exception {
        // 跨域预检请求直接放行
        if ("OPTIONS".equals(request.getMethod())) {
            return true;
        }
        
        String token = getTokenFromRequest(request);
        
        if (token != null) {
            try {
                // 检查Token是否即将过期（30分钟内）
                if (JwtUtil.isTokenExpiringSoon(token)) {
                    // 获取用户信息
                    Integer userId = JwtUtil.getUserId(token);
                    String username = JwtUtil.getUsername(token);
                    String userType = JwtUtil.getUserType(token);
                    
                    // 生成新的Access Token
                    String newToken = JwtUtil.generateAccessToken(userId, username, userType);
                    
                    // 在响应头中返回新Token
                    response.setHeader("New-Token", newToken);
                }
            } catch (Exception e) {
                // Token无效，不做处理，让AuthInterceptor处理
            }
        }
        
        return true;
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
