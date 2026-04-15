package services

import (
	"fmt"
	"log"
	"tagtime/config"
	"tagtime/models"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// AICrushService AI击溃服务
type AICrushService struct {
	dataExtractor    *DataExtractor
	behaviorAnalyzer *BehaviorAnalyzer
	dataSanitizer    *DataSanitizer
	promptBuilder    *PromptBuilder
	aiClient         *AIClient
	cacheManager     *CacheManager
	rateLimiter      *RateLimiter
	fallbackStrategy *FallbackStrategy
	config           *config.AIConfig
}

// NewAICrushService 创建AI击溃服务
func NewAICrushService(db *gorm.DB, redisClient *redis.Client, aiConfig *config.AIConfig) *AICrushService {
	return &AICrushService{
		dataExtractor:    NewDataExtractor(db),
		behaviorAnalyzer: NewBehaviorAnalyzer(db),
		dataSanitizer:    NewDataSanitizer(),
		promptBuilder:    NewPromptBuilder(),
		aiClient:         NewAIClient(aiConfig),
		cacheManager:     NewCacheManager(redisClient),
		rateLimiter:      NewRateLimiter(redisClient, aiConfig.DailyLimit, aiConfig.GlobalRateLimit),
		fallbackStrategy: NewFallbackStrategy(),
		config:           aiConfig,
	}
}

// GenerateCrushLine 生成击溃语
func (s *AICrushService) GenerateCrushLine(userID uint) (*models.CrushLineResponse, error) {
	startTime := time.Now()

	// 1. 检查全局限流
	if !s.rateLimiter.AllowGlobal() {
		log.Printf("[AI Crush] 全局限流触发，用户ID: %d", userID)
		return nil, fmt.Errorf("系统繁忙，请稍后再试")
	}

	// 2. 检查每日限制
	allowed, remaining, err := s.rateLimiter.CheckDailyLimit(userID)
	if err != nil {
		log.Printf("[AI Crush] 检查限流失败，用户ID: %d, 错误: %v", userID, err)
		return nil, fmt.Errorf("系统错误，请稍后再试")
	}

	if !allowed {
		log.Printf("[AI Crush] 用户超过每日限制，用户ID: %d", userID)
		return &models.CrushLineResponse{
			CrushLine:      "",
			RemainingCount: 0,
		}, fmt.Errorf("今日击溃次数已用完，明天再来吧")
	}

	// 3. 检查缓存
	cached, err := s.cacheManager.Get(userID)
	if err == nil && cached != "" {
		log.Printf("[AI Crush] 缓存命中，用户ID: %d", userID)
		return &models.CrushLineResponse{
			CrushLine:      cached,
			RemainingCount: remaining,
		}, nil
	}

	// 4. 提取用户数据
	log.Printf("[AI Crush] 开始提取用户数据，用户ID: %d", userID)
	summary, err := s.dataExtractor.BuildUserSummary(userID)
	if err != nil {
		log.Printf("[AI Crush] 提取用户数据失败，用户ID: %d, 错误: %v", userID, err)
		return nil, fmt.Errorf("数据提取失败，请稍后再试")
	}

	// 检查数据是否充足
	if summary.TaskStats.TotalTasks == 0 && len(summary.RecentNotes) == 0 {
		log.Printf("[AI Crush] 用户数据不足，用户ID: %d", userID)
		return nil, fmt.Errorf("数据不足，请先使用一段时间再来")
	}

	// 5. 分析行为模式
	log.Printf("[AI Crush] 开始分析行为模式，用户ID: %d", userID)
	pattern, err := s.behaviorAnalyzer.AnalyzePattern(userID)
	if err != nil {
		log.Printf("[AI Crush] 分析行为模式失败，用户ID: %d, 错误: %v", userID, err)
		// 继续执行，使用默认模式
	}

	summary.BehaviorPattern = pattern

	// 6. 数据脱敏
	summary = s.dataSanitizer.SanitizeSummary(summary)

	// 7. 构建提示词
	prompt := s.promptBuilder.BuildPrompt(summary, pattern)

	// 8. 调用AI服务
	log.Printf("[AI Crush] 开始调用AI服务，用户ID: %d", userID)
	crushLine, err := s.aiClient.GenerateCrushLine(prompt)

	// 如果AI调用失败，使用降级策略
	if err != nil {
		log.Printf("[AI Crush] AI服务调用失败，触发降级策略，用户ID: %d, 错误: %v", userID, err)
		crushLine = s.fallbackStrategy.GetFallbackLine(pattern)
	}

	// 9. 增加调用计数
	if err := s.rateLimiter.IncrementCount(userID); err != nil {
		log.Printf("[AI Crush] 增加调用计数失败，用户ID: %d, 错误: %v", userID, err)
	}

	// 10. 缓存结果
	if err := s.cacheManager.Set(userID, crushLine, s.config.CacheExpire); err != nil {
		log.Printf("[AI Crush] 缓存结果失败，用户ID: %d, 错误: %v", userID, err)
	}

	// 11. 获取剩余次数
	remaining, _ = s.rateLimiter.GetRemainingCount(userID)

	// 记录性能指标
	duration := time.Since(startTime)
	log.Printf("[AI Crush] 请求完成，用户ID: %d, 耗时: %v, 剩余次数: %d", userID, duration, remaining)

	return &models.CrushLineResponse{
		CrushLine:      crushLine,
		RemainingCount: remaining,
	}, nil
}

// GetRemainingCount 获取剩余调用次数
func (s *AICrushService) GetRemainingCount(userID uint) (int, error) {
	return s.rateLimiter.GetRemainingCount(userID)
}
