import type { Locale, SeedAccount } from '../types/workspace'

const seededAccountsByLocale: Record<Locale, SeedAccount[]> = {
  en: [
    { username: 'admin', password: 'admin123', nickname: 'Administrator', note: 'Acceptance and management walkthroughs' },
    { username: 'emma.chen', password: 'workpal123', nickname: 'Emma Chen', note: 'Operations and launch coordination' },
    { username: 'liam.wang', password: 'workpal123', nickname: 'Liam Wang', note: 'Engineering delivery and backend checks' },
    { username: 'sofia.zhao', password: 'workpal123', nickname: 'Sofia Zhao', note: 'Design QA and release readiness' },
  ],
  'zh-CN': [
    { username: 'admin', password: 'admin123', nickname: '管理员', note: '用于验收和管理视角联调' },
    { username: 'emma.chen', password: 'workpal123', nickname: 'Emma Chen', note: '运营协同与上线推进' },
    { username: 'liam.wang', password: 'workpal123', nickname: 'Liam Wang', note: '工程交付与后端联调' },
    { username: 'sofia.zhao', password: 'workpal123', nickname: 'Sofia Zhao', note: '设计验收与发布准备' },
  ],
}

export function getSeededAccounts(locale: Locale): SeedAccount[] {
  return seededAccountsByLocale[locale].map((account) => ({ ...account }))
}
