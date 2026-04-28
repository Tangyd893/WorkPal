export type Locale = 'en' | 'zh-CN'
export type ThemeMode = 'light' | 'dark'
export type WorkspaceSection = 'overview' | 'chat' | 'tasks' | 'schedule' | 'files' | 'directory'
export type TaskStatus = 'planned' | 'in_progress' | 'review' | 'done'
export type TaskPriority = 'high' | 'medium' | 'low'
export type DocumentStatus = 'draft' | 'review' | 'ready'

export interface WorkspaceUser {
  id: number
  username: string
  nickname: string
  avatar_url?: string
  email: string
  phone?: string
  status: number
  department_id: number
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
}

export interface SharedDocument {
  id: string
  title: string
  summary: string
  category: string
  ownerUsername: string
  updatedAt: string
  status: DocumentStatus
}

export interface TeamProfileMeta {
  role: string
  department: string
  location: string
  focus: string
}

export interface SeedAccount {
  username: string
  password: string
  nickname: string
  note: string
}
