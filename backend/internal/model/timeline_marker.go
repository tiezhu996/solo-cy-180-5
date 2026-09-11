package model

import "time"

// TimelineMarker 时间轴关键节点实体，标注录音片段中的关键时间点。
type TimelineMarker struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ProjectID       uint      `gorm:"index;not null" json:"project_id"`
	RecordingID     uint      `gorm:"index;not null" json:"recording_id"`
	TimestampSecond int       `gorm:"not null" json:"timestamp_second"`
	Label           string    `gorm:"size:128;not null" json:"label"`
	Note            string    `gorm:"size:512" json:"note"`
	CreatedBy       uint      `gorm:"not null" json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (TimelineMarker) TableName() string { return "timeline_markers" }
