package controllers

import (
	"net/http"
	"tagtime/config"
	"tagtime/models"
	"time"

	"github.com/gin-gonic/gin"
)

type TimelineItem struct {
	TaskID    uint       `json:"task_id"`
	TaskName  string     `json:"task_name"`
	TagName   string     `json:"tag_name"`
	TagColor  string     `json:"tag_color"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	Duration  int        `json:"duration"`
}

type TagRank struct {
	TagID         uint   `json:"tag_id"`
	TagName       string `json:"tag_name"`
	TagColor      string `json:"tag_color"`
	TotalDuration int64  `json:"total_duration"`
}

func GetTimeline(c *gin.Context) {
	userID := c.GetUint("user_id")
	date := c.Query("date")
	var targetDate time.Time
	if date == "" {
		targetDate = time.Now()
	} else {
		var err error
		targetDate, err = time.Parse("2006-01-02", date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "date format error"})
			return
		}
	}
	startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.Local)
	endOfDay := startOfDay.Add(24 * time.Hour)
	var entries []models.TimeEntry
	if err := config.DB.Where("user_id = ? AND start_time >= ? AND start_time < ?", userID, startOfDay, endOfDay).Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get timeline"})
		return
	}
	var timeline []TimelineItem
	for _, entry := range entries {
		var task models.Task
		config.DB.Preload("Tag").First(&task, entry.TaskID)
		timeline = append(timeline, TimelineItem{
			TaskID:    task.ID,
			TaskName:  task.Name,
			TagName:   task.Tag.Name,
			TagColor:  task.Tag.Color,
			StartTime: entry.StartTime,
			EndTime:   entry.EndTime,
			Duration:  entry.Duration,
		})
	}
	c.JSON(http.StatusOK, gin.H{"timeline": timeline})
}

func GetTagRanking(c *gin.Context) {
	userID := c.GetUint("user_id")
	var rankings []TagRank
	err := config.DB.Table("tasks").
		Select("tags.id as tag_id, tags.name as tag_name, tags.color as tag_color, SUM(tasks.total_duration) as total_duration").
		Joins("JOIN tags ON tags.id = tasks.tag_id").
		Where("tasks.user_id = ?", userID).
		Group("tags.id, tags.name, tags.color").
		Order("total_duration DESC").
		Scan(&rankings).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get ranking"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rankings": rankings})
}

func GetTaskStatistics(c *gin.Context) {
	userID := c.GetUint("user_id")
	period := c.Query("period")
	var startDate time.Time
	now := time.Now()
	switch period {
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	default:
		startDate = now.AddDate(0, -1, 0)
	}
	var totalTasks int64
	var completedTasks int64
	config.DB.Model(&models.Task{}).Where("user_id = ? AND created_at >= ?", userID, startDate).Count(&totalTasks)
	config.DB.Model(&models.Task{}).Where("user_id = ? AND status = 2 AND created_at >= ?", userID, startDate).Count(&completedTasks)
	completionRate := 0.0
	if totalTasks > 0 {
		completionRate = float64(completedTasks) / float64(totalTasks) * 100
	}
	c.JSON(http.StatusOK, gin.H{
		"total_tasks":     totalTasks,
		"completed_tasks": completedTasks,
		"completion_rate": completionRate,
	})
}
