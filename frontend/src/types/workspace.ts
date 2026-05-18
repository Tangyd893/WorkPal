export type Locale = 'en' | 'zh-CN'
export type ThemeMode = 'light' | 'dark'
export type WorkspaceSection = 'overview' | 'chat' | 'tasks' | 'schedule' | 'files' | 'directory' | 'projects' | 'docs' | 'calendar' | 'approvals'
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

export interface WorkflowTransition {
  from: string
  to: string
  conditions?: WorkflowCondition[]
  validators?: WorkflowValidator[]
  post_functions?: WorkflowPostFunction[]
}

export interface WorkflowCondition {
  field: string
  operator: string
  value?: string
}

export interface WorkflowValidator {
  class: string
  args?: Record<string, unknown>
}

export interface WorkflowPostFunction {
  class: string
  args?: Record<string, unknown>
}

export interface WorkflowDSL {
  statuses: string[]
  transitions: WorkflowTransition[]
}

export interface Workflow {
  id: string
  project_id: number
  name: string
  description: string
  is_active: boolean
  dsl_definition: WorkflowDSL
  created_at: string
}

export interface CreateWorkflowInput {
  name?: string
  description?: string
  is_active?: boolean
  dsl_definition?: WorkflowDSL
}

export interface AvailableStatuses {
  current_status: string
  statuses: string[]
}

export interface Role {
  id: number
  name: string
  description: string
  is_system: boolean
}

export interface Permission {
  id: number
  code: string
  name: string
  description: string
  resource_type: string
}

export interface ProjectRole {
  id: number
  project_id: number
  name: string
  description: string
  is_system: boolean
}

export interface ProjectMember {
  user_id: number
  project_id: number
  role_id: number
  role_name: string
}

export interface AssignRoleInput {
  user_id: number
  role_id: number
  project_id?: number | null
}

export interface AddProjectMemberInput {
  user_id: number
  project_role_id: number
}

export interface CustomFieldDef {
  id: number
  project_id: number
  name: string
  field_type: string
  options: string[]
  is_required: boolean
  sort_order: number
}

export interface CreateCustomFieldInput {
  name: string
  field_type: string
  options?: string[]
  is_required?: boolean
  sort_order?: number
}

export interface CustomFieldValue {
  id: number
  issue_id: number
  field_id: number
  field_name: string
  field_type: string
  value_text: string
  value_number: number
  value_date: string | null
}

export interface UpsertCustomFieldValueInput {
  field_id: number
  value_text?: string
  value_number?: number
  value_date?: string | null
}

export interface Document {
  id: number
  project_id: number | null
  parent_id: number | null
  title: string
  created_by: number
  updated_by: number
  is_folder: boolean
  sort_order: number
  content?: string
  created_at: string
  updated_at: string
}

export interface CreateDocumentInput {
  project_id?: number | null
  parent_id?: number | null
  title: string
  is_folder?: boolean
  content?: string
}

export interface DocumentRevision {
  id: number
  document_id: number
  version: number
  content: string
  created_by: number
  created_at: string
}

export interface CalendarEvent {
  id: number
  project_id: number | null
  title: string
  description: string
  starts_at: string
  ends_at: string
  is_all_day: boolean
  location: string
  organizer_id: number
  attendees?: CalendarAttendee[]
  created_at: string
}

export interface CalendarAttendee {
  id: number
  user_id: number
  status: string
}

export interface CreateCalendarInput {
  project_id?: number | null
  title: string
  description?: string
  starts_at: string
  ends_at: string
  is_all_day?: boolean
  location?: string
  attendee_ids?: number[]
}

export interface ApprovalTemplate {
  id: number
  project_id: number | null
  name: string
  description: string
  form_schema: string
  flow_definition: string
  is_active: boolean
}

export interface ApprovalInstance {
  id: number
  template_id: number
  title: string
  form_data: string
  status: string
  submitter_id: number
  current_node_id: string
  submitted_at: string
  actions?: ApprovalAction[]
}

export interface ApprovalAction {
  id: number
  node_id: string
  action: string
  comment: string
  user_id: number
  created_at: string
}

export interface CreateApprovalInput {
  template_id: number
  title: string
  form_data?: string
}

export interface ApprovalActionInput {
  action: string
  comment?: string
}

export interface Notification {
  id: number
  type: string
  title: string
  content: string
  entity_type: string
  entity_id: string
  is_read: boolean
  created_at: string
}
