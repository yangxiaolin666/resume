package com.resume.utils;

import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;

/**
 * BCrypt密码加密工具类
 */
public class PasswordUtil {
    
    private static final BCryptPasswordEncoder encoder = new BCryptPasswordEncoder();
    
    /**
     * 加密密码
     * @param rawPassword 原始密码
     * @return 加密后的密码
     */
    public static String encode(String rawPassword) {
        return encoder.encode(rawPassword);
    }
    
    /**
     * 验证密码
     * @param rawPassword 原始密码
     * @param encodedPassword 加密后的密码
     * @return true-匹配，false-不匹配
     */
    public static boolean matches(String rawPassword, String encodedPassword) {
        return encoder.matches(rawPassword, encodedPassword);
    }
    
    /**
     * 测试方法：生成BCrypt密码
     */
    public static void main(String[] args) {

        System.out.println("2024081125: " + encode("2024081125"));
        System.out.println("company01: " + encode("company01"));
        System.out.println("admin123: " + encode("admin123"));
        System.out.println("c1: " + encode("c1"));
    }
}
