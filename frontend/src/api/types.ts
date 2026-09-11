// 与后端 model/dto 对应的类型定义。
import type { ProjectStatus, RecordingStatus, Role } from '../constants'

export interface User {
  id: number
  username: string
  display_name: string
  email: string
  role: Role
  created_at: string
}

export interface LoginResponse {
  token: string
  user: User
}

export interface Project {
  id: number
  title: string
  interviewee_name: string
  birth_year: number
  background: string
  status: ProjectStatus
  created_by: number
  created_at: string
  updated_at: string
  questions?: Question[]
  recordings?: Recording[]
  timeline_markers?: TimelineMarker[]
}

export interface Question {
  id: number
  project_id: number
  content: string
  sort_order: number
  created_at: string
}

export interface Recording {
  id: number
  project_id: number
  question_id: number
  audio_key: string
  duration_seconds: number
  summary: string
  status: RecordingStatus
  created_by: number
  created_at: string
  updated_at: string
}

export interface TimelineMarker {
  id: number
  project_id: number
  recording_id: number
  timestamp_second: number
  label: string
  note: string
  created_by: number
  created_at: string
}

export interface AuditLog {
  id: number
  user_id: number
  username: string
  role: Role
  action: string
  entity_type: string
  entity_id: number
  detail: string
  ip: string
  request_id: string
  created_at: string
}

export interface Paged<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}
