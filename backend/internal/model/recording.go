package model

import "time"

// Recording 录音片段实体，audio_key 指向 MinIO 对象，status 承载状态机流转。
type Recording struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ProjectID       uint      `gorm:"index;not null" json:"project_id"`
	QuestionID      uint      `gorm:"index;not null" json:"question_id"`
	AudioKey        string    `gorm:"size:255" json:"audio_key"`
	DurationSeconds int       `gorm:"not null;default:0" json:"duration_seconds"`
	Summary         string    `gorm:"size:512" json:"summary"`
	Status          string    `gorm:"size:32;not null;default:recording" json:"status"`
	CreatedBy       uint      `gorm:"not null" json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (Recording) TableName() string { return "recordings" }
