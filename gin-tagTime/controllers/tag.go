package controllers

import (
	"net/http"
	"tagtime/config"
	"tagtime/models"

	"github.com/gin-gonic/gin"
)

type CreateTagRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

func GetTags(c *gin.Context) {
	userID := c.GetUint("user_id")

	var tags []models.Tag
	if err := config.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取标签失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

func CreateTag(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查标签名是否已存在
	var existingTag models.Tag
	if err := config.DB.Where("user_id = ? AND name = ?", userID, req.Name).First(&existingTag).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "标签名已存在"})
		return
	}

	color := req.Color
	if color == "" {
		color = "#4a90e2"
	}

	tag := models.Tag{
		UserID: userID,
		Name:   req.Name,
		Color:  color,
	}

	if err := config.DB.Create(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建标签失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"tag": tag})
}

func UpdateTag(c *gin.Context) {
	userID := c.GetUint("user_id")
	tagID := c.Param("id")

	var tag models.Tag
	if err := config.DB.Where("id = ? AND user_id = ?", tagID, userID).First(&tag).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "标签不存在"})
		return
	}

	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag.Name = req.Name
	if req.Color != "" {
		tag.Color = req.Color
	}

	if err := config.DB.Save(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新标签失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tag": tag})
}

type DeleteTagRequest struct {
	DeleteNotes bool `json:"delete_notes"` // true: 删除标签和笔记, false: 仅删除标签
}

func DeleteTag(c *gin.Context) {
	userID := c.GetUint("user_id")
	tagID := c.Param("id")

	// 先验证标签是否存在
	var tag models.Tag
	if err := config.DB.Where("id = ? AND user_id = ?", tagID, userID).First(&tag).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "标签不存在"})
		return
	}

	// 解析请求体，如果解析失败则默认为仅删除标签
	var req DeleteTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// DELETE 请求可能没有 body，默认为仅删除标签
		req.DeleteNotes = false
	}

	// 开始事务
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 首先处理关联的任务（无论哪种模式都需要处理）
	// 将关联任务的 tag_id 设置为 null，而不是删除任务
	if err := tx.Model(&models.Task{}).Where("tag_id = ?", tagID).Update("tag_id", nil).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新关联任务失败"})
		return
	}

	if req.DeleteNotes {
		// 模式1: 删除标签和所有关联的笔记
		// 1. 查找所有关联的笔记ID
		var noteIDs []uint
		if err := tx.Table("note_tags").Where("tag_id = ?", tagID).Pluck("note_id", &noteIDs).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询关联笔记失败"})
			return
		}

		// 2. 删除这些笔记（级联删除会自动处理 note_tags 和 note_tasks）
		if len(noteIDs) > 0 {
			if err := tx.Where("id IN ?", noteIDs).Delete(&models.Note{}).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "删除关联笔记失败"})
				return
			}
		}
	} else {
		// 模式2: 仅删除标签，从笔记内容中移除标签引用
		// 1. 查找所有关联的笔记
		var notes []models.Note
		if err := tx.Table("notes").
			Joins("JOIN note_tags ON notes.id = note_tags.note_id").
			Where("note_tags.tag_id = ? AND notes.user_id = ?", tagID, userID).
			Find(&notes).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询关联笔记失败"})
			return
		}

		// 2. 从每条笔记的 content 中移除 "#标签名 " 的内容（仅当有关联笔记时）
		if len(notes) > 0 {
			tagPattern := "#" + tag.Name + " "
			for _, note := range notes {
				// 使用字符串替换移除标签引用
				newContent := note.Content
				// 移除 "#标签名 " (后面有空格)
				newContent = replaceAll(newContent, tagPattern, "")
				// 移除 "#标签名" (后面没有空格，在行尾或文本末尾)
				newContent = replaceAll(newContent, "#"+tag.Name, "")

				// 更新笔记内容
				if err := tx.Model(&models.Note{}).Where("id = ?", note.ID).Update("content", newContent).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "更新笔记内容失败"})
					return
				}
			}
		}

		// 3. 删除 note_tags 关联关系（如果存在）
		// GORM 的 Delete 在没有记录时不会报错，直接执行即可
		tx.Where("tag_id = ?", tagID).Delete(&models.NoteTag{})
	}

	// 删除标签本身
	if err := tx.Delete(&tag).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除标签失败"})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// 辅助函数：替换所有匹配的字符串
func replaceAll(s, old, new string) string {
	result := s
	for {
		replaced := false
		for i := 0; i < len(result); i++ {
			if i+len(old) <= len(result) && result[i:i+len(old)] == old {
				result = result[:i] + new + result[i+len(old):]
				replaced = true
				break
			}
		}
		if !replaced {
			break
		}
	}
	return result
}
