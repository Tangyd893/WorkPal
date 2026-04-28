import { useEffect, useMemo, useState } from 'react'
import { Navigate, useNavigate, useParams } from 'react-router-dom'
import { workpalApi } from '../api/workpal'
import DirectoryPanel from '../components/workspace/DirectoryPanel'
import FilesPanel from '../components/workspace/FilesPanel'
import OverviewPanel from '../components/workspace/OverviewPanel'
import SchedulePanel from '../components/workspace/SchedulePanel'
import SettingsDrawer from '../components/workspace/SettingsDrawer'
import TasksPanel from '../components/workspace/TasksPanel'
import { buildDocuments, buildSchedule, buildSeedTasks } from '../data/workspace'
import { useAuthStore } from '../hooks/useAuthStore'
import { usePreferencesStore } from '../hooks/usePreferencesStore'
import { useI18n } from '../i18n'
import type {
  ConversationFile,
  CreateScheduleInput,
  CreateTaskInput,
  Department,
  ScheduleEvent,
  SharedDocument,
  TaskStatus,
  WorkspaceSection,
  WorkspaceTask,
  WorkspaceUser,
} from '../types/workspace'
import { copyText } from '../utils/clipboard'
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

function createClientID(prefix: string): string {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
}

function mergeLocalizedTasks(currentTasks: WorkspaceTask[], localizedTasks: WorkspaceTask[]): WorkspaceTask[] {
  const currentByID = new Map(currentTasks.map((task) => [task.id, task]))
  const customTasks = currentTasks.filter((task) => task.source === 'custom')

  return [
    ...localizedTasks.map((task) => ({
      ...task,
      status: currentByID.get(task.id)?.status ?? task.status,
      sharedCount: currentByID.get(task.id)?.sharedCount ?? task.sharedCount,
    })),
    ...customTasks,
  ]
}

function mergeLocalizedSchedule(currentEvents: ScheduleEvent[], localizedEvents: ScheduleEvent[]): ScheduleEvent[] {
  const currentByID = new Map(currentEvents.map((event) => [event.id, event]))
  const customEvents = currentEvents.filter((event) => event.source === 'custom')

  return [
    ...localizedEvents.map((event) => ({
      ...event,
      sharedCount: currentByID.get(event.id)?.sharedCount ?? event.sharedCount,
    })),
    ...customEvents,
  ]
}

function mapUploadedFileToDocument(
  file: ConversationFile,
  ownerUsername: string,
  shareCount: number,
  locale: 'en' | 'zh-CN',
): SharedDocument {
  return {
    id: `file-${file.id}`,
    title: file.name,
    summary: `${Math.max(1, Math.round(file.size / 1024))} KB`,
    category: locale === 'zh-CN' ? '上传' : 'Upload',
    ownerUsername,
    updatedAt: file.created_at,
    status: 'ready',
    sharedCount: shareCount,
    source: 'custom',
    fileId: file.id,
    attachmentName: file.name,
    attachmentUrl: file.download_path,
    downloadPath: file.download_path,
  }
}

