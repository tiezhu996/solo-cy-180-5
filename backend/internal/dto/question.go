package dto

// CreateQuestionRequest 添加采访问题请求。
type CreateQuestionRequest struct {
	Content   string `json:"content" binding:"required,min=1,max=512"`
	SortOrder int    `json:"sort_order" binding:"omitempty,min=0"`
}

// UpdateQuestionRequest 更新采访问题请求。
type UpdateQuestionRequest struct {
	Content   string `json:"content" binding:"omitempty,min=1,max=512"`
	SortOrder int    `json:"sort_order" binding:"omitempty,min=0"`
}

// QuestionResponse 采访问题响应。
type QuestionResponse struct {
	ID        uint   `json:"id"`
	ProjectID uint   `json:"project_id"`
	Content   string `json:"content"`
	SortOrder int    `json:"sort_order"`
	CreatedAt string `json:"created_at"`
}
