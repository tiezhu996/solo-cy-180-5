// 用户认证状态管理。
import { create } from 'zustand'
import type { User } from '../api/types'
import { fetchMe, login as apiLogin } from '../api/auth'
import { clearToken, getToken, setToken } from '../utils/request'

interface AuthState {
  user: User | null
  loading: boolean
  initialized: boolean
  login: (username: string, password: string) => Promise<void>
  loadMe: () => Promise<void>
  logout: () => void
  hasRole: (...roles: string[]) => boolean
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  loading: false,
  initialized: false,

  async login(username, password) {
    set({ loading: true })
    try {
      const resp = await apiLogin({ username, password })
      setToken(resp.token)
      set({ user: resp.user, loading: false, initialized: true })
    } catch (err) {
      set({ loading: false })
      throw err
    }
  },

  async loadMe() {
    if (!getToken()) {
      set({ user: null, initialized: true })
      return
    }
    try {
      const user = await fetchMe()
      set({ user, initialized: true })
    } catch {
      clearToken()
      set({ user: null, initialized: true })
    }
  },

  logout() {
    clearToken()
    set({ user: null })
  },

  hasRole(...roles) {
    const user = get().user
    return !!user && roles.includes(user.role)
  },
}))
