package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"tagtime/services"

	"github.com/gin-gonic/gin"
)

// AICrushController AI击溃控制器
type AICrushController struct {
	service *services.AICrushService
}

// NewAICrushController 创建AI击溃控制器
func NewAICrushController(service *services.AICrushService) *AICrushController {
	return &AICrushController{
		service: service,
	}
}

// GetCrushLine 获取击溃语
func (c *AICrushController) GetCrushLine(ctx *gin.Context) {
	// 1. 获取用户ID（从JWT中间件设置的上下文）
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户ID格式错误",
		})
		return
	}

	// 2. 调用服务层
	response, err := c.service.GenerateCrushLine(uid)

	// 3. 错误处理
	if err != nil {
		c.handleError(ctx, err)
		return
	}

	// 4. 设置响应头（包含剩余调用次数）
	ctx.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", response.RemainingCount))
	ctx.Header("X-RateLimit-Limit", "3")

	// 5. 返回成功响应
	log.Printf("[AI Crush] 返回响应给用户 %d: crush_line=%s, remaining=%d", uid, response.CrushLine, response.RemainingCount)
	ctx.JSON(http.StatusOK, response)
}

// GetRemainingCount 获取剩余调用次数
func (c *AICrushController) GetRemainingCount(ctx *gin.Context) {
	// 1. 获取用户ID（从JWT中间件设置的上下文）
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户ID格式错误",
		})
		return
	}

	// 2. 调用服务层获取剩余次数
	remaining, err := c.service.GetRemainingCount(uid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取剩余次数失败",
			"code":  "internal_error",
		})
		return
	}

	// 3. 设置响应头
	ctx.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
	ctx.Header("X-RateLimit-Limit", "3")

	// 4. 返回成功响应
	ctx.JSON(http.StatusOK, gin.H{
		"remaining_count": remaining,
	})
}

// handleError 处理错误并返回适当的响应
func (c *AICrushController) handleError(ctx *gin.Context, err error) {
	errMsg := err.Error()

	// 根据错误信息返回不同的状态码和提示
	switch {
	case strings.Contains(errMsg, "今日击溃次数已用完"):
		ctx.JSON(http.StatusTooManyRequests, gin.H{
			"error": errMsg,
			"code":  "rate_limit_exceeded",
		})

	case strings.Contains(errMsg, "数据不足"):
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": errMsg,
			"code":  "insufficient_data",
		})

	case strings.Contains(errMsg, "系统繁忙"):
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": errMsg,
			"code":  "service_busy",
		})

	case strings.Contains(errMsg, "数据提取失败") || strings.Contains(errMsg, "系统错误"):
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "系统繁忙，请稍后再试",
			"code":  "internal_error",
		})

	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "服务暂时不可用，请稍后再试",
			"code":  "unknown_error",
		})
	}
}
