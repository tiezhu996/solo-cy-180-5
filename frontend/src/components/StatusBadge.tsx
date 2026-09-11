// 通用状态徽标组件，跨页面复用。
import { PROJECT_STATUS_TEXT, RECORDING_STATUS_TEXT } from '../constants'

interface StatusBadgeProps {
  status: string
  type?: 'project' | 'recording'
}

const STYLES: Record<string, string> = {
  draft: 'badge-draft',
  in_progress: 'badge-progress',
  completed: 'badge-completed',
  archived: 'badge-archived',
  recording: 'badge-recording',
  processing: 'badge-processing',
  ready: 'badge-ready',
  failed: 'badge-failed',
}

export default function StatusBadge({ status, type = 'project' }: StatusBadgeProps) {
  const text = type === 'project' ? PROJECT_STATUS_TEXT[status] : RECORDING_STATUS_TEXT[status]
  return <span className={`status-badge ${STYLES[status] || 'badge-default'}`}>{text || status}</span>
}
