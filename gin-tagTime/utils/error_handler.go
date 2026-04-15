package utils

import (
	"errors"
	"fmt"
)

// 定义错误类型
var (
	ErrRateLimitExceeded = errors.New("rate_limit_exceeded")
	ErrInsufficientData  = errors.New("insufficient_data")
	ErrAIServiceTimeout  = errors.New("ai_service_timeout")
	ErrAIServiceError    = errors.New("ai_service_error")
	ErrDatabaseError     = errors.New("database_error")
	ErrCacheError        = errors.New("cache_error")
	ErrUnknownError      = errors.New("unknown_error")
)

// APIError API错误结构
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// Error 实现error接口
func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// ErrorHandler 错误处理器
type ErrorHandler struct{}

// NewErrorHandler 创建错误处理器
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// HandleError 处理错误并返回API错误
func (h *ErrorHandler) HandleError(err error) *APIError {
	switch {
	case errors.Is(err, ErrRateLimitExceeded):
		return &APIError{
			Code:    429,
			Message: "今日击溃次数已用完，明天再来吧",
			Type:    "rate_limit_exceeded",
		}

	case errors.Is(err, ErrInsufficientData):
		return &APIError{
			Code:    400,
			Message: "数据不足，请先使用一段时间再来",
			Type:    "insufficient_data",
		}

	case errors.Is(err, ErrAIServiceTimeout):
		return &APIError{
			Code:    504,
			Message: "AI服务响应超时，请稍后再试",
			Type:    "service_timeout",
		}

	case errors.Is(err, ErrAIServiceError):
		return &APIError{
			Code:    503,
			Message: "AI服务暂时不可用，请稍后再试",
			Type:    "service_unavailable",
		}

	case errors.Is(err, ErrDatabaseError):
		return &APIError{
			Code:    500,
			Message: "系统繁忙，请稍后再试",
			Type:    "internal_error",
		}

	case errors.Is(err, ErrCacheError):
		return &APIError{
			Code:    500,
			Message: "系统繁忙，请稍后再试",
			Type:    "internal_error",
		}

	default:
		return &APIError{
			Code:    500,
			Message: "服务暂时不可用，请稍后再试",
			Type:    "unknown_error",
		}
	}
}

// GetFriendlyMessage 获取友好的错误提示
func (h *ErrorHandler) GetFriendlyMessage(err error) string {
	apiErr := h.HandleError(err)
	return apiErr.Message
}
