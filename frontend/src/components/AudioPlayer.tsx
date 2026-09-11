// 录音播放组件：通过后端流接口拉取音频并播放，跨页面复用。
import { useCallback, useEffect, useRef, useState } from 'react'
import { fetchBlob } from '../utils/request'
import { formatDuration } from '../utils/format'

interface AudioPlayerProps {
  recordingId: number
  durationSeconds?: number
  onEnded?: () => void
}

export default function AudioPlayer({ recordingId, durationSeconds = 0, onEnded }: AudioPlayerProps) {
  const [url, setUrl] = useState<string>('')
  const [playing, setPlaying] = useState(false)
  const [current, setCurrent] = useState(0)
  const audioRef = useRef<HTMLAudioElement>(null)

  useEffect(() => {
    let objectUrl = ''
    let cancelled = false
    fetchBlob(`/recordings/${recordingId}/audio`)
      .then((blob) => {
        if (cancelled) return
        objectUrl = URL.createObjectURL(blob)
        setUrl(objectUrl)
      })
      .catch(() => setUrl(''))
    return () => {
      cancelled = true
      if (objectUrl) URL.revokeObjectURL(objectUrl)
    }
  }, [recordingId])

  const toggle = useCallback(() => {
    const audio = audioRef.current
    if (!audio) return
    if (audio.paused) {
      audio.play()
    } else {
      audio.pause()
    }
  }, [])

  if (!url) {
    return <span className="audio-missing">暂无音频文件</span>
  }

  return (
    <div className="audio-player">
      <audio
        ref={audioRef}
        src={url}
        onPlay={() => setPlaying(true)}
        onPause={() => setPlaying(false)}
        onTimeUpdate={(e) => setCurrent(e.currentTarget.currentTime)}
        onEnded={() => {
          setPlaying(false)
          setCurrent(0)
          onEnded?.()
        }}
      />
      <button className="btn btn-small" onClick={toggle}>
        {playing ? '⏸ 暂停' : '▶ 播放'}
      </button>
      <span className="audio-time">
        {formatDuration(Math.round(current))} / {formatDuration(durationSeconds)}
      </span>
    </div>
  )
}
