import { useMemo } from 'react'
import { usePreferencesStore } from './hooks/usePreferencesStore'
import type { Locale } from './types/workspace'

export const supportedLocales = ['en', 'zh-CN'] as const

export interface AppTranslations {
  common: {
    workpal: string
    cancel: string
    create: string
    search: string
    clear: string
    loading: string
    unavailable: string
    enabled: string
    disabled: string
    delete: string
    share: string
    save: string
    open: string
    upload: string
    add: string
    close: string
  }
  login: {
    title: string
    subtitle: string
    username: string
    password: string
    usernamePlaceholder: string
    passwordPlaceholder: string
    signIn: string
    signingIn: string
    switches: {
      language: string
      theme: string
    }
    seededAccounts: string
    seededAccountsHint: string
    useAccount: string
    helperTitle: string
    helperItems: string[]
  }
  navigation: Record<'overview' | 'chat' | 'tasks' | 'schedule' | 'files' | 'directory' | 'projects', string>
  shell: {
    subtitle: string
    welcome: string
    preferences: string
    signOut: string
    datePrefix: string
    liveData: string
    moduleSwitcher: string
    moduleSwitcherHint: string
    notifications: string
    noNotifications: string
    navGroups: Record<'overview' | 'collaboration' | 'work' | 'assets' | 'projects', string>
  }
  confirm: {
    confirmAction: string
    deleteTaskTitle: string
    deleteTaskMessage: string
    deleteScheduleTitle: string
    deleteScheduleMessage: string
    deleteFileTitle: string
    deleteFileMessage: string
    signOutTitle: string
    signOutMessage: string
  }
  validation: {
    titleRequired: string
    titleTooShort: string
    summaryTooLong: string
    ownerRequired: string
    futureStartRequired: string
  }
  settings: {
    title: string
    subtitle: string
    language: string
    languageHint: string
    theme: string
    themeHint: string
    light: string
    dark: string
    sound: string
    soundHint: string
    density: string
    densityHint: string
    comfortable: string
    compact: string
    reset: string
    close: string
  }
  overview: {
    title: string
    subtitle: string
    cards: {
      teammates: string
      activeTasks: string
      todayMeetings: string
      sharedFiles: string
    }
    sections: {
      priorities: string
      prioritiesHint: string
      agenda: string
      agendaHint: string
      docs: string
      docsHint: string
    }
    quickActions: string
    quickActionsHint: string
    openSection: string
    noUsers: string
  }
  tasks: {
    title: string
    subtitle: string
    addTask: string
    addHint: string
    titleLabel: string
    summaryLabel: string
    projectLabel: string
    priorityLabel: string
    createAction: string
    advance: string
    reset: string
    deleteAction: string
    shareAction: string
    empty: string
    sharedCount: string
    statuses: Record<'planned' | 'in_progress' | 'review' | 'done', string>
    priorities: Record<'high' | 'medium' | 'low', string>
    due: string
    owner: string
    teammates: string
  }
  schedule: {
    title: string
    subtitle: string
    addEvent: string
    addHint: string
    titleLabel: string
    detailLabel: string
    ownerLabel: string
    createAction: string
    deleteAction: string
    shareAction: string
    empty: string
    sharedCount: string
    starts: string
    duration: string
    room: string
    attendees: string
    minutes: string
    listView: string
    calendarView: string
  }
  files: {
    title: string
    subtitle: string
    uploadAction: string
    uploadHint: string
    owner: string
    updated: string
    attachmentLabel: string
    sharedCount: string
    deleteAction: string
    shareAction: string
    openAction: string
    previewAction: string
    previewTitle: string
    uploadProgress: string
    sourceSeed: string
    sourceUpload: string
    empty: string
    categories: Record<'draft' | 'review' | 'ready', string>
  }
  directory: {
    title: string
    subtitle: string
    searchPlaceholder: string
    searchHint: string
    allDepartments: string
    idLabel: string
    emailLabel: string
    phoneLabel: string
    roleLabel: string
    departmentLabel: string
    locationLabel: string
    focusLabel: string
    employeeNoLabel: string
    bioLabel: string
    noResults: string
    currentUser: string
  }
  chat: {
    title: string
    subtitle: string
    connectionOn: string
    connectionOff: string
    newConversation: string
    conversations: string
    noConversations: string
    selectConversation: string
    groupConversation: string
    directConversation: string
    groupChat: string
    directChatPrefix: string
    searchPlaceholder: string
    searchAction: string
    clearAction: string
    searching: string
    noSearchResults: string
    noMessages: string
    writeMessage: string
    send: string
    createTitle: string
    createSubtitle: string
    direct: string
    group: string
    directTarget: string
    directTargetHint: string
    groupName: string
    groupNamePlaceholder: string
    groupMembers: string
    groupMembersHint: string
    createAction: string
    creating: string
    invalidDirect: string
    invalidGroupName: string
    invalidGroup: string
    noTeamMembers: string
    memberCount: string
    announcementTitle: string
    announcementPlaceholder: string
    announcementSave: string
    announcementSaved: string
    groupFilesTitle: string
    uploadFile: string
    noFiles: string
    shareFile: string
    deleteFile: string
    uploadSuccess: string
    uploadFailure: string
    saveFailure: string
    editMessage: string
    recallMessage: string
    recallConfirmMessage: string
    deleteFileConfirm: string
  }
  projects: {
    title: string
    subtitle: string
    addProject: string
    addProjectHint: string
    projectKey: string
    projectName: string
    projectCategory: string
    createProject: string
    emptyProjects: string
    selectProject: string
    addIssue: string
    addIssueHint: string
    issueSummary: string
    issueDesc: string
    issuePriority: string
    issueType: string
    issueAssignee: string
    createIssue: string
    emptyIssues: string
    backlog: string
    inProgress: string
    inReview: string
    done: string
    deleteProject: string
    deleteIssue: string
    confirmDeleteProject: string
    confirmDeleteIssue: string
    viewChangelog: string
    noChangelogs: string
    priorities: Record<'Critical' | 'High' | 'Medium' | 'Low', string>
    issueTypes: Record<'Epic' | 'Story' | 'Task' | 'Sub-task' | 'Bug', string>
    sortBy: string
    summaryLabel: string
    keyLabel: string
  }
  status: {
    online: string
    away: string
  }
}

import { enTranslations } from './i18n/en'
import { zhCNTranslations } from './i18n/zh-CN'

export const translations: Record<Locale, AppTranslations> = {
  en: enTranslations,
  'zh-CN': zhCNTranslations,
}

export function useI18n() {
  const locale = usePreferencesStore((state) => state.locale)

  return useMemo(
    () => ({
      locale,
      t: translations[locale],
    }),
    [locale],
  )
}
