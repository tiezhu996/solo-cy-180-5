// 与后端 internal/constants 对应的共享枚举与错误码。

// 用户角色枚举
export const ROLE_ADMIN = 'admin'
export const ROLE_INTERVIEWER = 'interviewer'
export const ROLE_ARCHIVIST = 'archivist'

export const ROLE_OPTIONS = [
  { value: ROLE_INTERVIEWER, label: '采访员' },
  { value: ROLE_ARCHIVIST, label: '档案员' },
  { value: ROLE_ADMIN, label: '管理员' },
] as const

export const ROLE_TEXT: Record<string, string> = {
  [ROLE_ADMIN]: '管理员',
  [ROLE_INTERVIEWER]: '采访员',
  [ROLE_ARCHIVIST]: '档案员',
}

// 采访项目状态机枚举（与后端 constants/project_status.go 同步）
export const PROJECT_STATUS_DRAFT = 'draft'
export const PROJECT_STATUS_IN_PROGRESS = 'in_progress'
export const PROJECT_STATUS_COMPLETED = 'completed'
export const PROJECT_STATUS_ARCHIVED = 'archived'

export const PROJECT_STATUS_OPTIONS = [
  { value: PROJECT_STATUS_DRAFT, label: '草稿' },
  { value: PROJECT_STATUS_IN_PROGRESS, label: '进行中' },
  { value: PROJECT_STATUS_COMPLETED, label: '已完成' },
  { value: PROJECT_STATUS_ARCHIVED, label: '已归档' },
] as const

export const PROJECT_STATUS_TEXT: Record<string, string> = {
  [PROJECT_STATUS_DRAFT]: '草稿',
  [PROJECT_STATUS_IN_PROGRESS]: '进行中',
  [PROJECT_STATUS_COMPLETED]: '已完成',
  [PROJECT_STATUS_ARCHIVED]: '已归档',
}

// 录音状态机枚举（与后端 constants/recording_status.go 同步）
export const RECORDING_STATUS_RECORDING = 'recording'
export const RECORDING_STATUS_PROCESSING = 'processing'
export const RECORDING_STATUS_READY = 'ready'
export const RECORDING_STATUS_FAILED = 'failed'

export const RECORDING_STATUS_TEXT: Record<string, string> = {
  [RECORDING_STATUS_RECORDING]: '录制中',
  [RECORDING_STATUS_PROCESSING]: '处理中',
  [RECORDING_STATUS_READY]: '就绪',
  [RECORDING_STATUS_FAILED]: '失败',
}

// 错误码（与后端 constants/error_codes.go 同步）
export const ERROR_CODES = {
  OK: 0,
  BAD_REQUEST: 40000,
  UNAUTHORIZED: 40100,
  FORBIDDEN: 40300,
  NOT_FOUND: 40400,
  CONFLICT: 40900,
  INTERNAL: 50000,
  VALIDATION: 42200,
  RATE_LIMITED: 42900,
  DUPLICATE_NAME: 40901,
  LOGIN_FAILED: 40101,
  PROJECT_STATUS: 40902,
  RECORDING_STATUS: 40903,
  MARKER_CONFLICT: 40904,
} as const

export type ProjectStatus = typeof PROJECT_STATUS_DRAFT | typeof PROJECT_STATUS_IN_PROGRESS | typeof PROJECT_STATUS_COMPLETED | typeof PROJECT_STATUS_ARCHIVED
export type RecordingStatus = typeof RECORDING_STATUS_RECORDING | typeof RECORDING_STATUS_PROCESSING | typeof RECORDING_STATUS_READY | typeof RECORDING_STATUS_FAILED
export type Role = typeof ROLE_ADMIN | typeof ROLE_INTERVIEWER | typeof ROLE_ARCHIVIST
