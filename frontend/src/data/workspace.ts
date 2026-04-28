import type {
  Locale,
  ScheduleEvent,
  SeedAccount,
  SharedDocument,
  TeamProfileMeta,
  WorkspaceTask,
} from '../types/workspace'

export const seededAccountsByLocale: Record<Locale, SeedAccount[]> = {
  en: [
    { username: 'admin', password: 'admin123', nickname: 'Administrator', note: 'Acceptance and management walkthroughs' },
    { username: 'emma.chen', password: 'workpal123', nickname: 'Emma Chen', note: 'Operations and launch coordination' },
    { username: 'liam.wang', password: 'workpal123', nickname: 'Liam Wang', note: 'Engineering delivery and backend checks' },
    { username: 'sofia.zhao', password: 'workpal123', nickname: 'Sofia Zhao', note: 'Design QA and release readiness' },
  ],
  'zh-CN': [
    { username: 'admin', password: 'admin123', nickname: 'Administrator', note: '用于验收和管理视角联调' },
    { username: 'emma.chen', password: 'workpal123', nickname: 'Emma Chen', note: '运营协同与上线推进' },
    { username: 'liam.wang', password: 'workpal123', nickname: 'Liam Wang', note: '工程交付与后端联调' },
    { username: 'sofia.zhao', password: 'workpal123', nickname: 'Sofia Zhao', note: '设计验收与发布准备' },
  ],
}

const taskCatalog: Record<Locale, WorkspaceTask[]> = {
  en: [
    {
      id: 'launch-readiness',
      title: 'Prepare release readiness checklist',
      summary: 'Consolidate QA, rollout notes, and support handoff before the internal launch review.',
      project: 'Ops Launch',
      ownerUsername: 'emma.chen',
      teammates: ['admin', 'sofia.zhao'],
      dueDate: '2026-04-29',
      priority: 'high',
      status: 'in_progress',
    },
    {
      id: 'search-hardening',
      title: 'Verify search fallback on seeded data',
      summary: 'Confirm message search still behaves well after account bootstrap and new workspace modules.',
      project: 'Platform',
      ownerUsername: 'liam.wang',
      teammates: ['admin'],
      dueDate: '2026-04-30',
      priority: 'high',
      status: 'review',
    },
    {
      id: 'workspace-copy',
      title: 'Polish bilingual workspace copy',
      summary: 'Review the English and Simplified Chinese interface text for acceptance clarity.',
      project: 'Experience',
      ownerUsername: 'sofia.zhao',
      teammates: ['emma.chen'],
      dueDate: '2026-05-01',
      priority: 'medium',
      status: 'planned',
    },
    {
      id: 'acceptance-script',
      title: 'Capture acceptance walkthrough',
      summary: 'Document the recommended path through overview, chat, tasks, schedule, files, and directory.',
      project: 'Enablement',
      ownerUsername: 'admin',
      teammates: ['emma.chen', 'liam.wang', 'sofia.zhao'],
      dueDate: '2026-04-30',
      priority: 'medium',
      status: 'done',
    },
  ],
  'zh-CN': [
    {
      id: 'launch-readiness',
      title: '整理上线验收清单',
      summary: '在内部上线评审前，汇总 QA、发布说明和支持交接事项。',
      project: '运营上线',
      ownerUsername: 'emma.chen',
      teammates: ['admin', 'sofia.zhao'],
      dueDate: '2026-04-29',
      priority: 'high',
      status: 'in_progress',
    },
    {
      id: 'search-hardening',
      title: '验证预置账号后的搜索回退',
      summary: '确认消息搜索在账号引导与新工作台模块接入后仍然稳定。',
      project: '平台能力',
      ownerUsername: 'liam.wang',
      teammates: ['admin'],
      dueDate: '2026-04-30',
      priority: 'high',
      status: 'review',
    },
    {
      id: 'workspace-copy',
      title: '润色中英文工作台文案',
      summary: '检查界面文本在 English / 简体中文 两种语言下都足够清晰。',
      project: '体验升级',
      ownerUsername: 'sofia.zhao',
      teammates: ['emma.chen'],
      dueDate: '2026-05-01',
      priority: 'medium',
      status: 'planned',
    },
    {
      id: 'acceptance-script',
      title: '补齐验收演示路径',
      summary: '整理总览、沟通、任务、日程、文件、通讯录的推荐验收顺序。',
      project: '交付支持',
      ownerUsername: 'admin',
      teammates: ['emma.chen', 'liam.wang', 'sofia.zhao'],
      dueDate: '2026-04-30',
      priority: 'medium',
      status: 'done',
    },
  ],
}

