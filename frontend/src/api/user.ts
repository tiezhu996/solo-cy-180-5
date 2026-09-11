import { del, get, put } from '../utils/request'
import type { Paged, User } from './types'

export function listUsers(params: { page?: number; page_size?: number }) {
  return get<Paged<User>>('/users', params)
}

export function updateUserRole(id: number, role: string) {
  return put<null>(`/users/${id}/role`, { role })
}

export function deleteUser(id: number) {
  return del<null>(`/users/${id}`)
}
