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
      delete: 'Delete',
      share: 'Share',
      save: 'Save',
      open: 'Open',
      upload: 'Upload',
      add: 'Add',
      close: 'Close',
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
        'Toggle language, light and dark theme, message sound, and density preferences.',
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
      liveData: 'Live workspace data backed by the API',
    },
    settings: {
      title: 'Workspace preferences',
      subtitle: 'Tune the interface for testing, review, and everyday collaboration.',
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
      subtitle: 'A single place to scan team momentum before dropping into execution.',
      cards: {
        teammates: 'Teammates',
        activeTasks: 'Active tasks',
        todayMeetings: 'Today meetings',
        sharedFiles: 'Shared files',
      },
      sections: {
        priorities: 'Priority work',
        prioritiesHint: 'Tasks that need movement across the team this week.',
        agenda: 'Today agenda',
        agendaHint: 'Meetings and handoffs ready for execution.',
        docs: 'Shared assets',
        docsHint: 'Files, briefs, and references people can act on.',
      },
      quickActions: 'Quick actions',
      quickActionsHint: 'Every card and section can jump into a fuller workspace module.',
      openSection: 'Open module',
      noUsers: 'No teammate profiles were returned from the backend yet.',
    },
    tasks: {
      title: 'Task board',
      subtitle: 'Create, move, share, and clear work items without leaving the workspace.',
      addTask: 'Add task',
      addHint: 'Use this lightweight board for planning, execution, review, and completion.',
      titleLabel: 'Task title',
      summaryLabel: 'Summary',
      projectLabel: 'Project',
      priorityLabel: 'Priority',
      createAction: 'Create task',
      advance: 'Move forward',
      reset: 'Reset',
      deleteAction: 'Delete',
      shareAction: 'Share',
      empty: 'No tasks in this lane yet.',
      sharedCount: 'Shares',
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
      subtitle: 'Add, share, and trim events so the day view keeps handoffs visible.',
      addEvent: 'Add event',
      addHint: 'Seeded meetings cover acceptance flows, and you can create more for walkthroughs.',
      titleLabel: 'Event title',
      detailLabel: 'Details',
      ownerLabel: 'Host',
      createAction: 'Create event',
      deleteAction: 'Delete',
      shareAction: 'Share',
      empty: 'No events scheduled yet.',
      sharedCount: 'Shares',
      starts: 'Starts',
      duration: 'Duration',
      room: 'Room',
      attendees: 'Attendees',
      minutes: 'min',
    },
    files: {
      title: 'Files and knowledge',
      subtitle: 'Upload, open, share, and remove documents from the workspace library.',
      uploadAction: 'Upload file',
      uploadHint: 'All items in this library come from the backend file service.',
      owner: 'Owner',
      updated: 'Updated',
      attachmentLabel: 'Attachment',
      sharedCount: 'Shares',
      deleteAction: 'Delete',
      shareAction: 'Share',
      openAction: 'Open',
      sourceSeed: 'Seeded',
      sourceUpload: 'Uploaded',
      empty: 'No documents are available yet.',
      categories: {
        draft: 'Draft',
        review: 'Review',
        ready: 'Ready',
      },
    },
    directory: {
      title: 'People directory',
      subtitle: 'Search live employee records by name, phone, title, and department.',
      searchPlaceholder: 'Search by name, phone, title, employee number, or department',
      searchHint: 'The department filter narrows backend results before the free-text search applies.',
      allDepartments: 'All departments',
      idLabel: 'User ID',
      emailLabel: 'Email',
      phoneLabel: 'Phone',
      roleLabel: 'Title',
      departmentLabel: 'Department',
      locationLabel: 'Location',
      focusLabel: 'Bio',
      employeeNoLabel: 'Employee No.',
      bioLabel: 'Notes',
      noResults: 'No matching teammates.',
      currentUser: 'You',
    },
    chat: {
      title: 'Team chat',
      subtitle: 'Direct chats and group rooms now sit inside a broader collaboration workspace.',
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
      announcementTitle: 'Announcement',
      announcementPlaceholder: 'Write a group update, notice, or collaboration rule.',
      announcementSave: 'Save announcement',
      announcementSaved: 'Announcement saved.',
      groupFilesTitle: 'Group files',
      uploadFile: 'Upload to group',
      noFiles: 'No files in this group yet.',
      shareFile: 'Share link',
      deleteFile: 'Delete file',
      uploadSuccess: 'Group file uploaded.',
      uploadFailure: 'Unable to upload the group file.',
      saveFailure: 'Unable to update the group announcement.',
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
      delete: '删除',
      share: '分享',
      save: '保存',
      open: '打开',
      upload: '上传',
      add: '新增',
      close: '关闭',
    },
    login: {
      title: '登录 WorkPal',
      subtitle: '可直接使用预置验收账号，也可以登录你自己的工作区账号。',
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
        '可以在总览、沟通、任务、日程、文件、通讯录之间切换。',
        '可以使用预置员工创建私聊和群组会话。',
        '可以切换中英文、深浅色、消息提示音与界面密度设置。',
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
      liveData: '当前页面数据由后端 API 实时驱动',
    },
    settings: {
      title: '工作台偏好设置',
      subtitle: '针对验收、评审和日常协作，快速调整界面行为与视觉风格。',
      language: '语言',
      languageHint: '在 English 与 简体中文 之间切换界面文案。',
      theme: '界面风格',
      themeHint: '切换页面深色或浅色主题。',
      light: '浅色',
      dark: '深色',
      sound: '消息提示音',
      soundHint: '收到其他成员新消息时播放简短提示音。',
      density: '界面密度',
      densityHint: '需要更高信息密度时可压缩布局间距。',
      comfortable: '舒适',
      compact: '紧凑',
      reset: '重置偏好',
      close: '完成',
    },
    overview: {
      title: '工作台总览',
      subtitle: '先看清团队当下在推进什么，再进入具体执行模块。',
      cards: {
        teammates: '团队成员',
        activeTasks: '进行中任务',
        todayMeetings: '今日会议',
        sharedFiles: '共享文件',
      },
      sections: {
        priorities: '重点工作',
        prioritiesHint: '本周需要多人协同推进的事项。',
        agenda: '今日安排',
        agendaHint: '当前可以直接进入执行的会议与交接事项。',
        docs: '共享资料',
        docsHint: '团队可以直接打开、协作和传阅的文件与说明。',
      },
      quickActions: '快捷入口',
      quickActionsHint: '每张卡片和每个区块都可以跳转到对应模块。',
      openSection: '进入模块',
      noUsers: '后端暂时没有返回任何员工档案。',
    },
    tasks: {
      title: '任务看板',
      subtitle: '可以直接在工作台里新增、推进、分享和清理任务。',
      addTask: '新增任务',
      addHint: '适合放计划、执行、评审和完成这几类基础事项。',
      titleLabel: '任务标题',
      summaryLabel: '任务说明',
      projectLabel: '所属项目',
      priorityLabel: '优先级',
      createAction: '创建任务',
      advance: '推进状态',
      reset: '重置',
      deleteAction: '删除',
      shareAction: '分享',
      empty: '当前列还没有任务。',
      sharedCount: '分享次数',
      statuses: {
        planned: '计划中',
        in_progress: '进行中',
        review: '评审中',
        done: '已完成',
      },
      priorities: {
        high: '高',
        medium: '中',
        low: '低',
      },
      due: '截止日期',
      owner: '负责人',
      teammates: '协作者',
    },
    schedule: {
      title: '日程',
      subtitle: '新增、分享和整理日程安排，让协作节奏保持可见。',
      addEvent: '新增日程',
      addHint: '预置会议覆盖了验收流程，你也可以继续添加演示或联调安排。',
      titleLabel: '日程标题',
      detailLabel: '日程说明',
      ownerLabel: '主持人',
      createAction: '创建日程',
      deleteAction: '删除',
      shareAction: '分享',
      empty: '当前还没有安排日程。',
      sharedCount: '分享次数',
      starts: '开始时间',
      duration: '时长',
      room: '地点',
      attendees: '参与人',
      minutes: '分钟',
    },
    files: {
      title: '文件与资料',
      subtitle: '上传、打开、分享和删除工作台里的文件资料。',
      uploadAction: '上传文件',
      uploadHint: '文件列表只展示后端文件服务真实返回的数据。',
      owner: '负责人',
      updated: '更新时间',
      attachmentLabel: '附件',
      sharedCount: '分享次数',
      deleteAction: '删除',
      shareAction: '分享',
      openAction: '打开',
      sourceSeed: '预置资料',
      sourceUpload: '已上传',
      empty: '当前还没有任何文件。',
      categories: {
        draft: '草稿',
        review: '评审中',
        ready: '可发布',
      },
    },
    directory: {
      title: '通讯录',
      subtitle: '按姓名、电话、职称和部门搜索实时员工档案。',
      searchPlaceholder: '搜索姓名、电话、职称、工号或部门',
      searchHint: '左侧部门筛选会先收窄后端结果，再叠加关键词模糊匹配。',
      allDepartments: '全部部门',
      idLabel: '用户 ID',
      emailLabel: '邮箱',
      phoneLabel: '电话',
      roleLabel: '职称',
      departmentLabel: '部门',
      locationLabel: '办公地点',
      focusLabel: '简介',
      employeeNoLabel: '工号',
      bioLabel: '备注',
      noResults: '没有找到匹配的成员。',
      currentUser: '当前账号',
    },
    chat: {
      title: '沟通',
      subtitle: '私聊和群组已经放回到更完整的办公协作工作台里。',
      connectionOn: '已连接',
      connectionOff: '未连接',
      newConversation: '新建会话',
      conversations: '个会话',
      noConversations: '当前还没有会话。',
      selectConversation: '请选择一个会话开始沟通。',
      groupConversation: '群组会话',
      directConversation: '私聊会话',
      groupChat: '群聊',
      directChatPrefix: '私聊',
      searchPlaceholder: '搜索消息',
      searchAction: '搜索',
      clearAction: '清空',
      searching: '消息搜索中...',
      noSearchResults: '没有找到匹配消息。',
      noMessages: '当前还没有消息。',
      writeMessage: '输入消息',
      send: '发送',
      createTitle: '创建会话',
      createSubtitle: '可以发起私聊，也可以使用预置员工创建群组。',
      direct: '私聊',
      group: '群组',
      directTarget: '选择成员',
      directTargetHint: '选择一位员工账号，创建一对一私聊。',
      groupName: '群组名称',
      groupNamePlaceholder: '可选，留空则使用默认群名',
      groupMembers: '选择群成员',
      groupMembersHint: '至少选择一位成员加入该群组。',
      createAction: '创建',
      creating: '创建中...',
      invalidDirect: '请选择一位成员。',
      invalidGroup: '请至少选择一位成员。',
      noTeamMembers: '当前没有可用的团队成员账号。',
      memberCount: '名成员',
      announcementTitle: '群公告',
      announcementPlaceholder: '写一条群公告、协作规则或当前推进说明。',
      announcementSave: '保存公告',
      announcementSaved: '群公告已保存。',
      groupFilesTitle: '群文件',
      uploadFile: '上传到群组',
      noFiles: '当前群组还没有文件。',
      shareFile: '分享链接',
      deleteFile: '删除文件',
      uploadSuccess: '群文件上传成功。',
      uploadFailure: '群文件上传失败。',
      saveFailure: '保存群公告失败。',
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
