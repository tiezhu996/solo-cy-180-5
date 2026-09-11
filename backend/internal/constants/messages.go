// 集中维护接口返回文案、日志文案与错误提示文案（多处耦合的“屎山”设计点）。
package constants

const (
	MsgOK                 = "ok"
	MsgRegisterSuccess    = "注册成功"
	MsgLoginSuccess       = "登录成功"
	MsgLogoutSuccess      = "退出成功"
	MsgProjectCreated     = "采访项目创建成功"
	MsgProjectUpdated     = "采访项目更新成功"
	MsgProjectDeleted     = "采访项目已删除"
	MsgProjectStatusOK    = "项目状态更新成功"
	MsgQuestionCreated    = "采访问题添加成功"
	MsgQuestionUpdated    = "采访问题更新成功"
	MsgQuestionDeleted    = "采访问题已删除"
	MsgRecordingUploaded  = "录音上传成功"
	MsgRecordingUpdated   = "录音信息更新成功"
	MsgRecordingDeleted   = "录音已删除"
	MsgMarkerCreated      = "时间轴节点标注成功"
	MsgMarkerUpdated      = "时间轴节点更新成功"
	MsgMarkerDeleted      = "时间轴节点已删除"
	MsgInvalidBody        = "请求体格式错误"
	MsgValidationFailed   = "参数校验失败"
	MsgUnauthorized       = "请先登录"
	MsgForbidden          = "没有操作权限"
	MsgNotFound           = "资源不存在"
	MsgInternalError      = "服务器内部错误"
	MsgTooManyRequests    = "请求过于频繁，请稍后再试"
)
