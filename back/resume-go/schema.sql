-- resume-go 本地运行补充表结构（原基础库 resume.sql 中未包含这些表）
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
  INDEX `idx_com_id`(`com_id` ASC) USING BTREE,
  INDEX `idx_status_created`(`status` ASC, `created_at` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

CREATE TABLE IF NOT EXISTS `com_like_stu` (
  `id` int NOT NULL AUTO_INCREMENT,
  `stu_id` int NULL DEFAULT NULL,
  `com_id` int NULL DEFAULT NULL,
  `create_time` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_stu_com`(`stu_id`, `com_id`) USING BTREE,
  INDEX `idx_com_time`(`com_id`, `create_time`) USING BTREE
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

-- ----------------------------
-- 既有数据库的索引补充（幂等执行）
-- 说明：MySQL 8 的 CREATE INDEX 不支持 IF NOT EXISTS，
-- 所以这里查 information_schema.STATISTICS，不存在才执行 ALTER。
-- ----------------------------
SET @has_job_status_idx = (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'job_postings' AND INDEX_NAME = 'idx_status_created'
);
SET @add_job_status_idx = IF(@has_job_status_idx = 0,
  'ALTER TABLE `job_postings` ADD INDEX `idx_status_created` (`status`, `created_at`)',
  'SELECT 1');
PREPARE stmt_job_status_idx FROM @add_job_status_idx;
EXECUTE stmt_job_status_idx;
DEALLOCATE PREPARE stmt_job_status_idx;

SET @has_like_idx = (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'com_like_stu' AND INDEX_NAME = 'idx_com_time'
);
SET @add_like_idx = IF(@has_like_idx = 0,
  'ALTER TABLE `com_like_stu` ADD INDEX `idx_com_time` (`com_id`, `create_time`)',
  'SELECT 1');
PREPARE stmt_like_idx FROM @add_like_idx;
EXECUTE stmt_like_idx;
DEALLOCATE PREPARE stmt_like_idx;

-- ----------------------------
-- RabbitMQ 消费幂等表
-- event_id 是事件唯一标识（本项目为 "chat.message.created:{messageId}"）。
-- 消费者先 INSERT IGNORE，插入成功才执行业务，从数据库层面保证「只消费一次」。
-- ----------------------------
CREATE TABLE IF NOT EXISTS `mq_consumed_event` (
  `event_id` varchar(128) NOT NULL COMMENT '事件唯一ID',
  `consumed_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`event_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = 'MQ 消费幂等记录' ROW_FORMAT = Dynamic;

-- ----------------------------
-- 离线消息通知表
-- 由 RabbitMQ 消费者异步写入；即使丢失也不影响聊天记录本身（最终一致）。
-- ----------------------------
CREATE TABLE IF NOT EXISTS `chat_notification` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `message_id` bigint NOT NULL COMMENT '关联 chat_message.id',
  `receiver_id` int NOT NULL COMMENT '接收者 user_id',
  `receiver_type` varchar(20) NOT NULL COMMENT 'student/company',
  `sender_id` int NOT NULL COMMENT '发送者 user_id',
  `sender_type` varchar(20) NOT NULL COMMENT 'student/company',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '通知标题',
  `content` varchar(500) NOT NULL DEFAULT '' COMMENT '消息摘要',
  `is_read` tinyint NOT NULL DEFAULT 0 COMMENT '0未读 1已读',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_notification_message`(`message_id`, `receiver_id`, `receiver_type`) USING BTREE,
  INDEX `idx_notification_inbox`(`receiver_id`, `receiver_type`, `is_read`, `id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '离线消息通知' ROW_FORMAT = Dynamic;

-- ----------------------------
-- 简历投递表
-- 一条记录 = 一个学生投递给一个职位。
-- resume_url 保存「投递时的简历快照」，学生之后改简历不影响已投递记录。
-- ----------------------------
CREATE TABLE IF NOT EXISTS `job_application` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `job_id` int NOT NULL COMMENT 'job_postings.id',
  `com_id` int NOT NULL COMMENT '冗余企业ID，便于企业端按 com_id 查询和鉴权',
  `student_id` int NOT NULL COMMENT 'student.id',
  `resume_url` varchar(500) NOT NULL DEFAULT '' COMMENT '投递时的简历地址快照',
  `status` varchar(20) NOT NULL DEFAULT 'pending' COMMENT 'pending/viewed/interview/offer/rejected',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_job_student`(`job_id`, `student_id`) USING BTREE COMMENT '同一学生不能重复投递同一职位',
  INDEX `idx_student`(`student_id`, `id`) USING BTREE,
  INDEX `idx_company_status`(`com_id`, `status`, `id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '简历投递记录' ROW_FORMAT = Dynamic;
