export type Locale = 'en' | 'zh-CN'
export type ThemeMode = 'light' | 'dark'
export type WorkspaceSection = 'overview' | 'chat' | 'tasks' | 'schedule' | 'files' | 'directory'
export type TaskStatus = 'planned' | 'in_progress' | 'review' | 'done'
export type TaskPriority = 'high' | 'medium' | 'low'
export type DocumentStatus = 'draft' | 'review' | 'ready'

export interface Department {
  id: number
  code: string
  name: string
  description: string
  parent_id: number
  leader_id: number
  created_at: string
  updated_at: string
}

export interface WorkspaceUser {
  id: number
  username: string
  nickname: string
  avatar_url?: string
  email: string
  phone?: string
  status: number
  department_id: number
  department_name: string
  employee_id: number
  employee_no: string
  job_title: string
  office_location: string
  bio: string
  created_at: string
  updated_at: string
}

export interface WorkspaceTask {
  id: string
  title: string
  summary: string
  project: string
  ownerUsername: string
  teammates: string[]
  dueDate: string
  priority: TaskPriority
  status: TaskStatus
  sharedCount: number
  source: 'seed' | 'custom'
}

export interface CreateTaskInput {
  title: string
  summary: string
  project: string
  ownerUsername: string
  teammates: string[]
  dueDate: string
  priority: TaskPriority
}

export interface ScheduleEvent {
  id: string
  title: string
  detail: string
  ownerUsername: string
  startsAt: string
  durationMinutes: number
  attendees: string[]
  room: string
  sharedCount: number
  source: 'seed' | 'custom'
}

export interface CreateScheduleInput {
  title: string
  detail: string
  ownerUsername: string
  startsAt: string
  durationMinutes: number
  attendees: string[]
  room: string
}

export interface SharedDocument {
  id: string
  title: string
  summary: string
  category: string
  ownerUsername: string
  updatedAt: string
  status: DocumentStatus
  sharedCount: number
  source: 'seed' | 'custom'
  fileId?: number
  attachmentName?: string
  attachmentUrl?: string
  downloadPath?: string
}

export interface SeedAccount {
  username: string
  password: string
  nickname: string
  note: string
}

export interface ConversationFile {
  id: number
  user_id: number
  conv_id: number
  name: string
  size: number
  content_type: string
  mime_type: string
  created_at: string
  download_path: string
  share_path: string
  download_url: string
}
