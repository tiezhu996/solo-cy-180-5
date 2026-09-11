// 录音片段状态机枚举。
package constants

// 录音状态定义。
const (
	RecordingStatusRecording   = "recording"   // 录制中：浏览器端正在采集
	RecordingStatusProcessing  = "processing"  // 处理中：已上传等待转码
	RecordingStatusReady       = "ready"       // 就绪：可播放
	RecordingStatusFailed      = "failed"      // 失败：上传或处理失败
)

// ValidRecordingStatus 校验录音状态是否合法。
func ValidRecordingStatus(status string) bool {
	switch status {
	case RecordingStatusRecording, RecordingStatusProcessing, RecordingStatusReady, RecordingStatusFailed:
		return true
	default:
		return false
	}
}

// CanTransitionRecording 返回录音状态流转是否允许。
func CanTransitionRecording(from, to string) bool {
	switch from {
	case RecordingStatusRecording:
		return to == RecordingStatusProcessing || to == RecordingStatusFailed
	case RecordingStatusProcessing:
		return to == RecordingStatusReady || to == RecordingStatusFailed
	case RecordingStatusReady:
		return to == RecordingStatusFailed
	case RecordingStatusFailed:
		return false
	default:
		return false
	}
}
