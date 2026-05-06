import { lazy } from 'react'
import type { AppTranslations } from '../../i18n'
import type { Locale, SharedDocument, WorkspaceSection } from '../../types/workspace'
import type { ConfirmRequest } from '../../types/workspaceUi'
import type { useWorkspaceActions } from '../../hooks/useWorkspaceActions'
import type { useWorkspaceData } from '../../hooks/useWorkspaceData'

const OverviewPanel = lazy(() => import('./OverviewPanel'))
const ChatPage = lazy(() => import('../../pages/ChatPage'))
const TasksPanel = lazy(() => import('./TasksPanel'))
const SchedulePanel = lazy(() => import('./SchedulePanel'))
const FilesPanel = lazy(() => import('./FilesPanel'))
const DirectoryPanel = lazy(() => import('./DirectoryPanel'))

interface WorkspaceContentProps {
  activeSection: WorkspaceSection
  locale: Locale
  text: AppTranslations
  workspace: ReturnType<typeof useWorkspaceData>
  actions: ReturnType<typeof useWorkspaceActions>
  onOpenSection: (section: WorkspaceSection) => void
  onConfirm: (request: ConfirmRequest) => void
}

export default function WorkspaceContent({
  activeSection,
  locale,
  text,
  workspace,
  actions,
  onOpenSection,
  onConfirm,
}: WorkspaceContentProps) {
  const requestDeleteDocument = (document: SharedDocument) => {
    onConfirm({
      title: text.confirm.deleteFileTitle,
      message: text.confirm.deleteFileMessage,
      confirmText: text.common.delete,
      cancelText: text.common.cancel,
      variant: 'danger',
      onConfirm: () => actions.handleDeleteDocument(document),
    })
  }

  switch (activeSection) {
    case 'overview':
      return (
        <OverviewPanel
          users={workspace.teamMembers}
          tasks={workspace.tasks}
          events={workspace.schedule}
          documents={workspace.documents}
          text={text}
          getDisplayName={workspace.getDisplayName}
          onOpenSection={onOpenSection}
        />
      )
    case 'chat':
      return <ChatPage teamMembers={workspace.teamMembers} text={text} />
    case 'tasks':
      return (
        <TasksPanel
          users={workspace.teamMembers}
          tasks={workspace.tasks}
          text={text}
          getDisplayName={workspace.getDisplayName}
          onAdvanceTask={actions.handleAdvanceTask}
          onResetTask={actions.handleResetTask}
          onAddTask={actions.handleAddTask}
          onUpdateTaskStatus={actions.handleUpdateTaskStatus}
          onDeleteTask={(taskID) =>
            onConfirm({
              title: text.confirm.deleteTaskTitle,
              message: text.confirm.deleteTaskMessage,
              confirmText: text.common.delete,
              cancelText: text.common.cancel,
              variant: 'danger',
              onConfirm: () => actions.handleDeleteTask(taskID),
            })
          }
          onShareTask={(taskID) => void actions.handleShareTask(taskID)}
        />
      )
    case 'schedule':
      return (
        <SchedulePanel
          users={workspace.teamMembers}
          events={workspace.schedule}
          locale={locale}
          text={text}
          getDisplayName={workspace.getDisplayName}
          onAddEvent={actions.handleAddEvent}
          onDeleteEvent={(eventID) =>
            onConfirm({
              title: text.confirm.deleteScheduleTitle,
              message: text.confirm.deleteScheduleMessage,
              confirmText: text.common.delete,
              cancelText: text.common.cancel,
              variant: 'danger',
              onConfirm: () => actions.handleDeleteEvent(eventID),
            })
          }
          onShareEvent={(eventID) => void actions.handleShareEvent(eventID)}
        />
      )
    case 'files':
      return (
        <FilesPanel
          documents={workspace.documents}
          text={text}
          getDisplayName={workspace.getDisplayName}
          uploading={workspace.filesUploading}
          uploadProgress={workspace.uploadProgress}
          onUpload={actions.handleUploadDocument}
          onDelete={requestDeleteDocument}
          onShare={actions.handleShareDocument}
        />
      )
    case 'directory':
      return (
        <DirectoryPanel
          users={workspace.directoryUsers}
          departments={workspace.departments}
          query={workspace.directoryQuery}
          selectedDepartmentId={workspace.directoryDepartmentID}
          currentUserId={workspace.currentUser?.id ?? null}
          text={text}
          loading={workspace.directoryLoading}
          onQueryChange={workspace.setDirectoryQuery}
          onDepartmentChange={workspace.setDirectoryDepartmentID}
        />
      )
  }
}
