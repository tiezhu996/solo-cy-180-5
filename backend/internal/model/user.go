// Package model 定义数据库实体。
package model

import "time"

// User 用户实体，role 字段承载 RBAC 权限。
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	DisplayName  string    `gorm:"size:64;not null" json:"display_name"`
	Email        string    `gorm:"size:128" json:"email"`
	Role         string    `gorm:"size:32;not null;default:interviewer" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (User) TableName() string { return "users" }
