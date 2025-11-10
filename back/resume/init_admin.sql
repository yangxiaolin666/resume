-- 管理员账号初始化脚本
-- 创建管理员账号，用户名: admin, 密码: admin123

-- 删除已存在的管理员账号（如果有）
DELETE FROM user WHERE user_type = 'admin';

-- 插入管理员账号
-- 注意：密码需要使用BCrypt加密，这里使用Spring Security BCryptPasswordEncoder加密后的密码
-- 原始密码: admin123
-- BCrypt加密后: $2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iAt6Z2EHeYoFCSU7Vvv4CVJgSGHq
INSERT INTO user (username, password, user_type, create_time, update_time) 
VALUES ('admin', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iAt6Z2EHeYoFCSU7Vvv4CVJgSGHq', 'admin', NOW(), NOW());

-- 验证插入结果
SELECT * FROM user WHERE user_type = 'admin';
