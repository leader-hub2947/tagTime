package controllers

import (
	"net/http"
	"tagtime/config"
	"tagtime/models"

	"github.com/gin-gonic/gin"
)

type UpdateSettingsRequest struct {
	AutoArchiveTime    string `json:"auto_archive_time"`
	AutoArchiveEnabled *bool  `json:"auto_archive_enabled"`
}

// GetSettings 获取用户设置
func GetSettings(c *gin.Context) {
	userID := c.GetUint("user_id")

	var settings models.UserSettings
	if err := config.DB.Where("user_id = ?", userID).First(&settings).Error; err != nil {
		// 如果设置不存在，创建默认设置
		settings = models.UserSettings{
			UserID:             userID,
			AutoArchiveTime:    "00:00",
			AutoArchiveEnabled: true,
		}
		if err := config.DB.Create(&settings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建设置失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

// UpdateSettings 更新用户设置
func UpdateSettings(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var settings models.UserSettings
	if err := config.DB.Where("user_id = ?", userID).First(&settings).Error; err != nil {
		// 如果设置不存在，创建新设置
		settings = models.UserSettings{
			UserID:             userID,
			AutoArchiveTime:    "00:00",
			AutoArchiveEnabled: true,
		}
	}

	if req.AutoArchiveTime != "" {
		settings.AutoArchiveTime = req.AutoArchiveTime
	}
	if req.AutoArchiveEnabled != nil {
		settings.AutoArchiveEnabled = *req.AutoArchiveEnabled
	}

	if err := config.DB.Save(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新设置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings, "message": "设置已更新"})
}
