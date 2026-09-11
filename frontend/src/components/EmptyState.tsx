// 空状态组件，跨页面复用。
interface EmptyStateProps {
  title?: string
  description?: string
  action?: React.ReactNode
}

export default function EmptyState({ title = '暂无数据', description = '还没有任何记录，点击下方按钮开始创建', action }: EmptyStateProps) {
  return (
    <div className="empty-state">
      <div className="empty-icon">📄</div>
      <div className="empty-title">{title}</div>
      <div className="empty-desc">{description}</div>
      {action && <div className="empty-action">{action}</div>}
    </div>
  )
}
