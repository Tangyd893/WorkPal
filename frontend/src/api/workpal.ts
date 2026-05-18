import { apiDelete, apiGet, apiPost, apiPut } from './request'
import type {
  ChatMessage,
  Conversation,
  CreateConversationDraft,
  LoginResponse,
  PageData,
  SearchResult,
} from '../types/chat'
import type {
  ConversationFile,
  CreateIssueInput,
  CreateProjectInput,
  CreateScheduleInput,
  CreateTaskInput,
  Department,
  Issue,
  IssueChangelog,
  IssueType,
  Project,
  ScheduleEvent,
  TaskStatus,
  Workflow,
  CreateWorkflowInput,
  AvailableStatuses,
  Role,
  Permission,
  ProjectRole,
  ProjectMember,
  AssignRoleInput,
  AddProjectMemberInput,
  CustomFieldDef,
  CreateCustomFieldInput,
  CustomFieldValue,
  UpsertCustomFieldValueInput,
  Document,
  CreateDocumentInput,
  DocumentRevision,
  CalendarEvent,
  CreateCalendarInput,
  ApprovalTemplate,
  ApprovalInstance,
  CreateApprovalInput,
  ApprovalActionInput,
  WorkspaceTask,
  WorkspaceUser,
} from '../types/workspace'

interface LoginPayload {
  username: string
  password: string
}

interface SendMessagePayload {
  type: number
  content: string
  idempotency_key: string
}

interface UpdateConversationPayload {
  name?: string
  announcement?: string
}

interface EditMessagePayload {
  content: string
}

interface AddMemberPayload {
  user_id: number
}

type CreateConversationPayload =
  | {
      type: 1
      target_uid: number
    }
  | {
      type: 2
      name: string
      member_ids: number[]
    }

function unwrapPageData<T>(value: T[] | PageData<T>): T[] {
  return Array.isArray(value) ? value : value.items
}

