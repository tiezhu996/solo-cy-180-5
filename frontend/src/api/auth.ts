import { get, post } from '../utils/request'
import type { LoginResponse, User } from './types'

export function register(payload: {
  username: string
  password: string
  display_name: string
  email?: string
  role?: string
}) {
  return post<User>('/auth/register', payload)
}

export function login(payload: { username: string; password: string }) {
  return post<LoginResponse>('/auth/login', payload)
}

export function fetchMe() {
  return get<User>('/auth/me')
}
