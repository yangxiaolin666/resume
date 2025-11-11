/*
 Navicat Premium Data Transfer

 Source Server         : winmysql
 Source Server Type    : MySQL
 Source Server Version : 80036
 Source Host           : localhost:3306
 Source Schema         : resume

 Target Server Type    : MySQL
 Target Server Version : 80036
 File Encoding         : 65001

 Date: 11/11/2025 12:57:07
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for com_collect
-- ----------------------------
DROP TABLE IF EXISTS `com_collect`;
CREATE TABLE `com_collect`  (
  `com_id` int NOT NULL,
  `stu_id` int NULL DEFAULT NULL,
  PRIMARY KEY (`com_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of com_collect
-- ----------------------------

-- ----------------------------
-- Table structure for company
-- ----------------------------
DROP TABLE IF EXISTS `company`;
CREATE TABLE `company`  (
  `id` int(10) UNSIGNED ZEROFILL NOT NULL AUTO_INCREMENT,
  `com_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `com_region` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `user_id` int NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `user_com`(`user_id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of company
-- ----------------------------
INSERT INTO `company` VALUES (0000000001, '大连公司', '大连', NULL);
INSERT INTO `company` VALUES (0000000003, '阿里巴巴', '杭州', 7);
INSERT INTO `company` VALUES (0000000004, '字节跳动', '北京', 8);

-- ----------------------------
-- Table structure for role
-- ----------------------------
DROP TABLE IF EXISTS `role`;
CREATE TABLE `role`  (
  `id` int NOT NULL,
  `type` varchar(25) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of role
-- ----------------------------

-- ----------------------------
-- Table structure for stu_certificate
-- ----------------------------
DROP TABLE IF EXISTS `stu_certificate`;
CREATE TABLE `stu_certificate`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `stu_id` bigint NOT NULL COMMENT '学生主键，关联已有学生表.id',
  `cert_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '证书全称，如“软件技术资格水平认证-中级软件设计师”',
  `level` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '级别：初级/中级/高级',
  `cert_no` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '证书编号',
  `issue_date` date NULL DEFAULT NULL COMMENT '获证日期',
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 14 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '学生计算机类证书' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of stu_certificate
-- ----------------------------
INSERT INTO `stu_certificate` VALUES (1, 1, '软件技术资格水平认证-中级软件设计师', '中级', NULL, NULL, '2025-10-16 16:29:22', '2025-10-16 16:29:22');
INSERT INTO `stu_certificate` VALUES (2, 2, '软件技术资格水平认证-中级软件设计师', '中级', NULL, NULL, '2025-10-16 16:29:41', '2025-10-16 16:29:41');
INSERT INTO `stu_certificate` VALUES (11, 3, '软考中级软件设计师', '中级', NULL, NULL, '2025-11-09 16:11:15', '2025-11-09 16:11:15');

-- ----------------------------
-- Table structure for stu_internship
-- ----------------------------
DROP TABLE IF EXISTS `stu_internship`;
CREATE TABLE `stu_internship`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `company` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '实习公司',
  `experience` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '实习经历',
  `stu_id` int NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 30 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of stu_internship
-- ----------------------------
INSERT INTO `stu_internship` VALUES (28, '腾讯', '后端开发实习', 3);
INSERT INTO `stu_internship` VALUES (29, '腾讯', '后端开发实习', 3);

-- ----------------------------
-- Table structure for stu_job
-- ----------------------------
DROP TABLE IF EXISTS `stu_job`;
CREATE TABLE `stu_job`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `post` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `stu_id` int NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 37 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of stu_job
-- ----------------------------
INSERT INTO `stu_job` VALUES (1, '后端开发', 1);
INSERT INTO `stu_job` VALUES (2, '前端开发', 2);
INSERT INTO `stu_job` VALUES (35, '后端开发', 3);
INSERT INTO `stu_job` VALUES (36, '全栈开发', 3);

-- ----------------------------
-- Table structure for stu_language
-- ----------------------------
DROP TABLE IF EXISTS `stu_language`;
CREATE TABLE `stu_language`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `stu_id` bigint NOT NULL COMMENT '学生主键，关联已有学生表.id',
  `language` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '语言，如 English / Japanese / Korean',
  `level` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '等级，如 CET-4 / N1 / TOPIK II 等',
  `cert_no` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `issue_date` date NULL DEFAULT NULL,
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 30 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '学生语言等级/证书' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of stu_language
-- ----------------------------
INSERT INTO `stu_language` VALUES (1, 1, 'English', '大学四级', NULL, NULL, '2025-10-16 16:30:58', '2025-10-16 16:30:58');
INSERT INTO `stu_language` VALUES (24, 3, '英语', '六级', NULL, NULL, '2025-11-09 16:11:15', '2025-11-09 16:11:15');
INSERT INTO `stu_language` VALUES (25, 3, '日语', 'N2', NULL, NULL, '2025-11-09 16:11:15', '2025-11-09 16:11:15');

-- ----------------------------
-- Table structure for stu_region
-- ----------------------------
DROP TABLE IF EXISTS `stu_region`;
CREATE TABLE `stu_region`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `stu_id` int NOT NULL,
  `province` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `city` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 54 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of stu_region
-- ----------------------------
INSERT INTO `stu_region` VALUES (1, 1, '辽宁', '大连');
INSERT INTO `stu_region` VALUES (2, 2, NULL, '北京');
INSERT INTO `stu_region` VALUES (3, 2, '辽宁', '沈阳');
INSERT INTO `stu_region` VALUES (42, 3, NULL, '北京');
INSERT INTO `stu_region` VALUES (43, 3, NULL, '上海');
INSERT INTO `stu_region` VALUES (44, 3, '广东', NULL);
INSERT INTO `stu_region` VALUES (45, 3, '江苏', NULL);

-- ----------------------------
-- Table structure for stu_resume
-- ----------------------------
DROP TABLE IF EXISTS `stu_resume`;
CREATE TABLE `stu_resume`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `stu_id` int NULL DEFAULT NULL,
  `resume` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 4 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of stu_resume
-- ----------------------------
INSERT INTO `stu_resume` VALUES (1, 3, 'uploads\\resume\\20230001_20251109164749_5a645a16.pdf');
INSERT INTO `stu_resume` VALUES (2, 3, 'uploads\\resume\\20230001_20251109173432_2a8be2ae.pdf');
INSERT INTO `stu_resume` VALUES (3, 3, 'uploads\\resume\\20230001_20251109173954_494fd89e.pdf');

-- ----------------------------
-- Table structure for stu_speciality
-- ----------------------------
DROP TABLE IF EXISTS `stu_speciality`;
CREATE TABLE `stu_speciality`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `stu_id` int NULL DEFAULT NULL,
  `speciality` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '专业技能',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 46 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of stu_speciality
-- ----------------------------
INSERT INTO `stu_speciality` VALUES (1, 1, 'java');
INSERT INTO `stu_speciality` VALUES (2, 1, 'python');
INSERT INTO `stu_speciality` VALUES (3, 2, 'cpp');
INSERT INTO `stu_speciality` VALUES (43, 3, 'Java');
INSERT INTO `stu_speciality` VALUES (44, 3, 'Python');
INSERT INTO `stu_speciality` VALUES (45, 3, 'MySQL');

-- ----------------------------
-- Table structure for student
-- ----------------------------
DROP TABLE IF EXISTS `student`;
CREATE TABLE `student`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `stu_id` int NOT NULL,
  `name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `age` int NULL DEFAULT NULL,
  `gender` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `nation` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `major` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '所属专业',
  `degree` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '学位',
  `graduation_year` int NULL DEFAULT NULL COMMENT '毕业年份',
  `internship` int NULL DEFAULT NULL COMMENT '实习次数',
  `user_id` int NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `user_stu`(`user_id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 4 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of student
-- ----------------------------
INSERT INTO `student` VALUES (5, 2024081104, '王网', 21, '女', '汉族', '软件工程', '本科生', 2028, 1, 5);

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user`  (
  `user_id` int NOT NULL AUTO_INCREMENT,
  `username` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `password` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `user_type` varchar(25) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `create_time` datetime NULL DEFAULT NULL,
  `update_time` datetime NULL DEFAULT NULL,
  PRIMARY KEY (`user_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 4 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of user
-- ----------------------------
INSERT INTO `user` VALUES (2, '大连公司', '123456', 'company', NULL, NULL);
INSERT INTO `user` VALUES (3, 'admin', '$2a$10$Bgd/WqoNybgbpJRBnXpv3.yVF1i.vHe03pBeadSY53onOjM0eGZGy', 'admin', NULL, NULL);
INSERT INTO `user` VALUES (5, '2024081104', '$2a$10$gdOan5YroYz1.hZ6AsEZZOELlWRGU3lHDrr/ReCyHu8btadhaY/l.', 'student', '2025-11-10 20:25:16', '2025-11-10 20:25:16');
INSERT INTO `user` VALUES (6, '腾讯科技', '$2a$10$FlMAmkggs695DgcP.Cnl1.feVaOFkRYuAMfks2UCCKK5LDPuHhodC', 'company', '2025-11-10 20:54:12', '2025-11-10 20:54:12');
INSERT INTO `user` VALUES (7, '阿里巴巴', '$2a$10$DQU3.REhGEIQl4p5otM.w.Hg0ZMZnu88fFHna0qUS/.RQ4dJDeYIe', 'company', '2025-11-10 20:54:12', '2025-11-10 20:54:12');
INSERT INTO `user` VALUES (8, '字节跳动', '$2a$10$9bzZnprLMGSUDkuDtjYqfe83G2dvcqSklNGh6Dt7cy/OfeoGAiArS', 'company', '2025-11-10 20:54:12', '2025-11-10 20:54:12');

SET FOREIGN_KEY_CHECKS = 1;
