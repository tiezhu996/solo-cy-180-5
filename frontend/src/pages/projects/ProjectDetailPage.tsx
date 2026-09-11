// 项目详情页：基本信息、采访问题、时间线（录音片段 + 关键节点 + 一句话摘要）。
import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import AudioPlayer from '../../components/AudioPlayer'
import ConfirmDialog from '../../components/ConfirmDialog'
import EmptyState from '../../components/EmptyState'
import StatusBadge from '../../components/StatusBadge'
import {
  PROJECT_STATUS_ARCHIVED,
  PROJECT_STATUS_COMPLETED,
  PROJECT_STATUS_IN_PROGRESS,
} from '../../constants'
import { useProjectStore } from '../../stores/projectStore'
import { useQuestionStore } from '../../stores/questionStore'
import { useRecordingStore } from '../../stores/recordingStore'
import { useTimelineStore } from '../../stores/timelineStore'
import { formatDateTime, formatDuration } from '../../utils/format'
import type { Recording, TimelineMarker } from '../../api/types'

export default function ProjectDetailPage() {
  const { id } = useParams()
  const projectId = Number(id)
  const navigate = useNavigate()
  const { detail, fetchDetail, transitionStatus, remove } = useProjectStore()
  const { questions, fetchByProject, create: createQuestion, remove: removeQuestion } = useQuestionStore()
  const { recordings, fetchByProject: fetchRecordings } = useRecordingStore()
  const { markers, fetchByProject: fetchMarkers, create: createMarker } = useTimelineStore()
  const [newQuestion, setNewQuestion] = useState('')
  const [message, setMessage] = useState('')

  useEffect(() => {
    if (projectId) {
      fetchDetail(projectId)
      fetchByProject(projectId)
      fetchRecordings(projectId)
      fetchMarkers(projectId)
    }
  }, [projectId, fetchDetail, fetchByProject, fetchRecordings, fetchMarkers])

  const handleAddQuestion = useCallback(async () => {
    if (!newQuestion.trim()) return
    await createQuestion(projectId, newQuestion.trim())
    setNewQuestion('')
    setMessage('采访问题已添加')
    setTimeout(() => setMessage(''), 3000)
  }, [createQuestion, newQuestion, projectId])

  const markersOf = (recordingId: number) => markers.filter((m) => m.recording_id === recordingId)

  if (!detail) {
    return <div className="page">加载中…</div>
  }

  return (
    <div className="page">
      {message && <div className="toast success">{message}</div>}
      <div className="page-header">
        <button className="btn btn-plain" onClick={() => navigate('/')}>
          ← 返回列表
        </button>
        <h2>{detail.title}</h2>
        <StatusBadge status={detail.status} type="project" />
      </div>

      <section className="card">
        <div className="card-title">项目信息</div>
        <div className="detail-grid">
          <div>
            <div className="detail-label">受访者</div>
            <div className="detail-value">{detail.interviewee_name}</div>
          </div>
          <div>
            <div className="detail-label">出生年份</div>
            <div className="detail-value">{detail.birth_year}</div>
          </div>
          <div>
            <div className="detail-label">创建时间</div>
            <div className="detail-value">{formatDateTime(detail.created_at)}</div>
          </div>
          <div>
            <div className="detail-label">背景简介</div>
            <div className="detail-value">{detail.background || '-'}</div>
          </div>
        </div>
        <div className="row-actions" style={{ marginTop: 12 }}>
          {detail.status !== PROJECT_STATUS_ARCHIVED && (
            <button
              className="btn btn-primary btn-small"
              onClick={async () => {
                const next =
                  detail.status === PROJECT_STATUS_IN_PROGRESS ? PROJECT_STATUS_COMPLETED : PROJECT_STATUS_IN_PROGRESS
                await transitionStatus(projectId, next)
                setMessage('项目状态已更新')
                setTimeout(() => setMessage(''), 3000)
              }}
            >
              {detail.status === PROJECT_STATUS_IN_PROGRESS ? '标记为已完成' : '开始采访'}
            </button>
          )}
          <ConfirmDialog
            title="删除采访项目"
            message="确定删除该项目吗？此操作不可恢复。"
            danger
            confirmText="删除"
            onConfirm={async () => {
              await remove(projectId)
              navigate('/')
            }}
          >
            <button className="btn btn-danger btn-small">删除项目</button>
          </ConfirmDialog>
          <LinkToInterview projectId={projectId} />
        </div>
      </section>

      <section className="card">
        <div className="card-title">采访问题</div>
        {questions.length === 0 ? (
          <EmptyState title="还没有采访问题" description="添加采访问题，作为录音的提纲" />
        ) : (
          <ul className="question-list">
            {questions.map((q) => (
              <li key={q.id} className="question-item">
                <span className="question-index">{q.sort_order + 1}</span>
                <span className="question-content">{q.content}</span>
                <ConfirmDialog
                  title="删除采访问题"
                  message="删除问题将同时删除其下的录音片段，确定继续？"
                  danger
                  confirmText="删除"
                  onConfirm={() => removeQuestion(q.id)}
                >
                  <button className="btn btn-plain btn-small">删除</button>
                </ConfirmDialog>
              </li>
            ))}
          </ul>
        )}
        <div className="inline-form">
          <input value={newQuestion} onChange={(e) => setNewQuestion(e.target.value)} placeholder="输入新的采访问题" />
          <button className="btn btn-primary" onClick={handleAddQuestion} disabled={!newQuestion.trim()}>
            添加问题
          </button>
        </div>
      </section>

      <section className="card">
        <div className="card-title">时间线 · 采访片段</div>
        {recordings.length === 0 ? (
          <EmptyState title="还没有录音片段" description="前往采访工作台开始录音，片段将按时间线展示" />
        ) : (
          <div className="timeline">
            {recordings.map((r) => (
              <TimelineItem key={r.id} recording={r} markers={markersOf(r.id)} onCreateMarker={createMarker} />
            ))}
          </div>
        )}
      </section>
    </div>
  )
}

