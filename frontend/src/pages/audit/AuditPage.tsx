// 审计日志页（仅管理员）。
import { useCallback, useEffect, useState } from 'react'
import DataTable from '../../components/DataTable'
import { useAuth } from '../../hooks/useAuth'
import { usePagination } from '../../hooks/usePagination'
import { listAuditLogs } from '../../api/audit'
import { ROLE_ADMIN, ROLE_TEXT } from '../../constants'
import { formatDateTime } from '../../utils/format'
import type { AuditLog } from '../../api/types'

export default function AuditPage() {
  useAuth([ROLE_ADMIN])
  const { page, pageSize, total, setTotal, setPage } = usePagination(1, 20)
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [loading, setLoading] = useState(false)
  const [username, setUsername] = useState('')
  const [error, setError] = useState('')

  const fetchLogs = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const res = await listAuditLogs({ page, page_size: pageSize, username: username || undefined })
      setLogs(res.list)
      setTotal(res.total)
    } catch (e) {
      setError(e instanceof Error ? e.message : '审计日志加载失败')
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, username, setTotal])

  useEffect(() => {
    fetchLogs()
  }, [fetchLogs])

  return (
    <div className="page">
      <div className="page-header">
        <h2>操作审计日志</h2>
      </div>
      <div className="filter-bar">
        <input placeholder="按用户名筛选" value={username} onChange={(e) => setUsername(e.target.value)} />
        <button className="btn btn-plain btn-small" onClick={() => setPage(1)}>
          查询
        </button>
      </div>
      {error && <div className="toast error">{error}</div>}
      <DataTable<AuditLog>
        loading={loading}
        rows={logs}
        rowKey={(l) => l.id}
        columns={[
          { key: 'id', title: 'ID', render: (l) => l.id },
          { key: 'username', title: '操作人', render: (l) => `${l.username}（${ROLE_TEXT[l.role] || l.role}）` },
          { key: 'action', title: '动作', render: (l) => <code>{l.action}</code> },
          { key: 'entity', title: '对象', render: (l) => `${l.entity_type}#${l.entity_id}` },
          { key: 'detail', title: '详情', render: (l) => l.detail || '-' },
          { key: 'ip', title: 'IP', render: (l) => l.ip || '-' },
          { key: 'request_id', title: 'RequestID', render: (l) => <code className="muted">{l.request_id.slice(0, 8)}</code> },
          { key: 'created_at', title: '时间', render: (l) => formatDateTime(l.created_at) },
        ]}
        emptyText="暂无审计日志"
      />
      <div className="pagination">
        <button className="btn btn-plain btn-small" disabled={page <= 1} onClick={() => setPage(page - 1)}>
          上一页
        </button>
        <span>
          第 {page} / {Math.max(1, Math.ceil(total / pageSize))} 页，共 {total} 条
        </span>
        <button className="btn btn-plain btn-small" disabled={page >= Math.ceil(total / pageSize)} onClick={() => setPage(page + 1)}>
          下一页
        </button>
      </div>
    </div>
  )
}
