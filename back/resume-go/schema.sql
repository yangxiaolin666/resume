-- resume-go 本地运行补充表结构（原 Java 项目中未包含在 resume.sql 里的表）
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `job_postings` (
  `id` int NOT NULL AUTO_INCREMENT,
  `com_id` int NULL DEFAULT NULL,
  `title` varchar(255) NULL DEFAULT NULL,
  `description` text NULL,
  `external_link` varchar(500) NULL DEFAULT NULL,
  `status` varchar(20) NULL DEFAULT 'active',
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_com_id`(`com_id` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

CREATE TABLE IF NOT EXISTS `com_like_stu` (
  `id` int NOT NULL AUTO_INCREMENT,
  `stu_id` int NULL DEFAULT NULL,
  `com_id` int NULL DEFAULT NULL,
  `create_time` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

CREATE TABLE IF NOT EXISTS `admin_all_skills` (
  `id` int NOT NULL AUTO_INCREMENT,
  `skillName` varchar(100) NULL DEFAULT NULL,
  `create_time` date NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

CREATE TABLE IF NOT EXISTS `language` (
  `language_name` varchar(50) NOT NULL,
  PRIMARY KEY (`language_name`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

CREATE TABLE IF NOT EXISTS `certificate` (
  `certificate_name` varchar(100) NOT NULL,
  PRIMARY KEY (`certificate_name`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

INSERT IGNORE INTO `admin_all_skills` (`skillName`, `create_time`) VALUES
('Java', CURDATE()), ('Python', CURDATE()), ('C++', CURDATE()), ('MySQL', CURDATE()), ('Vue', CURDATE());

INSERT IGNORE INTO `language` (`language_name`) VALUES
('英语'), ('日语'), ('韩语'), ('法语'), ('德语');

INSERT IGNORE INTO `certificate` (`certificate_name`) VALUES
('计算机二级'), ('软件设计师'), ('英语六级'), ('日语N2');

-- ----------------------------
-- 聊天消息表（学生 <-> 企业 私聊）
-- sender_type / receiver_type 只存 student / company
-- ----------------------------
CREATE TABLE IF NOT EXISTS `chat_message` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `sender_id` int NOT NULL COMMENT '发送者 user_id',
  `sender_type` varchar(20) NOT NULL COMMENT '发送者类型 student/company',
  `receiver_id` int NOT NULL COMMENT '接收者 user_id',
  `receiver_type` varchar(20) NOT NULL COMMENT '接收者类型 student/company',
  `content` text NOT NULL COMMENT '消息内容：文本，或图片地址',
  `is_read` tinyint NOT NULL DEFAULT 0 COMMENT '0未读 1已读',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_chat_pair`(`sender_id`, `sender_type`, `receiver_id`, `receiver_type`, `id`) USING BTREE,
  INDEX `idx_chat_inbox`(`receiver_id`, `receiver_type`, `is_read`, `id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '学生与企业聊天消息' ROW_FORMAT = Dynamic;

-- 兼容旧库：补充消息类型字段（text / image）
SET @has_msg_type = (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'chat_message' AND COLUMN_NAME = 'msg_type'
);
SET @add_msg_type = IF(@has_msg_type = 0,
  'ALTER TABLE `chat_message` ADD COLUMN `msg_type` varchar(20) NOT NULL DEFAULT ''text'' COMMENT ''消息类型 text/image'' AFTER `content`',
  'SELECT 1');
PREPARE stmt_msg_type FROM @add_msg_type;
EXECUTE stmt_msg_type;
DEALLOCATE PREPARE stmt_msg_type;
