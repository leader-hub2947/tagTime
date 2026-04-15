package services

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// CacheManager 缓存管理器
type CacheManager struct {
	redis *redis.Client
	ctx   context.Context
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(redisClient *redis.Client) *CacheManager {
	return &CacheManager{
		redis: redisClient,
		ctx:   context.Background(),
	}
}

// Get 获取缓存
func (m *CacheManager) Get(userID uint) (string, error) {
	key := fmt.Sprintf("ai_crush:%d", userID)
	result, err := m.redis.Get(m.ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("缓存不存在")
	}
	if err != nil {
		return "", fmt.Errorf("获取缓存失败: %w", err)
	}
	return result, nil
}

// Set 设置缓存
func (m *CacheManager) Set(userID uint, crushLine string, expiration time.Duration) error {
	key := fmt.Sprintf("ai_crush:%d", userID)
	err := m.redis.Set(m.ctx, key, crushLine, expiration).Err()
	if err != nil {
		return fmt.Errorf("设置缓存失败: %w", err)
	}
	return nil
}

// Delete 删除缓存
func (m *CacheManager) Delete(userID uint) error {
	key := fmt.Sprintf("ai_crush:%d", userID)
	err := m.redis.Del(m.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("删除缓存失败: %w", err)
	}
	return nil
}
