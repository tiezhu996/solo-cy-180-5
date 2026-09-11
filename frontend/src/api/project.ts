import { del, get, post, put } from '../utils/request'
import type { Paged, Project } from './types'

export interface CreateProjectPayload {
  title: string
  interviewee_name: string
  birth_year: number
  background?: string
  status?: string
}

export interface UpdateProjectPayload {
  title?: string
  interviewee_name?: string
  birth_year?: number
  background?: string
}

export function listProjects(params: { page?: number; page_size?: number; status?: string }) {
  return get<Paged<Project>>('/projects', params)
}

export function listMyProjects(params: { page?: number; page_size?: number }) {
  return get<Paged<Project>>('/projects/mine', params)
}

export function getProject(id: number) {
  return get<Project>(`/projects/${id}`)
}

export function createProject(payload: CreateProjectPayload) {
  return post<Project>('/projects', payload)
}

export function updateProject(id: number, payload: UpdateProjectPayload) {
  return put<Project>(`/projects/${id}`, payload)
}

export function transitionProjectStatus(id: number, status: string) {
  return put<Project>(`/projects/${id}/status`, { status })
}

export function deleteProject(id: number) {
  return del<null>(`/projects/${id}`)
}

export function getProjectStats() {
  return get<{ project_total: number }>('/projects/stats')
}
