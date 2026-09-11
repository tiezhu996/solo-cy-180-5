// 集中维护日志格式模板，业务字段变更必须同步修改对应模板与调用处。
package constants

// 日志模板（>= 25 条）。
const (
	LogStartServer         = "server starting on port %s"
	LogDBConnected         = "database connected host=%s db=%s"
	LogDBMigrated          = "database migrated tables=%d"
	LogRedisConnected      = "redis connected addr=%s"
	LogMinIOConnected      = "minio connected endpoint=%s bucket=%s"
	LogUserRegister        = "user register username=%s role=%s"
	LogUserLogin           = "user login username=%s success=%t"
	LogUserCreate          = "user create by=%s target=%s role=%s"
	LogUserUpdate          = "user update by=%s target=%s"
	LogProjectCreate       = "project create by=%s title=%s interviewee=%s status=%s"
	LogProjectUpdate       = "project update by=%s project=%d title=%s status=%s"
	LogProjectStatus       = "project status transition by=%s project=%d from=%s to=%s"
	LogProjectDelete       = "project delete by=%s project=%d"
	LogQuestionCreate      = "question create by=%s project=%d content=%s"
	LogQuestionUpdate      = "question update by=%s question=%d content=%s"
	LogQuestionDelete      = "question delete by=%s question=%d"
	LogRecordingUpload     = "recording upload by=%s project=%d question=%d duration=%d status=%s"
	LogRecordingSummary    = "recording summary by=%s recording=%d summary=%s"
	LogRecordingStatus     = "recording status transition by=%s recording=%d from=%s to=%s"
	LogRecordingDelete     = "recording delete by=%s recording=%d"
	LogMarkerCreate        = "timeline marker create by=%s project=%d recording=%d ts=%d label=%s"
	LogMarkerUpdate        = "timeline marker update by=%s marker=%d label=%s"
	LogMarkerDelete        = "timeline marker delete by=%s marker=%d"
	LogAuditQuery          = "audit log query by=%s page=%d pageSize=%d"
	LogAuthDenied          = "auth denied user=%s path=%s role=%s required=%s"
	LogRequest             = "request request_id=%s method=%s path=%s status=%d latency_ms=%d"
	LogPanicRecovered      = "panic recovered request_id=%s error=%v"
	LogRateLimited         = "rate limited client=%s key=%s limit=%d"
)