function LinkToInterview({ projectId }: { projectId: number }) {
  return (
    <a className="btn btn-plain btn-small" href={`#/interview?project_id=${projectId}`}>
      前往采访工作台
    </a>
  )
}

function TimelineItem({
  recording,
  markers,
  onCreateMarker,
}: {
  recording: Recording
  markers: TimelineMarker[]
  onCreateMarker: (payload: {
    project_id: number
    recording_id: number
    timestamp_second: number
    label: string
    note?: string
  }) => Promise<void>
}) {
  const [label, setLabel] = useState('')
  const question = useQuestionStore((s) => s.questions.find((q) => q.id === recording.question_id))

  return (
    <div className="timeline-item">
      <div className="timeline-dot" />
      <div className="timeline-content">
        <div className="timeline-head">
          <span className="timeline-q">{question ? `问题：${question.content}` : `问题 #${recording.question_id}`}</span>
          <StatusBadge status={recording.status} type="recording" />
          <span className="timeline-duration">{formatDuration(recording.duration_seconds)}</span>
        </div>
        <AudioPlayer recordingId={recording.id} durationSeconds={recording.duration_seconds} />
        <div className="timeline-summary">
          <span className="summary-label">一句话摘要：</span>
          {recording.summary || <span className="muted">暂无摘要</span>}
        </div>
        {markers.length > 0 && (
          <div className="marker-list">
            {markers.map((m) => (
              <span key={m.id} className="marker-chip">
                ⏱ {formatDuration(m.timestamp_second)} · {m.label}
                {m.note ? `（${m.note}）` : ''}
              </span>
            ))}
          </div>
        )}
        <div className="inline-form">
          <input
            value={label}
            onChange={(e) => setLabel(e.target.value)}
            placeholder="标注关键节点，如：回忆童年故居"
          />
          <button
            className="btn btn-plain btn-small"
            disabled={!label.trim()}
            onClick={async () => {
              await onCreateMarker({
                project_id: recording.project_id,
                recording_id: recording.id,
                timestamp_second: recording.duration_seconds > 0 ? Math.floor(recording.duration_seconds / 2) : 0,
                label: label.trim(),
              })
              setLabel('')
            }}
          >
            ＋ 标注节点
          </button>
        </div>
      </div>
    </div>
  )
}