export const workpalApi = {
  login(payload: LoginPayload) {
    return apiPost<LoginResponse, LoginPayload>('/auth/login', payload)
  },

  getMe() {
    return apiGet<WorkspaceUser>('/users/me')
  },

  async listUsers(pageSize = 100, query = '', departmentId?: number, signal?: AbortSignal): Promise<WorkspaceUser[]> {
    const config = {
      params: {
        page: 1,
        page_size: pageSize,
        q: query || undefined,
        department_id: departmentId || undefined,
      },
    }
    const data = await apiGet<WorkspaceUser[] | PageData<WorkspaceUser>>(
      '/users',
      signal ? { ...config, signal } : config,
    )
    return unwrapPageData(data)
  },

  listDepartments() {
    return apiGet<Department[]>('/departments')
  },

  listUserFiles() {
    return apiGet<ConversationFile[]>('/files')
  },

  async listConversations(): Promise<Conversation[]> {
    const data = await apiGet<Conversation[] | PageData<Conversation>>('/conversations')
    return unwrapPageData(data)
  },

  getConversationMessages(convID: number) {
    return apiGet<ChatMessage[]>(`/conversations/${convID}/messages`)
  },

  sendMessage(convID: number, content: string) {
    const idempotencyKey =
      typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
        ? crypto.randomUUID()
        : `idem-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`
    const payload: SendMessagePayload = { type: 1, content, idempotency_key: idempotencyKey }
    return apiPost<ChatMessage, SendMessagePayload>(`/conversations/${convID}/messages`, payload)
  },

  searchMessages(query: string, convID?: number, page = 1, pageSize = 20) {
    return apiGet<SearchResult>('/search/messages', {
      params: {
        q: query,
        conv_id: convID,
        page,
        page_size: pageSize,
      },
    })
  },

  createConversation(draft: CreateConversationDraft) {
    const payload: CreateConversationPayload =
      draft.mode === 'private'
        ? {
            type: 1,
            target_uid: draft.targetUserId,
          }
        : {
            type: 2,
            name: draft.name,
            member_ids: draft.memberIds,
          }

    return apiPost<Conversation, CreateConversationPayload>('/conversations', payload)
  },

  updateConversation(convID: number, payload: UpdateConversationPayload) {
    return apiPut<Conversation, UpdateConversationPayload>(`/conversations/${convID}`, payload)
  },

  updateConversationAnnouncement(convID: number, announcement: string) {
    return apiPut<Conversation, Pick<UpdateConversationPayload, 'announcement'>>(`/conversations/${convID}/announcement`, {
      announcement,
    })
  },

  listConversationFiles(convID: number) {
    return apiGet<ConversationFile[]>(`/conversations/${convID}/files`)
  },

  uploadConversationFile(convID: number, file: File, onProgress?: (progress: number) => void) {
    const form = new FormData()
    form.set('conv_id', String(convID))
    form.set('file', file)
    return apiPost<ConversationFile, FormData>('/files/upload', form, {
      onUploadProgress: (event) => {
        if (!onProgress || !event.total) {
          return
        }

        onProgress(Math.round((event.loaded / event.total) * 100))
      },
    })
  },

  uploadUserFile(file: File, onProgress?: (progress: number) => void) {
    const form = new FormData()
    form.set('file', file)
    return apiPost<ConversationFile, FormData>('/files/upload', form, {
      onUploadProgress: (event) => {
        if (!onProgress || !event.total) {
          return
        }

        onProgress(Math.round((event.loaded / event.total) * 100))
      },
    })
  },

  deleteFile(fileID: number) {
    return apiDelete<ConversationFile>(`/files/${fileID}`)
  },

  shareFile(fileID: number) {
    return apiPost<{ file_id: number; name: string; download_path: string; share_text: string }, undefined>(
      `/files/${fileID}/share`,
    )
  },

  listTasks() {
    return apiGet<WorkspaceTask[]>('/tasks')
  },

  createTask(payload: CreateTaskInput) {
    return apiPost<WorkspaceTask, CreateTaskInput>('/tasks', payload)
  },

  updateTaskStatus(taskID: string, status: TaskStatus) {
    return apiPut<WorkspaceTask, { status: TaskStatus }>(`/tasks/${taskID}/status`, { status })
  },

  deleteTask(taskID: string) {
    return apiDelete<null>(`/tasks/${taskID}`)
  },

  shareTask(taskID: string) {
    return apiPost<WorkspaceTask, undefined>(`/tasks/${taskID}/share`)
  },

  listSchedule() {
    return apiGet<ScheduleEvent[]>('/schedule')
  },

  createScheduleEvent(payload: CreateScheduleInput) {
    return apiPost<ScheduleEvent, CreateScheduleInput>('/schedule', payload)
  },

  deleteScheduleEvent(eventID: string) {
    return apiDelete<null>(`/schedule/${eventID}`)
  },

  shareScheduleEvent(eventID: string) {
    return apiPost<ScheduleEvent, undefined>(`/schedule/${eventID}/share`)
  },

  editMessage(messageID: number, content: string) {
    return apiPut<ChatMessage, EditMessagePayload>(`/messages/${messageID}`, { content })
  },

  recallMessage(messageID: number) {
    return apiDelete<{ id: number; recalled: boolean }>(`/messages/${messageID}`)
  },

  markRead(messageID: number) {
    return apiPost<{ message_id: number; read: boolean }, undefined>(`/messages/${messageID}/read`)
  },

  markAllRead(convID: number) {
    return apiPost<{ conv_id: number; read: boolean }, undefined>(`/conversations/${convID}/read-all`)
  },

  getConversation(convID: number) {
    return apiGet<Conversation>(`/conversations/${convID}`)
  },

  deleteConversation(convID: number) {
    return apiDelete<{ id: number; deleted: boolean }>(`/conversations/${convID}`)
  },

  addConversationMember(convID: number, userId: number) {
    return apiPost<Conversation, AddMemberPayload>(`/conversations/${convID}/members`, { user_id: userId })
  },

  removeConversationMember(convID: number, userId: number) {
    return apiDelete<Conversation>(`/conversations/${convID}/members/${userId}`)
  },

  // --- 项目空间 API ---
  listProjects() {
    return apiGet<Project[]>('/projects')
  },
  createProject(payload: CreateProjectInput) {
    return apiPost<Project, CreateProjectInput>('/projects', payload)
  },
  getProject(projectID: string) {
    return apiGet<Project>(`/projects/${projectID}`)
  },
  deleteProject(projectID: string) {
    return apiDelete<null>(`/projects/${projectID}`)
  },
  listIssues(projectID: string) {
    return apiGet<Issue[]>(`/projects/${projectID}/issues`)
  },
  createIssue(projectID: string, payload: CreateIssueInput) {
    return apiPost<Issue, CreateIssueInput>(`/projects/${projectID}/issues`, payload)
  },
  getIssue(issueID: string) {
    return apiGet<Issue>(`/issues/${issueID}`)
  },
  updateIssue(issueID: string, payload: Partial<CreateIssueInput>) {
    return apiPut<Issue, Partial<CreateIssueInput>>(`/issues/${issueID}`, payload)
  },
  updateIssueStatus(issueID: string, status: string) {
    return apiPut<Issue, { status: string }>(`/issues/${issueID}/status`, { status })
  },
  deleteIssue(issueID: string) {
    return apiDelete<null>(`/issues/${issueID}`)
  },
  listChangelogs(issueID: string) {
    return apiGet<IssueChangelog[]>(`/issues/${issueID}/changelogs`)
  },
  listIssueTypes(projectID: string) {
    return apiGet<IssueType[]>(`/projects/${projectID}/issue-types`)
  },
  listWorkflows(projectID: string) {
    return apiGet<Workflow[]>(`/projects/${projectID}/workflows`)
  },
  createWorkflow(projectID: string, payload: CreateWorkflowInput) {
    return apiPost<Workflow, CreateWorkflowInput>(`/projects/${projectID}/workflows`, payload)
  },
  getWorkflow(workflowID: string) {
    return apiGet<Workflow>(`/workflows/${workflowID}`)
  },
  updateWorkflow(workflowID: string, payload: Partial<CreateWorkflowInput>) {
    return apiPut<Workflow, Partial<CreateWorkflowInput>>(`/workflows/${workflowID}`, payload)
  },
  deleteWorkflow(workflowID: string) {
    return apiDelete<null>(`/workflows/${workflowID}`)
  },
  getAvailableStatuses(issueID: string) {
    return apiGet<AvailableStatuses>(`/issues/${issueID}/available-statuses`)
  },
  listRoles() {
    return apiGet<Role[]>('/roles')
  },
  listPermissions() {
    return apiGet<Permission[]>('/permissions')
  },
  assignRole(payload: AssignRoleInput) {
    return apiPost<null, AssignRoleInput>('/user-roles', payload)
  },
  removeRole(payload: AssignRoleInput) {
    return apiDelete<null>('/user-roles', { data: payload })
  },
  getUserPermissions(userID: number) {
    return apiGet<string[]>(`/users/${userID}/permissions`)
  },
  listProjectRoles(projectID: string) {
    return apiGet<ProjectRole[]>(`/projects/${projectID}/roles`)
  },
  createProjectRole(projectID: string, payload: { name: string; description: string }) {
    return apiPost<ProjectRole, { name: string; description: string }>(`/projects/${projectID}/roles`, payload)
  },
  addProjectMember(projectID: string, payload: AddProjectMemberInput) {
    return apiPost<null, AddProjectMemberInput>(`/projects/${projectID}/members`, payload)
  },
  removeProjectMember(projectID: string, userID: number) {
    return apiDelete<null>(`/projects/${projectID}/members/${userID}`)
  },
  listProjectMembers(projectID: string) {
    return apiGet<ProjectMember[]>(`/projects/${projectID}/members`)
  },
  listCustomFieldDefs(projectID: string) {
    return apiGet<CustomFieldDef[]>(`/projects/${projectID}/custom-fields`)
  },
  createCustomFieldDef(projectID: string, payload: CreateCustomFieldInput) {
    return apiPost<CustomFieldDef, CreateCustomFieldInput>(`/projects/${projectID}/custom-fields`, payload)
  },
  updateCustomFieldDef(fieldID: string, payload: Partial<CreateCustomFieldInput>) {
    return apiPut<CustomFieldDef, Partial<CreateCustomFieldInput>>(`/custom-fields/${fieldID}`, payload)
  },
  deleteCustomFieldDef(fieldID: string) {
    return apiDelete<null>(`/custom-fields/${fieldID}`)
  },
  getIssueCustomFieldValues(issueID: string) {
    return apiGet<CustomFieldValue[]>(`/issues/${issueID}/custom-fields`)
  },
  upsertCustomFieldValue(issueID: string, payload: UpsertCustomFieldValueInput) {
    return apiPut<null, UpsertCustomFieldValueInput>(`/issues/${issueID}/custom-fields`, payload)
  },
  listDocuments(params?: { project_id?: number; parent_id?: number }) {
    return apiGet<Document[]>('/documents', { params })
  },
  getDocument(docID: number) {
    return apiGet<Document>(`/documents/${docID}`)
  },
  createDocument(payload: CreateDocumentInput) {
    return apiPost<Document, CreateDocumentInput>('/documents', payload)
  },
  updateDocument(docID: number, payload: Partial<CreateDocumentInput>) {
    return apiPut<Document, Partial<CreateDocumentInput>>(`/documents/${docID}`, payload)
  },
  deleteDocument(docID: number) {
    return apiDelete<null>(`/documents/${docID}`)
  },
  listDocumentRevisions(docID: number) {
    return apiGet<DocumentRevision[]>(`/documents/${docID}/revisions`)
  },
  listCalendarEvents(params?: { from?: string; to?: string; organizer_id?: number }) {
    return apiGet<CalendarEvent[]>('/calendar', { params })
  },
  createCalendarEvent(payload: CreateCalendarInput) {
    return apiPost<CalendarEvent, CreateCalendarInput>('/calendar', payload)
  },
  updateCalendarEvent(eventID: number, payload: Partial<CreateCalendarInput>) {
    return apiPut<CalendarEvent, Partial<CreateCalendarInput>>(`/calendar/${eventID}`, payload)
  },
  deleteCalendarEvent(eventID: number) {
    return apiDelete<null>(`/calendar/${eventID}`)
  },
  listApprovalTemplates(projectID?: number) {
    return apiGet<ApprovalTemplate[]>('/approvals/templates', { params: { project_id: projectID } })
  },
  createApprovalTemplate(payload: { name: string; description?: string; form_schema?: string; flow_definition?: string; project_id?: number }) {
    return apiPost<ApprovalTemplate, typeof payload>('/approvals/templates', payload)
  },
  listApprovalInstances(params?: { submitter_id?: number; status?: string }) {
    return apiGet<ApprovalInstance[]>('/approvals/instances', { params })
  },
  getApprovalInstance(id: number) {
    return apiGet<ApprovalInstance>(`/approvals/instances/${id}`)
  },
  createApprovalInstance(payload: CreateApprovalInput) {
    return apiPost<ApprovalInstance, CreateApprovalInput>('/approvals/instances', payload)
  },
  processApprovalAction(id: number, payload: ApprovalActionInput) {
    return apiPost<ApprovalInstance, ApprovalActionInput>(`/approvals/instances/${id}/action`, payload)
  },
}
