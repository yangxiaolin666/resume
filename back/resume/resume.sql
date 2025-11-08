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

 Date: 07/11/2025 21:27:33
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
  `id` int NOT NULL,
  `com_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `com_region` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `user_id` int NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `user_com`(`user_id` ASC) USING BTREE,
  CONSTRAINT `user_com` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of company
-- ----------------------------

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
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '学生计算机类证书' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of stu_certificate
-- ----------------------------
INSERT INTO `stu_certificate` VALUES (1, 1, '软件技术资格水平认证-中级软件设计师', '中级', NULL, NULL, '2025-10-16 16:29:22', '2025-10-16 16:29:22');
INSERT INTO `stu_certificate` VALUES (2, 2, '软件技术资格水平认证-中级软件设计师', '中级', NULL, NULL, '2025-10-16 16:29:41', '2025-10-16 16:29:41');

-- ----------------------------
-- Table structure for stu_internship
-- ----------------------------
DROP TABLE IF EXISTS `stu_internship`;
CREATE TABLE `stu_internship`  (
  `id` int NOT NULL,
  `company` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '实习公司',
  `experience` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '实习经历',
  `stu_id` int NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of stu_internship
-- ----------------------------

-- ----------------------------
-- Table structure for stu_job
-- ----------------------------
DROP TABLE IF EXISTS `stu_job`;
CREATE TABLE `stu_job`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `post` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `stu_id` int NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of stu_job
-- ----------------------------
INSERT INTO `stu_job` VALUES (1, '后端开发', 1);
INSERT INTO `stu_job` VALUES (2, '前端开发', 2);

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
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '学生语言等级/证书' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of stu_language
-- ----------------------------
INSERT INTO `stu_language` VALUES (1, 1, 'English', '大学四级', NULL, NULL, '2025-10-16 16:30:58', '2025-10-16 16:30:58');

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
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of stu_region
-- ----------------------------
INSERT INTO `stu_region` VALUES (1, 1, '辽宁', '大连');
INSERT INTO `stu_region` VALUES (2, 2, NULL, '北京');
INSERT INTO `stu_region` VALUES (3, 2, '辽宁', '沈阳');

-- ----------------------------
-- Table structure for stu_resume
-- ----------------------------
DROP TABLE IF EXISTS `stu_resume`;
CREATE TABLE `stu_resume`  (
  `id` int NOT NULL,
  `stu_id` int NULL DEFAULT NULL,
  `resume` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of stu_resume
-- ----------------------------

-- ----------------------------
-- Table structure for stu_speciality
-- ----------------------------
DROP TABLE IF EXISTS `stu_speciality`;
CREATE TABLE `stu_speciality`  (
  `id` int NOT NULL,
  `stu_id` int NULL DEFAULT NULL,
  `speciality` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '专业技能',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of stu_speciality
-- ----------------------------
INSERT INTO `stu_speciality` VALUES (1, 1, 'java');
INSERT INTO `stu_speciality` VALUES (2, 1, 'python');
INSERT INTO `stu_speciality` VALUES (3, 2, 'cpp');

-- ----------------------------
-- Table structure for student
-- ----------------------------
DROP TABLE IF EXISTS `student`;
CREATE TABLE `student`  (
  `id` int NOT NULL,
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
  INDEX `user_stu`(`user_id` ASC) USING BTREE,
  CONSTRAINT `user_stu` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of student
-- ----------------------------
INSERT INTO `student` VALUES (1, 2024081101, '张三', 18, '男', '汉', '计算机科学与技术', '本科生', 2028, NULL, NULL);
INSERT INTO `student` VALUES (2, 2024081102, '李四', 19, '女', '汉', '软件工程', '本科生', 2028, NULL, NULL);

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user`  (
  `user_id` int NOT NULL AUTO_INCREMENT,
  `username` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `password` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `user_type` varchar(25) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `create_time` datetime NULL DEFAULT NULL,
  `update_time` datetime NULL DEFAULT NULL,
  PRIMARY KEY (`user_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of user
-- ----------------------------

SET FOREIGN_KEY_CHECKS = 1;
