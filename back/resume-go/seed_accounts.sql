-- 本地演示账号（可重复执行）
-- 学生 x1 / x1    企业 q1 / q1    管理员 g1 / g1
SET NAMES utf8mb4;

-- 学生 x1
INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'x1', '$2a$10$lcPfyxS5OJvMSeOxpMj9FuW3EkZzCysGDS.lPJE4Qlr2xaBzLu5PW', 'student', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'x1');

SET @x1_uid = (SELECT user_id FROM user WHERE username = 'x1' LIMIT 1);

INSERT INTO student (id, stu_id, name, age, gender, nation, major, degree, graduation_year, internship, user_id)
SELECT @x1_uid, 1, '王一', 22, '男', '汉族', '软件工程', '本科生', 2027, 1, @x1_uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM student WHERE user_id = @x1_uid);

-- 企业 q1
INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'q1', '$2a$10$n.eEAklTrOLBZK4O4uuCLedwFWO.rkEAdhFQfR9Ng7ayWMJyeEJ8u', 'company', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'q1');

SET @q1_uid = (SELECT user_id FROM user WHERE username = 'q1' LIMIT 1);

INSERT INTO company (com_name, com_region, user_id)
SELECT '企业一', '北京', @q1_uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM company WHERE user_id = @q1_uid);

-- 管理员 g1
INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'g1', '$2a$10$tnn5QnL8dfC0udKGw7002.e.v888KqUL550cWvOjspXfy7.GcGAhK', 'admin', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'g1');

-- ----------------------------
-- 批量演示账号：学生 x2~x10（王二~王十）
-- 密码与账号同名（x2 的密码就是 x2），与 x1 的规则一致
-- ----------------------------
INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'x2', '$2a$10$DRQthfiMcYFL3hemEaQgautwucAPmhIJVpdA5Da./B196HUvjk7by', 'student', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'x2');
SET @uid = (SELECT user_id FROM user WHERE username = 'x2' LIMIT 1);
INSERT INTO student (stu_id, name, age, gender, nation, major, degree, graduation_year, internship, user_id)
SELECT 2, '王二', 22, '男', '汉族', '软件工程', '本科生', 2027, 0, @uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM student WHERE user_id = @uid);

INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'x3', '$2a$10$2mzrKce1/eUtgvbALp4sFOWZSXRIPigKLftUipKIMWBzOOD8ieFJe', 'student', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'x3');
SET @uid = (SELECT user_id FROM user WHERE username = 'x3' LIMIT 1);
INSERT INTO student (stu_id, name, age, gender, nation, major, degree, graduation_year, internship, user_id)
SELECT 3, '王三', 22, '女', '汉族', '计算机科学与技术', '本科生', 2027, 1, @uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM student WHERE user_id = @uid);

INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'x4', '$2a$10$qPpOAOJ6pCVl6cVD7IA.9ul5mcLH9eh5v0OeTU2njvvVgPBrBD/Eq', 'student', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'x4');
SET @uid = (SELECT user_id FROM user WHERE username = 'x4' LIMIT 1);
INSERT INTO student (stu_id, name, age, gender, nation, major, degree, graduation_year, internship, user_id)
SELECT 4, '王四', 21, '男', '汉族', '人工智能', '本科生', 2028, 1, @uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM student WHERE user_id = @uid);

INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'x5', '$2a$10$al.8vznhdaMVwDX1PiB4UOD.J1PHfz/1ndE8UWusPuj8KSI62LrE.', 'student', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'x5');
SET @uid = (SELECT user_id FROM user WHERE username = 'x5' LIMIT 1);
INSERT INTO student (stu_id, name, age, gender, nation, major, degree, graduation_year, internship, user_id)
SELECT 5, '王五', 23, '女', '汉族', '数据科学与大数据技术', '硕士生', 2026, 2, @uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM student WHERE user_id = @uid);

INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'x6', '$2a$10$f1AFb5zhguuiMn5d44Mtq.Aw7TwjV4xjCmxwTEieVuhgYIfXe6kD6', 'student', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'x6');
SET @uid = (SELECT user_id FROM user WHERE username = 'x6' LIMIT 1);
INSERT INTO student (stu_id, name, age, gender, nation, major, degree, graduation_year, internship, user_id)
SELECT 6, '王六', 22, '男', '汉族', '网络工程', '本科生', 2027, 0, @uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM student WHERE user_id = @uid);

INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'x7', '$2a$10$ZShH04zGDFovUHH8FCS82.wbsw811ldzViwgE5TYNO.7KJPgHtAui', 'student', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'x7');
SET @uid = (SELECT user_id FROM user WHERE username = 'x7' LIMIT 1);
INSERT INTO student (stu_id, name, age, gender, nation, major, degree, graduation_year, internship, user_id)
SELECT 7, '王七', 21, '女', '汉族', '信息安全', '本科生', 2028, 0, @uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM student WHERE user_id = @uid);

INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'x8', '$2a$10$y6FSEVII/AVlF7gSl9m72O.7kWkV2WSoZONJSHMJhJYty1BEcPPaO', 'student', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'x8');
SET @uid = (SELECT user_id FROM user WHERE username = 'x8' LIMIT 1);
INSERT INTO student (stu_id, name, age, gender, nation, major, degree, graduation_year, internship, user_id)
SELECT 8, '王八', 24, '男', '汉族', '软件工程', '硕士生', 2026, 3, @uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM student WHERE user_id = @uid);

INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'x9', '$2a$10$VK708cNGDbPS4T6NSvVcJu95Jud/1X9U3Ak9C0QALMBaPOTpWhyfi', 'student', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'x9');
SET @uid = (SELECT user_id FROM user WHERE username = 'x9' LIMIT 1);
INSERT INTO student (stu_id, name, age, gender, nation, major, degree, graduation_year, internship, user_id)
SELECT 9, '王九', 22, '女', '汉族', '物联网工程', '本科生', 2027, 1, @uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM student WHERE user_id = @uid);

INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'x10', '$2a$10$Ht00knEsOKkNbSaf49CZ5.vyPlsR/EuGibGl68wRf/SIbxGxXV8BK', 'student', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'x10');
SET @uid = (SELECT user_id FROM user WHERE username = 'x10' LIMIT 1);
INSERT INTO student (stu_id, name, age, gender, nation, major, degree, graduation_year, internship, user_id)
SELECT 10, '王十', 22, '男', '汉族', '软件工程', '本科生', 2027, 0, @uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM student WHERE user_id = @uid);

-- ----------------------------
-- 批量演示账号：企业 q2~q10（企业二~企业十）
-- 密码与账号同名（q2 的密码就是 q2），与 q1 的规则一致
-- ----------------------------
INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'q2', '$2a$10$.oFBBO9Kw1upSWqUZnWgPeL3dScaUn2fE1zIlq5vrEu8XxsTwLP3G', 'company', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'q2');
SET @uid = (SELECT user_id FROM user WHERE username = 'q2' LIMIT 1);
INSERT INTO company (com_name, com_region, user_id)
SELECT '企业二', '北京', @uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM company WHERE user_id = @uid);

INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'q3', '$2a$10$k6c26Petm46QtrMKpetkKugxfaooNDyEle3lmws8On7xYhkh5qj2W', 'company', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'q3');
SET @uid = (SELECT user_id FROM user WHERE username = 'q3' LIMIT 1);
INSERT INTO company (com_name, com_region, user_id)
SELECT '企业三', '上海', @uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM company WHERE user_id = @uid);

INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'q4', '$2a$10$7Rr.Mt/JDDJ9UlAmI9GS8eda6dMYqk2PrmWSPQHkfvM.7FinAzVfK', 'company', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'q4');
SET @uid = (SELECT user_id FROM user WHERE username = 'q4' LIMIT 1);
INSERT INTO company (com_name, com_region, user_id)
SELECT '企业四', '广州', @uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM company WHERE user_id = @uid);

INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'q5', '$2a$10$fSrIQNMx7VdxytwvyEqZMOoX2t9gammPODCXoe.Qxcwmj9dVQa02S', 'company', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'q5');
SET @uid = (SELECT user_id FROM user WHERE username = 'q5' LIMIT 1);
INSERT INTO company (com_name, com_region, user_id)
SELECT '企业五', '深圳', @uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM company WHERE user_id = @uid);

INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'q6', '$2a$10$dIK4JSaHoBYle9aHXpSBpO/Y.exsPh/TL3lbL2kvbfQer7TCxQuCC', 'company', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'q6');
SET @uid = (SELECT user_id FROM user WHERE username = 'q6' LIMIT 1);
INSERT INTO company (com_name, com_region, user_id)
SELECT '企业六', '杭州', @uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM company WHERE user_id = @uid);

INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'q7', '$2a$10$QFoSC82NQhWlW5NYviE5YuKADm0WAQhocZIjyGSI6B44joYcoeseK', 'company', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'q7');
SET @uid = (SELECT user_id FROM user WHERE username = 'q7' LIMIT 1);
INSERT INTO company (com_name, com_region, user_id)
SELECT '企业七', '成都', @uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM company WHERE user_id = @uid);

INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'q8', '$2a$10$UbiIRWtYECKcLtmab8unouUBVdTmh.DHnK7U2clNFq2alSVo2OMUu', 'company', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'q8');
SET @uid = (SELECT user_id FROM user WHERE username = 'q8' LIMIT 1);
INSERT INTO company (com_name, com_region, user_id)
SELECT '企业八', '武汉', @uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM company WHERE user_id = @uid);

INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'q9', '$2a$10$6IrZ/B/QtHBbTGfeisH.N.TtOeu6651AyAlmhoWVXLcTLbKqD2lHm', 'company', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'q9');
SET @uid = (SELECT user_id FROM user WHERE username = 'q9' LIMIT 1);
INSERT INTO company (com_name, com_region, user_id)
SELECT '企业九', '南京', @uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM company WHERE user_id = @uid);

INSERT INTO user (username, password, user_type, create_time, update_time)
SELECT 'q10', '$2a$10$rqXGDHXmQJ77d4MZMyZN2eVnoRVFgWYiY.g.10Tv2SgnZefZE/BTC', 'company', NOW(), NOW()
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'q10');
SET @uid = (SELECT user_id FROM user WHERE username = 'q10' LIMIT 1);
INSERT INTO company (com_name, com_region, user_id)
SELECT '企业十', '西安', @uid
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM company WHERE user_id = @uid);

-- ----------------------------
-- 演示学生的技能标签（学生卡片上的“能力标签”，可重复执行）
-- ----------------------------
INSERT INTO stu_speciality (stu_id, speciality)
SELECT s.id, v.skill
FROM (
    SELECT 'x1' AS username, 'Java' AS skill
    UNION ALL SELECT 'x1', 'MySQL'
    UNION ALL SELECT 'x1', 'Spring Boot'
    UNION ALL SELECT 'x2', 'Java'
    UNION ALL SELECT 'x2', 'Vue'
    UNION ALL SELECT 'x2', 'MySQL'
    UNION ALL SELECT 'x3', 'Python'
    UNION ALL SELECT 'x3', 'MySQL'
    UNION ALL SELECT 'x3', '数据分析'
    UNION ALL SELECT 'x4', 'Python'
    UNION ALL SELECT 'x4', '机器学习'
    UNION ALL SELECT 'x4', 'TensorFlow'
    UNION ALL SELECT 'x5', 'Java'
    UNION ALL SELECT 'x5', 'Redis'
    UNION ALL SELECT 'x5', 'Docker'
    UNION ALL SELECT 'x6', 'C++'
    UNION ALL SELECT 'x6', 'Linux'
    UNION ALL SELECT 'x6', '计算机网络'
    UNION ALL SELECT 'x7', 'Python'
    UNION ALL SELECT 'x7', 'Linux'
    UNION ALL SELECT 'x7', '网络安全'
    UNION ALL SELECT 'x8', 'Java'
    UNION ALL SELECT 'x8', 'Spring Boot'
    UNION ALL SELECT 'x8', 'Redis'
    UNION ALL SELECT 'x8', 'MySQL'
    UNION ALL SELECT 'x9', 'C'
    UNION ALL SELECT 'x9', '嵌入式开发'
    UNION ALL SELECT 'x9', 'Linux'
    UNION ALL SELECT 'x10', 'Vue'
    UNION ALL SELECT 'x10', 'JavaScript'
    UNION ALL SELECT 'x10', 'Node.js'
) v
JOIN user u ON u.username = v.username
JOIN student s ON s.user_id = u.user_id
WHERE NOT EXISTS (
    SELECT 1 FROM stu_speciality t WHERE t.stu_id = s.id AND t.speciality = v.skill
);
