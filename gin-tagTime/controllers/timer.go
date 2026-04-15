package controllers

import (
	"net/http"
	"tagtime/config"
	"tagtime/models"
	"time"

	"github.com/gin-gonic/gin"
)

func StartTimer(c *gin.Context) {
	userID := c.GetUint("user_id")
	taskID := c.Param("id")

	var req struct {
		TimerMode    string `json:"timer_mode"`    // free 或 pomodoro
		WorkMinutes  int    `json:"work_minutes"`  // 番茄钟工作时长（分钟）
		BreakMinutes int    `json:"break_minutes"` // 番茄钟休息时长（分钟）
	}
	c.ShouldBindJSON(&req)
	if req.TimerMode == "" {
		req.TimerMode = "free"
	}
	// 设置默认值
	if req.WorkMinutes == 0 {
		req.WorkMinutes = 25
	}
	if req.BreakMinutes == 0 {
		req.BreakMinutes = 5
	}

	// 检查任务是否存在
	var task models.Task
	if err := config.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// 检查是否有正在进行的计时
	var activeEntry models.TimeEntry
	if err := config.DB.Where("user_id = ? AND end_time IS NULL", userID).First(&activeEntry).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "已有正在进行的计时任务", "active_entry": activeEntry})
		return
	}

	// 创建新的计时记录
	entry := models.TimeEntry{
		TaskID:       task.ID,
		UserID:       userID,
		StartTime:    time.Now(),
		TimerMode:    req.TimerMode,
		WorkMinutes:  req.WorkMinutes,
		BreakMinutes: req.BreakMinutes,
	}

	if err := config.DB.Create(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "开始计时失败"})
		return
	}

	// 更新任务状态为进行中
	task.Status = 1
	config.DB.Save(&task)

	c.JSON(http.StatusOK, gin.H{
		"message": "计时已开始",
		"entry":   entry,
		"task":    task,
	})
}

func PauseTimer(c *gin.Context) {
	userID := c.GetUint("user_id")

	var entry models.TimeEntry
	if err := config.DB.Where("user_id = ? AND end_time IS NULL AND is_paused = ?", userID, false).First(&entry).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "没有正在进行的计时"})
		return
	}

	now := time.Now()
	entry.IsPaused = true
	entry.LastPauseTime = &now

	if err := config.DB.Save(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "暂停失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "计时已暂停",
		"entry":   entry,
	})
}

func ResumeTimer(c *gin.Context) {
	userID := c.GetUint("user_id")

	var entry models.TimeEntry
	if err := config.DB.Where("user_id = ? AND end_time IS NULL AND is_paused = ?", userID, true).First(&entry).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "没有暂停中的计时"})
		return
	}

	if entry.LastPauseTime != nil {
		pausedSeconds := int(time.Since(*entry.LastPauseTime).Seconds())
		entry.PausedDuration += pausedSeconds
	}

	entry.IsPaused = false
	entry.LastPauseTime = nil

	if err := config.DB.Save(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "恢复失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "计时已恢复",
		"entry":   entry,
	})
}

func EndTimer(c *gin.Context) {
	userID := c.GetUint("user_id")

	var entry models.TimeEntry
	if err := config.DB.Where("user_id = ? AND end_time IS NULL", userID).First(&entry).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "计时记录不存在或已结束"})
		return
	}

	now := time.Now()
	entry.EndTime = &now

	// 如果当前是暂停状态，先计算暂停时长
	if entry.IsPaused && entry.LastPauseTime != nil {
		pausedSeconds := int(now.Sub(*entry.LastPauseTime).Seconds())
		entry.PausedDuration += pausedSeconds
	}

	// 总时长 = 结束时间 - 开始时间 - 暂停时长
	totalSeconds := int(now.Sub(entry.StartTime).Seconds())
	entry.Duration = totalSeconds - entry.PausedDuration
	entry.IsPaused = false

	if err := config.DB.Save(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "结束计时失败"})
		return
	}

	// 更新任务的总时长
	var task models.Task
	config.DB.First(&task, entry.TaskID)
	task.TotalDuration += int64(entry.Duration)
	config.DB.Save(&task)

	c.JSON(http.StatusOK, gin.H{
		"message": "计时已结束",
		"entry":   entry,
		"task":    task,
	})
}

