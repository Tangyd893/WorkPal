import { useEffect, useMemo, useState } from 'react'
import { Navigate, useNavigate, useParams } from 'react-router-dom'
import { workpalApi } from '../api/workpal'
import DirectoryPanel from '../components/workspace/DirectoryPanel'
import FilesPanel from '../components/workspace/FilesPanel'
import OverviewPanel from '../components/workspace/OverviewPanel'
import SchedulePanel from '../components/workspace/SchedulePanel'
import SettingsDrawer from '../components/workspace/SettingsDrawer'
import TasksPanel from '../components/workspace/TasksPanel'
import { buildDocuments, buildSchedule, buildSeedTasks, getTeamProfileMeta } from '../data/workspace'
import { useAuthStore } from '../hooks/useAuthStore'
import { usePreferencesStore } from '../hooks/usePreferencesStore'
import { useI18n } from '../i18n'
import type { TaskStatus, WorkspaceSection, WorkspaceTask, WorkspaceUser } from '../types/workspace'
import ChatPage from './ChatPage'

const sectionOrder: WorkspaceSection[] = ['overview', 'chat', 'tasks', 'schedule', 'files', 'directory']

const nextTaskStatus: Record<TaskStatus, TaskStatus> = {
  planned: 'in_progress',
  in_progress: 'review',
  review: 'done',
  done: 'done',
}

function isWorkspaceSection(value: string | undefined): value is WorkspaceSection {
  return sectionOrder.includes(value as WorkspaceSection)
}

function matchesUserQuery(user: WorkspaceUser, query: string): boolean {
  if (!query.trim()) {
    return true
  }

  const keyword = query.trim().toLowerCase()
  return [user.username, user.nickname, user.email].some((value) => value?.toLowerCase().includes(keyword))
}

