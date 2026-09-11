package util

import (
	"fmt"
	"time"

	"github.com/oralhistory/oralhistory/internal/constants"
)

// FormatDateTime 格式化时间为 "2006-01-02 15:04:05"。
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ProjectStatusText 返回项目状态的展示文本（与前端徽标枚举同步）。
func ProjectStatusText(status string) string {
	switch status {
	case constants.ProjectStatusDraft:
		return "草稿"
	case constants.ProjectStatusInProgress:
		return "进行中"
	case constants.ProjectStatusCompleted:
		return "已完成"
	case constants.ProjectStatusArchived:
		return "已归档"
	default:
		return "未知"
	}
}

// RecordingStatusText 返回录音状态的展示文本。
func RecordingStatusText(status string) string {
	switch status {
	case constants.RecordingStatusRecording:
		return "录制中"
	case constants.RecordingStatusProcessing:
		return "处理中"
	case constants.RecordingStatusReady:
		return "就绪"
	case constants.RecordingStatusFailed:
		return "失败"
	default:
		return "未知"
	}
}

// RoleText 返回角色的展示文本。
func RoleText(role string) string {
	switch role {
	case constants.RoleAdmin:
		return "管理员"
	case constants.RoleInterviewer:
		return "采访员"
	case constants.RoleArchivist:
		return "档案员"
	default:
		return "未知"
	}
}

// FormatDuration 将秒数格式化为 mm:ss。
func FormatDuration(seconds int) string {
	return fmt.Sprintf("%02d:%02d", seconds/60, seconds%60)
}