func GetCurrentTimer(c *gin.Context) {
	userID := c.GetUint("user_id")

	var entry models.TimeEntry
	if err := config.DB.Where("user_id = ? AND end_time IS NULL", userID).First(&entry).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"entry": nil})
		return
	}

	var task models.Task
	config.DB.Preload("Tag").First(&task, entry.TaskID)

	c.JSON(http.StatusOK, gin.H{
		"entry": entry,
		"task":  task,
	})
}

func SwitchTimer(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		NewTaskID    uint   `json:"new_task_id" binding:"required"`
		TimerMode    string `json:"timer_mode"`
		WorkMinutes  int    `json:"work_minutes"`
		BreakMinutes int    `json:"break_minutes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.TimerMode == "" {
		req.TimerMode = "free"
	}
	if req.WorkMinutes == 0 {
		req.WorkMinutes = 25
	}
	if req.BreakMinutes == 0 {
		req.BreakMinutes = 5
	}

	// 结束当前计时
	var currentEntry models.TimeEntry
	if err := config.DB.Where("user_id = ? AND end_time IS NULL", userID).First(&currentEntry).Error; err == nil {
		now := time.Now()

		// 如果当前是暂停状态，先计算暂停时长
		if currentEntry.IsPaused && currentEntry.LastPauseTime != nil {
			pausedSeconds := int(now.Sub(*currentEntry.LastPauseTime).Seconds())
			currentEntry.PausedDuration += pausedSeconds
		}

		currentEntry.EndTime = &now
		totalSeconds := int(now.Sub(currentEntry.StartTime).Seconds())
		currentEntry.Duration = totalSeconds - currentEntry.PausedDuration
		currentEntry.IsPaused = false
		config.DB.Save(&currentEntry)

		var oldTask models.Task
		config.DB.First(&oldTask, currentEntry.TaskID)
		oldTask.TotalDuration += int64(currentEntry.Duration)
		config.DB.Save(&oldTask)
	}

	// 开始新任务计时
	var newTask models.Task
	if err := config.DB.Where("id = ? AND user_id = ?", req.NewTaskID, userID).Preload("Tag").First(&newTask).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "新任务不存在"})
		return
	}

	newEntry := models.TimeEntry{
		TaskID:       newTask.ID,
		UserID:       userID,
		StartTime:    time.Now(),
		TimerMode:    req.TimerMode,
		WorkMinutes:  req.WorkMinutes,
		BreakMinutes: req.BreakMinutes,
	}

	if err := config.DB.Create(&newEntry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "切换任务失败"})
		return
	}

	newTask.Status = 1
	config.DB.Save(&newTask)

	c.JSON(http.StatusOK, gin.H{
		"message": "任务已切换",
		"entry":   newEntry,
		"task":    newTask,
	})
}

func CompleteTask(c *gin.Context) {
	userID := c.GetUint("user_id")
	taskID := c.Param("id")

	var req struct {
		CompletedAt string `json:"completed_at"` // ISO 8601 格式的时间字符串
	}
	c.ShouldBindJSON(&req)

	var task models.Task
	if err := config.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// 如果有正在进行的计时，先结束它
	var activeEntry models.TimeEntry
	if err := config.DB.Where("task_id = ? AND end_time IS NULL", task.ID).First(&activeEntry).Error; err == nil {
		now := time.Now()

		if activeEntry.IsPaused && activeEntry.LastPauseTime != nil {
			pausedSeconds := int(now.Sub(*activeEntry.LastPauseTime).Seconds())
			activeEntry.PausedDuration += pausedSeconds
		}

		activeEntry.EndTime = &now
		totalSeconds := int(now.Sub(activeEntry.StartTime).Seconds())
		activeEntry.Duration = totalSeconds - activeEntry.PausedDuration
		activeEntry.IsPaused = false
		config.DB.Save(&activeEntry)

		task.TotalDuration += int64(activeEntry.Duration)
	}

	// 解析完成时间，如果没有提供则使用当前时间
	var completedAt time.Time
	if req.CompletedAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, req.CompletedAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "完成时间格式错误"})
			return
		}
		completedAt = parsedTime
	} else {
		completedAt = time.Now()
	}

	task.Status = 2
	task.CompletedAt = &completedAt
	config.DB.Save(&task)

	c.JSON(http.StatusOK, gin.H{
		"message": "任务已完成",
		"task":    task,
	})
}
