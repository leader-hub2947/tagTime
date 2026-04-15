package models

// UserSummary 用户数据摘要
type UserSummary struct {
	TopTags         []TagStat        `json:"top_tags"`
	TaskStats       TaskStatistics   `json:"task_stats"`
	TimingStats     TimingStatistics `json:"timing_stats"`
	RecentNotes     []string         `json:"recent_notes"`
	KeyWords        []string         `json:"keywords"`
	BehaviorPattern BehaviorPattern  `json:"behavior_pattern"`
}

// TagStat 标签统计
type TagStat struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Count int    `json:"count"`
}

// TaskStatistics 任务统计
type TaskStatistics struct {
	TotalTasks      int      `json:"total_tasks"`
	CompletedTasks  int      `json:"completed_tasks"`
	CompletionRate  float64  `json:"completion_rate"`
	UnfinishedTasks []string `json:"unfinished_tasks"`
	OngoingTasks    []string `json:"ongoing_tasks"`
}

// TimingStatistics 计时统计
type TimingStatistics struct {
	TopTagTimings       []TagTiming `json:"top_tag_timings"`
	Last7DaysTotalHours float64     `json:"last_7days_total_hours"`
	PeakDayHours        float64     `json:"peak_day_hours"`
}

// TagTiming 标签计时
type TagTiming struct {
	TagName string  `json:"tag_name"`
	Hours   float64 `json:"hours"`
}

// BehaviorPattern 行为模式
type BehaviorPattern struct {
	WorkTimePreference   string   `json:"work_time_preference"`  // 工作时间偏好
	TaskSwitchRate       float64  `json:"task_switch_rate"`      // 任务切换频率
	ProcrastinationScore float64  `json:"procrastination_score"` // 拖延指数 0-100
	ConsistencyScore     float64  `json:"consistency_score"`     // 坚持指数 0-100
	StressIndicators     []string `json:"stress_indicators"`     // 压力指标
}

// CrushLineResponse API响应
type CrushLineResponse struct {
	CrushLine      string `json:"crush_line"`
	RemainingCount int    `json:"remaining_count"`
}

// AIRequest AI请求
type AIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
	Stream      bool      `json:"stream"` // 是否流式输出
}

// Message AI消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIResponse AI响应
type AIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice AI选择
type Choice struct {
	Index            int     `json:"index"`
	Message          Message `json:"message"`
	FinishReason     string  `json:"finish_reason"`
	ReasoningContent string  `json:"reasoning_content,omitempty"` // DeepSeek Reasoner 特有字段
}

// Usage token使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
