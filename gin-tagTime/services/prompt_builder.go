package services

import (
	"fmt"
	"strings"
	"tagtime/models"
)

// PromptBuilder 提示词构建器
type PromptBuilder struct {
	systemPrompts map[string]string
}

// NewPromptBuilder 创建提示词构建器
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		systemPrompts: map[string]string{
			"procrastination": "你是一个洞察人心的 AI 助手，擅长识别拖延和自我欺骗的模式。",
			"inconsistency":   "你是一个洞察人心的 AI 助手，擅长识别三分钟热度和缺乏坚持的模式。",
			"stress":          "你是一个洞察人心的 AI 助手，擅长识别压力和焦虑的根源。",
			"default":         "你是一个洞察人心的 AI 助手，能看穿用户的内心矛盾和逃避的事实。",
		},
	}
}

// BuildPrompt 构建完整提示词
func (b *PromptBuilder) BuildPrompt(summary *models.UserSummary, pattern models.BehaviorPattern) string {
	systemPrompt := b.selectSystemPrompt(pattern)
	fewShotExamples := b.getFewShotExamples()
	userPrompt := b.formatUserSummary(summary, pattern)

	return fmt.Sprintf(`%s

%s

现在，基于以下用户数据生成一句话：

%s

要求：
1. 仅输出一句话，不要任何解释
2. 直接针对数据中暴露的模式
3. 语气冷酷但真诚
4. 避免空泛的鸡汤
5. 字数控制在50字以内`, systemPrompt, fewShotExamples, userPrompt)
}

// selectSystemPrompt 根据行为模式选择系统提示词
func (b *PromptBuilder) selectSystemPrompt(pattern models.BehaviorPattern) string {
	if pattern.ProcrastinationScore > 70 {
		return b.systemPrompts["procrastination"]
	} else if pattern.ConsistencyScore < 30 {
		return b.systemPrompts["inconsistency"]
	} else if len(pattern.StressIndicators) > 2 {
		return b.systemPrompts["stress"]
	}
	return b.systemPrompts["default"]
}

// getFewShotExamples 获取Few-Shot示例
func (b *PromptBuilder) getFewShotExamples() string {
	return `示例：

用户数据：常用标签"学习"，但7天仅计时2小时；便签中多次出现"明天开始"。
输出：你的"明天"已经说了三个月，但今天依然是昨天的重复。

用户数据：任务完成率15%，未完成任务堆积45个；便签中频繁出现"焦虑"。
输出：你不是被任务压垮的，你是被自己的逃避压垮的。

用户数据：深夜工作时长占比60%，便签中出现"累"、"撑不住"；任务切换频率高。
输出：你在用忙碌掩盖焦虑，但焦虑从未离开。

用户数据：坚持指数25分，连续计时天数仅2天；便签中出现"又放弃了"。
输出：三分钟热度是你唯一坚持的事。`
}

// formatUserSummary 格式化用户摘要为自然语言
func (b *PromptBuilder) formatUserSummary(summary *models.UserSummary, pattern models.BehaviorPattern) string {
	var builder strings.Builder

	builder.WriteString("用户数据摘要：\n\n")

	// 1. 常用标签
	if len(summary.TopTags) > 0 {
		builder.WriteString("常用标签：")
		for i, tag := range summary.TopTags {
			if i > 0 {
				builder.WriteString("、")
			}
			builder.WriteString(fmt.Sprintf("%s", tag.Name))
		}
		builder.WriteString("\n\n")
	}

	// 2. 任务情况
	builder.WriteString(fmt.Sprintf("任务情况：共创建 %d 个任务，已完成 %d 个，完成率 %.1f%%。",
		summary.TaskStats.TotalTasks,
		summary.TaskStats.CompletedTasks,
		summary.TaskStats.CompletionRate))

	if len(summary.TaskStats.UnfinishedTasks) > 0 {
		builder.WriteString("未完成任务包括：")
		for i, task := range summary.TaskStats.UnfinishedTasks {
			if i > 0 {
				builder.WriteString("、")
			}
			builder.WriteString(fmt.Sprintf("\"%s\"", task))
			if i >= 2 { // 最多显示3个
				break
			}
		}
		builder.WriteString("。")
	}
	builder.WriteString("\n\n")

	// 3. 计时数据
	if len(summary.TimingStats.TopTagTimings) > 0 {
		builder.WriteString("计时最多标签：")
		for i, timing := range summary.TimingStats.TopTagTimings {
			if i > 0 {
				builder.WriteString("、")
			}
			builder.WriteString(fmt.Sprintf("%s %.1f小时", timing.TagName, timing.Hours))
		}
		builder.WriteString("。")
	}

	builder.WriteString(fmt.Sprintf("最近一周计时：共 %.1f 小时。\n\n",
		summary.TimingStats.Last7DaysTotalHours))

	// 4. 便签片段
	if len(summary.RecentNotes) > 0 {
		builder.WriteString("最近便签片段：\n")
		for i, note := range summary.RecentNotes {
			if i >= 3 { // 最多显示3条
				break
			}
			// 截断过长的便签
			if len(note) > 100 {
				note = note[:100] + "..."
			}
			builder.WriteString(fmt.Sprintf("- \"%s\"\n", note))
		}
		builder.WriteString("\n")
	}

	// 5. 高频关键词
	if len(summary.KeyWords) > 0 {
		builder.WriteString("高频关键词：")
		for i, keyword := range summary.KeyWords {
			if i > 0 {
				builder.WriteString("、")
			}
			builder.WriteString(keyword)
			if i >= 7 { // 最多显示8个
				break
			}
		}
		builder.WriteString("。\n\n")
	}

	// 6. 行为模式
	builder.WriteString(fmt.Sprintf("行为模式：工作时间偏好为%s，", pattern.WorkTimePreference))

	// 添加任务切换频率
	if pattern.TaskSwitchRate > 0 {
		builder.WriteString(fmt.Sprintf("任务切换频率 %.1f 次/小时，", pattern.TaskSwitchRate))
	}

	builder.WriteString(fmt.Sprintf("拖延指数 %.0f 分，", pattern.ProcrastinationScore))
	builder.WriteString(fmt.Sprintf("坚持指数 %.0f 分。", pattern.ConsistencyScore))

	if len(pattern.StressIndicators) > 0 {
		builder.WriteString("压力指标：")
		for i, indicator := range pattern.StressIndicators {
			if i > 0 {
				builder.WriteString("、")
			}
			builder.WriteString(indicator)
		}
		builder.WriteString("。")
	}

	return builder.String()
}
