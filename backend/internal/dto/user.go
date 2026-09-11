package dto

// RegisterRequest 注册请求。
type RegisterRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=64"`
	Password    string `json:"password" binding:"required,min=6,max=64"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=64"`
	Email       string `json:"email" binding:"omitempty,email,max=128"`
	Role        string `json:"role" binding:"omitempty,oneof=admin interviewer archivist"`
}

// LoginRequest 登录请求。
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserResponse 用户信息响应。
type UserResponse struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	CreatedAt   string `json:"created_at"`
}

// LoginResponse 登录响应。
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
