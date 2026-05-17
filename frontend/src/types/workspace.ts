export type Locale = 'en' | 'zh-CN'
export type ThemeMode = 'light' | 'dark'
export type WorkspaceSection = 'overview' | 'chat' | 'tasks' | 'schedule' | 'files' | 'directory' | 'projects'
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

// --- 项目空间类型 (WorkPal 2.0) ---

export type ProjectStatus = 'active' | 'archived'

export interface Project {
  id: string
  key: string
  name: string
  description: string
  lead_id: number
  icon: string
  category: string
  is_archived: boolean
}

export interface CreateProjectInput {
  key: string
  name: string
  description: string
  lead_id: number
  icon: string
  category: string
}

export interface Issue {
  id: string
  project_id: number
  issue_type_id: number
  issue_type_name: string
  parent_id: number | null
  key: string
  summary: string
  description: string
  status: string
  priority: string
  assignee_id: number | null
  reporter_id: number
  due_date: string | null
  story_points: number | null
  resolution: string
  version_ids: number[]
  fix_version_ids: number[]
  time_estimate: number
  time_spent: number
  sort_order: number
  created_at: string
  updated_at: string
  changelogs?: IssueChangelog[]
}

export interface CreateIssueInput {
  project_id: number
  issue_type_id: number
  parent_id: number | null
  summary: string
  description: string
  priority: string
  assignee_id: number | null
  reporter_id: number
  due_date: string | null
  story_points: number | null
  version_ids: number[]
  time_estimate: number
}

export interface IssueChangelog {
  id: number
  field: string
  old_value: string
  new_value: string
  changed_by: number
  created_at: string
}

export interface IssueType {
  id: number
  project_id: number
  name: string
  description: string
  icon_url: string
  hierarchy_level: number
  is_standard: boolean
}
