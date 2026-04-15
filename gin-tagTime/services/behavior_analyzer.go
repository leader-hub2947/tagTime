package services

import (
	"math"
	"strings"
	"tagtime/models"
	"time"

	"gorm.io/gorm"
)

// BehaviorAnalyzer 行为分析器
type BehaviorAnalyzer struct {
	db *gorm.DB
}

// NewBehaviorAnalyzer 创建行为分析器
func NewBehaviorAnalyzer(db *gorm.DB) *BehaviorAnalyzer {
	return &BehaviorAnalyzer{db: db}
}

// AnalyzePattern 分析用户行为模式
func (a *BehaviorAnalyzer) AnalyzePattern(userID uint) (models.BehaviorPattern, error) {
	pattern := models.BehaviorPattern{}

	// 1. 分析工作时间偏好
	pattern.WorkTimePreference = a.classifyWorkTime(userID)

	// 2. 计算任务切换频率
	pattern.TaskSwitchRate = a.calculateTaskSwitchRate(userID)

	// 3. 计算拖延指数
	pattern.ProcrastinationScore = a.calculateProcrastination(userID)

	// 4. 计算坚持指数
	pattern.ConsistencyScore = a.calculateConsistency(userID)

	// 5. 识别压力指标
	pattern.StressIndicators = a.identifyStressSignals(userID)

	return pattern, nil
}

// classifyWorkTime 分类工作时间偏好
func (a *BehaviorAnalyzer) classifyWorkTime(userID uint) string {
	// 统计各时间段的计时时长
	type hourStat struct {
		Hour     int   //开始时间
		Duration int64 //持续时长
	}

	var hourStats []hourStat
	query := `
		SELECT HOUR(start_time) as hour, SUM(duration) as duration
		FROM time_entries
		WHERE user_id = ? AND start_time >= DATE_SUB(NOW(), INTERVAL 30 DAY)
		GROUP BY HOUR(start_time)
	`

	a.db.Raw(query, userID).Scan(&hourStats)

	if len(hourStats) == 0 {
		return "数据不足"
	}

	// 计算各时间段总时长
	var morningHours int64   // 6-12
	var afternoonHours int64 // 12-18
	var eveningHours int64   // 18-24
	var nightHours int64     // 0-6

	for _, stat := range hourStats {
		if stat.Hour >= 6 && stat.Hour < 12 {
			morningHours += stat.Duration
		} else if stat.Hour >= 12 && stat.Hour < 18 {
			afternoonHours += stat.Duration
		} else if stat.Hour >= 18 && stat.Hour < 24 {
			eveningHours += stat.Duration
		} else {
			nightHours += stat.Duration
		}
	}

	// 判断工作时间偏好
	if nightHours > morningHours && nightHours > afternoonHours {
		return "夜猫子"
	} else if morningHours > afternoonHours && morningHours > eveningHours {
		return "早起鸟"
	} else {
		return "正常作息"
	}
}

// calculateTaskSwitchRate 计算任务切换频率
func (a *BehaviorAnalyzer) calculateTaskSwitchRate(userID uint) float64 {
	// 查询最近30天的计时记录数和总时长
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	var entryCount int64
	var totalHours float64

	a.db.Model(&models.TimeEntry{}).
		Where("user_id = ? AND start_time >= ?", userID, thirtyDaysAgo).
		Count(&entryCount)

	var totalSeconds int64
	a.db.Model(&models.TimeEntry{}).
		Where("user_id = ? AND start_time >= ?", userID, thirtyDaysAgo).
		Select("COALESCE(SUM(duration), 0)").
		Scan(&totalSeconds)

	totalHours = float64(totalSeconds) / 3600.0

	if totalHours == 0 {
		return 0
	}

	// 切换频率 = 计时记录数 / 总时长（次/小时）
	return float64(entryCount) / totalHours
}

// calculateProcrastination 计算拖延指数
func (a *BehaviorAnalyzer) calculateProcrastination(userID uint) float64 {
	score := 0.0

	// 因素1: 未完成任务占比 (40%)
	var total, unfinished int64
	a.db.Model(&models.Task{}).Where("user_id = ?", userID).Count(&total)
	a.db.Model(&models.Task{}).Where("user_id = ? AND status != 2", userID).Count(&unfinished)

	if total > 0 {
		score += (float64(unfinished) / float64(total)) * 40
	}

	// 因素2: 任务创建到开始的平均延迟 (30%)
	var avgDelayHours *float64
	query := `
		SELECT AVG(TIMESTAMPDIFF(HOUR, t.created_at, te.start_time)) as avg_delay
		FROM tasks t
		JOIN time_entries te ON t.id = te.task_id
		WHERE t.user_id = ? AND te.start_time > t.created_at
		LIMIT 1
	`
	a.db.Raw(query, userID).Scan(&avgDelayHours)

	// 处理NULL值情况
	if avgDelayHours != nil {
		if *avgDelayHours > 72 { // 超过3天
			score += 30
		} else if *avgDelayHours > 24 { // 超过1天
			score += 15
		}
	}

	// 因素3: 长期未完成任务数量 (30%)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var longTermUnfinished int64
	a.db.Model(&models.Task{}).
		Where("user_id = ? AND status != 2 AND created_at < ?", userID, thirtyDaysAgo).
		Count(&longTermUnfinished)

	if longTermUnfinished > 10 {
		score += 30
	} else if longTermUnfinished > 5 {
		score += 20
	} else if longTermUnfinished > 2 {
		score += 10
	}

	return math.Min(score, 100)
}

