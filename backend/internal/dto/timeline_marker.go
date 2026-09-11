package dto

// CreateTimelineMarkerRequest 标注时间轴节点请求。
type CreateTimelineMarkerRequest struct {
	ProjectID       uint   `json:"project_id" binding:"required"`
	RecordingID     uint   `json:"recording_id" binding:"required"`
	TimestampSecond int    `json:"timestamp_second" binding:"required,min=0"`
	Label           string `json:"label" binding:"required,min=1,max=128"`
	Note            string `json:"note" binding:"omitempty,max=512"`
}

// UpdateTimelineMarkerRequest 更新时间轴节点请求。
type UpdateTimelineMarkerRequest struct {
	TimestampSecond int    `json:"timestamp_second" binding:"omitempty,min=0"`
	Label           string `json:"label" binding:"omitempty,min=1,max=128"`
	Note            string `json:"note" binding:"omitempty,max=512"`
}

// TimelineMarkerResponse 时间轴节点响应。
type TimelineMarkerResponse struct {
	ID              uint   `json:"id"`
	ProjectID       uint   `json:"project_id"`
	RecordingID     uint   `json:"recording_id"`
	TimestampSecond int    `json:"timestamp_second"`
	Label           string `json:"label"`
	Note            string `json:"note"`
	CreatedBy       uint   `json:"created_by"`
	CreatedAt       string `json:"created_at"`
}
