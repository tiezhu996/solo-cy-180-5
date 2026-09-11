import {
  PROJECT_STATUS_TEXT,
  RECORDING_STATUS_TEXT,
  ROLE_TEXT,
} from '../constants'

export function formatDateTime(value?: string): string {
  if (!value) return '-'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

export function formatDuration(seconds: number): string {
  if (!seconds || seconds <= 0) return '00:00'
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

export function projectStatusText(status: string): string {
  return PROJECT_STATUS_TEXT[status] || status
}

export function recordingStatusText(status: string): string {
  return RECORDING_STATUS_TEXT[status] || status
}

export function roleText(role: string): string {
  return ROLE_TEXT[role] || role
}
