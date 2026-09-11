// 应用主布局：顶部导航 + 内容区。
import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'
import { ROLE_ADMIN, ROLE_TEXT } from '../constants'

export default function Layout() {
  const { user, logout, hasRole } = useAuthStore()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="brand">
          <span className="brand-logo">🎙️</span>
          <span>口述历史采集工具</span>
        </div>
        <nav className="nav">
          <NavLink to="/" className={({ isActive }) => (isActive ? 'nav-link active' : 'nav-link')} end>
            采访项目
          </NavLink>
          <NavLink to="/interview" className={({ isActive }) => (isActive ? 'nav-link active' : 'nav-link')}>
            采访工作台
          </NavLink>
          {hasRole(ROLE_ADMIN) && (
            <NavLink to="/audit" className={({ isActive }) => (isActive ? 'nav-link active' : 'nav-link')}>
              审计日志
            </NavLink>
          )}
        </nav>
        <div className="user-box">
          <span className="user-role">{user ? ROLE_TEXT[user.role] || user.role : ''}</span>
          <span className="user-name">{user?.display_name || user?.username || ''}</span>
          <button className="btn btn-plain btn-small" onClick={handleLogout}>
            退出
          </button>
        </div>
      </header>
      <main className="main">
        <Outlet />
      </main>
    </div>
  )
}
