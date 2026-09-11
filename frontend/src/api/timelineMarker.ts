import { del, get, post, put } from '../utils/request'
import type { TimelineMarker } from './types'

export function listMarkers(params: { project_id?: number; recording_id?: number }) {
  return get<{ list: TimelineMarker[] }>('/timeline-markers', params)
}

export function createMarker(payload: {
  project_id: number
  recording_id: number
  timestamp_second: number
  label: string
  note?: string
}) {
  return post<TimelineMarker>('/timeline-markers', payload)
}

export function updateMarker(id: number, payload: { timestamp_second?: number; label?: string; note?: string }) {
  return put<TimelineMarker>(`/timeline-markers/${id}`, payload)
}

export function deleteMarker(id: number) {
  return del<null>(`/timeline-markers/${id}`)
}
