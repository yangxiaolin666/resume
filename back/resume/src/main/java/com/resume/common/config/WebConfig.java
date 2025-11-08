package com.resume.common.config;

import com.resume.common.interceptor.AuthInterceptor;
import com.resume.common.interceptor.TokenRefreshInterceptor;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.InterceptorRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

/**
 * Web配置类
 * 注册拦截器
 */
@Configuration
public class WebConfig implements WebMvcConfigurer {
    
    @Autowired
    private AuthInterceptor authInterceptor;
    
    @Autowired
    private TokenRefreshInterceptor tokenRefreshInterceptor;
    
    @Override
    public void addInterceptors(InterceptorRegistry registry) {
        // 第一层：Token刷新拦截器（先执行，检查是否需要刷新）
        registry.addInterceptor(tokenRefreshInterceptor)
                .addPathPatterns("/com/**", "/stu/**", "/admin/**")
                .excludePathPatterns("/login", "/refresh-token")
                .order(1);
        
        // 第二层：认证拦截器（后执行，验证Token有效性）
        registry.addInterceptor(authInterceptor)
                .addPathPatterns("/com/**", "/stu/**", "/admin/**")
                .excludePathPatterns("/login", "/refresh-token")
                .order(2);
    }
}