const scheduleCatalog: Record<Locale, ScheduleEvent[]> = {
  en: [
    {
      id: 'daily-sync',
      title: 'Workspace launch sync',
      detail: 'Review seeded accounts, chat flows, and acceptance checkpoints.',
      ownerUsername: 'admin',
      startsAt: '2026-04-28T09:30:00+08:00',
      durationMinutes: 30,
      attendees: ['admin', 'emma.chen', 'liam.wang', 'sofia.zhao'],
      room: 'War Room A',
    },
    {
      id: 'design-review',
      title: 'UI fit-and-finish review',
      detail: 'Validate bilingual copy, density, and dark theme coverage.',
      ownerUsername: 'sofia.zhao',
      startsAt: '2026-04-28T14:00:00+08:00',
      durationMinutes: 45,
      attendees: ['admin', 'sofia.zhao', 'emma.chen'],
      room: 'Design Bay',
    },
    {
      id: 'retro',
      title: 'Delivery retrospective',
      detail: 'Collect follow-up items for backend-backed modules and future iterations.',
      ownerUsername: 'liam.wang',
      startsAt: '2026-04-28T17:30:00+08:00',
      durationMinutes: 30,
      attendees: ['admin', 'emma.chen', 'liam.wang', 'sofia.zhao'],
      room: 'Focus Pod',
    },
  ],
  'zh-CN': [
    {
      id: 'daily-sync',
      title: '工作台上线同步会',
      detail: '对齐预置账号、聊天流程与验收检查点。',
      ownerUsername: 'admin',
      startsAt: '2026-04-28T09:30:00+08:00',
      durationMinutes: 30,
      attendees: ['admin', 'emma.chen', 'liam.wang', 'sofia.zhao'],
      room: 'A 作战室',
    },
    {
      id: 'design-review',
      title: '界面收口评审',
      detail: '检查双语文案、界面密度和深色主题覆盖。',
      ownerUsername: 'sofia.zhao',
      startsAt: '2026-04-28T14:00:00+08:00',
      durationMinutes: 45,
      attendees: ['admin', 'sofia.zhao', 'emma.chen'],
      room: '设计区',
    },
    {
      id: 'retro',
      title: '交付复盘',
      detail: '沉淀后续需要后端化的模块与下一轮迭代项。',
      ownerUsername: 'liam.wang',
      startsAt: '2026-04-28T17:30:00+08:00',
      durationMinutes: 30,
      attendees: ['admin', 'emma.chen', 'liam.wang', 'sofia.zhao'],
      room: '专注舱',
    },
  ],
}

