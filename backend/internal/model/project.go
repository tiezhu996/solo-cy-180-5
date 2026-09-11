package model

import "time"

// Project 采访项目实体，status 承载状态机流转。
type Project struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Title           string    `gorm:"size:128;not null" json:"title"`
	IntervieweeName string    `gorm:"size:64;not null" json:"interviewee_name"`
	BirthYear       int       `gorm:"not null" json:"birth_year"`
	Background      string    `gorm:"type:text" json:"background"`
	Status          string    `gorm:"size:32;not null;default:draft" json:"status"`
	CreatedBy       uint      `gorm:"not null" json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Questions       []Question      `gorm:"foreignKey:ProjectID" json:"questions,omitempty"`
	Recordings      []Recording     `gorm:"foreignKey:ProjectID" json:"recordings,omitempty"`
	TimelineMarkers []TimelineMarker `gorm:"foreignKey:ProjectID" json:"timeline_markers,omitempty"`
}

// TableName 指定表名。
func (Project) TableName() string { return "projects" }
