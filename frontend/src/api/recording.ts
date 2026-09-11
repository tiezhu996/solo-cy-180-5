import { del, get, post, put, upload } from '../utils/request'
import type { Recording } from './types'

export function listRecordings(params: { project_id?: number; question_id?: number }) {
  return get<{ list: Recording[] }>('/recordings', params)
}

export function createRecording(payload: {
  project_id: number
  question_id: number
  duration_seconds?: number
  summary?: string
}) {
  return post<Recording>('/recordings', payload)
}

export function getRecording(id: number) {
  return get<Recording>(`/recordings/${id}`)
}

export function updateRecording(id: number, payload: { duration_seconds?: number; summary?: string; status?: string }) {
  return put<Recording>(`/recordings/${id}`, payload)
}

export function updateRecordingSummary(id: number, summary: string) {
  return put<Recording>(`/recordings/${id}/summary`, { summary })
}

export function uploadRecordingAudio(id: number, file: Blob, durationSeconds: number, onProgress?: (p: number) => void) {
  const form = new FormData()
  form.append('file', file, `recording_${id}.webm`)
  form.append('duration_seconds', String(durationSeconds))
  return upload<Recording>(`/recordings/${id}/audio`, form, onProgress)
}

export function deleteRecording(id: number) {
  return del<null>(`/recordings/${id}`)
}
