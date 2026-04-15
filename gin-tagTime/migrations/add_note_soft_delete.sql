-- 为 notes 表添加软删除字段
-- 执行时间：2024-XX-XX

-- 添加软删除标记字段
ALTER TABLE `notes` ADD COLUMN `is_deleted` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '软删除标记：0-正常，1-已删除' AFTER `images`;

-- 添加删除时间字段
ALTER TABLE `notes` ADD COLUMN `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间' AFTER `is_deleted`;

-- 为软删除字段添加索引，提高查询性能
ALTER TABLE `notes` ADD INDEX `idx_is_deleted` (`is_deleted`);

-- 为删除时间字段添加索引
ALTER TABLE `notes` ADD INDEX `idx_deleted_at` (`deleted_at`);

-- 查询验证
-- SELECT * FROM notes WHERE is_deleted = 0; -- 查询正常笔记
-- SELECT * FROM notes WHERE is_deleted = 1; -- 查询回收站笔记
