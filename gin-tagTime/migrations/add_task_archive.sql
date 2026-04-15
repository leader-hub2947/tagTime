-- 添加任务归档功能
-- 1. 添加 archived_at 字段
ALTER TABLE `tasks` ADD COLUMN `archived_at` DATETIME DEFAULT NULL COMMENT '归档时间';

-- 2. 添加索引以优化归档任务查询
ALTER TABLE `tasks` ADD INDEX `idx_archived_at` (`archived_at`);

-- 3. 更新状态字段注释（状态：0-未开始，1-进行中，2-已完成，3-已归档）
-- 注意：MySQL 不支持直接修改注释，这里仅作说明
-- 实际使用时，状态 3 表示已归档

-- 4. 可选：将现有已完成超过30天的任务自动归档（根据需求决定是否执行）
-- UPDATE `tasks` 
-- SET `status` = 3, `archived_at` = NOW() 
-- WHERE `status` = 2 
-- AND `completed_at` IS NOT NULL 
-- AND `completed_at` < DATE_SUB(NOW(), INTERVAL 30 DAY);
