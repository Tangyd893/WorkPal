import { Suspense, useCallback, useEffect, useMemo, useState } from 'react'
import { Navigate, useNavigate, useParams } from 'react-router-dom'
import ConfirmDialog from '../components/ConfirmDialog'
import ErrorBoundary from '../components/ErrorBoundary'
import ModuleSwitcher from '../components/ModuleSwitcher'
import Sidebar from '../components/Sidebar'
import ToastViewport from '../components/Toast'
import Topbar from '../components/Topbar'
import SettingsDrawer from '../components/workspace/SettingsDrawer'
import WorkspaceContent from '../components/workspace/WorkspaceContent'
import { useAuthStore } from '../hooks/useAuthStore'
import { usePreferencesStore } from '../hooks/usePreferencesStore'
import { useToastStore, type ToastType } from '../hooks/useToastStore'
import { useWorkspaceActions } from '../hooks/useWorkspaceActions'
import { useWorkspaceData } from '../hooks/useWorkspaceData'
import { useI18n } from '../i18n'
import type { WorkspaceSection } from '../types/workspace'
import type { ConfirmRequest } from '../types/workspaceUi'

const sectionOrder: WorkspaceSection[] = ['overview', 'chat', 'tasks', 'schedule', 'files', 'directory']

function isWorkspaceSection(value: string | undefined): value is WorkspaceSection {
  return sectionOrder.includes(value as WorkspaceSection)
}

