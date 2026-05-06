import { useEffect } from 'react'
import { BrowserRouter, Navigate, Outlet, Route, Routes } from 'react-router-dom'
import ErrorBoundary from './components/ErrorBoundary'
import Layout from './components/Layout'
import { useAuthStore } from './hooks/useAuthStore'
import { usePreferencesStore } from './hooks/usePreferencesStore'
import LoginPage from './pages/LoginPage'
import NotFoundPage from './pages/NotFoundPage'
import WorkspacePage from './pages/WorkspacePage'

function ProtectedRoute() {
  const token = useAuthStore((state) => state.token)
  if (!token) {
    return <Navigate to="/login" replace />
  }

  return <Outlet />
}

function LoginRoute() {
  const token = useAuthStore((state) => state.token)
  if (token) {
    return <Navigate to="/workspace/overview" replace />
  }

  return <LoginPage />
}

function PreferenceBridge() {
  const locale = usePreferencesStore((state) => state.locale)
  const theme = usePreferencesStore((state) => state.theme)
  const compactMode = usePreferencesStore((state) => state.compactMode)

  useEffect(() => {
    document.documentElement.dataset.theme = theme
    document.documentElement.dataset.density = compactMode ? 'compact' : 'comfortable'
    document.documentElement.lang = locale
  }, [compactMode, locale, theme])

  return null
}

export default function App() {
  return (
    <ErrorBoundary title="WorkPal 加载失败" message="应用渲染时出现异常，可以重试当前页面。">
      <BrowserRouter>
        <PreferenceBridge />
        <Routes>
          <Route path="/login" element={<LoginRoute />} />
          <Route element={<ProtectedRoute />}>
            <Route element={<Layout />}>
              <Route index element={<Navigate to="/workspace/overview" replace />} />
              <Route path="/chat" element={<Navigate to="/workspace/chat" replace />} />
              <Route path="/workspace/:section" element={<WorkspacePage />} />
              <Route path="/workspace/chat/:conversationId" element={<WorkspacePage />} />
            </Route>
          </Route>
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </BrowserRouter>
    </ErrorBoundary>
  )
}
