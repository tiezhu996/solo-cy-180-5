import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import { useAuthStore } from './stores/authStore'
import './index.css'

// 启动时恢复登录态。
useAuthStore.getState().loadMe()

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
