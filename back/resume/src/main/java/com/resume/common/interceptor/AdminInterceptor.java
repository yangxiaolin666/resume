package com.resume.common.interceptor;

import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.HandlerInterceptor;

/**
 * 管理员权限拦截器
 * 验证用户是否为管理员角色
 */
@Component
public class AdminInterceptor implements HandlerInterceptor {
    
    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) throws Exception {
        // 跨域预检请求直接放行
        if ("OPTIONS".equals(request.getMethod())) {
            return true;
        }
        
        // 从request中获取用户类型（由AuthInterceptor设置）
        String userType = (String) request.getAttribute("userType");
        
        if (userType == null) {
            response.setStatus(HttpServletResponse.SC_UNAUTHORIZED);
            response.setContentType("application/json;charset=UTF-8");
            response.getWriter().write("{\"code\":401,\"message\":\"未登录\",\"data\":null}");
            return false;
        }
        
        // 验证是否为管理员
        if (!"admin".equals(userType)) {
            response.setStatus(HttpServletResponse.SC_FORBIDDEN);
            response.setContentType("application/json;charset=UTF-8");
            response.getWriter().write("{\"code\":403,\"message\":\"权限不足，需要管理员权限\",\"data\":null}");
            return false;
        }
        
        return true;
    }
}
