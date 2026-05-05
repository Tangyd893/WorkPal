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
  CreateScheduleInput,
  CreateTaskInput,
  Department,
  ScheduleEvent,
  TaskStatus,
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

  async listUsers(pageSize = 100, query = '', departmentId?: number): Promise<WorkspaceUser[]> {
    const data = await apiGet<WorkspaceUser[] | PageData<WorkspaceUser>>('/users', {
      params: {
        page: 1,
        page_size: pageSize,
        q: query || undefined,
        department_id: departmentId || undefined,
      },
    })
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

  uploadConversationFile(convID: number, file: File) {
    const form = new FormData()
    form.set('conv_id', String(convID))
    form.set('file', file)
    return apiPost<ConversationFile, FormData>('/files/upload', form)
  },

  uploadUserFile(file: File) {
    const form = new FormData()
    form.set('file', file)
    return apiPost<ConversationFile, FormData>('/files/upload', form)
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
}
