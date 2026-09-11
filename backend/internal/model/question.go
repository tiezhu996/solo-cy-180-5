package model

import "time"

// Question 采访问题实体，隶属于采访项目。
type Question struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `gorm:"index;not null" json:"project_id"`
	Content   string    `gorm:"size:512;not null" json:"content"`
	SortOrder int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (Question) TableName() string { return "questions" }
