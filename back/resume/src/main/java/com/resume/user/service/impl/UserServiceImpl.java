package com.resume.user.service.impl;

import com.resume.user.service.UserService;
import com.resume.model.dto.LoginDTO;
import com.resume.model.dto.LoginVO;
import com.resume.model.entity.User;
import com.resume.user.mapper.UserMapper;
import com.resume.utils.JwtUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.stereotype.Service;

import java.util.concurrent.TimeUnit;

/**
 * 用户服务实现类
 */
@Service
public class UserServiceImpl implements UserService {
    
    @Autowired
    private UserMapper userMapper;
    
    @Autowired
    private StringRedisTemplate redisTemplate;
    
    // BCrypt密码加密器
    private final BCryptPasswordEncoder passwordEncoder = new BCryptPasswordEncoder();
    
    // 登录尝试次数限制key前缀
    private static final String LOGIN_ATTEMPT_PREFIX = "login_attempt:";
    // 最大登录尝试次数
    private static final int MAX_LOGIN_ATTEMPTS = 5;
    // 登录尝试锁定时间（30分钟）
    private static final long LOCK_TIME = 30;
    
    @Override
    public LoginVO login(LoginDTO loginDTO) {
        String username = loginDTO.getUsername();
        String password = loginDTO.getPassword();
        String userType = loginDTO.getUserType();

        // 检查登录尝试次数
        String attemptKey = LOGIN_ATTEMPT_PREFIX + username;
        String attemptCount = redisTemplate.opsForValue().get(attemptKey);
        //redisTemplate.delete(attemptKey);
//        if (attemptCount != null && Integer.parseInt(attemptCount) >= MAX_LOGIN_ATTEMPTS) {
//            throw new RuntimeException("42901:登录尝试次数过多，请30分钟后再试");
//        }
        
        // 查询用户
        User user = userMapper.findByUsername(username);
        
        // 用户不存在
        if (user == null) {

            throw new RuntimeException("40005:用户名或密码错误");
        }

        
        // 校验用户类型是否一致
        if (!user.getUserType().equals(userType)) {
            throw new RuntimeException("401:用户身份不匹配，请选择正确的身份登录");
        }
        System.out.println(passwordEncoder.encode(password));
        // 验证密码（使用BCrypt）
        if (!passwordEncoder.matches(password, user.getPassword())) {
            // 增加登录尝试次数
            incrementLoginAttempt(attemptKey);
            throw new RuntimeException("40005:用户名或密码错误");
        }

//        if(!password.equals(user.getPassword())){
//            throw new RuntimeException("40005:密码错误");
//        }
        
        // 登录成功，清除登录尝试记录
        redisTemplate.delete(attemptKey);
        
        // 生成Access Token和Refresh Token
        String accessToken = JwtUtil.generateAccessToken(user.getUserId(), user.getUsername(), user.getUserType());
        String refreshToken = JwtUtil.generateRefreshToken(user.getUserId(), user.getUsername(), user.getUserType());
        
        // 构造返回结果
        LoginVO loginVO = new LoginVO();
        loginVO.setUserId(user.getUserId());
        loginVO.setUsername(user.getUsername());
        loginVO.setUserType(user.getUserType());
        loginVO.setToken(accessToken);
        loginVO.setRefreshToken(refreshToken);
        
        return loginVO;
    }
    
    /**
     * 增加登录尝试次数
     */
    private void incrementLoginAttempt(String key) {
        String count = redisTemplate.opsForValue().get(key);
        if (count == null) {
            redisTemplate.opsForValue().set(key, "1", LOCK_TIME, TimeUnit.MINUTES);
        } else {
            redisTemplate.opsForValue().increment(key);
        }
    }
    
    @Override
    public String refreshToken(String refreshToken) {
        try {
            // 验证Refresh Token
            com.auth0.jwt.interfaces.DecodedJWT jwt = JwtUtil.verifyToken(refreshToken);
            
            // 检查Token类型
            String tokenType = JwtUtil.getTokenType(refreshToken);
            if (!"refresh".equals(tokenType)) {
                throw new RuntimeException("401:Token类型错误");
            }
            
            // 获取用户信息
            Integer userId = JwtUtil.getUserId(refreshToken);
            String username = JwtUtil.getUsername(refreshToken);
            String userType = JwtUtil.getUserType(refreshToken);
            
            // 生成新的Access Token
            return JwtUtil.generateAccessToken(userId, username, userType);
        } catch (Exception e) {
            throw new RuntimeException("401:Refresh Token无效或已过期");
        }
    }
}
