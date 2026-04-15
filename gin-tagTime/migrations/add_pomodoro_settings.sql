-- 添加番茄钟配置字段到 time_entries 表
-- 执行日期: 2024-03-24

ALTER TABLE `time_entries` 
ADD COLUMN `work_minutes` INT NOT NULL DEFAULT 25 COMMENT '番茄钟工作时长（分钟）' AFTER `timer_mode`,
ADD COLUMN `break_minutes` INT NOT NULL DEFAULT 5 COMMENT '番茄钟休息时长（分钟）' AFTER `work_minutes`;

-- 说明：
-- work_minutes: 番茄钟工作时长，默认25分钟
-- break_minutes: 番茄钟休息时长，默认5分钟
-- 这两个字段允许用户自定义番茄钟的工作和休息时间