function sortUsers(users: WorkspaceUser[]): WorkspaceUser[] {
  return [...users].sort((left, right) => (left.nickname || left.username).localeCompare(right.nickname || right.username))
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
  const [actionNotice, setActionNotice] = useState('')
  const [actionError, setActionError] = useState('')
  const [directoryQuery, setDirectoryQuery] = useState('')
  const [directoryDepartmentID, setDirectoryDepartmentID] = useState(0)
  const [directoryLoading, setDirectoryLoading] = useState(false)
  const [currentUser, setCurrentUser] = useState<WorkspaceUser | null>(null)
  const [teamMembers, setTeamMembers] = useState<WorkspaceUser[]>([])
  const [directoryUsers, setDirectoryUsers] = useState<WorkspaceUser[]>([])
  const [departments, setDepartments] = useState<Department[]>([])
  const [tasks, setTasks] = useState<WorkspaceTask[]>(() => buildSeedTasks(locale))
  const [schedule, setSchedule] = useState<ScheduleEvent[]>(() => buildSchedule(locale))
  const [seedDocuments, setSeedDocuments] = useState<SharedDocument[]>(() => buildDocuments(locale))
  const [uploadedFiles, setUploadedFiles] = useState<ConversationFile[]>([])
  const [uploadShareCounts, setUploadShareCounts] = useState<Record<number, number>>({})
  const [filesUploading, setFilesUploading] = useState(false)

  const activeSection = isWorkspaceSection(section) ? section : null

  useEffect(() => {
    setTasks((currentTasks) => mergeLocalizedTasks(currentTasks, buildSeedTasks(locale)))
    setSchedule((currentEvents) => mergeLocalizedSchedule(currentEvents, buildSchedule(locale)))
    setSeedDocuments(buildDocuments(locale))
  }, [locale])

  useEffect(() => {
    let disposed = false

    const loadWorkspaceData = async () => {
      setLoading(true)
      setLoadError('')

      try {
        const [me, users, departmentList, files] = await Promise.all([
          workpalApi.getMe(),
          workpalApi.listUsers(),
          workpalApi.listDepartments(),
          workpalApi.listUserFiles(),
        ])
        if (disposed) {
          return
        }

        const sortedUsers = sortUsers(users)
        setCurrentUser(me)
        setTeamMembers(sortedUsers)
        setDirectoryUsers(sortedUsers)
        setDepartments(departmentList)
        setUploadedFiles(files)
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

  useEffect(() => {
    if (loading) {
      return
    }

    let disposed = false
    const timer = window.setTimeout(async () => {
      setDirectoryLoading(true)
      try {
        const users = await workpalApi.listUsers(100, directoryQuery, directoryDepartmentID || undefined)
        if (!disposed) {
          setDirectoryUsers(sortUsers(users))
          setActionError('')
        }
      } catch (error) {
        if (!disposed) {
          setActionError(error instanceof Error ? error.message : 'Unable to search the directory.')
        }
      } finally {
        if (!disposed) {
          setDirectoryLoading(false)
        }
      }
    }, 180)

    return () => {
      disposed = true
      window.clearTimeout(timer)
    }
  }, [directoryDepartmentID, directoryQuery, loading])

  const documents = useMemo(() => {
    const ownerUsername = currentUser?.username || username || 'admin'
    return [
      ...seedDocuments,
      ...uploadedFiles.map((file) => mapUploadedFileToDocument(file, ownerUsername, uploadShareCounts[file.id] ?? 0, locale)),
    ].sort((left, right) => new Date(right.updatedAt).getTime() - new Date(left.updatedAt).getTime())
  }, [currentUser, locale, seedDocuments, uploadShareCounts, uploadedFiles, username])

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

  const openSection = (targetSection: WorkspaceSection) => {
    navigate(`/workspace/${targetSection}`)
  }

  const notify = (message: string) => {
    setActionNotice(message)
    setActionError('')
  }

  const fail = (message: string) => {
    setActionNotice('')
    setActionError(message)
  }

  const handleAdvanceTask = (taskID: string) => {
    setTasks((currentTasks) =>
      currentTasks.map((task) =>
        task.id === taskID
          ? {
              ...task,
              status: nextTaskStatus[task.status],
            }
          : task,
      ),
    )
  }

  const handleResetTask = (taskID: string) => {
    setTasks((currentTasks) =>
      currentTasks.map((task) =>
        task.id === taskID
          ? {
              ...task,
              status: 'planned',
            }
          : task,
      ),
    )
  }

  const handleAddTask = (draft: CreateTaskInput) => {
    setTasks((currentTasks) => [
      {
        id: createClientID('task'),
        title: draft.title,
        summary: draft.summary,
        project: draft.project,
        ownerUsername: draft.ownerUsername,
        teammates: draft.teammates,
        dueDate: draft.dueDate,
        priority: draft.priority,
        status: 'planned',
        sharedCount: 0,
        source: 'custom',
      },
      ...currentTasks,
    ])
    notify(draft.title)
  }

  const handleDeleteTask = (taskID: string) => {
    setTasks((currentTasks) => currentTasks.filter((task) => task.id !== taskID))
  }

  const handleShareTask = async (taskID: string) => {
    const task = tasks.find((item) => item.id === taskID)
    if (!task) {
      return
    }

    const copied = await copyText(`${task.title}\n${task.summary}\n${task.project}`)
    if (!copied) {
      fail('Unable to copy the task details.')
      return
    }

    setTasks((currentTasks) =>
      currentTasks.map((item) => (item.id === taskID ? { ...item, sharedCount: item.sharedCount + 1 } : item)),
    )
    notify(task.title)
  }

  const handleAddEvent = (draft: CreateScheduleInput) => {
    setSchedule((currentEvents) => [
      {
        id: createClientID('event'),
        title: draft.title,
        detail: draft.detail,
        ownerUsername: draft.ownerUsername,
        startsAt: draft.startsAt,
        durationMinutes: draft.durationMinutes,
        attendees: draft.attendees,
        room: draft.room,
        sharedCount: 0,
        source: 'custom',
      },
      ...currentEvents,
    ])
    notify(draft.title)
  }

  const handleDeleteEvent = (eventID: string) => {
    setSchedule((currentEvents) => currentEvents.filter((event) => event.id !== eventID))
  }

  const handleShareEvent = async (eventID: string) => {
    const event = schedule.find((item) => item.id === eventID)
    if (!event) {
      return
    }

    const copied = await copyText(`${event.title}\n${event.detail}\n${event.room}`)
    if (!copied) {
      fail('Unable to copy the schedule details.')
      return
    }

    setSchedule((currentEvents) =>
      currentEvents.map((item) => (item.id === eventID ? { ...item, sharedCount: item.sharedCount + 1 } : item)),
    )
    notify(event.title)
  }

  const handleUploadDocument = async (file: File) => {
    setFilesUploading(true)
    try {
      const uploaded = await workpalApi.uploadUserFile(file)
      setUploadedFiles((currentFiles) => [uploaded, ...currentFiles])
      notify(uploaded.name)
    } catch (error) {
      fail(error instanceof Error ? error.message : 'Unable to upload the file.')
    } finally {
      setFilesUploading(false)
    }
  }

  const handleDeleteDocument = async (document: SharedDocument) => {
    if (document.fileId) {
      try {
        await workpalApi.deleteFile(document.fileId)
        setUploadedFiles((currentFiles) => currentFiles.filter((file) => file.id !== document.fileId))
        notify(document.title)
      } catch (error) {
        fail(error instanceof Error ? error.message : 'Unable to delete the file.')
      }
      return
    }

    setSeedDocuments((currentDocuments) => currentDocuments.filter((item) => item.id !== document.id))
  }

  const handleShareDocument = async (document: SharedDocument) => {
    if (document.fileId) {
      try {
        const shareInfo = await workpalApi.shareFile(document.fileId)
        const copied = await copyText(shareInfo.share_text)
        if (!copied) {
          fail('Unable to copy the file share link.')
          return
        }

        setUploadShareCounts((current) => ({
          ...current,
          [document.fileId as number]: (current[document.fileId as number] ?? 0) + 1,
        }))
        notify(shareInfo.share_text)
      } catch (error) {
        fail(error instanceof Error ? error.message : 'Unable to share the file.')
      }
      return
    }

    const copied = await copyText(`${document.title}\n${document.summary}`)
    if (!copied) {
      fail('Unable to copy the document summary.')
      return
    }

    setSeedDocuments((currentDocuments) =>
      currentDocuments.map((item) =>
        item.id === document.id
          ? {
              ...item,
              sharedCount: item.sharedCount + 1,
            }
          : item,
      ),
    )
    notify(document.title)
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
          onOpenSection={openSection}
        />
      )
      break
    case 'chat':
      sectionContent = <ChatPage teamMembers={teamMembers} text={t} />
      break
    case 'tasks':
      sectionContent = (
        <TasksPanel
          users={teamMembers}
          tasks={tasks}
          text={t}
          getDisplayName={getDisplayName}
          onAdvanceTask={handleAdvanceTask}
          onResetTask={handleResetTask}
          onAddTask={handleAddTask}
          onDeleteTask={handleDeleteTask}
          onShareTask={(taskID) => {
            void handleShareTask(taskID)
          }}
        />
      )
      break
    case 'schedule':
      sectionContent = (
        <SchedulePanel
          users={teamMembers}
          events={schedule}
          locale={locale}
          text={t}
          getDisplayName={getDisplayName}
          onAddEvent={handleAddEvent}
          onDeleteEvent={handleDeleteEvent}
          onShareEvent={(eventID) => {
            void handleShareEvent(eventID)
          }}
        />
      )
      break
    case 'files':
      sectionContent = (
        <FilesPanel
          documents={documents}
          text={t}
          getDisplayName={getDisplayName}
          uploading={filesUploading}
          onUpload={handleUploadDocument}
          onDelete={handleDeleteDocument}
          onShare={handleShareDocument}
        />
      )
      break
    case 'directory':
      sectionContent = (
        <DirectoryPanel
          users={directoryUsers}
          departments={departments}
          query={directoryQuery}
          selectedDepartmentId={directoryDepartmentID}
          currentUserId={currentUser?.id ?? null}
          text={t}
          loading={directoryLoading}
          onQueryChange={setDirectoryQuery}
          onDepartmentChange={setDirectoryDepartmentID}
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
          {loadError ? <div className="banner-error">{loadError}</div> : null}
          {actionError ? <div className="banner-error">{actionError}</div> : null}
          {actionNotice ? <div className="banner-info">{actionNotice}</div> : null}
          {!loading && teamMembers.length === 0 ? <div className="banner-info">{t.overview.noUsers}</div> : null}
          {loading ? <div className="module-surface empty-panel">{t.common.loading}</div> : sectionContent}
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
