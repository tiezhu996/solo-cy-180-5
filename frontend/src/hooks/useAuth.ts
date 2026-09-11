// 认证相关 hook：路由守卫与按钮显隐。
import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'

export function useAuth(requiredRoles?: string[]) {
  const { user, initialized, loadMe, hasRole } = useAuthStore()
  const navigate = useNavigate()

  useEffect(() => {
    if (!initialized) {
      loadMe()
    }
  }, [initialized, loadMe])

  useEffect(() => {
    if (!initialized) return
    if (!user) {
      navigate('/login', { replace: true })
      return
    }
    if (requiredRoles && requiredRoles.length > 0 && !hasRole(...requiredRoles)) {
      navigate('/', { replace: true })
    }
  }, [initialized, user, requiredRoles, navigate, hasRole])

  return { user, initialized, hasRole }
}
