// 采访项目状态机枚举。
package constants

// 项目状态定义。
const (
	ProjectStatusDraft      = "draft"       // 草稿：项目刚创建
	ProjectStatusInProgress = "in_progress" // 进行中：正在采访录音
	ProjectStatusCompleted  = "completed"   // 已完成：录音与摘要整理完成
	ProjectStatusArchived   = "archived"    // 已归档：进入档案库
)

// ValidProjectStatus 校验项目状态是否合法。
func ValidProjectStatus(status string) bool {
	switch status {
	case ProjectStatusDraft, ProjectStatusInProgress, ProjectStatusCompleted, ProjectStatusArchived:
		return true
	default:
		return false
	}
}

// CanTransitionProject 返回项目状态流转是否允许。
func CanTransitionProject(from, to string) bool {
	switch from {
	case ProjectStatusDraft:
		return to == ProjectStatusInProgress || to == ProjectStatusArchived
	case ProjectStatusInProgress:
		return to == ProjectStatusCompleted || to == ProjectStatusArchived
	case ProjectStatusCompleted:
		return to == ProjectStatusArchived || to == ProjectStatusInProgress
	case ProjectStatusArchived:
		return false
	default:
		return false
	}
}
