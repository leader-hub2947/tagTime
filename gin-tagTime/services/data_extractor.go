package services

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"tagtime/models"
	"time"

	"gorm.io/gorm"
)

// DataExtractor 数据提取器
type DataExtractor struct {
	db *gorm.DB
}

// NewDataExtractor 创建数据提取器
func NewDataExtractor(db *gorm.DB) *DataExtractor {
	return &DataExtractor{db: db}
}

// BuildUserSummary 构建用户摘要（使用并发查询优化性能）
func (e *DataExtractor) BuildUserSummary(userID uint) (*models.UserSummary, error) {
	summary := &models.UserSummary{}
	var wg sync.WaitGroup
	errChan := make(chan error, 5)

	// 并发查询5个维度的数据
	wg.Add(5)

	// 1. 查询Top标签
	go func() {
		defer wg.Done()
		tags, err := e.getTopTags(userID, 5)
		if err != nil {
			errChan <- fmt.Errorf("获取标签失败: %w", err)
			return
		}
		summary.TopTags = tags
	}()

	// 2. 查询任务统计
	go func() {
		defer wg.Done()
		stats, err := e.getTaskStatistics(userID)
		if err != nil {
			errChan <- fmt.Errorf("获取任务统计失败: %w", err)
			return
		}
		summary.TaskStats = stats
	}()

	// 3. 查询计时统计
	go func() {
		defer wg.Done()
		timing, err := e.getTimingStatistics(userID)
		if err != nil {
			errChan <- fmt.Errorf("获取计时统计失败: %w", err)
			return
		}
		summary.TimingStats = timing
	}()

	// 4. 查询最近便签
	go func() {
		defer wg.Done()
		notes, err := e.getRecentNotes(userID)
		if err != nil {
			errChan <- fmt.Errorf("获取便签失败: %w", err)
			return
		}
		summary.RecentNotes = notes
	}()

	// 5. 提取关键词
	go func() {
		defer wg.Done()
		// 等待便签数据
		time.Sleep(100 * time.Millisecond)
		keywords := e.extractKeywords(summary.RecentNotes)
		summary.KeyWords = keywords
	}()

	wg.Wait()
	close(errChan)

	// 检查是否有错误
	if len(errChan) > 0 {
		return nil, <-errChan
	}

	return summary, nil
}

// getTopTags 获取使用频率最高的标签
func (e *DataExtractor) getTopTags(userID uint, limit int) ([]models.TagStat, error) {
	var tags []models.TagStat

	// 联合查询便签和任务关联的标签
	query := `
		SELECT t.name, t.color, COUNT(*) as count
		FROM tags t
		LEFT JOIN note_tags nt ON t.id = nt.tag_id
		LEFT JOIN tasks tk ON t.id = tk.tag_id
		WHERE t.user_id = ?
		GROUP BY t.id, t.name, t.color
		ORDER BY count DESC
		LIMIT ?
	`

	err := e.db.Raw(query, userID, limit).Scan(&tags).Error
	return tags, err
}

// getTaskStatistics 获取任务统计
func (e *DataExtractor) getTaskStatistics(userID uint) (models.TaskStatistics, error) {
	stats := models.TaskStatistics{}

	// 查询总任务数和已完成任务数
	var total, completed int64
	e.db.Model(&models.Task{}).Where("user_id = ?", userID).Count(&total)
	e.db.Model(&models.Task{}).Where("user_id = ? AND status = 2", userID).Count(&completed)

	stats.TotalTasks = int(total)
	stats.CompletedTasks = int(completed)

	if total > 0 {
		stats.CompletionRate = float64(completed) / float64(total) * 100
	}

	// 查询未完成任务（最多5个）
	var unfinishedTasks []models.Task
	e.db.Where("user_id = ? AND status != 2", userID).
		Order("created_at DESC").
		Limit(5).
		Find(&unfinishedTasks)

	stats.UnfinishedTasks = make([]string, 0, len(unfinishedTasks))
	for _, task := range unfinishedTasks {
		stats.UnfinishedTasks = append(stats.UnfinishedTasks, task.Name)
	}

	// 查询进行中的任务
	var ongoingTasks []models.Task
	e.db.Where("user_id = ? AND status = 1", userID).
		Limit(3).
		Find(&ongoingTasks)

	stats.OngoingTasks = make([]string, 0, len(ongoingTasks))
	for _, task := range ongoingTasks {
		stats.OngoingTasks = append(stats.OngoingTasks, task.Name)
	}

	return stats, nil
}

