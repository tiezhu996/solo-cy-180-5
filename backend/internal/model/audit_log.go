package model

import "time"

// AuditLog 操作审计日志实体。
type AuditLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	Username   string    `gorm:"size:64;not null" json:"username"`
	Role       string    `gorm:"size:32" json:"role"`
	Action     string    `gorm:"size:64;not null" json:"action"`
	EntityType string    `gorm:"size:32;not null" json:"entity_type"`
	EntityID   uint      `json:"entity_id"`
	Detail     string    `gorm:"size:512" json:"detail"`
	IP         string    `gorm:"size:64" json:"ip"`
	RequestID  string    `gorm:"size:64" json:"request_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 指定表名。
func (AuditLog) TableName() string { return "audit_logs" }
