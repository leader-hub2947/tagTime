package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"size:128;uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"size:128;not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserSettings 用户设置
type UserSettings struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	UserID             uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	AutoArchiveTime    string    `gorm:"size:5;default:00:00" json:"auto_archive_time"` // 格式: HH:MM
	AutoArchiveEnabled bool      `gorm:"default:true" json:"auto_archive_enabled"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	Color     string    `gorm:"size:16;default:#4a90e2" json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Note struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `gorm:"not null;index" json:"user_id"`
	Content   string     `gorm:"type:text;not null" json:"content"`
	Images    *string    `gorm:"type:json" json:"images,omitempty"`
	IsDeleted bool       `gorm:"default:false;index" json:"is_deleted"` // 软删除标记
	DeletedAt *time.Time `gorm:"index" json:"deleted_at"`               // 删除时间
	CreatedAt time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Tags      []Tag      `gorm:"many2many:note_tags;" json:"tags"`
	Tasks     []Task     `gorm:"many2many:note_tasks;" json:"tasks"`
}

type NoteTag struct {
	NoteID uint `gorm:"primaryKey" json:"note_id"`
	TagID  uint `gorm:"primaryKey" json:"tag_id"`
}

type NoteTask struct {
	NoteID uint `gorm:"primaryKey" json:"note_id"`
	TaskID uint `gorm:"primaryKey" json:"task_id"`
}

type Task struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	UserID        uint       `gorm:"not null;index" json:"user_id"`
	TagID         *uint      `gorm:"index" json:"tag_id"` // 改为可空，删除标签时设置为null
	Name          string     `gorm:"size:128;not null" json:"name"`
	Description   string     `gorm:"type:text" json:"description"`
	Status        int8       `gorm:"default:0;index" json:"status"`   // 0-未开始，1-进行中，2-已完成，3-已归档
	TotalDuration int64      `gorm:"default:0" json:"total_duration"` // 秒
	CreatedAt     time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	ArchivedAt    *time.Time `gorm:"index" json:"archived_at"` // 归档时间
	Tag           *Tag       `gorm:"foreignKey:TagID" json:"tag,omitempty"`
}

type TimeEntry struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	TaskID         uint       `gorm:"not null;index" json:"task_id"`
	UserID         uint       `gorm:"not null;index" json:"user_id"`
	StartTime      time.Time  `gorm:"not null;index" json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	Duration       int        `gorm:"default:0" json:"duration"`              // 秒
	PausedDuration int        `gorm:"default:0" json:"paused_duration"`       // 累计暂停时长（秒）
	IsPaused       bool       `gorm:"default:false" json:"is_paused"`         // 是否暂停中
	LastPauseTime  *time.Time `json:"last_pause_time"`                        // 最后一次暂停时间
	TimerMode      string     `gorm:"size:20;default:free" json:"timer_mode"` // free-自由计时，pomodoro-番茄钟
	WorkMinutes    int        `gorm:"default:25" json:"work_minutes"`         // 番茄钟工作时长（分钟）
	BreakMinutes   int        `gorm:"default:5" json:"break_minutes"`         // 番茄钟休息时长（分钟）
	PomodoroCount  int        `gorm:"default:0" json:"pomodoro_count"`        // 番茄钟完成次数
	CreatedAt      time.Time  `json:"created_at"`
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&UserSettings{},
		&Tag{},
		&Note{},
		&NoteTag{},
		&NoteTask{},
		&Task{},
		&TimeEntry{},
	)
}