// getTimingStatistics 获取计时统计
func (e *DataExtractor) getTimingStatistics(userID uint) (models.TimingStatistics, error) {
	stats := models.TimingStatistics{}

	// 查询标签时长排行榜（前3名）
	query := `
		SELECT t.name as tag_name, SUM(tk.total_duration) / 3600.0 as hours
		FROM tasks tk
		JOIN tags t ON tk.tag_id = t.id
		WHERE tk.user_id = ?
		GROUP BY t.id, t.name
		ORDER BY hours DESC
		LIMIT 3
	`

	err := e.db.Raw(query, userID).Scan(&stats.TopTagTimings).Error
	if err != nil {
		return stats, err
	}

	// 查询最近7天总计时长
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	var totalSeconds int64
	e.db.Model(&models.TimeEntry{}).
		Where("user_id = ? AND start_time >= ?", userID, sevenDaysAgo).
		Select("COALESCE(SUM(duration), 0)").
		Scan(&totalSeconds)

	stats.Last7DaysTotalHours = float64(totalSeconds) / 3600.0

	// 查询最高计时日的时长
	var peakDayHours *float64
	query = `
		SELECT MAX(daily_hours) as peak_day_hours
		FROM (
			SELECT DATE(start_time) as day, SUM(duration) / 3600.0 as daily_hours
			FROM time_entries
			WHERE user_id = ? AND start_time >= ?
			GROUP BY DATE(start_time)
		) as daily_stats
	`

	e.db.Raw(query, userID, sevenDaysAgo).Scan(&peakDayHours)

	// 处理NULL值
	if peakDayHours != nil {
		stats.PeakDayHours = *peakDayHours
	} else {
		stats.PeakDayHours = 0
	}

	return stats, nil
}

// getRecentNotes 获取最近的便签（动态调整采样）
func (e *DataExtractor) getRecentNotes(userID uint) ([]string, error) {
	var count int64
	e.db.Model(&models.Note{}).Where("user_id = ? AND is_deleted = false", userID).Count(&count)

	// 根据总量动态调整采样数量和截断长度
	limit := 10
	maxLength := 200

	if count > 1000 {
		limit = 5
		maxLength = 100
	} else if count < 50 {
		limit = 20
		maxLength = 300
	}

	// 查询最近的便签
	var notes []models.Note
	err := e.db.Where("user_id = ? AND is_deleted = false", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&notes).Error

	if err != nil {
		return nil, err
	}

	// 提取内容并截断
	contents := make([]string, 0, len(notes))
	for _, note := range notes {
		content := note.Content
		if len(content) > maxLength {
			content = content[:maxLength] + "..."
		}
		contents = append(contents, content)
	}

	return contents, nil
}

// extractKeywords 提取关键词（使用TF-IDF算法）
func (e *DataExtractor) extractKeywords(notes []string) []string {
	if len(notes) == 0 {
		return []string{}
	}

	// 停用词列表
	stopWords := map[string]bool{
		"的": true, "了": true, "在": true, "是": true,
		"我": true, "有": true, "和": true, "就": true,
		"不": true, "人": true, "都": true, "一": true,
		"个": true, "上": true, "也": true, "很": true,
		"到": true, "说": true, "要": true, "去": true,
		"你": true, "会": true, "着": true, "没": true,
		"看": true, "好": true, "自己": true, "这": true,
		"那": true, "里": true, "为": true, "以": true,
		"他": true, "时候": true, "可以": true, "但": true,
	}

	// 情感词权重加成
	emotionWords := map[string]float64{
		"焦虑": 2.0, "拖延": 2.0, "压力": 2.0,
		"迷茫": 2.0, "疲惫": 2.0, "放弃": 2.0,
		"坚持": 1.5, "努力": 1.5, "改变": 1.5,
		"累": 1.8, "烦": 1.8, "难": 1.5,
		"痛苦": 2.0, "无助": 2.0, "孤独": 1.8,
	}

	// 统计词频
	wordFreq := make(map[string]float64)

	for _, note := range notes {
		// 简单分词（按字符分割）
		words := segmentText(note)
		for _, word := range words {
			if stopWords[word] || len(word) < 2 {
				continue
			}

			weight := 1.0
			if w, ok := emotionWords[word]; ok {
				weight = w
			}

			wordFreq[word] += weight
		}
	}

	// 排序取Top 10
	return getTopWords(wordFreq, 10)
}

// segmentText 简单分词（按标点和空格分割）
func segmentText(text string) []string {
	// 定义需要替换的标点符号
	punctuationMap := map[rune]bool{
		'，': true, '。': true, '！': true, '？': true,
		'；': true, '：': true, '"': true, // 英文双引号
		'\'': true,            // 英文单引号（转义写法）
		'‘':  true, '’': true, // 中文左/右单引号
		'“': true, '”': true, // 中文左/右双引号（可选）
		'（': true, '）': true,
		'、': true, '《': true, '》': true,
	}

	// 替换标点符号为空格
	var builder strings.Builder
	for _, r := range text {
		if punctuationMap[r] || r == '\n' || r == '\r' || r == '\t' {
			builder.WriteRune(' ')
		} else {
			builder.WriteRune(r)
		}
	}
	text = builder.String()

	// 分割并过滤空字符串
	words := strings.Fields(text)
	result := make([]string, 0, len(words))

	for _, word := range words {
		word = strings.TrimSpace(word)
		if word != "" {
			result = append(result, word)
		}
	}

	return result
}

// getTopWords 获取频率最高的词
func getTopWords(wordFreq map[string]float64, limit int) []string {
	type wordCount struct {
		word  string
		count float64
	}

	// 转换为切片
	words := make([]wordCount, 0, len(wordFreq))
	for word, count := range wordFreq {
		words = append(words, wordCount{word, count})
	}

	// 排序
	sort.Slice(words, func(i, j int) bool {
		return words[i].count > words[j].count
	})

	// 取前N个
	if len(words) > limit {
		words = words[:limit]
	}

	result := make([]string, len(words))
	for i, wc := range words {
		result[i] = wc.word
	}

	return result
}
