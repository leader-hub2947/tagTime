package controllers

import (
	"net/http"
	"tagtime/config"
	"tagtime/models"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateTaskRequest struct {
	TagID       *uint  `json:"tag_id"` // 改为可选
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateTaskRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      *int8  `json:"status"`
	TagID       *uint  `json:"tag_id"`
}

func GetTasks(c *gin.Context) {
	userID := c.GetUint("user_id")
	tagID := c.Query("tag_id")
	status := c.Query("status")
	archived := c.Query("archived") // "true" 表示获取归档任务

	query := config.DB.Where("user_id = ?", userID)

	if tagID != "" {
		query = query.Where("tag_id = ?", tagID)
	}

	// 归档任务单独查询
	if archived == "true" {
		query = query.Where("status = ?", 3)
	} else {
		// 显示所有非归档任务（不限制日期）
		query = query.Where("status != ?", 3)

		if status != "" {
			query = query.Where("status = ?", status)
		}
	}

	var tasks []models.Task
	if err := query.Preload("Tag").Order("created_at DESC").Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取任务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func GetTask(c *gin.Context) {
	userID := c.GetUint("user_id")
	taskID := c.Param("id")

	var task models.Task
	if err := config.DB.Where("id = ? AND user_id = ?", taskID, userID).
		Preload("Tag").First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

func CreateTask(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 如果提供了标签ID，验证标签是否属于当前用户
	if req.TagID != nil {
		var tag models.Tag
		if err := config.DB.Where("id = ? AND user_id = ?", *req.TagID, userID).First(&tag).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "标签不存在"})
			return
		}
	}

	task := models.Task{
		UserID:      userID,
		TagID:       req.TagID,
		Name:        req.Name,
		Description: req.Description,
		Status:      0,
	}

	if err := config.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败"})
		return
	}

	config.DB.Preload("Tag").First(&task, task.ID)

	c.JSON(http.StatusCreated, gin.H{"task": task})
}

func UpdateTask(c *gin.Context) {
	userID := c.GetUint("user_id")
	taskID := c.Param("id")

	var task models.Task
	if err := config.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		task.Name = req.Name
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Status != nil {
		task.Status = *req.Status
		if *req.Status == 2 {
			now := time.Now()
			task.CompletedAt = &now
		}
	}
	if req.TagID != nil {
		var tag models.Tag
		if err := config.DB.Where("id = ? AND user_id = ?", *req.TagID, userID).First(&tag).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "标签不存在"})
			return
		}
		task.TagID = req.TagID
	}

	if err := config.DB.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新任务失败"})
		return
	}

	config.DB.Preload("Tag").First(&task, task.ID)

	c.JSON(http.StatusOK, gin.H{"task": task})
}

func DeleteTask(c *gin.Context) {
	userID := c.GetUint("user_id")
	taskID := c.Param("id")

	var task models.Task
	if err := config.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	if err := config.DB.Delete(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除任务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ArchiveTask 归档任务
func ArchiveTask(c *gin.Context) {
	userID := c.GetUint("user_id")
	taskID := c.Param("id")

	var task models.Task
	if err := config.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// 只有已完成的任务才能归档
	if task.Status != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只有已完成的任务才能归档"})
		return
	}

	now := time.Now()
	task.Status = 3
	task.ArchivedAt = &now

	if err := config.DB.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "归档任务失败"})
		return
	}

	config.DB.Preload("Tag").First(&task, task.ID)

	c.JSON(http.StatusOK, gin.H{"task": task, "message": "归档成功"})
}

// UnarchiveTask 取消归档任务
func UnarchiveTask(c *gin.Context) {
	userID := c.GetUint("user_id")
	taskID := c.Param("id")

	var task models.Task
	if err := config.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	if task.Status != 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该任务未归档"})
		return
	}

	task.Status = 2
	task.ArchivedAt = nil

	if err := config.DB.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消归档失败"})
		return
	}

	config.DB.Preload("Tag").First(&task, task.ID)

	c.JSON(http.StatusOK, gin.H{"task": task, "message": "取消归档成功"})
}

// GetArchivedTasks 获取归档任务（按天分组）
func GetArchivedTasks(c *gin.Context) {
	userID := c.GetUint("user_id")

	var tasks []models.Task
	if err := config.DB.Where("user_id = ? AND status = ?", userID, 3).
		Preload("Tag").
		Order("archived_at DESC").
		Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取归档任务失败"})
		return
	}

	// 按归档日期分组
	groupedTasks := make(map[string][]models.Task)
	for _, task := range tasks {
		if task.ArchivedAt != nil {
			dateKey := task.ArchivedAt.Format("2006-01-02")
			groupedTasks[dateKey] = append(groupedTasks[dateKey], task)
		}
	}

	c.JSON(http.StatusOK, gin.H{"grouped_tasks": groupedTasks})
}

// GetTaskNotes 获取任务的相关便签
func GetTaskNotes(c *gin.Context) {
	userID := c.GetUint("user_id")
	taskID := c.Param("id")

	// 验证任务是否存在且属于当前用户
	var task models.Task
	if err := config.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// 查询关联的便签
	var notes []models.Note
	if err := config.DB.Joins("JOIN note_tasks ON note_tasks.note_id = notes.id").
		Where("note_tasks.task_id = ? AND notes.user_id = ?", taskID, userID).
		Preload("Tags").Preload("Tasks").
		Order("notes.created_at DESC").
		Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取相关便签失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notes": notes, "count": len(notes)})
}
