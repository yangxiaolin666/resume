package com.resume;

import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.crypto.password.PasswordEncoder;

public class test {
    //@Autowired
    private PasswordEncoder passwordEncoder; // 使用Spring容器中的Bean

    @Test
    void updatePassword() {
        String rawPassword = "yourPassword";
        String encodedPassword = passwordEncoder.encode(rawPassword);
        System.out.println(encodedPassword); // 将此字符串更新到数据库
    }
}
