import type { Dispatch, SetStateAction } from 'react'
import { workpalApi } from '../api/workpal'
import type {
  ConversationFile,
  CreateIssueInput,
  CreateProjectInput,
  CreateScheduleInput,
  CreateTaskInput,
  Issue,
  Project,
  ScheduleEvent,
  SharedDocument,
  TaskStatus,
  WorkspaceTask,
} from '../types/workspace'
import { copyText } from '../utils/clipboard'
import type { ToastType } from './useToastStore'

interface WorkspaceActionsInput {
  tasks: WorkspaceTask[]
  schedule: ScheduleEvent[]
  projects: Project[]
  projectIssues: Issue[]
  selectedProjectId: string | null
  setTasks: Dispatch<SetStateAction<WorkspaceTask[]>>
  setSchedule: Dispatch<SetStateAction<ScheduleEvent[]>>
  setProjects: Dispatch<SetStateAction<Project[]>>
  setProjectIssues: Dispatch<SetStateAction<Issue[]>>
  setUploadedFiles: Dispatch<SetStateAction<ConversationFile[]>>
  setUploadShareCounts: Dispatch<SetStateAction<Record<number, number>>>
  setFilesUploading: Dispatch<SetStateAction<boolean>>
  setUploadProgress: Dispatch<SetStateAction<number>>
  notify: (type: ToastType, message: string) => void
}

const nextTaskStatus: Record<TaskStatus, TaskStatus> = {
  planned: 'in_progress',
  in_progress: 'review',
  review: 'done',
  done: 'done',
}

function getErrorMessage(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback
}

