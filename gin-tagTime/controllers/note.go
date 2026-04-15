package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"tagtime/config"
	"tagtime/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateNoteRequest struct {
	Content string `json:"content" binding:"required"`
	Images  string `json:"images"`
	TagIDs  []uint `json:"tag_ids"`
	TaskIDs []uint `json:"task_ids"` // 手动指定的任务ID
}

// parseTaskReferences 从便签内容中解析任务引用
// 支持 #任务名称 和 @任务名称 两种格式
func parseTaskReferences(content string, userID uint) []uint {
	var taskIDs []uint
	taskNames := make(map[string]bool)

	// 匹配 #任务名称 或 @任务名称
	// 任务名称可以包含中文、英文、数字、下划线、连字符
	re := regexp.MustCompile(`[#@]([\p{Han}\w\-]+)`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			taskName := strings.TrimSpace(match[1])
			if taskName != "" {
				taskNames[taskName] = true
			}
		}
	}

	// 根据任务名称查找任务ID
	if len(taskNames) > 0 {
		names := make([]string, 0, len(taskNames))
		for name := range taskNames {
			names = append(names, name)
		}

		var tasks []models.Task
		config.DB.Where("user_id = ? AND name IN ? AND status != ?", userID, names, 3).
			Select("id").Find(&tasks)

		for _, task := range tasks {
			taskIDs = append(taskIDs, task.ID)
		}
	}

	return taskIDs
}

func GetNotes(c *gin.Context) {
	userID := c.GetUint("user_id")
	tagID := c.Query("tag_id")
	taskID := c.Query("task_id")
	showDeleted := c.Query("show_deleted") == "true" // 是否显示回收站笔记

	query := config.DB.Where("user_id = ?", userID)

	// 根据参数决定是否显示已删除的笔记
	if showDeleted {
		query = query.Where("is_deleted = ?", true)
	} else {
		query = query.Where("is_deleted = ?", false)
	}

	if tagID != "" {
		query = query.Joins("JOIN note_tags ON note_tags.note_id = notes.id").
			Where("note_tags.tag_id = ?", tagID)
	}

	if taskID != "" {
		query = query.Joins("JOIN note_tasks ON note_tasks.note_id = notes.id").
			Where("note_tasks.task_id = ?", taskID)
	}

	var notes []models.Note
	if err := query.Preload("Tags").Preload("Tasks").Order("created_at DESC").Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取便签失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notes": notes})
}

func GetNote(c *gin.Context) {
	userID := c.GetUint("user_id")
	noteID := c.Param("id")

	var note models.Note
	if err := config.DB.Where("id = ? AND user_id = ?", noteID, userID).
		Preload("Tags").Preload("Tasks").First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "便签不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"note": note})
}

func CreateNote(c *gin.Context) {
	userID := c.GetUint("user_id")

	// 调试日志
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	var req CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 验证任务引用：从内容中解析任务引用并检查是否存在
	parsedTaskIDs := parseTaskReferences(req.Content, userID)
	taskIDMap := make(map[uint]bool)

	// 合并手动指定的任务ID
	for _, id := range req.TaskIDs {
		taskIDMap[id] = true
	}

	// 合并从内容中解析的任务ID
	for _, id := range parsedTaskIDs {
		taskIDMap[id] = true
	}

	// 验证所有任务是否存在
	if len(taskIDMap) > 0 {
		taskIDs := make([]uint, 0, len(taskIDMap))
		for id := range taskIDMap {
			taskIDs = append(taskIDs, id)
		}

		var tasks []models.Task
		if err := config.DB.Where("id IN ? AND user_id = ?", taskIDs, userID).Find(&tasks).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询任务失败: " + err.Error()})
			return
		}

		// 检查是否所有任务都存在
		if len(tasks) != len(taskIDs) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "部分任务不存在，无法创建笔记"})
			return
		}
	}

	note := models.Note{
		UserID:  userID,
		Content: req.Content,
	}

	// 只有当 images 不为空且不是空字符串时才设置
	if req.Images != "" && req.Images != `""` && req.Images != "[]" {
		note.Images = &req.Images
	}

	if err := config.DB.Create(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建便签失败: " + err.Error()})
		return
	}

	// 关联标签
	if len(req.TagIDs) > 0 {
		var tags []models.Tag
		if err := config.DB.Where("id IN ? AND user_id = ?", req.TagIDs, userID).Find(&tags).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询标签失败: " + err.Error()})
			return
		}
		if len(tags) > 0 {
			if err := config.DB.Model(&note).Association("Tags").Append(tags); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "关联标签失败: " + err.Error()})
				return
			}
		}
	}

	// 关联任务
	if len(taskIDMap) > 0 {
		taskIDs := make([]uint, 0, len(taskIDMap))
		for id := range taskIDMap {
			taskIDs = append(taskIDs, id)
		}

		var tasks []models.Task
		if err := config.DB.Where("id IN ? AND user_id = ?", taskIDs, userID).Find(&tasks).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询任务失败: " + err.Error()})
			return
		}
		if len(tasks) > 0 {
			if err := config.DB.Model(&note).Association("Tasks").Append(tasks); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "关联任务失败: " + err.Error()})
				return
			}
		}
	}

	if err := config.DB.Preload("Tags").Preload("Tasks").First(&note, note.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载便签失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"note": note})
}

