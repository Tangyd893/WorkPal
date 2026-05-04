import { describe, expect, it } from 'vitest'
import { getSeededAccounts } from './data/workspace'
import { translations } from './i18n'

describe('translations', () => {
  it('exposes both supported locales', () => {
    expect(translations.en.navigation.chat).toBe('Chat')
    expect(translations['zh-CN'].navigation.chat).toBe('沟通')
  })

  it('provides localized seeded accounts for each locale', () => {
    expect(getSeededAccounts('en')).toHaveLength(4)
    expect(getSeededAccounts('en')[0].nickname).not.toBe(getSeededAccounts('zh-CN')[0].nickname)
  })
})
