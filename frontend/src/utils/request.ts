// 全局请求封装：携带 JWT、统一错误处理、请求追踪。
import axios, { AxiosError } from 'axios'

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

const TOKEN_KEY = 'oralhistory_token'

export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
})

request.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    if (response.config.responseType === 'blob') {
      return response
    }
    const body = response.data as ApiResponse
    if (body && body.code !== 0) {
      return Promise.reject(new Error(body.message || '请求失败'))
    }
    return response
  },
  (error: AxiosError<ApiResponse>) => {
    const status = error.response?.status
    if (status === 401) {
      clearToken()
      if (!window.location.hash.startsWith('#/login')) {
        window.location.hash = '#/login'
      }
    }
    const message = error.response?.data?.message || error.message || '网络错误'
    return Promise.reject(new Error(message))
  },
)

export async function get<T>(url: string, params?: Record<string, unknown>): Promise<T> {
  const res = await request.get<ApiResponse<T>>(url, { params })
  return res.data.data
}

export async function post<T>(url: string, data?: unknown): Promise<T> {
  const res = await request.post<ApiResponse<T>>(url, data)
  return res.data.data
}

export async function put<T>(url: string, data?: unknown): Promise<T> {
  const res = await request.put<ApiResponse<T>>(url, data)
  return res.data.data
}

export async function del<T>(url: string): Promise<T> {
  const res = await request.delete<ApiResponse<T>>(url)
  return res.data.data
}

export async function upload<T>(url: string, formData: FormData, onProgress?: (p: number) => void): Promise<T> {
  const res = await request.post<ApiResponse<T>>(url, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: (e) => {
      if (onProgress && e.total) {
        onProgress(Math.round((e.loaded / e.total) * 100))
      }
    },
  })
  return res.data.data
}

export async function fetchBlob(url: string): Promise<Blob> {
  const res = await request.get<Blob>(url, { responseType: 'blob' })
  return res.data
}

export default request
