package dto

// CreateProjectRequest 创建采访项目请求。
type CreateProjectRequest struct {
	Title           string `json:"title" binding:"required,min=1,max=128"`
	IntervieweeName string `json:"interviewee_name" binding:"required,min=1,max=64"`
	BirthYear       int    `json:"birth_year" binding:"required,min=1900,max=2100"`
	Background      string `json:"background" binding:"omitempty,max=2000"`
	Status          string `json:"status" binding:"omitempty,oneof=draft in_progress completed archived"`
}

// UpdateProjectRequest 更新采访项目请求。
type UpdateProjectRequest struct {
	Title           string `json:"title" binding:"omitempty,min=1,max=128"`
	IntervieweeName string `json:"interviewee_name" binding:"omitempty,min=1,max=64"`
	BirthYear       int    `json:"birth_year" binding:"omitempty,min=1900,max=2100"`
	Background      string `json:"background" binding:"omitempty,max=2000"`
}

// UpdateProjectStatusRequest 项目状态流转请求。
type UpdateProjectStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=draft in_progress completed archived"`
}

// ProjectResponse 项目响应。
type ProjectResponse struct {
	ID              uint   `json:"id"`
	Title           string `json:"title"`
	IntervieweeName string `json:"interviewee_name"`
	BirthYear       int    `json:"birth_year"`
	Background      string `json:"background"`
	Status          string `json:"status"`
	CreatedBy       uint   `json:"created_by"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}
