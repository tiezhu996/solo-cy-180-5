package dto

// AuditLogResponse 审计日志响应。
type AuditLogResponse struct {
	ID         uint   `json:"id"`
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	Role       string `json:"role"`
	Action     string `json:"action"`
	EntityType string `json:"entity_type"`
	EntityID   uint   `json:"entity_id"`
	Detail     string `json:"detail"`
	IP         string `json:"ip"`
	RequestID  string `json:"request_id"`
	CreatedAt  string `json:"created_at"`
}
