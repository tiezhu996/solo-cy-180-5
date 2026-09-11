// 时间轴节点状态管理。
import { create } from 'zustand'
import { createMarker, deleteMarker, listMarkers } from '../api/timelineMarker'
import type { TimelineMarker } from '../api/types'

interface TimelineState {
  markers: TimelineMarker[]
  loading: boolean
  fetchByProject: (projectId: number) => Promise<void>
  fetchByRecording: (recordingId: number) => Promise<TimelineMarker[]>
  create: (payload: {
    project_id: number
    recording_id: number
    timestamp_second: number
    label: string
    note?: string
  }) => Promise<void>
  remove: (id: number) => Promise<void>
}

export const useTimelineStore = create<TimelineState>((set) => ({
  markers: [],
  loading: false,

  async fetchByProject(projectId) {
    set({ loading: true })
    try {
      const res = await listMarkers({ project_id: projectId })
      set({ markers: res.list, loading: false })
    } catch (e) {
      console.error('fetch markers failed', e)
      set({ loading: false })
    }
  },

  async fetchByRecording(recordingId) {
    const res = await listMarkers({ recording_id: recordingId })
    return res.list
  },

  async create(payload) {
    const marker = await createMarker(payload)
    set((s) => ({ markers: [...s.markers, marker] }))
  },

  async remove(id) {
    await deleteMarker(id)
    set((s) => ({ markers: s.markers.filter((m) => m.id !== id) }))
  },
}))
