-- 检查 company 表中的数据
-- 执行这个 SQL 语句来查看数据库中是否有企业记录

-- 1. 查看 company 表中的所有数据
SELECT id, com_name, user_id
FROM company;

-- 2. 如果有企业数据，检查 userId 是否在 user 表中存在
-- SELECT c.id, c.com_name, c.user_id, u.id as user_id_in_user_table
-- FROM company c
-- LEFT JOIN user u ON c.user_id = u.id
-- ORDER BY c.id;

-- 3. 查看 user 表中的数据
-- SELECT id, username, user_type
-- FROM user
-- WHERE user_type = 'company' OR user_type = 'admin';

-- 4. 统计每个企业类型的用户数量
-- SELECT user_type, COUNT(*) as count
-- FROM user
-- GROUP BY user_type;
