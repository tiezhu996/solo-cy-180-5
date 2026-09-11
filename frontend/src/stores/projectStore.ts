// 采访项目状态管理。
import { create } from 'zustand'
import {
  createProject,
  deleteProject,
  getProject,
  listMyProjects,
  listProjects,
  transitionProjectStatus,
  updateProject,
} from '../api/project'
import type { CreateProjectPayload, UpdateProjectPayload } from '../api/project'
import type { Paged, Project } from '../api/types'

interface ProjectState {
  projects: Project[]
  total: number
  loading: boolean
  detail: Project | null
  fetchList: (params?: { page?: number; page_size?: number; status?: string; mine?: boolean }) => Promise<void>
  fetchDetail: (id: number) => Promise<Project>
  create: (payload: CreateProjectPayload) => Promise<Project>
  update: (id: number, payload: UpdateProjectPayload) => Promise<Project>
  transitionStatus: (id: number, status: string) => Promise<Project>
  remove: (id: number) => Promise<void>
  clearDetail: () => void
}

export const useProjectStore = create<ProjectState>((set) => ({
  projects: [],
  total: 0,
  loading: false,
  detail: null,

  async fetchList(params = {}) {
    set({ loading: true })
    try {
      const paged: Paged<Project> = params.mine
        ? await listMyProjects({ page: params.page || 1, page_size: params.page_size || 20 })
        : await listProjects({
            page: params.page || 1,
            page_size: params.page_size || 20,
            status: params.status || '',
          })
      set({ projects: paged.list, total: paged.total, loading: false })
    } catch (e) {
      console.error('fetch projects failed', e)
      set({ loading: false })
    }
  },

  async fetchDetail(id) {
    const detail = await getProject(id)
    set({ detail })
    return detail
  },

  async create(payload) {
    const project = await createProject(payload)
    set((s) => ({ projects: [project, ...s.projects], total: s.total + 1 }))
    return project
  },

  async update(id, payload) {
    const project = await updateProject(id, payload)
    set((s) => ({
      projects: s.projects.map((p) => (p.id === id ? project : p)),
      detail: s.detail?.id === id ? project : s.detail,
    }))
    return project
  },

  async transitionStatus(id, status) {
    const project = await transitionProjectStatus(id, status)
    set((s) => ({
      projects: s.projects.map((p) => (p.id === id ? project : p)),
      detail: s.detail?.id === id ? project : s.detail,
    }))
    return project
  },

  async remove(id) {
    await deleteProject(id)
    set((s) => ({ projects: s.projects.filter((p) => p.id !== id), total: Math.max(0, s.total - 1) }))
  },

  clearDetail() {
    set({ detail: null })
  },
}))