export function useWorkspaceActions({
  tasks,
  schedule,
  projects: _projects,
  projectIssues: _projectIssues,
  selectedProjectId: _selectedProjectId,
  setTasks,
  setSchedule,
  setProjects,
  setProjectIssues,
  setUploadedFiles,
  setUploadShareCounts,
  setFilesUploading,
  setUploadProgress,
  notify,
}: WorkspaceActionsInput) {
  const replaceTask = (updatedTask: WorkspaceTask) => {
    setTasks((currentTasks) => currentTasks.map((task) => (task.id === updatedTask.id ? updatedTask : task)))
  }

  const handleUpdateTaskStatus = async (taskID: string, status: TaskStatus) => {
    try {
      const updatedTask = await workpalApi.updateTaskStatus(taskID, status)
      replaceTask(updatedTask)
      notify('success', updatedTask.title)
    } catch (error) {
      notify('error', getErrorMessage(error, 'Unable to update the task.'))
    }
  }

  const handleAdvanceTask = async (taskID: string) => {
    const task = tasks.find((item) => item.id === taskID)
    if (!task) {
      return
    }

    await handleUpdateTaskStatus(taskID, nextTaskStatus[task.status])
  }

  const handleResetTask = async (taskID: string) => {
    await handleUpdateTaskStatus(taskID, 'planned')
  }

  const handleAddTask = async (draft: CreateTaskInput) => {
    try {
      const createdTask = await workpalApi.createTask(draft)
      setTasks((currentTasks) => [createdTask, ...currentTasks])
      notify('success', createdTask.title)
    } catch (error) {
      notify('error', getErrorMessage(error, 'Unable to create the task.'))
    }
  }

  const handleDeleteTask = async (taskID: string) => {
    try {
      await workpalApi.deleteTask(taskID)
      setTasks((currentTasks) => currentTasks.filter((task) => task.id !== taskID))
      notify('success', taskID)
    } catch (error) {
      notify('error', getErrorMessage(error, 'Unable to delete the task.'))
    }
  }

  const handleShareTask = async (taskID: string) => {
    const task = tasks.find((item) => item.id === taskID)
    if (!task) {
      return
    }

    const copied = await copyText(`${task.title}\n${task.summary}\n${task.project}`)
    if (!copied) {
      notify('error', 'Unable to copy the task details.')
      return
    }

    try {
      const updatedTask = await workpalApi.shareTask(taskID)
      replaceTask(updatedTask)
      notify('success', task.title)
    } catch (error) {
      notify('error', getErrorMessage(error, 'Unable to share the task.'))
    }
  }

  const replaceEvent = (updatedEvent: ScheduleEvent) => {
    setSchedule((currentEvents) => currentEvents.map((event) => (event.id === updatedEvent.id ? updatedEvent : event)))
  }

  const handleAddEvent = async (draft: CreateScheduleInput) => {
    try {
      const createdEvent = await workpalApi.createScheduleEvent(draft)
      setSchedule((currentEvents) => [createdEvent, ...currentEvents])
      notify('success', createdEvent.title)
    } catch (error) {
      notify('error', getErrorMessage(error, 'Unable to create the schedule event.'))
    }
  }

  const handleDeleteEvent = async (eventID: string) => {
    try {
      await workpalApi.deleteScheduleEvent(eventID)
      setSchedule((currentEvents) => currentEvents.filter((event) => event.id !== eventID))
      notify('success', eventID)
    } catch (error) {
      notify('error', getErrorMessage(error, 'Unable to delete the schedule event.'))
    }
  }

  const handleShareEvent = async (eventID: string) => {
    const event = schedule.find((item) => item.id === eventID)
    if (!event) {
      return
    }

    const copied = await copyText(`${event.title}\n${event.detail}\n${event.room}`)
    if (!copied) {
      notify('error', 'Unable to copy the schedule details.')
      return
    }

    try {
      const updatedEvent = await workpalApi.shareScheduleEvent(eventID)
      replaceEvent(updatedEvent)
      notify('success', event.title)
    } catch (error) {
      notify('error', getErrorMessage(error, 'Unable to share the schedule event.'))
    }
  }

  const handleUploadDocument = async (file: File) => {
    setFilesUploading(true)
    setUploadProgress(0)
    try {
      const uploaded = await workpalApi.uploadUserFile(file, setUploadProgress)
      setUploadProgress(100)
      setUploadedFiles((currentFiles) => [uploaded, ...currentFiles])
      notify('success', uploaded.name)
    } catch (error) {
      notify('error', getErrorMessage(error, 'Unable to upload the file.'))
    } finally {
      window.setTimeout(() => setUploadProgress(0), 300)
      setFilesUploading(false)
    }
  }

  const handleAddProject = async (draft: CreateProjectInput) => {
    try {
      const project = await workpalApi.createProject(draft)
      setProjects((current) => [project, ...current])
      notify('success', project.name)
    } catch (error) {
      notify('error', getErrorMessage(error, 'Unable to create project.'))
    }
  }

  const handleDeleteProject = async (projectID: string) => {
    try {
      await workpalApi.deleteProject(projectID)
      setProjects((current) => current.filter((p) => p.id !== projectID))
      notify('success', projectID)
    } catch (error) {
      notify('error', getErrorMessage(error, 'Unable to delete project.'))
    }
  }

  const handleAddIssue = async (projectID: string, draft: CreateIssueInput) => {
    try {
      const issue = await workpalApi.createIssue(projectID, draft)
      setProjectIssues((current) => [issue, ...current])
      notify('success', issue.key)
    } catch (error) {
      notify('error', getErrorMessage(error, 'Unable to create issue.'))
    }
  }

  const handleUpdateIssueStatus = async (issueID: string, status: string) => {
    try {
      const updated = await workpalApi.updateIssueStatus(issueID, status)
      setProjectIssues((current) => current.map((iss) => iss.id === updated.id ? updated : iss))
      notify('success', updated.key)
    } catch (error) {
      notify('error', getErrorMessage(error, 'Unable to update issue.'))
    }
  }

  const handleDeleteIssue = async (issueID: string) => {
    try {
      await workpalApi.deleteIssue(issueID)
      setProjectIssues((current) => current.filter((iss) => iss.id !== issueID))
      notify('success', issueID)
    } catch (error) {
      notify('error', getErrorMessage(error, 'Unable to delete issue.'))
    }
  }

  const handleDeleteDocument = async (document: SharedDocument) => {
    if (!document.fileId) {
      return
    }

    try {
      await workpalApi.deleteFile(document.fileId)
      setUploadedFiles((currentFiles) => currentFiles.filter((file) => file.id !== document.fileId))
      notify('success', document.title)
    } catch (error) {
      notify('error', getErrorMessage(error, 'Unable to delete the file.'))
    }
  }

  const handleShareDocument = async (document: SharedDocument) => {
    if (!document.fileId) {
      return
    }

    try {
      const shareInfo = await workpalApi.shareFile(document.fileId)
      const copied = await copyText(shareInfo.share_text)
      if (!copied) {
        notify('warning', shareInfo.download_path)
        return
      }

      setUploadShareCounts((current) => ({
        ...current,
        [document.fileId as number]: (current[document.fileId as number] ?? 0) + 1,
      }))
      notify('success', shareInfo.share_text)
    } catch (error) {
      notify('error', getErrorMessage(error, 'Unable to share the file.'))
    }
  }

  return {
    handleAddEvent,
    handleAddIssue,
    handleAddProject,
    handleAddTask,
    handleAdvanceTask,
    handleDeleteDocument,
    handleDeleteEvent,
    handleDeleteIssue,
    handleDeleteProject,
    handleDeleteTask,
    handleUploadDocument,
    handleResetTask,
    handleShareDocument,
    handleShareEvent,
    handleShareTask,
    handleUpdateIssueStatus,
    handleUpdateTaskStatus,
  }
}
