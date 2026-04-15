-- ============================================
-- AI洞察功能性能优化索引
-- 用于优化用户数据提取和行为分析查询
-- ============================================

-- 优化标签统计查询
CREATE INDEX IF NOT EXISTS idx_note_tags_tag_id ON note_tags(tag_id);
CREATE INDEX IF NOT EXISTS idx_tasks_tag_id ON tasks(tag_id);

-- 优化用户数据查询
CREATE INDEX IF NOT EXISTS idx_notes_user_created ON notes(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tasks_user_status ON tasks(user_id, status);
CREATE INDEX IF NOT EXISTS idx_time_entries_user_start ON time_entries(user_id, start_time DESC);

-- 优化行为分析查询
CREATE INDEX IF NOT EXISTS idx_tasks_user_completed ON tasks(user_id, completed_at);
CREATE INDEX IF NOT EXISTS idx_time_entries_task_start ON time_entries(task_id, start_time);
