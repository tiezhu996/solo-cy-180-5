// 项目创建/编辑表单组件，列表页与详情页复用。
import { useState } from 'react'

export interface ProjectFormValues {
  title: string
  interviewee_name: string
  birth_year: number
  background: string
}

interface ProjectFormProps {
  initial?: Partial<ProjectFormValues>
  onSubmit: (values: ProjectFormValues) => Promise<void> | void
  submitText?: string
}

export default function ProjectForm({ initial = {}, onSubmit, submitText = '保存' }: ProjectFormProps) {
  const [values, setValues] = useState<ProjectFormValues>({
    title: initial.title || '',
    interviewee_name: initial.interviewee_name || '',
    birth_year: initial.birth_year || new Date().getFullYear() - 60,
    background: initial.background || '',
  })
  const [saving, setSaving] = useState(false)

  const handleSubmit = async () => {
    if (!values.title.trim() || !values.interviewee_name.trim()) {
      alert('请填写项目标题与受访者姓名')
      return
    }
    setSaving(true)
    try {
      await onSubmit(values)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="project-form">
      <div className="form-row">
        <label>项目标题</label>
        <input value={values.title} onChange={(e) => setValues({ ...values, title: e.target.value })} placeholder="如：老城记忆口述史" />
      </div>
      <div className="form-row">
        <label>受访者姓名</label>
        <input value={values.interviewee_name} onChange={(e) => setValues({ ...values, interviewee_name: e.target.value })} placeholder="受访者姓名" />
      </div>
      <div className="form-row">
        <label>出生年份</label>
        <input
          type="number"
          value={values.birth_year}
          onChange={(e) => setValues({ ...values, birth_year: Number(e.target.value) })}
          min={1900}
          max={2100}
        />
      </div>
      <div className="form-row">
        <label>背景简介</label>
        <textarea value={values.background} onChange={(e) => setValues({ ...values, background: e.target.value })} rows={3} placeholder="受访者背景简介" />
      </div>
      <button className="btn btn-primary" onClick={handleSubmit} disabled={saving}>
        {saving ? '提交中…' : submitText}
      </button>
    </div>
  )
}