export default function WorkspacePage() {
  const { section } = useParams<{ section: string }>()
  const navigate = useNavigate()
  const { logout, username } = useAuthStore()
  const { locale, t } = useI18n()
  const theme = usePreferencesStore((state) => state.theme)
  const soundEnabled = usePreferencesStore((state) => state.soundEnabled)
  const compactMode = usePreferencesStore((state) => state.compactMode)
  const setLocale = usePreferencesStore((state) => state.setLocale)
  const setTheme = usePreferencesStore((state) => state.setTheme)
  const setSoundEnabled = usePreferencesStore((state) => state.setSoundEnabled)
  const setCompactMode = usePreferencesStore((state) => state.setCompactMode)
  const resetPreferences = usePreferencesStore((state) => state.reset)
  const addToast = useToastStore((state) => state.addToast)

  const activeSection = isWorkspaceSection(section) ? section : 'overview'
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [moduleSwitcherOpen, setModuleSwitcherOpen] = useState(false)
  const [confirmRequest, setConfirmRequest] = useState<ConfirmRequest | null>(null)
  const [confirmBusy, setConfirmBusy] = useState(false)

  const workspace = useWorkspaceData(activeSection, username, locale)
  const notify = useCallback((type: ToastType, message: string) => addToast({ type, message }), [addToast])
  const actions = useWorkspaceActions({
    tasks: workspace.tasks,
    schedule: workspace.schedule,
    setTasks: workspace.setTasks,
    setSchedule: workspace.setSchedule,
    setUploadedFiles: workspace.setUploadedFiles,
    setUploadShareCounts: workspace.setUploadShareCounts,
    setFilesUploading: workspace.setFilesUploading,
    setUploadProgress: workspace.setUploadProgress,
    notify,
  })

  useEffect(() => {
    document.title = `${t.common.workpal} - ${t.navigation[activeSection]}`
  }, [activeSection, t])

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      const commandPressed = event.ctrlKey || event.metaKey
      if (commandPressed && event.key.toLocaleLowerCase() === 'k') {
        event.preventDefault()
        setModuleSwitcherOpen(true)
      }
      if (commandPressed && event.key === '/') {
        event.preventDefault()
        setDrawerOpen(true)
      }
      if (event.key === 'Escape') {
        setModuleSwitcherOpen(false)
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [])

  const formattedDate = useMemo(
    () =>
      new Intl.DateTimeFormat(locale === 'zh-CN' ? 'zh-CN' : 'en-US', {
        weekday: 'long',
        month: 'long',
        day: 'numeric',
      }).format(new Date()),
    [locale],
  )

  const notifications = useMemo(() => {
    const activeTasks = workspace.tasks.filter((task) => task.status !== 'done').length
    const upcomingEvents = workspace.schedule.filter((event) => new Date(event.startsAt).getTime() >= Date.now()).length
    return [
      activeTasks > 0 ? `${t.overview.cards.activeTasks}: ${activeTasks}` : '',
      upcomingEvents > 0 ? `${t.overview.cards.todayMeetings}: ${upcomingEvents}` : '',
    ].filter(Boolean)
  }, [t.overview.cards.activeTasks, t.overview.cards.todayMeetings, workspace.schedule, workspace.tasks])

  const openSection = (targetSection: WorkspaceSection) => navigate(`/workspace/${targetSection}`)
  const requestConfirm = (request: ConfirmRequest) => setConfirmRequest(request)

  const handleConfirm = async () => {
    if (!confirmRequest) return
    setConfirmBusy(true)
    try {
      await confirmRequest.onConfirm()
      setConfirmRequest(null)
    } finally {
      setConfirmBusy(false)
    }
  }

  const requestLogout = () => {
    requestConfirm({
      title: t.confirm.signOutTitle,
      message: t.confirm.signOutMessage,
      confirmText: t.shell.signOut,
      cancelText: t.common.cancel,
      onConfirm: () => {
        logout()
        navigate('/login', { replace: true })
      },
    })
  }

  if (!isWorkspaceSection(section)) return <Navigate to="/workspace/overview" replace />

  return (
    <div className="workspace-shell">
      <Sidebar
        activeSection={activeSection}
        userInfo={workspace.currentUser}
        username={username}
        labels={t}
        onNavigate={openSection}
        onOpenSettings={() => setDrawerOpen(true)}
      />

      <main className="workspace-main">
        <Topbar
          activeSection={activeSection}
          formattedDate={formattedDate}
          locale={locale}
          theme={theme}
          labels={t}
          notifications={notifications}
          onLocaleChange={setLocale}
          onThemeChange={setTheme}
          onOpenSettings={() => setDrawerOpen(true)}
          onLogout={requestLogout}
        />

        <div className="workspace-content">
          {workspace.loadError ? <div className="banner-error" role="alert">{workspace.loadError}</div> : null}
          {!workspace.loading && workspace.teamMembers.length === 0 ? (
            <div className="banner-info" role="status">{t.overview.noUsers}</div>
          ) : null}
          <ErrorBoundary resetKey={activeSection} title="模块加载失败">
            <Suspense fallback={<div className="module-surface empty-panel skeleton-panel" role="status">{t.common.loading}</div>}>
              {workspace.loading ? <div className="module-surface empty-panel skeleton-panel" role="status">{t.common.loading}</div> : (
                <WorkspaceContent
                  activeSection={activeSection}
                  locale={locale}
                  text={t}
                  workspace={workspace}
                  actions={actions}
                  onOpenSection={openSection}
                  onConfirm={requestConfirm}
                />
              )}
            </Suspense>
          </ErrorBoundary>
        </div>
      </main>

      <SettingsDrawer
        open={drawerOpen}
        locale={locale}
        theme={theme}
        soundEnabled={soundEnabled}
        compactMode={compactMode}
        text={t}
        onClose={() => setDrawerOpen(false)}
        onLocaleChange={setLocale}
        onThemeChange={setTheme}
        onSoundChange={setSoundEnabled}
        onCompactModeChange={setCompactMode}
        onReset={resetPreferences}
      />
      <ModuleSwitcher
        open={moduleSwitcherOpen}
        activeSection={activeSection}
        sections={sectionOrder}
        labels={t}
        onNavigate={openSection}
        onClose={() => setModuleSwitcherOpen(false)}
      />
      <ConfirmDialog
        open={Boolean(confirmRequest)}
        title={confirmRequest?.title ?? ''}
        message={confirmRequest?.message ?? ''}
        confirmText={confirmRequest?.confirmText ?? t.confirm.confirmAction}
        cancelText={confirmRequest?.cancelText ?? t.common.cancel}
        {...(confirmRequest?.variant ? { variant: confirmRequest.variant } : {})}
        busy={confirmBusy}
        onConfirm={() => void handleConfirm()}
        onCancel={() => setConfirmRequest(null)}
      />
      <ToastViewport />
    </div>
  )
}
