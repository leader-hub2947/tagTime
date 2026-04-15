package config

import (
	"time"
)

// AIConfig AI服务配置
type AIConfig struct {
	Provider        string        // AI服务提供商: openai, zhipu, deepseek
	APIKey          string        // API密钥
	Endpoint        string        // API端点
	Model           string        // 模型名称
	Timeout         time.Duration // 请求超时时间
	DailyLimit      int           // 每日调用次数限制
	CacheExpire     time.Duration // 缓存过期时间
	GlobalRateLimit int           // 全局限流（次/分钟）
}

// LoadAIConfig 加载AI配置
func LoadAIConfig() *AIConfig {
	return &AIConfig{
		Provider:        getEnv("AI_PROVIDER", "deepseek"),
		APIKey:          getEnv("AI_API_KEY", ""),
		Endpoint:        getEnv("AI_ENDPOINT", "https://api.deepseek.com/chat/completions"),
		Model:           getEnv("AI_MODEL", "deepseek-reasoner"),
		Timeout:         getDurationEnv("AI_TIMEOUT", 30*time.Second),
		DailyLimit:      getIntEnv("AI_DAILY_LIMIT", 3),
		CacheExpire:     getDurationEnv("AI_CACHE_EXPIRE", 10*time.Minute),
		GlobalRateLimit: getIntEnv("AI_GLOBAL_RATE_LIMIT", 10),
	}
}
