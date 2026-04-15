package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"tagtime/config"
	"tagtime/models"
)

// AIClient AI客户端
type AIClient struct {
	config     *config.AIConfig
	httpClient *http.Client
}

// NewAIClient 创建AI客户端
func NewAIClient(cfg *config.AIConfig) *AIClient {
	return &AIClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// GenerateCrushLine 生成击溃语
func (c *AIClient) GenerateCrushLine(prompt string) (string, error) {
	switch c.config.Provider {
	case "openai":
		return c.callOpenAI(prompt)
	case "zhipu":
		return c.callZhipuAI(prompt)
	case "deepseek":
		return c.callDeepSeek(prompt)
	default:
		return c.callOpenAI(prompt)
	}
}

// callOpenAI 调用OpenAI API
func (c *AIClient) callOpenAI(prompt string) (string, error) {
	// 构建请求
	request := models.AIRequest{
		Model: c.config.Model,
		Messages: []models.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.8,
		MaxTokens:   200,
		Stream:      false,
	}

	// 序列化请求
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", c.config.Endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API返回错误: %d, %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response models.AIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 提取击溃语
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("AI返回空响应")
	}

	crushLine := response.Choices[0].Message.Content

	// 截断过长的文本
	if len(crushLine) > 200 {
		crushLine = crushLine[:200]
	}

	return crushLine, nil
}

// callZhipuAI 调用智谱清言API（示例实现）
func (c *AIClient) callZhipuAI(prompt string) (string, error) {
	// 智谱清言的API格式可能不同，这里提供一个基本框架
	// 实际使用时需要根据智谱清言的API文档调整

	type ZhipuRequest struct {
		Model       string           `json:"model"`
		Messages    []models.Message `json:"messages"`
		Temperature float64          `json:"temperature"`
		MaxTokens   int              `json:"max_tokens"`
		Stream      bool             `json:"stream"`
	}

	request := ZhipuRequest{
		Model: c.config.Model,
		Messages: []models.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.8,
		MaxTokens:   200,
		Stream:      false,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", c.config.Endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API返回错误: %d, %s", resp.StatusCode, string(body))
	}

	// 解析响应（格式可能与OpenAI不同）
	var response models.AIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("AI返回空响应")
	}

	crushLine := response.Choices[0].Message.Content

	if len(crushLine) > 200 {
		crushLine = crushLine[:200]
	}

	return crushLine, nil
}

// callDeepSeek 调用DeepSeek API
func (c *AIClient) callDeepSeek(prompt string) (string, error) {
	// 构建请求
	// DeepSeek Reasoner 模型需要更多 token：推理过程 + 最终回复
	request := models.AIRequest{
		Model: c.config.Model,
		Messages: []models.Message{
			{
				Role:    "system",
				Content: "你是一个智能助手，擅长根据用户的工作数据生成简短、有力、激励性的话语。请直接输出一句话，不要解释推理过程。",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.8,
		MaxTokens:   2000,  // 增加到 2000，确保推理过程和最终回复都有足够空间
		Stream:      false, // 明确设置为非流式输出
	}

	// 序列化请求
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	fmt.Printf("[DEBUG] DeepSeek 请求体: %s\n", string(requestBody))

	// 创建HTTP请求
	req, err := http.NewRequest("POST", c.config.Endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API返回错误: %d, %s", resp.StatusCode, string(body))
	}

	// 添加调试日志
	fmt.Printf("[DEBUG] DeepSeek API 原始响应: %s\n", string(body))

	// 解析响应
	var response models.AIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("解析响应失败: %w, 响应内容: %s", err, string(body))
	}

	// 调试日志
	fmt.Printf("[DEBUG] 解析后的响应: Choices数量=%d\n", len(response.Choices))
	if len(response.Choices) > 0 {
		fmt.Printf("[DEBUG] Choice[0].Message.Content长度=%d\n", len(response.Choices[0].Message.Content))
		fmt.Printf("[DEBUG] Choice[0].Message.Content前100字符=%s\n",
			func() string {
				content := response.Choices[0].Message.Content
				if len(content) > 100 {
					return content[:100]
				}
				return content
			}())
		fmt.Printf("[DEBUG] Choice[0].ReasoningContent长度=%d\n", len(response.Choices[0].ReasoningContent))
	}

	// 提取击溃语
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("AI返回空响应")
	}

	// DeepSeek Reasoner 模型的响应在 Message.Content 中
	// 如果 Content 为空，尝试从 ReasoningContent 提取最后一句话
	crushLine := response.Choices[0].Message.Content
	if crushLine == "" && response.Choices[0].ReasoningContent != "" {
		fmt.Printf("[DEBUG] Content为空，从ReasoningContent提取最后一句\n")
		// 从推理内容中提取最后一句完整的话（通常是结论）
		reasoningContent := response.Choices[0].ReasoningContent
		// 尝试找到最后一个句号、问号或感叹号之后的内容
		lastSentenceStart := -1
		runes := []rune(reasoningContent)
		for i := len(runes) - 1; i >= 0; i-- {
			if runes[i] == '。' || runes[i] == '？' || runes[i] == '！' || runes[i] == '.' {
				if i+1 < len(runes) { // 确保后面还有内容
					lastSentenceStart = i + 1
					break
				}
			}
		}
		if lastSentenceStart > 0 && lastSentenceStart < len(runes) {
			crushLine = string(runes[lastSentenceStart:])
		} else {
			// 如果找不到句号，取最后100个字符
			if len(runes) > 100 {
				crushLine = string(runes[len(runes)-100:])
			} else {
				crushLine = reasoningContent
			}
		}
	}

	// 检查是否仍然为空
	if crushLine == "" {
		return "", fmt.Errorf("AI返回的内容为空")
	}

	// 清理和截断文本
	// 移除首尾空白
	crushLine = string(bytes.TrimSpace([]byte(crushLine)))

	// 如果文本过长，截断到最后一个完整句子（按字符数，不是字节数）
	runes := []rune(crushLine)
	if len(runes) > 100 {
		// 尝试在100字符内找到最后一个句号
		truncated := runes[:100]
		lastPeriod := -1
		for i := len(truncated) - 1; i >= 0; i-- {
			if truncated[i] == '。' || truncated[i] == '？' || truncated[i] == '！' || truncated[i] == '.' {
				lastPeriod = i + 1
				break
			}
		}
		if lastPeriod > 0 {
			crushLine = string(truncated[:lastPeriod])
		} else {
			crushLine = string(truncated)
		}
	}

	fmt.Printf("[DEBUG] 最终返回的击溃语: %s\n", crushLine)

	return crushLine, nil
}
