package com.resume.common.config;

import com.resume.common.interceptor.AdminInterceptor;
import com.resume.common.interceptor.AuthInterceptor;
import com.resume.common.interceptor.TokenRefreshInterceptor;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.CorsRegistry;
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
    
    @Autowired
    private AdminInterceptor adminInterceptor;

    @Override
    public void addInterceptors(InterceptorRegistry registry) {
        // 第一层：Token刷新拦截器（先执行，检查是否需要刷新）
        registry.addInterceptor(tokenRefreshInterceptor)
                .addPathPatterns("/com/**", "/stu/**", "/admin/**")
                .excludePathPatterns("/login", "/refresh-token")
                .excludePathPatterns("/swagger-ui.html", "/swagger-resources/**", "/v2/api-docs", "/webjars/**") // 如果有Swagger也排除
                .order(1);

        // 第二层：认证拦截器（后执行，验证Token有效性）
        registry.addInterceptor(authInterceptor)
                .addPathPatterns("/com/**", "/stu/**", "/admin/**")
                .excludePathPatterns("/login", "/refresh-token")
                .excludePathPatterns("/swagger-ui.html", "/swagger-resources/**", "/v2/api-docs", "/webjars/**")
                .order(2);

        // 第三层：管理员权限拦截器（最后执行，验证管理员权限）
        registry.addInterceptor(adminInterceptor)
                .addPathPatterns("/admin/**")
                .order(3);
    }

    //跨域
    @Override
    public void addCorsMappings(CorsRegistry registry) {
        registry.addMapping("/**")
                .allowedOriginPatterns("*")  // 更安全的通配符处理
                .allowedMethods("GET", "POST", "PUT", "DELETE", "OPTIONS")
                .allowedHeaders("*")
                .allowCredentials(true)  // 如需携带凭证则设为true
                .maxAge(3600);
    }

}
