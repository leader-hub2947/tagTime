package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"tagtime/config"
	"tagtime/models"
	"tagtime/services"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetRemainingCount(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 跳过测试如果没有Redis连接
	if config.RedisClient == nil {
		t.Skip("跳过测试：Redis未配置")
	}

	// 创建服务和控制器
	aiConfig := config.LoadAIConfig()
	aiService := services.NewAICrushService(config.DB, config.RedisClient, aiConfig)
	controller := NewAICrushController(aiService)

	// 创建测试路由
	router := gin.New()
	router.GET("/api/v1/ai/crush/remaining", func(c *gin.Context) {
		// 模拟JWT中间件设置的user_id
		c.Set("user_id", uint(1))
		controller.GetRemainingCount(c)
	})

	// 创建测试请求
	req := httptest.NewRequest("GET", "/api/v1/ai/crush/remaining", nil)
	w := httptest.NewRecorder()

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	// 解析响应
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证响应包含remaining_count字段
	remainingCount, exists := response["remaining_count"]
	assert.True(t, exists, "响应应包含remaining_count字段")

	// 验证remaining_count是数字且在合理范围内
	count, ok := remainingCount.(float64)
	assert.True(t, ok, "remaining_count应该是数字")
	assert.GreaterOrEqual(t, count, float64(0), "remaining_count应该大于等于0")
	assert.LessOrEqual(t, count, float64(3), "remaining_count应该小于等于3")

	// 验证响应头
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"), "响应头应包含X-RateLimit-Remaining")
	assert.Equal(t, "3", w.Header().Get("X-RateLimit-Limit"), "响应头应包含X-RateLimit-Limit")
}

func TestGetRemainingCount_Unauthorized(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 跳过测试如果没有Redis连接
	if config.RedisClient == nil {
		t.Skip("跳过测试：Redis未配置")
	}

	// 创建服务和控制器
	aiConfig := config.LoadAIConfig()
	aiService := services.NewAICrushService(config.DB, config.RedisClient, aiConfig)
	controller := NewAICrushController(aiService)

	// 创建测试路由（不设置user_id）
	router := gin.New()
	router.GET("/api/v1/ai/crush/remaining", controller.GetRemainingCount)

	// 创建测试请求
	req := httptest.NewRequest("GET", "/api/v1/ai/crush/remaining", nil)
	w := httptest.NewRecorder()

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 解析响应
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证错误信息
	assert.Contains(t, response["error"], "未授权访问")
}

func TestGetCrushLine_ResponseHeaders(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 跳过测试如果没有Redis连接
	if config.RedisClient == nil {
		t.Skip("跳过测试：Redis未配置")
	}

	// 创建服务和控制器
	aiConfig := config.LoadAIConfig()
	aiService := services.NewAICrushService(config.DB, config.RedisClient, aiConfig)
	controller := NewAICrushController(aiService)

	// 创建测试路由
	router := gin.New()
	router.POST("/api/v1/ai/crush", func(c *gin.Context) {
		// 模拟JWT中间件设置的user_id
		c.Set("user_id", uint(1))
		controller.GetCrushLine(c)
	})

	// 创建测试请求
	req := httptest.NewRequest("POST", "/api/v1/ai/crush", nil)
	w := httptest.NewRecorder()

	// 执行请求
	router.ServeHTTP(w, req)

	// 如果成功（200），验证响应头
	if w.Code == http.StatusOK {
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"), "响应头应包含X-RateLimit-Remaining")
		assert.Equal(t, "3", w.Header().Get("X-RateLimit-Limit"), "响应头应包含X-RateLimit-Limit")

		// 验证响应体
		var response models.CrushLineResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, response.RemainingCount, 0)
		assert.LessOrEqual(t, response.RemainingCount, 3)
	}
}
