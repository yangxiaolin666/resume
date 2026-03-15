package com.resume.utils;

import com.auth0.jwt.JWT;
import com.auth0.jwt.algorithms.Algorithm;
import com.auth0.jwt.interfaces.DecodedJWT;
import com.auth0.jwt.interfaces.JWTVerifier;
import org.springframework.stereotype.Component;

import java.time.LocalDateTime;
import java.time.ZoneId;
import java.util.Date;

/**
 * JWT工具类
 */
@Component
public class JwtUtil {
    
    // JWT密钥
    private static final String JWT_SECRET = "resume_secret_key_2024";
    // Access Token有效期（2小时）
    private static final long ACCESS_TOKEN_EXPIRE_HOURS = 2;
    // Refresh Token有效期（7天）
    private static final long REFRESH_TOKEN_EXPIRE_DAYS = 7;
    
    /**
     * 生成Access Token（短期）
     * @param userId 用户ID
     * @param username 用户名
     * @param userType 用户类型
     * @return Token字符串
     */
    public static String generateAccessToken(Integer userId, String username, String userType) {
        LocalDateTime expireTime = LocalDateTime.now().plusHours(ACCESS_TOKEN_EXPIRE_HOURS);
        Date expireDate = Date.from(expireTime.atZone(ZoneId.systemDefault()).toInstant());
        
        return JWT.create()
                .withClaim("userId", userId)
                .withClaim("username", username)
                .withClaim("userType", userType)
                .withClaim("tokenType", "access")
                .withExpiresAt(expireDate)
                .sign(Algorithm.HMAC256(JWT_SECRET));
    }
    
    /**
     * 生成Refresh Token（长期）
     * @param userId 用户ID
     * @param username 用户名
     * @param userType 用户类型
     * @return Token字符串
     */
    public static String generateRefreshToken(Integer userId, String username, String userType) {
        LocalDateTime expireTime = LocalDateTime.now().plusDays(REFRESH_TOKEN_EXPIRE_DAYS);
        Date expireDate = Date.from(expireTime.atZone(ZoneId.systemDefault()).toInstant());
        
        return JWT.create()
                .withClaim("userId", userId)
                .withClaim("username", username)
                .withClaim("userType", userType)
                .withClaim("tokenType", "refresh")
                .withExpiresAt(expireDate)
                .sign(Algorithm.HMAC256(JWT_SECRET));
    }
    
    /**
     * 验证Token
     * @param token Token字符串
     * @return 解码后的JWT
     */
    public static DecodedJWT verifyToken(String token) {
        JWTVerifier verifier = JWT.require(Algorithm.HMAC256(JWT_SECRET)).build();
        return verifier.verify(token);
    }
    
    /**
     * 检查Token是否即将过期（在后30分钟内过期）
     * @param token Token字符串
     * @return true-即将过期，false-未过期
     */
    public static boolean isTokenExpiringSoon(String token) {
        try {
            DecodedJWT jwt = verifyToken(token);
            Date expiresAt = jwt.getExpiresAt();
            long currentTime = System.currentTimeMillis();
            long expiresTime = expiresAt.getTime();
            // 如果在30分钟内过期，返回true
            return (expiresTime - currentTime) < 30 * 60 * 1000;
        } catch (Exception e) {
            return true;
        }
    }
    
    /**
     * 从 Token中获取Token类型
     * @param token Token字符串
     * @return Token类型
     */
    public static String getTokenType(String token) {
        DecodedJWT jwt = verifyToken(token);
        return jwt.getClaim("tokenType").asString();
    }
    
    /**
     * 从Token中获取用户ID
     * @param token Token字符串
     * @return 用户ID
     */
    public static Integer getUserId(String token) {
        DecodedJWT jwt = verifyToken(token);
        return jwt.getClaim("userId").asInt();
    }
    
    /**
     * 从Token中获取用户名
     * @param token Token字符串
     * @return 用户名
     */
    public static String getUsername(String token) {
        DecodedJWT jwt = verifyToken(token);
        return jwt.getClaim("username").asString();
    }
    
    /**
     * 从Token中获取用户类型
     * @param token Token字符串
     * @return 用户类型
     */
    public static String getUserType(String token) {
        DecodedJWT jwt = verifyToken(token);
        return jwt.getClaim("userType").asString();
    }
}
