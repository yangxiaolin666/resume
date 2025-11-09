-- 修复表结构，添加AUTO_INCREMENT
-- 执行此脚本前请确保对应表中没有重要数据，或先备份数据

-- 修复 stu_internship 表
ALTER TABLE `stu_internship` MODIFY `id` int NOT NULL AUTO_INCREMENT;

-- 修复 stu_speciality 表
ALTER TABLE `stu_speciality` MODIFY `id` int NOT NULL AUTO_INCREMENT;

-- 修复 stu_resume 表
ALTER TABLE `stu_resume` MODIFY `id` int NOT NULL AUTO_INCREMENT;

-- 修复 student 表（如果需要）
ALTER TABLE `student` MODIFY `id` int NOT NULL AUTO_INCREMENT;
