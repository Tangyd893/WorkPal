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
  navigation: Record<'overview' | 'chat' | 'tasks' | 'schedule' | 'files' | 'directory', string>
  shell: {
    subtitle: string
    welcome: string
    preferences: string
    signOut: string
    datePrefix: string
    liveData: string
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
    noUsers: string
  }
  tasks: {
    title: string
    subtitle: string
    advance: string
    reset: string
    statuses: Record<'planned' | 'in_progress' | 'review' | 'done', string>
    priorities: Record<'high' | 'medium' | 'low', string>
    due: string
    owner: string
    teammates: string
  }
  schedule: {
    title: string
    subtitle: string
    starts: string
    duration: string
    room: string
    attendees: string
    minutes: string
  }
  files: {
    title: string
    subtitle: string
    updated: string
    owner: string
    categories: Record<'draft' | 'review' | 'ready', string>
  }
  directory: {
    title: string
    subtitle: string
    searchPlaceholder: string
    idLabel: string
    emailLabel: string
    phoneLabel: string
    roleLabel: string
    departmentLabel: string
    locationLabel: string
    focusLabel: string
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
    invalidGroup: string
    noTeamMembers: string
    memberCount: string
  }
  status: {
    online: string
    away: string
  }
}

export const translations: Record<Locale, AppTranslations> = {
  en: {
    common: {
      workpal: 'WorkPal',
      cancel: 'Cancel',
      create: 'Create',
      search: 'Search',
      clear: 'Clear',
      loading: 'Loading...',
      unavailable: 'Unavailable',
      enabled: 'Enabled',
      disabled: 'Disabled',
    },
    login: {
      title: 'Sign in to WorkPal',
      subtitle: 'Use a seeded account for acceptance or sign in with your own workspace account.',
      username: 'Username',
      password: 'Password',
      usernamePlaceholder: 'Enter your username',
      passwordPlaceholder: 'Enter your password',
      signIn: 'Sign in',
      signingIn: 'Signing in...',
      switches: {
        language: 'Language',
        theme: 'Theme',
      },
      seededAccounts: 'Seeded test accounts',
      seededAccountsHint: 'These accounts are recreated on every backend start in development mode.',
      useAccount: 'Fill account',
      helperTitle: 'What you can validate after login',
      helperItems: [
        'Switch between Overview, Chat, Tasks, Schedule, Files, and Directory.',
        'Create direct chats and group rooms with seeded employees.',
        'Toggle language, light/dark theme, message sound, and density preferences.',
      ],
    },
    navigation: {
      overview: 'Overview',
      chat: 'Chat',
      tasks: 'Tasks',
      schedule: 'Schedule',
      files: 'Files',
      directory: 'Directory',
    },
    shell: {
      subtitle: 'Office collaboration workspace',
      welcome: 'Welcome back',
      preferences: 'Preferences',
      signOut: 'Sign out',
      datePrefix: 'Today',
      liveData: 'Live users from backend',
    },
    settings: {
      title: 'Workspace preferences',
      subtitle: 'Tune the interface for different testing and daily collaboration scenarios.',
      language: 'Language',
      languageHint: 'Switch UI copy between English and Simplified Chinese.',
      theme: 'Appearance',
      themeHint: 'Choose the page theme you want to validate.',
      light: 'Light',
      dark: 'Dark',
      sound: 'Message sound',
      soundHint: 'Play a short tone for incoming chat messages from teammates.',
      density: 'Density',
      densityHint: 'Reduce spacing when you want a denser operations layout.',
      comfortable: 'Comfortable',
      compact: 'Compact',
      reset: 'Reset preferences',
      close: 'Done',
    },
    overview: {
      title: 'Workspace overview',
      subtitle: 'A single place to scan team momentum before you drop into execution.',
      cards: {
        teammates: 'Teammates',
        activeTasks: 'Active tasks',
        todayMeetings: 'Today meetings',
        sharedFiles: 'Shared assets',
      },
      sections: {
        priorities: 'Priority work',
        prioritiesHint: 'Tasks that need movement across the team this week.',
        agenda: 'Today agenda',
        agendaHint: 'Shared meetings seeded for collaboration walkthroughs.',
        docs: 'Ready-to-share docs',
        docsHint: 'Artifacts that make the workspace feel more than just chat.',
      },
      noUsers: 'No teammate profiles were returned from the backend yet.',
    },
    tasks: {
      title: 'Task board',
      subtitle: 'A lightweight execution lane for planning, delivery, review, and completion.',
      advance: 'Move forward',
      reset: 'Reset',
      statuses: {
        planned: 'Planned',
        in_progress: 'In progress',
        review: 'Review',
        done: 'Done',
      },
      priorities: {
        high: 'High',
        medium: 'Medium',
        low: 'Low',
      },
      due: 'Due',
      owner: 'Owner',
      teammates: 'Collaborators',
    },
    schedule: {
      title: 'Schedule',
      subtitle: 'The day view keeps meetings and handoffs visible across the workspace.',
      starts: 'Starts',
      duration: 'Duration',
      room: 'Room',
      attendees: 'Attendees',
      minutes: 'min',
    },
    files: {
      title: 'Files and knowledge',
      subtitle: 'Shared documents, briefs, and operating references for the team.',
      updated: 'Updated',
      owner: 'Owner',
      categories: {
        draft: 'Draft',
        review: 'Review',
        ready: 'Ready',
      },
    },
    directory: {
      title: 'People directory',
      subtitle: 'Live user records from the backend, enriched with seeded team context for acceptance.',
      searchPlaceholder: 'Search by username, nickname, or email',
      idLabel: 'User ID',
      emailLabel: 'Email',
      phoneLabel: 'Phone',
      roleLabel: 'Role',
      departmentLabel: 'Department',
      locationLabel: 'Location',
      focusLabel: 'Focus',
      noResults: 'No matching teammates.',
      currentUser: 'You',
    },
    chat: {
      title: 'Team chat',
      subtitle: 'Messaging now lives inside a broader workspace instead of being the whole product.',
      connectionOn: 'Connected',
      connectionOff: 'Disconnected',
      newConversation: 'New conversation',
      conversations: 'conversations',
      noConversations: 'No conversations yet.',
      selectConversation: 'Select a conversation to start chatting.',
      groupConversation: 'Group conversation',
      directConversation: 'Direct conversation',
      groupChat: 'Group chat',
      directChatPrefix: 'Direct chat',
      searchPlaceholder: 'Search messages',
      searchAction: 'Search',
      clearAction: 'Clear',
      searching: 'Searching messages...',
      noSearchResults: 'No matching messages found.',
      noMessages: 'No messages yet.',
      writeMessage: 'Write a message',
      send: 'Send',
      createTitle: 'Create conversation',
      createSubtitle: 'Start a direct thread or launch a group room with seeded teammates.',
      direct: 'Direct',
      group: 'Group',
      directTarget: 'Choose teammate',
      directTargetHint: 'Pick one employee account to open a direct thread.',
      groupName: 'Group name',
      groupNamePlaceholder: 'Optional group name',
      groupMembers: 'Choose members',
      groupMembersHint: 'Select at least one teammate for the room.',
      createAction: 'Create',
      creating: 'Creating...',
      invalidDirect: 'Please choose a teammate.',
      invalidGroup: 'Please select at least one teammate.',
      noTeamMembers: 'No teammate accounts are available yet.',
      memberCount: 'members',
    },
    status: {
      online: 'Online',
      away: 'Focus',
    },
  },
  'zh-CN': {
    common: {
      workpal: 'WorkPal',
      cancel: '取消',
      create: '创建',
      search: '搜索',
      clear: '清除',
      loading: '加载中...',
      unavailable: '暂无',
      enabled: '开启',
      disabled: '关闭',
    },
    login: {
      title: '登录 WorkPal',
      subtitle: '可直接使用预置验收账号，或登录你自己的工作区账号。',
      username: '用户名',
      password: '密码',
      usernamePlaceholder: '请输入用户名',
      passwordPlaceholder: '请输入密码',
      signIn: '登录',
      signingIn: '登录中...',
      switches: {
        language: '语言',
        theme: '主题',
      },
      seededAccounts: '预置测试账号',
      seededAccountsHint: '开发模式下，后端每次启动都会重新确保这些账号可用。',
      useAccount: '填入账号',
      helperTitle: '登录后可验收的内容',
      helperItems: [
        '可在总览、聊天、任务、日程、文件、通讯录之间切换。',
        '可使用预置员工创建私聊和办公群组。',
        '可切换中英文、深浅色、消息提示音与界面密度设置。',
      ],
    },
    navigation: {
      overview: '总览',
      chat: '沟通',
      tasks: '任务',
      schedule: '日程',
      files: '文件',
      directory: '通讯录',
    },
    shell: {
      subtitle: '办公协作工作台',
      welcome: '欢迎回来',
      preferences: '偏好设置',
      signOut: '退出登录',
      datePrefix: '今天',
      liveData: '后端实时用户数据',
    },
    settings: {
      title: '工作台偏好设置',
      subtitle: '针对验收与日常协作，快速调整界面行为与视觉风格。',
      language: '语言',
      languageHint: '在 English 和 简体中文 之间切换。',
      theme: '界面风格',
      themeHint: '切换页面深浅色主题。',
      light: '浅色',
      dark: '深色',
      sound: '消息提示音',
      soundHint: '收到他人新消息时播放简短提示音。',
      density: '界面密度',
      densityHint: '在需要更高信息密度时压缩页面间距。',
      comfortable: '舒适',
      compact: '紧凑',
      reset: '恢复默认设置',
      close: '完成',
    },
    overview: {
      title: '工作台总览',
      subtitle: '先扫一眼团队状态，再进入具体执行模块。',
      cards: {
        teammates: '团队成员',
        activeTasks: '进行中任务',
        todayMeetings: '今日会议',
        sharedFiles: '共享资料',
      },
      sections: {
        priorities: '重点事项',
        prioritiesHint: '本周需要团队协同推进的任务。',
        agenda: '今日安排',
        agendaHint: '用于联调验收的共享会议与交接节奏。',
        docs: '可共享文档',
        docsHint: '让系统不止有聊天，还具备办公协作的文档感。',
      },
      noUsers: '后端暂未返回可展示的成员资料。',
    },
    tasks: {
      title: '任务看板',
      subtitle: '用轻量执行流覆盖规划、推进、评审与完成。',
      advance: '推进到下一列',
      reset: '重置',
      statuses: {
        planned: '待规划',
        in_progress: '进行中',
        review: '待评审',
        done: '已完成',
      },
      priorities: {
        high: '高',
        medium: '中',
        low: '低',
      },
      due: '截止',
      owner: '负责人',
      teammates: '协作者',
    },
    schedule: {
      title: '日程',
      subtitle: '把会议与交接放在工作台里，协作节奏才完整。',
      starts: '开始时间',
      duration: '时长',
      room: '会议地点',
      attendees: '参会人',
      minutes: '分钟',
    },
    files: {
      title: '文件与知识',
      subtitle: '共享文档、运营简报与团队知识都能在这里集中查看。',
      updated: '更新于',
      owner: '负责人',
      categories: {
        draft: '草稿',
        review: '评审中',
        ready: '可用',
      },
    },
    directory: {
      title: '通讯录',
      subtitle: '基于后端实时用户列表，并补足验收所需的团队角色信息。',
      searchPlaceholder: '按用户名、昵称或邮箱搜索',
      idLabel: '用户 ID',
      emailLabel: '邮箱',
      phoneLabel: '电话',
      roleLabel: '角色',
      departmentLabel: '部门',
      locationLabel: '地点',
      focusLabel: '当前关注',
      noResults: '没有匹配的成员。',
      currentUser: '当前账号',
    },
    chat: {
      title: '团队沟通',
      subtitle: '聊天被纳入完整工作台，而不是产品的全部。',
      connectionOn: '已连接',
      connectionOff: '未连接',
      newConversation: '新建会话',
      conversations: '个会话',
      noConversations: '还没有会话。',
      selectConversation: '选择一个会话开始沟通。',
      groupConversation: '群组会话',
      directConversation: '私聊会话',
      groupChat: '群聊',
      directChatPrefix: '私聊',
      searchPlaceholder: '搜索消息',
      searchAction: '搜索',
      clearAction: '清除',
      searching: '正在搜索消息...',
      noSearchResults: '没有找到匹配的消息。',
      noMessages: '暂时没有消息。',
      writeMessage: '输入消息',
      send: '发送',
      createTitle: '创建会话',
      createSubtitle: '可直接发起私聊，或用预置员工账号创建办公群组。',
      direct: '私聊',
      group: '群组',
      directTarget: '选择成员',
      directTargetHint: '选择一位员工，快速建立私聊。',
      groupName: '群组名称',
      groupNamePlaceholder: '可选的群组名称',
      groupMembers: '选择群成员',
      groupMembersHint: '至少选择一位成员加入群组。',
      createAction: '创建',
      creating: '创建中...',
      invalidDirect: '请选择一位成员。',
      invalidGroup: '请至少选择一位成员。',
      noTeamMembers: '暂时没有可用的成员账号。',
      memberCount: '名成员',
    },
    status: {
      online: '在线',
      away: '专注中',
    },
  },
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
