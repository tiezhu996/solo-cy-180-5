// 采访工作台：选择项目 → 按问题录音 → 自动关联 → 一句话摘要 → 时间轴标注。
import { useCallback, useEffect, useRef, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import AudioPlayer from '../../components/AudioPlayer'
import EmptyState from '../../components/EmptyState'
import StatusBadge from '../../components/StatusBadge'
import { useProjectStore } from '../../stores/projectStore'
import { useQuestionStore } from '../../stores/questionStore'
import { useRecordingStore } from '../../stores/recordingStore'
import { useTimelineStore } from '../../stores/timelineStore'
import { formatDuration } from '../../utils/format'

export default function InterviewPage() {
  const [params, setParams] = useSearchParams()
  const selectedProject = Number(params.get('project_id')) || 0
  const { projects, fetchList } = useProjectStore()
  const { questions, fetchByProject } = useQuestionStore()
  const { fetchByProject: fetchRecordings } = useRecordingStore()
  const [activeQuestion, setActiveQuestion] = useState(0)
  const [message, setMessage] = useState('')

  useEffect(() => {
    fetchList({ page: 1, page_size: 100 })
  }, [fetchList])

  useEffect(() => {
    if (selectedProject) {
      fetchByProject(selectedProject)
      fetchRecordings(selectedProject)
      setActiveQuestion(0)
    } else {
      setActiveQuestion(0)
    }
  }, [selectedProject, fetchByProject, fetchRecordings])

  const chooseProject = (projectId: number) => {
    const next = new URLSearchParams(params)
    if (projectId) {
      next.set('project_id', String(projectId))
    } else {
      next.delete('project_id')
    }
    setParams(next)
  }

  return (
    <div className="page">
      <div className="page-header">
        <h2>采访工作台</h2>
      </div>
      {message && <div className="toast success">{message}</div>}

      <section className="card">
        <div className="card-title">选择采访项目</div>
        <select value={selectedProject} onChange={(e) => chooseProject(Number(e.target.value))}>
          <option value={0}>请选择项目</option>
          {projects.map((p) => (
            <option key={p.id} value={p.id}>
              {p.title}（{p.interviewee_name}）
            </option>
          ))}
        </select>
      </section>

      {selectedProject === 0 ? (
        <EmptyState title="请先选择采访项目" description="选择一个项目后即可开始录音" />
      ) : (
        <>
          <section className="card">
            <div className="card-title">采访问题列表</div>
            {questions.length === 0 ? (
              <EmptyState title="该项目还没有采访问题" description="请先到项目详情页添加采访问题" />
            ) : (
              <div className="question-tabs">
                {questions.map((q, idx) => (
                  <button
                    key={q.id}
                    className={`question-tab ${activeQuestion === q.id ? 'active' : ''}`}
                    onClick={() => setActiveQuestion(q.id)}
                  >
                    {idx + 1}. {q.content}
                  </button>
                ))}
              </div>
            )}
          </section>

          {activeQuestion > 0 && (
            <RecorderPanel
              projectId={selectedProject}
              questionId={activeQuestion}
              onRecorded={(summary) => {
                fetchRecordings(selectedProject)
                setMessage(summary)
                setTimeout(() => setMessage(''), 4000)
              }}
            />
          )}
        </>
      )}
    </div>
  )
}

function RecorderPanel({
  projectId,
  questionId,
  onRecorded,
}: {
  projectId: number
  questionId: number
  onRecorded: (msg: string) => void
}) {
  const { create, uploadAudio, updateSummary, fetchByQuestion } = useRecordingStore()
  const { markers, fetchByRecording, create: createMarker } = useTimelineStore()
  const [recording, setRecording] = useState(false)
  const [seconds, setSeconds] = useState(0)
  const [uploading, setUploading] = useState(0)
  const [recordings, setRecordings] = useState<Awaited<ReturnType<typeof fetchByQuestion>>>([])
  const [summaryDraft, setSummaryDraft] = useState('')
  const mediaRecorderRef = useRef<MediaRecorder | null>(null)
  const chunksRef = useRef<Blob[]>([])
  const timerRef = useRef<number | null>(null)

  const reload = useCallback(async () => {
    setRecordings(await fetchByQuestion(questionId))
  }, [fetchByQuestion, questionId])

  useEffect(() => {
    reload()
  }, [reload])

  useEffect(() => {
    if (recording) {
      timerRef.current = window.setInterval(() => setSeconds((s) => s + 1), 1000)
    } else if (timerRef.current) {
      window.clearInterval(timerRef.current)
      timerRef.current = null
    }
    return () => {
      if (timerRef.current) window.clearInterval(timerRef.current)
    }
  }, [recording])

  const startRecording = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      const recorder = new MediaRecorder(stream)
      chunksRef.current = []
      recorder.ondataavailable = (e) => {
        if (e.data.size > 0) chunksRef.current.push(e.data)
      }
      recorder.onstop = async () => {
        stream.getTracks().forEach((t) => t.stop())
        const blob = new Blob(chunksRef.current, { type: 'audio/webm' })
        setRecording(false)
        setUploading(1)
        try {
          const created = await create({ project_id: projectId, question_id: questionId, duration_seconds: seconds })
          await uploadAudio(created.id, blob, seconds, (p) => setUploading(p))
          await reload()
          onRecorded('录音上传成功，已自动关联到当前问题')
        } catch (e) {
          onRecorded(e instanceof Error ? e.message : '录音上传失败')
        } finally {
          setUploading(0)
          setSeconds(0)
        }
      }
      mediaRecorderRef.current = recorder
      recorder.start()
      setRecording(true)
      setSeconds(0)
    } catch {
      alert('无法访问麦克风，请在浏览器中授权麦克风权限')
    }
  }

  const stopRecording = () => {
    mediaRecorderRef.current?.stop()
  }

  const addMarker = async (recordingId: number, label: string) => {
    await createMarker({
      project_id: projectId,
      recording_id: recordingId,
      timestamp_second: 0,
      label,
    })
    await fetchByRecording(recordingId)
  }

  return (
    <section className="card">
      <div className="card-title">录音面板</div>
      <div className="recorder-box">
        {uploading > 0 ? (
          <div className="upload-progress">
            上传中… {uploading}%
            <div className="progress-bar">
              <div className="progress-inner" style={{ width: `${uploading}%` }} />
            </div>
          </div>
        ) : recording ? (
          <>
            <div className="recording-indicator">
              <span className="rec-dot" /> 正在录音 {formatDuration(seconds)}
            </div>
            <button className="btn btn-danger" onClick={stopRecording}>
              ⏹ 停止并保存
            </button>
          </>
        ) : (
          <button className="btn btn-primary" onClick={startRecording}>
            ⏺ 开始录音
          </button>
        )}
      </div>

      <div className="card-title" style={{ marginTop: 20 }}>
        本问题已录片段（{recordings.length}）
      </div>
      {recordings.length === 0 ? (
        <EmptyState title="还没有录音" description="点击上方开始录音" />
      ) : (
        <div className="recording-list">
          {recordings.map((r) => (
            <div key={r.id} className="recording-row">
              <div className="recording-meta">
                <StatusBadge status={r.status} type="recording" />
                <span className="muted">{formatDuration(r.duration_seconds)}</span>
              </div>
              <AudioPlayer recordingId={r.id} durationSeconds={r.duration_seconds} />
              <div className="summary-edit">
                <input
                  value={summaryDraft || r.summary}
                  placeholder="写一句话摘要"
                  onChange={(e) => setSummaryDraft(e.target.value)}
                />
                <button
                  className="btn btn-plain btn-small"
                  disabled={!summaryDraft.trim()}
                  onClick={async () => {
                    await updateSummary(r.id, summaryDraft.trim())
                    setSummaryDraft('')
                    reload()
                  }}
                >
                  保存摘要
                </button>
              </div>
              <div className="marker-actions">
                <span className="muted">时间轴节点：</span>
                {markers
                  .filter((m) => m.recording_id === r.id)
                  .map((m) => (
                    <span key={m.id} className="marker-chip">
                      {m.label}
                    </span>
                  ))}
                <input
                  placeholder="新增节点，如：讲到参军经历"
                  style={{ maxWidth: 220 }}
                  id={`marker-input-${r.id}`}
                />
                <button
                  className="btn btn-plain btn-small"
                  onClick={() => {
                    const input = document.getElementById(`marker-input-${r.id}`) as HTMLInputElement
                    if (input?.value.trim()) {
                      addMarker(r.id, input.value.trim())
                      input.value = ''
                    }
                  }}
                >
                  ＋ 标注
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
