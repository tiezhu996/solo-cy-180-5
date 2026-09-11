// 路由配置：登录页公开，其余页面受路由守卫保护。
import { createHashRouter, Navigate } from 'react-router-dom'
import Layout from '../components/Layout'
import AuditPage from '../pages/audit/AuditPage'
import InterviewPage from '../pages/interview/InterviewPage'
import LoginPage from '../pages/login/LoginPage'
import ProjectDetailPage from '../pages/projects/ProjectDetailPage'
import ProjectListPage from '../pages/projects/ProjectListPage'

export const router = createHashRouter([
  { path: '/login', element: <LoginPage /> },
  {
    path: '/',
    element: <Layout />,
    children: [
      { index: true, element: <ProjectListPage /> },
      { path: 'projects/:id', element: <ProjectDetailPage /> },
      { path: 'interview', element: <InterviewPage /> },
      { path: 'audit', element: <AuditPage /> },
      { path: '*', element: <Navigate to="/" replace /> },
    ],
  },
])