// calculateConsistency 计算坚持指数
func (a *BehaviorAnalyzer) calculateConsistency(userID uint) float64 {
	score := 0.0

	// 因素1: 任务完成率 (40%)
	var total, completed int64
	a.db.Model(&models.Task{}).Where("user_id = ?", userID).Count(&total)
	a.db.Model(&models.Task{}).Where("user_id = ? AND status = 2", userID).Count(&completed)

	if total > 0 {
		score += (float64(completed) / float64(total)) * 40
	}

	// 因素2: 连续计时天数 (30%)
	consecutiveDays := a.getConsecutiveTimingDays(userID)
	if consecutiveDays >= 30 {
		score += 30
	} else if consecutiveDays >= 7 {
		score += 20
	} else if consecutiveDays >= 3 {
		score += 10
	}

	// 因素3: 长期任务完成情况 (30%)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var longTermTotal, longTermCompleted int64

	a.db.Model(&models.Task{}).
		Where("user_id = ? AND created_at < ?", userID, thirtyDaysAgo).
		Count(&longTermTotal)

	a.db.Model(&models.Task{}).
		Where("user_id = ? AND created_at < ? AND status = 2", userID, thirtyDaysAgo).
		Count(&longTermCompleted)

	if longTermTotal > 0 {
		longTermRate := float64(longTermCompleted) / float64(longTermTotal)
		score += longTermRate * 30
	}

	return math.Min(score, 100)
}

// getConsecutiveTimingDays 获取连续计时天数
func (a *BehaviorAnalyzer) getConsecutiveTimingDays(userID uint) int {
	// 查询最近90天有计时记录的日期
	var dates []time.Time
	query := `
		SELECT DISTINCT DATE(start_time) as date
		FROM time_entries
		WHERE user_id = ? AND start_time >= DATE_SUB(NOW(), INTERVAL 90 DAY)
		ORDER BY date DESC
	`

	a.db.Raw(query, userID).Scan(&dates)

	if len(dates) == 0 {
		return 0
	}

	// 计算连续天数
	consecutive := 1
	for i := 0; i < len(dates)-1; i++ {
		diff := dates[i].Sub(dates[i+1]).Hours() / 24
		if diff <= 1.5 { // 允许1天的误差
			consecutive++
		} else {
			break
		}
	}

	return consecutive
}

// identifyStressSignals 识别压力指标
func (a *BehaviorAnalyzer) identifyStressSignals(userID uint) []string {
	signals := []string{}

	// 1. 检查便签中的情感词
	var notes []models.Note
	a.db.Where("user_id = ? AND is_deleted = false", userID).
		Order("created_at DESC").
		Limit(20).
		Find(&notes)

	stressWords := []string{"焦虑", "压力", "累", "疲惫", "崩溃", "撑不住", "烦", "痛苦"}
	stressCount := 0

	for _, note := range notes {
		for _, word := range stressWords {
			if strings.Contains(note.Content, word) {
				stressCount++
				break
			}
		}
	}

	if stressCount > 5 {
		signals = append(signals, "便签中频繁出现负面情绪词")
	}

	// 2. 检查任务完成率骤降
	recentRate := a.getRecentCompletionRate(userID, 7)
	overallRate := a.getOverallCompletionRate(userID)

	if recentRate < overallRate*0.5 && overallRate > 0 {
		signals = append(signals, "近期任务完成率骤降")
	}

	// 3. 检查深夜工作频率
	lateNightHours := a.getTimingInHourRange(userID, 0, 6)
	if lateNightHours > 10 {
		signals = append(signals, "频繁深夜工作")
	}

	// 4. 检查任务堆积
	var unfinishedCount int64
	a.db.Model(&models.Task{}).
		Where("user_id = ? AND status != 2", userID).
		Count(&unfinishedCount)

	if unfinishedCount > 20 {
		signals = append(signals, "任务严重堆积")
	}

	return signals
}

// getRecentCompletionRate 获取最近N天的任务完成率
func (a *BehaviorAnalyzer) getRecentCompletionRate(userID uint, days int) float64 {
	startDate := time.Now().AddDate(0, 0, -days)

	var total, completed int64
	a.db.Model(&models.Task{}).
		Where("user_id = ? AND created_at >= ?", userID, startDate).
		Count(&total)

	a.db.Model(&models.Task{}).
		Where("user_id = ? AND created_at >= ? AND status = 2", userID, startDate).
		Count(&completed)

	if total == 0 {
		return 0
	}

	return float64(completed) / float64(total)
}

// getOverallCompletionRate 获取总体任务完成率
func (a *BehaviorAnalyzer) getOverallCompletionRate(userID uint) float64 {
	var total, completed int64
	a.db.Model(&models.Task{}).Where("user_id = ?", userID).Count(&total)
	a.db.Model(&models.Task{}).Where("user_id = ? AND status = 2", userID).Count(&completed)

	if total == 0 {
		return 0
	}

	return float64(completed) / float64(total)
}

// getTimingInHourRange 获取指定时间段的计时时长（小时）
func (a *BehaviorAnalyzer) getTimingInHourRange(userID uint, startHour, endHour int) float64 {
	var totalSeconds int64
	query := `
		SELECT COALESCE(SUM(duration), 0)
		FROM time_entries
		WHERE user_id = ? 
		AND start_time >= DATE_SUB(NOW(), INTERVAL 30 DAY)
		AND HOUR(start_time) >= ? 
		AND HOUR(start_time) < ?
	`

	a.db.Raw(query, userID, startHour, endHour).Scan(&totalSeconds)

	return float64(totalSeconds) / 3600.0
}
