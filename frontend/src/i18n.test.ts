import { describe, expect, it } from 'vitest'
import { buildDocuments, buildSchedule, buildSeedTasks, getSeededAccounts } from './data/workspace'
import { translations } from './i18n'

describe('translations', () => {
  it('exposes both supported locales', () => {
    expect(translations.en.navigation.chat).toBe('Chat')
    expect(translations['zh-CN'].navigation.chat).toBe('沟通')
  })

  it('provides localized seed data for each locale', () => {
    expect(buildSeedTasks('en')[0].title).not.toBe(buildSeedTasks('zh-CN')[0].title)
    expect(buildSchedule('en')).toHaveLength(3)
    expect(buildDocuments('zh-CN')).toHaveLength(3)
    expect(getSeededAccounts('en')).toHaveLength(4)
  })
})
