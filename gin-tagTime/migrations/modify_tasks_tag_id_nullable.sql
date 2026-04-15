-- 修改 tasks 表的 tag_id 字段为可空
-- 这样删除标签时可以将关联任务的 tag_id 设置为 null，而不是删除任务

-- 1. 先删除外键约束
ALTER TABLE `tasks` DROP FOREIGN KEY `fk_tasks_tag`;

-- 2. 修改 tag_id 字段为可空
ALTER TABLE `tasks` MODIFY COLUMN `tag_id` BIGINT UNSIGNED NULL;

-- 3. 重新添加外键约束（允许 NULL 值，删除标签时设置为 NULL）
ALTER TABLE `tasks` 
ADD CONSTRAINT `fk_tasks_tag` 
FOREIGN KEY (`tag_id`) 
REFERENCES `tags` (`id`) 
ON DELETE SET NULL;
