// 登录页。
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { register } from '../../api/auth'
import { useAuthStore } from '../../stores/authStore'

export default function LoginPage() {
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [error, setError] = useState('')
  const { login } = useAuthStore()
  const navigate = useNavigate()

  const handleLogin = async () => {
    setError('')
    try {
      await login(username, password)
      navigate('/')
    } catch (e) {
      setError(e instanceof Error ? e.message : '登录失败')
    }
  }

  const handleRegister = async () => {
    setError('')
    try {
      await register({ username, password, display_name: displayName })
      await login(username, password)
      navigate('/')
    } catch (e) {
      setError(e instanceof Error ? e.message : '注册失败')
    }
  }

  return (
    <div className="login-page">
      <div className="login-card">
        <div className="login-logo">🎙️</div>
        <h1 className="login-title">口述历史采集工具</h1>
        <p className="login-sub">记录每一段不该被遗忘的记忆</p>
        <div className="login-tabs">
          <button className={mode === 'login' ? 'tab active' : 'tab'} onClick={() => setMode('login')}>
            登录
          </button>
          <button className={mode === 'register' ? 'tab active' : 'tab'} onClick={() => setMode('register')}>
            注册
          </button>
        </div>
        {mode === 'register' && (
          <div className="form-row">
            <label>昵称</label>
            <input value={displayName} onChange={(e) => setDisplayName(e.target.value)} placeholder="您的昵称" />
          </div>
        )}
        <div className="form-row">
          <label>用户名</label>
          <input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="用户名" autoComplete="username" />
        </div>
        <div className="form-row">
          <label>密码</label>
          <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="密码" autoComplete="current-password" />
        </div>
        {error && <div className="form-error">{error}</div>}
        <button className="btn btn-primary btn-block" onClick={mode === 'login' ? handleLogin : handleRegister}>
          {mode === 'login' ? '登 录' : '注册并登录'}
        </button>
        <div className="login-hint">默认管理员：admin / admin123456</div>
      </div>
    </div>
  )
}
