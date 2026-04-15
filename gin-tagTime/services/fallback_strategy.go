package services

import (
	"math/rand"
	"tagtime/models"
	"time"
)

// FallbackStrategy 降级策略
type FallbackStrategy struct {
	presetLines map[string][]string
	rng         *rand.Rand
}

// NewFallbackStrategy 创建降级策略
func NewFallbackStrategy() *FallbackStrategy {
	return &FallbackStrategy{
		presetLines: map[string][]string{
			"high_procrastination": {
				"你的待办清单越来越长，但完成的永远是最简单的那几个。",
				"你很擅长制定计划，只是从不执行。",
				"你的'明天'已经说了三个月，但今天依然是昨天的重复。",
				"拖延不是时间管理问题，是你在逃避面对自己。",
				"你不是没时间，你只是不想开始。",
			},
			"low_consistency": {
				"三分钟热度是你唯一坚持的事。",
				"你不是不知道该做什么，你只是不想开始。",
				"每次信誓旦旦，每次半途而废，你已经习惯了对自己失望。",
				"你的坚持，从来没有超过一周。",
				"你总是在开始和放弃之间循环，从未真正前进。",
			},
			"stress_overload": {
				"你在用忙碌掩盖焦虑，但焦虑从未离开。",
				"你的疲惫不是因为做得太多，而是想得太多。",
				"你不是被任务压垮的，你是被自己的逃避压垮的。",
				"你清晰地看着自己沉沦，却连伸自己一把的勇气都没有。",
				"你的焦虑来自于你知道自己在浪费时间，却不愿改变。",
			},
			"default": {
				"你很聪明，只是从不逼自己。",
				"你明明很缺安全感，却总在假装无所谓。",
				"你的问题不是能力不够，而是从不全力以赴。",
				"你总是在等待完美的时机，但完美的时机从不存在。",
				"你不是做不到，你只是不想承认自己害怕失败。",
				"你的舒适区，正在慢慢变成你的牢笼。",
			},
		},
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetFallbackLine 根据行为模式获取降级击溃语
func (f *FallbackStrategy) GetFallbackLine(pattern models.BehaviorPattern) string {
	category := f.classifyPattern(pattern)
	lines := f.presetLines[category]

	if len(lines) == 0 {
		lines = f.presetLines["default"]
	}

	// 随机选择一条
	index := f.rng.Intn(len(lines))
	return lines[index]
}

// classifyPattern 分类用户模式
func (f *FallbackStrategy) classifyPattern(pattern models.BehaviorPattern) string {
	// 优先级：拖延 > 压力 > 不坚持 > 默认

	if pattern.ProcrastinationScore > 70 {
		return "high_procrastination"
	}

	if len(pattern.StressIndicators) > 2 {
		return "stress_overload"
	}

	if pattern.ConsistencyScore < 30 {
		return "low_consistency"
	}

	return "default"
}
