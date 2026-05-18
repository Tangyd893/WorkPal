import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { workpalApi } from '../api/workpal'
import type {
  ConversationFile,
  Department,
  Issue,
  IssueType,
  Project,
  ScheduleEvent,
  SharedDocument,
  WorkspaceSection,
  WorkspaceTask,
  WorkspaceUser,
} from '../types/workspace'

type ResourceKey = 'tasks' | 'schedule' | 'files' | 'projects'

const resourcesBySection: Record<WorkspaceSection, ResourceKey[]> = {
  overview: ['tasks', 'schedule', 'files'],
  chat: [],
  tasks: ['tasks'],
  schedule: ['schedule'],
  files: ['files'],
  directory: [],
  projects: ['projects'],
  docs: [],
  calendar: [],
  approvals: [],
}

function sortUsers(users: WorkspaceUser[]): WorkspaceUser[] {
  return [...users].sort((left, right) => (left.nickname || left.username).localeCompare(right.nickname || right.username))
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

export function useWorkspaceData(activeSection: WorkspaceSection, username: string | null, locale: 'en' | 'zh-CN') {
  const [baseLoading, setBaseLoading] = useState(true)
  const [loadError, setLoadError] = useState('')
  const [directoryQuery, setDirectoryQuery] = useState('')
  const [directoryDepartmentID, setDirectoryDepartmentID] = useState(0)
  const [directoryLoading, setDirectoryLoading] = useState(false)
  const [currentUser, setCurrentUser] = useState<WorkspaceUser | null>(null)
  const [teamMembers, setTeamMembers] = useState<WorkspaceUser[]>([])
  const [directoryUsers, setDirectoryUsers] = useState<WorkspaceUser[]>([])
  const [departments, setDepartments] = useState<Department[]>([])
  const [tasks, setTasks] = useState<WorkspaceTask[]>([])
  const [schedule, setSchedule] = useState<ScheduleEvent[]>([])
  const [uploadedFiles, setUploadedFiles] = useState<ConversationFile[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(null)
  const [projectIssues, setProjectIssues] = useState<Issue[]>([])
  const [projectIssueTypes, setProjectIssueTypes] = useState<IssueType[]>([])
  const [projectIssuesLoading, setProjectIssuesLoading] = useState(false)
  const [uploadShareCounts, setUploadShareCounts] = useState<Record<number, number>>({})
  const [filesUploading, setFilesUploading] = useState(false)
  const [uploadProgress, setUploadProgress] = useState(0)
  const [loadedResources, setLoadedResources] = useState<Record<ResourceKey, boolean>>({
    tasks: false,
    schedule: false,
    files: false,
    projects: false,
  })
  const [pendingResources, setPendingResources] = useState<Record<ResourceKey, boolean>>({
    tasks: false,
    schedule: false,
    files: false,
    projects: false,
  })

  const directoryRequestRef = useRef(0)

  useEffect(() => {
    let disposed = false

    const loadBaseData = async () => {
      setBaseLoading(true)
      setLoadError('')

      try {
        const [me, users, departmentList] = await Promise.all([
          workpalApi.getMe(),
          workpalApi.listUsers(),
          workpalApi.listDepartments(),
        ])
        if (disposed) {
          return
        }

        const sortedUsers = sortUsers(users)
        setCurrentUser(me)
        setTeamMembers(sortedUsers)
        setDirectoryUsers(sortedUsers)
        setDepartments(departmentList)
      } catch (error) {
        if (!disposed) {
          setLoadError(error instanceof Error ? error.message : 'Unable to load workspace data.')
        }
      } finally {
        if (!disposed) {
          setBaseLoading(false)
        }
      }
    }

    void loadBaseData()
    return () => {
      disposed = true
    }
  }, [])

  const markResourcePending = useCallback((resource: ResourceKey, pending: boolean) => {
    setPendingResources((current) => ({ ...current, [resource]: pending }))
  }, [])

  const markResourceLoaded = useCallback((resource: ResourceKey) => {
    setLoadedResources((current) => ({ ...current, [resource]: true }))
  }, [])

  const loadTasks = useCallback(async () => {
    if (loadedResources.tasks || pendingResources.tasks) {
      return
    }

    markResourcePending('tasks', true)
    try {
      setTasks(await workpalApi.listTasks())
      markResourceLoaded('tasks')
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : 'Unable to load tasks.')
    } finally {
      markResourcePending('tasks', false)
    }
  }, [loadedResources.tasks, markResourceLoaded, markResourcePending, pendingResources.tasks])

  const loadSchedule = useCallback(async () => {
    if (loadedResources.schedule || pendingResources.schedule) {
      return
    }

    markResourcePending('schedule', true)
    try {
      setSchedule(await workpalApi.listSchedule())
      markResourceLoaded('schedule')
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : 'Unable to load schedule.')
    } finally {
      markResourcePending('schedule', false)
    }
  }, [loadedResources.schedule, markResourceLoaded, markResourcePending, pendingResources.schedule])

  const loadFiles = useCallback(async () => {
    if (loadedResources.files || pendingResources.files) {
      return
    }

    markResourcePending('files', true)
    try {
      setUploadedFiles(await workpalApi.listUserFiles())
      markResourceLoaded('files')
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : 'Unable to load files.')
    } finally {
      markResourcePending('files', false)
    }
  }, [loadedResources.files, markResourceLoaded, markResourcePending, pendingResources.files])

  const loadProjects = useCallback(async () => {
    if (loadedResources.projects || pendingResources.projects) return
    markResourcePending('projects', true)
    try {
      setProjects(await workpalApi.listProjects())
      markResourceLoaded('projects')
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : 'Unable to load projects.')
    } finally {
      markResourcePending('projects', false)
    }
  }, [loadedResources.projects, markResourceLoaded, markResourcePending, pendingResources.projects])

  useEffect(() => {
    if (baseLoading) {
      return
    }

    const requiredResources = resourcesBySection[activeSection]
    if (requiredResources.includes('tasks')) {
      void loadTasks()
    }
    if (requiredResources.includes('schedule')) {
      void loadSchedule()
    }
    if (requiredResources.includes('files')) {
      void loadFiles()
    }
    if (requiredResources.includes('projects')) {
      void loadProjects()
    }
  }, [activeSection, baseLoading, loadFiles, loadProjects, loadSchedule, loadTasks])

  useEffect(() => {
    if (baseLoading || activeSection !== 'directory') {
      return undefined
    }

    const requestID = directoryRequestRef.current + 1
    directoryRequestRef.current = requestID
    const controller = new AbortController()
    const timer = window.setTimeout(async () => {
      setDirectoryLoading(true)
      try {
        const users = await workpalApi.listUsers(100, directoryQuery, directoryDepartmentID || undefined, controller.signal)
        if (directoryRequestRef.current === requestID) {
          setDirectoryUsers(sortUsers(users))
        }
      } catch (error) {
        if (!controller.signal.aborted && directoryRequestRef.current === requestID) {
          setLoadError(error instanceof Error ? error.message : 'Unable to search the directory.')
        }
      } finally {
        if (!controller.signal.aborted && directoryRequestRef.current === requestID) {
          setDirectoryLoading(false)
        }
      }
    }, 180)

    return () => {
      controller.abort()
      window.clearTimeout(timer)
    }
  }, [activeSection, baseLoading, directoryDepartmentID, directoryQuery])

  const loadProjectIssues = useCallback(async (projectId: string) => {
    setProjectIssuesLoading(true)
    try {
      const [issues, types] = await Promise.all([
        workpalApi.listIssues(projectId),
        workpalApi.listIssueTypes(projectId),
      ])
      setProjectIssues(issues)
      setProjectIssueTypes(types)
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : 'Unable to load issues.')
    } finally {
      setProjectIssuesLoading(false)
    }
  }, [])

  const documents = useMemo(() => {
    const ownerUsername = currentUser?.username || username || 'admin'
    return uploadedFiles
      .map((file) => mapUploadedFileToDocument(file, ownerUsername, uploadShareCounts[file.id] ?? 0, locale))
      .sort((left, right) => new Date(right.updatedAt).getTime() - new Date(left.updatedAt).getTime())
  }, [currentUser, locale, uploadShareCounts, uploadedFiles, username])

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

  const requiredResources = resourcesBySection[activeSection]
  const loading =
    baseLoading ||
    requiredResources.some((resource) => pendingResources[resource] || !loadedResources[resource])

  return {
    currentUser,
    departments,
    directoryDepartmentID,
    directoryLoading,
    directoryQuery,
    directoryUsers,
    documents,
    filesUploading,
    getDisplayName: (accountUsername: string) => displayNameByUsername.get(accountUsername) ?? accountUsername,
    loading,
    loadError,
    projects,
    projectIssueTypes,
    projectIssues,
    projectIssuesLoading,
    selectedProjectId,
    schedule,
    setDirectoryDepartmentID,
    setDirectoryQuery,
    setFilesUploading,
    setProjectIssues,
    setProjects,
    setSchedule,
    setSelectedProjectId,
    setTasks,
    setUploadProgress,
    setUploadShareCounts,
    setUploadedFiles,
    tasks,
    teamMembers,
    uploadProgress,
    uploadShareCounts,
    uploadedFiles,
    loadProjectIssues,
  }
}