const documentCatalog: Record<Locale, SharedDocument[]> = {
  en: [
    {
      id: 'ops-runbook',
      title: 'Operations rollout runbook',
      summary: 'Step-by-step checklist for release support, risk logging, and escalation routing.',
      category: 'Operations',
      ownerUsername: 'emma.chen',
      updatedAt: '2026-04-28T08:20:00+08:00',
      status: 'ready',
    },
    {
      id: 'qa-brief',
      title: 'Cross-module QA brief',
      summary: 'Expected behaviors for chat, directory, settings, and workspace navigation.',
      category: 'QA',
      ownerUsername: 'sofia.zhao',
      updatedAt: '2026-04-27T18:40:00+08:00',
      status: 'review',
    },
    {
      id: 'platform-notes',
      title: 'Platform verification notes',
      summary: 'Smoke results for API login, seeded employees, and runtime validation.',
      category: 'Engineering',
      ownerUsername: 'liam.wang',
      updatedAt: '2026-04-28T10:05:00+08:00',
      status: 'draft',
    },
  ],
  'zh-CN': [
    {
      id: 'ops-runbook',
      title: '运营上线手册',
      summary: '覆盖发布支持、风险登记与升级路径的逐步清单。',
      category: '运营',
      ownerUsername: 'emma.chen',
      updatedAt: '2026-04-28T08:20:00+08:00',
      status: 'ready',
    },
    {
      id: 'qa-brief',
      title: '跨模块验收说明',
      summary: '整理聊天、通讯录、设置和导航模块的预期行为。',
      category: '测试',
      ownerUsername: 'sofia.zhao',
      updatedAt: '2026-04-27T18:40:00+08:00',
      status: 'review',
    },
    {
      id: 'platform-notes',
      title: '平台联调记录',
      summary: '记录 API 登录、预置员工账号与运行时验证结果。',
      category: '工程',
      ownerUsername: 'liam.wang',
      updatedAt: '2026-04-28T10:05:00+08:00',
      status: 'draft',
    },
  ],
}

const teamMetaCatalog: Record<Locale, Record<string, TeamProfileMeta>> = {
  en: {
    admin: {
      role: 'Workspace owner',
      department: 'Program office',
      location: 'Shanghai',
      focus: 'Acceptance walkthrough and release confidence',
    },
    'emma.chen': {
      role: 'Operations lead',
      department: 'Operations',
      location: 'Shanghai',
      focus: 'Launch sequencing and stakeholder updates',
    },
    'liam.wang': {
      role: 'Platform engineer',
      department: 'Engineering',
      location: 'Hangzhou',
      focus: 'API stability and runtime verification',
    },
    'sofia.zhao': {
      role: 'Product designer',
      department: 'Design',
      location: 'Shenzhen',
      focus: 'Bilingual polish and visual QA',
    },
  },
  'zh-CN': {
    admin: {
      role: '工作台负责人',
      department: '项目办公室',
      location: '上海',
      focus: '验收演示与发布把关',
    },
    'emma.chen': {
      role: '运营负责人',
      department: '运营部',
      location: '上海',
      focus: '上线节奏与干系人同步',
    },
    'liam.wang': {
      role: '平台工程师',
      department: '工程部',
      location: '杭州',
      focus: 'API 稳定性与运行验证',
    },
    'sofia.zhao': {
      role: '产品设计师',
      department: '设计部',
      location: '深圳',
      focus: '双语细节与视觉验收',
    },
  },
}

export function buildSeedTasks(locale: Locale): WorkspaceTask[] {
  return taskCatalog[locale].map((task) => ({ ...task, teammates: [...task.teammates] }))
}

export function buildSchedule(locale: Locale): ScheduleEvent[] {
  return scheduleCatalog[locale].map((event) => ({ ...event, attendees: [...event.attendees] }))
}

export function buildDocuments(locale: Locale): SharedDocument[] {
  return documentCatalog[locale].map((document) => ({ ...document }))
}

export function getSeededAccounts(locale: Locale): SeedAccount[] {
  return seededAccountsByLocale[locale]
}

export function getTeamProfileMeta(locale: Locale, username: string): TeamProfileMeta {
  return (
    teamMetaCatalog[locale][username] ?? {
      role: locale === 'en' ? 'Team member' : '团队成员',
      department: locale === 'en' ? 'General' : '通用',
      location: locale === 'en' ? 'Remote' : '远程',
      focus: locale === 'en' ? 'General collaboration' : '通用协作',
    }
  )
}