function mergeTaskStatus(currentTasks: WorkspaceTask[], nextTasks: WorkspaceTask[]): WorkspaceTask[] {
  const statusById = new Map(currentTasks.map((task) => [task.id, task.status]))
  return nextTasks.map((task) => ({
    ...task,
    status: statusById.get(task.id) ?? task.status,
  }))
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

  const [drawerOpen, setDrawerOpen] = useState(false)
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')
  const [directoryQuery, setDirectoryQuery] = useState('')
  const [currentUser, setCurrentUser] = useState<WorkspaceUser | null>(null)
  const [teamMembers, setTeamMembers] = useState<WorkspaceUser[]>([])
  const [tasks, setTasks] = useState<WorkspaceTask[]>(() => buildSeedTasks(locale))

  const activeSection = isWorkspaceSection(section) ? section : null

  useEffect(() => {
    setTasks((currentTasks) => mergeTaskStatus(currentTasks, buildSeedTasks(locale)))
  }, [locale])

  useEffect(() => {
    let disposed = false

    const loadWorkspaceData = async () => {
      setLoading(true)
      setLoadError('')

      try {
        const [me, users] = await Promise.all([workpalApi.getMe(), workpalApi.listUsers()])
        if (disposed) {
          return
        }

        const sortedUsers = [...users].sort((left, right) => left.username.localeCompare(right.username))
        setCurrentUser(me)
        setTeamMembers(sortedUsers)
      } catch (error) {
        if (disposed) {
          return
        }

        setLoadError(error instanceof Error ? error.message : 'Unable to load workspace data.')
      } finally {
        if (!disposed) {
          setLoading(false)
        }
      }
    }

    void loadWorkspaceData()
    return () => {
      disposed = true
    }
  }, [])

  useEffect(() => {
    if (!activeSection) {
      return
    }

    document.title = `${t.common.workpal} - ${t.navigation[activeSection]}`
  }, [activeSection, t])

  const schedule = useMemo(() => buildSchedule(locale), [locale])
  const documents = useMemo(() => buildDocuments(locale), [locale])
  const filteredUsers = useMemo(
    () => teamMembers.filter((user) => matchesUserQuery(user, directoryQuery)),
    [directoryQuery, teamMembers],
  )
  const formattedDate = useMemo(
    () =>
      new Intl.DateTimeFormat(locale === 'zh-CN' ? 'zh-CN' : 'en-US', {
        weekday: 'long',
        month: 'long',
        day: 'numeric',
      }).format(new Date()),
    [locale],
  )

  const displayNameByUsername = useMemo(() => {
    const nameMap = new Map<string, string>()
    teamMembers.forEach((user) => {
      nameMap.set(user.username, user.nickname || user.username)
    })
    if (currentUser) {
      nameMap.set(currentUser.username, currentUser.nickname || currentUser.username)
    }
    return nameMap
  }, [currentUser, teamMembers])

  const getDisplayName = (accountUsername: string): string => displayNameByUsername.get(accountUsername) ?? accountUsername

  if (!activeSection) {
    return <Navigate to="/workspace/overview" replace />
  }

  const handleLogout = () => {
    logout()
    navigate('/login', { replace: true })
  }

  const handleAdvanceTask = (taskId: string) => {
    setTasks((currentTasks) =>
      currentTasks.map((task) =>
        task.id === taskId
          ? {
              ...task,
              status: nextTaskStatus[task.status],
            }
          : task,
      ),
    )
  }

  const handleResetTask = (taskId: string) => {
    setTasks((currentTasks) =>
      currentTasks.map((task) =>
        task.id === taskId
          ? {
              ...task,
              status: 'planned',
            }
          : task,
      ),
    )
  }

  let sectionContent: JSX.Element
  switch (activeSection) {
    case 'overview':
      sectionContent = (
        <OverviewPanel
          users={teamMembers}
          tasks={tasks}
          events={schedule}
          documents={documents}
          text={t}
          getDisplayName={getDisplayName}
        />
      )
      break
    case 'chat':
      sectionContent = <ChatPage teamMembers={teamMembers} text={t} />
      break
    case 'tasks':
      sectionContent = (
        <TasksPanel
          tasks={tasks}
          text={t}
          getDisplayName={getDisplayName}
          onAdvanceTask={handleAdvanceTask}
          onResetTask={handleResetTask}
        />
      )
      break
    case 'schedule':
      sectionContent = <SchedulePanel events={schedule} locale={locale} text={t} getDisplayName={getDisplayName} />
      break
    case 'files':
      sectionContent = <FilesPanel documents={documents} text={t} getDisplayName={getDisplayName} />
      break
    case 'directory':
      sectionContent = (
        <DirectoryPanel
          users={filteredUsers}
          query={directoryQuery}
          currentUserId={currentUser?.id ?? null}
          text={t}
          onQueryChange={setDirectoryQuery}
          getProfileMeta={(accountUsername) => getTeamProfileMeta(locale, accountUsername)}
        />
      )
      break
  }

  return (
    <div className="workspace-shell">
      <aside className="workspace-sidebar">
        <div className="brand-block">
          <strong>{t.common.workpal}</strong>
          <span>{t.shell.subtitle}</span>
        </div>

        <nav className="workspace-nav">
          {sectionOrder.map((item) => (
            <button
              key={item}
              type="button"
              className={item === activeSection ? 'nav-button active' : 'nav-button'}
              onClick={() => navigate(`/workspace/${item}`)}
            >
              <span>{t.navigation[item]}</span>
            </button>
          ))}
        </nav>

        <div className="sidebar-footer">
          <div className="profile-card">
            <strong>{currentUser?.nickname || username || t.common.unavailable}</strong>
            <span>@{currentUser?.username || username || 'guest'}</span>
          </div>
          <button type="button" className="secondary-button block-button" onClick={() => setDrawerOpen(true)}>
            {t.shell.preferences}
          </button>
        </div>
      </aside>

      <div className="workspace-main">
        <header className="workspace-topbar">
          <div>
            <span className="eyebrow">
              {t.shell.datePrefix} {formattedDate}
            </span>
            <h1>{t.navigation[activeSection]}</h1>
            <p>{t.shell.liveData}</p>
          </div>

          <div className="topbar-actions">
            <div className="segmented-control">
              <button
                type="button"
                className={locale === 'en' ? 'segment-button active' : 'segment-button'}
                onClick={() => setLocale('en')}
              >
                English
              </button>
              <button
                type="button"
                className={locale === 'zh-CN' ? 'segment-button active' : 'segment-button'}
                onClick={() => setLocale('zh-CN')}
              >
                简体中文
              </button>
            </div>

            <div className="segmented-control">
              <button
                type="button"
                className={theme === 'light' ? 'segment-button active' : 'segment-button'}
                onClick={() => setTheme('light')}
              >
                {t.settings.light}
              </button>
              <button
                type="button"
                className={theme === 'dark' ? 'segment-button active' : 'segment-button'}
                onClick={() => setTheme('dark')}
              >
                {t.settings.dark}
              </button>
            </div>

            <button type="button" className="secondary-button" onClick={() => setDrawerOpen(true)}>
              {t.shell.preferences}
            </button>
            <button type="button" className="secondary-button" onClick={handleLogout}>
              {t.shell.signOut}
            </button>
          </div>
        </header>

        <div className="workspace-content">
          {loading ? <div className="module-surface empty-panel">{t.common.loading}</div> : sectionContent}
          {loadError ? <div className="banner-error">{loadError}</div> : null}
          {!loading && teamMembers.length === 0 ? <div className="banner-info">{t.overview.noUsers}</div> : null}
        </div>
      </div>

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
    </div>
  )
}
