-- 便签-任务关联表
CREATE TABLE IF NOT EXISTS `note_tasks` (
    `note_id` BIGINT UNSIGNED NOT NULL COMMENT '便签ID',
    `task_id` BIGINT UNSIGNED NOT NULL COMMENT '任务ID',
    PRIMARY KEY (`note_id`, `task_id`),
    FOREIGN KEY (`note_id`) REFERENCES `notes`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`task_id`) REFERENCES `tasks`(`id`) ON DELETE CASCADE,
    INDEX `idx_task_id` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='便签-任务关联表';
