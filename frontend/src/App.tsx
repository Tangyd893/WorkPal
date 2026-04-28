import { useEffect } from 'react'
import { BrowserRouter, Navigate, Outlet, Route, Routes } from 'react-router-dom'
import Layout from './components/Layout'
import { useAuthStore } from './hooks/useAuthStore'
import { usePreferencesStore } from './hooks/usePreferencesStore'
import LoginPage from './pages/LoginPage'
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
    <BrowserRouter>
      <PreferenceBridge />
      <Routes>
        <Route path="/login" element={<LoginRoute />} />
        <Route element={<ProtectedRoute />}>
          <Route element={<Layout />}>
            <Route index element={<Navigate to="/workspace/overview" replace />} />
            <Route path="/chat" element={<Navigate to="/workspace/chat" replace />} />
            <Route path="/workspace/:section" element={<WorkspacePage />} />
          </Route>
        </Route>
        <Route path="*" element={<Navigate to="/workspace/overview" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
