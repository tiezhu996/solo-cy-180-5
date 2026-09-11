// 采访问题状态管理。
import { create } from 'zustand'
import { createQuestion, deleteQuestion, listQuestions, updateQuestion } from '../api/question'
import type { Question } from '../api/types'

interface QuestionState {
  questions: Question[]
  loading: boolean
  fetchByProject: (projectId: number) => Promise<void>
  create: (projectId: number, content: string) => Promise<Question>
  update: (id: number, content: string) => Promise<void>
  remove: (id: number) => Promise<void>
}

export const useQuestionStore = create<QuestionState>((set, get) => ({
  questions: [],
  loading: false,

  async fetchByProject(projectId) {
    set({ loading: true })
    try {
      const res = await listQuestions(projectId)
      set({ questions: res.list, loading: false })
    } catch (e) {
      console.error('fetch questions failed', e)
      set({ loading: false })
    }
  },

  async create(projectId, content) {
    const question = await createQuestion(projectId, { content, sort_order: get().questions.length })
    set((s) => ({ questions: [...s.questions, question] }))
    return question
  },

  async update(id, content) {
    await updateQuestion(id, { content })
    set((s) => ({ questions: s.questions.map((q) => (q.id === id ? { ...q, content } : q)) }))
  },

  async remove(id) {
    await deleteQuestion(id)
    set((s) => ({ questions: s.questions.filter((q) => q.id !== id) }))
  },
}))
