import { beforeEach, describe, expect, it } from 'vitest'
import { usePreferencesStore } from '../hooks/usePreferencesStore'
import {
  defaultPreferencesState,
  loadPreferencesState,
  PREFERENCES_STORAGE_KEY,
  savePreferencesState,
} from './preferencesStorage'

beforeEach(() => {
  localStorage.clear()
  usePreferencesStore.getState().reset()
})

describe('preferences storage', () => {
  it('loads defaults when storage is empty', () => {
    expect(loadPreferencesState()).toEqual(defaultPreferencesState)
  })

  it('round-trips preference values', () => {
    const next = {
      locale: 'en' as const,
      theme: 'dark' as const,
      soundEnabled: false,
      compactMode: true,
    }

    savePreferencesState(next)
    expect(loadPreferencesState()).toEqual(next)
  })

  it('store actions persist to localStorage', () => {
    const store = usePreferencesStore.getState()

    store.setLocale('en')
    store.setTheme('dark')
    store.setSoundEnabled(false)
    store.setCompactMode(true)

    const raw = localStorage.getItem(PREFERENCES_STORAGE_KEY)
    expect(raw).not.toBeNull()
    expect(loadPreferencesState()).toEqual({
      locale: 'en',
      theme: 'dark',
      soundEnabled: false,
      compactMode: true,
    })
  })
})
