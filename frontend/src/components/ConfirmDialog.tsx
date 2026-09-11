// 通用确认弹窗组件，跨页面复用。
import { useState } from 'react'

interface ConfirmDialogProps {
  title: string
  message: string
  confirmText?: string
  danger?: boolean
  onConfirm: () => void | Promise<void>
  onCancel?: () => void
  children?: React.ReactNode
}

export default function ConfirmDialog({
  title,
  message,
  confirmText = '确认',
  danger = false,
  onConfirm,
  onCancel,
  children,
}: ConfirmDialogProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)

  const handleConfirm = async () => {
    setLoading(true)
    try {
      await onConfirm()
      setOpen(false)
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <span onClick={() => setOpen(true)} style={{ display: 'inline-flex', cursor: 'pointer' }}>
        {children}
      </span>
      {open && (
        <div className="modal-mask" onClick={() => setOpen(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-title">{title}</div>
            <div className="modal-body">{message}</div>
            <div className="modal-footer">
              <button className="btn btn-plain" onClick={() => { setOpen(false); onCancel?.() }}>
                取消
              </button>
              <button className={`btn ${danger ? 'btn-danger' : 'btn-primary'}`} onClick={handleConfirm} disabled={loading}>
                {loading ? '处理中…' : confirmText}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
