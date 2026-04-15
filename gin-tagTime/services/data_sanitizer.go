package services

import (
	"regexp"
	"tagtime/models"
)

// DataSanitizer 数据脱敏器
type DataSanitizer struct {
	phoneRegex  *regexp.Regexp
	emailRegex  *regexp.Regexp
	idCardRegex *regexp.Regexp
	bankRegex   *regexp.Regexp
}

// NewDataSanitizer 创建数据脱敏器
func NewDataSanitizer() *DataSanitizer {
	return &DataSanitizer{
		phoneRegex:  regexp.MustCompile(`1[3-9]\d{9}`),
		emailRegex:  regexp.MustCompile(`[\w.-]+@[\w.-]+\.\w+`),
		idCardRegex: regexp.MustCompile(`\d{17}[\dXx]`),
		bankRegex:   regexp.MustCompile(`\d{16,19}`),
	}
}

// SanitizeContent 脱敏单个文本内容
func (s *DataSanitizer) SanitizeContent(content string) string {
	// 1. 移除手机号
	content = s.phoneRegex.ReplaceAllString(content, "[手机号]")

	// 2. 移除邮箱
	content = s.emailRegex.ReplaceAllString(content, "[邮箱]")

	// 3. 移除身份证号
	content = s.idCardRegex.ReplaceAllString(content, "[身份证]")

	// 4. 移除银行卡号
	content = s.bankRegex.ReplaceAllString(content, "[卡号]")

	return content
}

// SanitizeSummary 脱敏用户摘要
func (s *DataSanitizer) SanitizeSummary(summary *models.UserSummary) *models.UserSummary {
	// 脱敏便签内容
	for i, note := range summary.RecentNotes {
		summary.RecentNotes[i] = s.SanitizeContent(note)
	}

	// 脱敏任务名称
	for i, task := range summary.TaskStats.UnfinishedTasks {
		summary.TaskStats.UnfinishedTasks[i] = s.SanitizeContent(task)
	}

	for i, task := range summary.TaskStats.OngoingTasks {
		summary.TaskStats.OngoingTasks[i] = s.SanitizeContent(task)
	}

	// 脱敏关键词（虽然关键词一般不包含敏感信息，但为了安全起见）
	for i, keyword := range summary.KeyWords {
		summary.KeyWords[i] = s.SanitizeContent(keyword)
	}

	return summary
}
