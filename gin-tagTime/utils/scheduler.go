package utils

import (
	"log"
	"tagtime/config"
	"tagtime/models"
	"time"

	"github.com/robfig/cron/v3"
)

var cronScheduler *cron.Cron

// StartScheduler 启动定时任务调度器
func StartScheduler() {
	cronScheduler = cron.New(cron.WithSeconds())

	// 每分钟检查一次是否需要执行自动归档
	_, err := cronScheduler.AddFunc("0 * * * * *", checkAndAutoArchive)
	if err != nil {
		log.Printf("添加定时任务失败: %v", err)
		return
	}

	cronScheduler.Start()
	log.Println("定时任务调度器已启动")
}

// StopScheduler 停止定时任务调度器
func StopScheduler() {
	if cronScheduler != nil {
		cronScheduler.Stop()
		log.Println("定时任务调度器已停止")
	}
}

// checkAndAutoArchive 检查并执行自动归档
func checkAndAutoArchive() {
	now := time.Now()
	currentTime := now.Format("15:04")

	// 查询所有启用自动归档且时间匹配的用户
	var settings []models.UserSettings
	if err := config.DB.Where("auto_archive_enabled = ? AND auto_archive_time = ?", true, currentTime).Find(&settings).Error; err != nil {
		log.Printf("查询用户设置失败: %v", err)
		return
	}

	for _, setting := range settings {
		autoArchiveUserTasks(setting.UserID)
	}
}

// autoArchiveUserTasks 自动归档用户的所有已完成任务
func autoArchiveUserTasks(userID uint) {
	// 查询用户所有已完成但未归档的任务
	var tasks []models.Task
	if err := config.DB.Where("user_id = ? AND status = ?", userID, 2).Find(&tasks).Error; err != nil {
		log.Printf("查询用户 %d 的任务失败: %v", userID, err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	// 批量归档任务
	now := time.Now()
	for _, task := range tasks {
		task.Status = 3
		task.ArchivedAt = &now
		if err := config.DB.Save(&task).Error; err != nil {
			log.Printf("归档任务 %d 失败: %v", task.ID, err)
		}
	}

	log.Printf("用户 %d 自动归档了 %d 个任务", userID, len(tasks))
}
