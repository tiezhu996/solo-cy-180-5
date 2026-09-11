import { get } from '../utils/request'
import type { AuditLog, Paged } from './types'

export function listAuditLogs(params: { page?: number; page_size?: number; username?: string }) {
  return get<Paged<AuditLog>>('/audit-logs', params)
}
