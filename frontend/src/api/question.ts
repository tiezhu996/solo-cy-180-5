import { del, get, post, put } from '../utils/request'
import type { Question } from './types'

export function listQuestions(projectId: number) {
  return get<{ list: Question[] }>(`/projects/${projectId}/questions`)
}

export function createQuestion(projectId: number, payload: { content: string; sort_order?: number }) {
  return post<Question>(`/projects/${projectId}/questions`, payload)
}

export function updateQuestion(id: number, payload: { content?: string; sort_order?: number }) {
  return put<Question>(`/questions/${id}`, payload)
}

export function deleteQuestion(id: number) {
  return del<null>(`/questions/${id}`)
}
