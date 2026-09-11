// 采访项目列表页：创建、筛选、状态流转、删除。
import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import ConfirmDialog from '../../components/ConfirmDialog'
import DataTable from '../../components/DataTable'
import EmptyState from '../../components/EmptyState'
import ProjectForm, { type ProjectFormValues } from '../../components/ProjectForm'
import StatusBadge from '../../components/StatusBadge'
import {
  PROJECT_STATUS_ARCHIVED,
  PROJECT_STATUS_COMPLETED,
  PROJECT_STATUS_DRAFT,
  PROJECT_STATUS_IN_PROGRESS,
  PROJECT_STATUS_OPTIONS,
} from '../../constants'
import { useProjectStore } from '../../stores/projectStore'
import { formatDateTime } from '../../utils/format'
import type { Project } from '../../api/types'

export default function ProjectListPage() {
  const { projects, total, loading, fetchList, create, transitionStatus, remove } = useProjectStore()
  const [statusFilter, setStatusFilter] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [message, setMessage] = useState('')

  useEffect(() => {
    fetchList({ page: 1, page_size: 50, status: statusFilter })
  }, [fetchList, statusFilter])

  const handleCreate = useCallback(
    async (values: ProjectFormValues) => {
      await create(values)
      setShowCreate(false)
      setMessage('采访项目创建成功')
      setTimeout(() => setMessage(''), 3000)
    },
    [create],
  )

  const nextStatus = (status: string): string => {
    switch (status) {
      case PROJECT_STATUS_DRAFT:
        return PROJECT_STATUS_IN_PROGRESS
      case PROJECT_STATUS_IN_PROGRESS:
        return PROJECT_STATUS_COMPLETED
      case PROJECT_STATUS_COMPLETED:
        return PROJECT_STATUS_ARCHIVED
      default:
        return ''
    }
  }

  return (
    <div className="page">
      <div className="page-header">
        <h2>采访项目</h2>
        <button className="btn btn-primary" onClick={() => setShowCreate(true)}>
          ＋ 新建采访项目
        </button>
      </div>
      {message && <div className="toast success">{message}</div>}

      <div className="filter-bar">
        <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
          <option value="">全部状态</option>
          {PROJECT_STATUS_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
        <span className="filter-count">共 {total} 个项目</span>
      </div>

      <DataTable<Project>
        loading={loading}
        rows={projects}
        rowKey={(p) => p.id}
        columns={[
          {
            key: 'title',
            title: '项目标题',
            render: (p) => (
              <Link className="link" to={`/projects/${p.id}`}>
                {p.title}
              </Link>
            ),
          },
          { key: 'interviewee', title: '受访者', render: (p) => `${p.interviewee_name}（${p.birth_year}年生）` },
          { key: 'status', title: '状态', render: (p) => <StatusBadge status={p.status} type="project" /> },
          { key: 'created_at', title: '创建时间', render: (p) => formatDateTime(p.created_at) },
          {
            key: 'actions',
            title: '操作',
            render: (p) => (
              <div className="row-actions">
                <Link className="btn btn-plain btn-small" to={`/projects/${p.id}`}>
                  详情
                </Link>
                {p.status !== PROJECT_STATUS_ARCHIVED && (
                  <button
                    className="btn btn-plain btn-small"
                    onClick={async () => {
                      await transitionStatus(p.id, nextStatus(p.status))
                      setMessage('项目状态已更新')
                      setTimeout(() => setMessage(''), 3000)
                    }}
                  >
                    流转至{PROJECT_STATUS_OPTIONS.find((o) => o.value === nextStatus(p.status))?.label}
                  </button>
                )}
                <ConfirmDialog
                  title="删除采访项目"
                  message={`确定删除项目「${p.title}」吗？其问题、录音与时间轴节点将一并删除。`}
                  confirmText="删除"
                  danger
                  onConfirm={async () => {
                    await remove(p.id)
                    setMessage('项目已删除')
                    setTimeout(() => setMessage(''), 3000)
                  }}
                >
                  <button className="btn btn-danger btn-small">删除</button>
                </ConfirmDialog>
              </div>
            ),
          },
        ]}
        emptyText="暂无采访项目"
      />
      {projects.length === 0 && !loading && (
        <EmptyState
          title="还没有采访项目"
          description="创建一个口述历史采访项目，开始记录受访者的故事"
          action={
            <button className="btn btn-primary" onClick={() => setShowCreate(true)}>
              新建采访项目
            </button>
          }
        />
      )}

      {showCreate && (
        <div className="modal-mask" onClick={() => setShowCreate(false)}>
          <div className="modal modal-lg" onClick={(e) => e.stopPropagation()}>
            <div className="modal-title">新建采访项目</div>
            <ProjectForm onSubmit={handleCreate} submitText="创建项目" />
          </div>
        </div>
      )}
    </div>
  )
}