func UpdateNote(c *gin.Context) {
	userID := c.GetUint("user_id")
	noteID := c.Param("id")

	var note models.Note
	if err := config.DB.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "便签不存在"})
		return
	}

	var req CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证任务引用：从内容中解析任务引用并检查是否存在
	parsedTaskIDs := parseTaskReferences(req.Content, userID)
	taskIDMap := make(map[uint]bool)

	// 合并手动指定的任务ID
	for _, id := range req.TaskIDs {
		taskIDMap[id] = true
	}

	// 合并从内容中解析的任务ID
	for _, id := range parsedTaskIDs {
		taskIDMap[id] = true
	}

	// 验证所有任务是否存在
	if len(taskIDMap) > 0 {
		taskIDs := make([]uint, 0, len(taskIDMap))
		for id := range taskIDMap {
			taskIDs = append(taskIDs, id)
		}

		var tasks []models.Task
		if err := config.DB.Where("id IN ? AND user_id = ?", taskIDs, userID).Find(&tasks).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询任务失败: " + err.Error()})
			return
		}

		// 检查是否所有任务都存在
		if len(tasks) != len(taskIDs) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "部分任务不存在，无法更新笔记"})
			return
		}
	}

	note.Content = req.Content
	// 只有当 images 不为空且不是空字符串时才设置
	if req.Images != "" && req.Images != `""` && req.Images != "[]" {
		note.Images = &req.Images
	} else {
		// 清空图片字段（设置为 NULL）
		note.Images = nil
	}

	if err := config.DB.Save(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新便签失败"})
		return
	}

	// 更新标签关联
	config.DB.Model(&note).Association("Tags").Clear()
	if len(req.TagIDs) > 0 {
		var tags []models.Tag
		config.DB.Where("id IN ? AND user_id = ?", req.TagIDs, userID).Find(&tags)
		if len(tags) > 0 {
			config.DB.Model(&note).Association("Tags").Append(tags)
		}
	}

	// 更新任务关联
	config.DB.Model(&note).Association("Tasks").Clear()
	if len(taskIDMap) > 0 {
		taskIDs := make([]uint, 0, len(taskIDMap))
		for id := range taskIDMap {
			taskIDs = append(taskIDs, id)
		}

		var tasks []models.Task
		config.DB.Where("id IN ? AND user_id = ?", taskIDs, userID).Find(&tasks)
		if len(tasks) > 0 {
			config.DB.Model(&note).Association("Tasks").Append(tasks)
		}
	}

	config.DB.Preload("Tags").Preload("Tasks").First(&note, note.ID)

	c.JSON(http.StatusOK, gin.H{"note": note})
}

func DeleteNote(c *gin.Context) {
	userID := c.GetUint("user_id")
	noteID := c.Param("id")

	var note models.Note
	if err := config.DB.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "便签不存在"})
		return
	}

	// 如果笔记已经在回收站，则物理删除
	if note.IsDeleted {
		if err := config.DB.Delete(&note).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除便签失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "永久删除成功"})
		return
	}

	// 否则执行软删除
	now := time.Now()
	note.IsDeleted = true
	note.DeletedAt = &now

	if err := config.DB.Save(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除便签失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已移至回收站"})
}

// RestoreNote 从回收站恢复笔记
func RestoreNote(c *gin.Context) {
	userID := c.GetUint("user_id")
	noteID := c.Param("id")

	var note models.Note
	if err := config.DB.Where("id = ? AND user_id = ? AND is_deleted = ?", noteID, userID, true).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "便签不存在或未在回收站中"})
		return
	}

	note.IsDeleted = false
	note.DeletedAt = nil

	if err := config.DB.Save(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "恢复便签失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "恢复成功"})
}

// EmptyTrash 清空回收站
func EmptyTrash(c *gin.Context) {
	userID := c.GetUint("user_id")

	// 查找所有回收站中的笔记
	var notes []models.Note
	if err := config.DB.Where("user_id = ? AND is_deleted = ?", userID, true).Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询回收站失败"})
		return
	}

	if len(notes) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "回收站已空"})
		return
	}

	// 物理删除所有回收站笔记
	if err := config.DB.Where("user_id = ? AND is_deleted = ?", userID, true).Delete(&models.Note{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清空回收站失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("已清空 %d 条笔记", len(notes))})
}

func GetNoteCalendar(c *gin.Context) {
	userID := c.GetUint("user_id")
	year := c.Param("year")
	month := c.Param("month")

	yearInt, _ := strconv.Atoi(year)
	monthInt, _ := strconv.Atoi(month)

	startDate := time.Date(yearInt, time.Month(monthInt), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)

	var notes []models.Note
	// 只统计未删除的笔记
	if err := config.DB.Where("user_id = ? AND is_deleted = ? AND created_at >= ? AND created_at < ?",
		userID, false, startDate, endDate).Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取日历数据失败"})
		return
	}

	// 统计每天的便签数量
	dateMap := make(map[string]int)
	for _, note := range notes {
		date := note.CreatedAt.Format("2006-01-02")
		dateMap[date]++
	}

	c.JSON(http.StatusOK, gin.H{"calendar": dateMap})
}

// UploadImage 上传图片
func UploadImage(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择图片文件"})
		return
	}

	// 验证文件类型
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只支持 jpg, jpeg, png, gif, webp 格式的图片"})
		return
	}

	// 验证文件大小（限制为 5MB）
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片大小不能超过 5MB"})
		return
	}

	// 创建上传目录
	uploadDir := "uploads/images"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%s_%s%s", time.Now().Format("20060102150405"), uuid.New().String()[:8], ext)
	filepath := filepath.Join(uploadDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存图片失败"})
		return
	}

	// 返回图片 URL
	imageURL := fmt.Sprintf("/uploads/images/%s", filename)
	c.JSON(http.StatusOK, gin.H{"url": imageURL})
}

// DeleteImage 删除图片
func DeleteImage(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	var req struct {
		URL string `json:"url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 从 URL 中提取文件路径
	// URL 格式: /uploads/images/filename.jpg
	filepath := "." + req.URL

	// 删除文件
	if err := os.Remove(filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除图片失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
