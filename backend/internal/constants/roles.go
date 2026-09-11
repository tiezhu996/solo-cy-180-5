// 用户角色枚举。
package constants

// 角色定义。
const (
	RoleAdmin      = "admin"      // 管理员：管理全部项目与用户、查看审计日志
	RoleInterviewer = "interviewer" // 采访员：创建项目、提问、录音
	RoleArchivist  = "archivist"  // 档案员：整理时间轴、撰写摘要、归档项目
)

// ValidRoles 校验角色是否合法。
func ValidRoles(role string) bool {
	switch role {
	case RoleAdmin, RoleInterviewer, RoleArchivist:
		return true
	default:
		return false
	}
}
