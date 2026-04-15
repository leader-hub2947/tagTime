package services

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"
)

// RateLimiter 限流器
type RateLimiter struct {
	redis       *redis.Client
	ctx         context.Context
	dailyLimit  int
	globalLimit *rate.Limiter
}

// NewRateLimiter 创建限流器
func NewRateLimiter(redisClient *redis.Client, dailyLimit int, globalRateLimit int) *RateLimiter {
	// 创建全局限流器（每分钟允许的请求数）
	limiter := rate.NewLimiter(rate.Limit(globalRateLimit)/60.0, globalRateLimit)

	return &RateLimiter{
		redis:       redisClient,
		ctx:         context.Background(),
		dailyLimit:  dailyLimit,
		globalLimit: limiter,
	}
}

// CheckDailyLimit 检查每日调用次数限制
func (l *RateLimiter) CheckDailyLimit(userID uint) (bool, int, error) {
	remaining, err := l.GetRemainingCount(userID)
	if err != nil {
		return false, 0, err
	}

	if remaining <= 0 {
		return false, 0, nil
	}

	return true, remaining, nil
}

// IncrementCount 增加调用计数
func (l *RateLimiter) IncrementCount(userID uint) error {
	key := l.getDailyKey(userID)

	// 增加计数
	count, err := l.redis.Incr(l.ctx, key).Result()
	if err != nil {
		return fmt.Errorf("增加计数失败: %w", err)
	}

	// 如果是第一次调用，设置过期时间为24小时
	if count == 1 {
		tomorrow := time.Now().Add(24 * time.Hour)
		midnight := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())
		expiration := midnight.Sub(time.Now())

		err = l.redis.Expire(l.ctx, key, expiration).Err()
		if err != nil {
			return fmt.Errorf("设置过期时间失败: %w", err)
		}
	}

	return nil
}

// GetRemainingCount 获取剩余调用次数
func (l *RateLimiter) GetRemainingCount(userID uint) (int, error) {
	key := l.getDailyKey(userID)

	count, err := l.redis.Get(l.ctx, key).Int()
	if err == redis.Nil {
		// 没有记录，说明今天还没调用过
		return l.dailyLimit, nil
	}
	if err != nil {
		return 0, fmt.Errorf("获取计数失败: %w", err)
	}

	remaining := l.dailyLimit - count
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// AllowGlobal 检查全局限流
func (l *RateLimiter) AllowGlobal() bool {
	return l.globalLimit.Allow()
}

// getDailyKey 获取每日计数的Redis键
func (l *RateLimiter) getDailyKey(userID uint) string {
	today := time.Now().Format("2006-01-02")
	return fmt.Sprintf("ai_crush_daily:%d:%s", userID, today)
}
