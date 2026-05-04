import type { Locale, ThemeMode } from '../types/workspace'

export const PREFERENCES_STORAGE_KEY = 'workpal-preferences'

export interface PersistedPreferencesState {
  locale: Locale
  theme: ThemeMode
  soundEnabled: boolean
  compactMode: boolean
}

export const defaultPreferencesState: PersistedPreferencesState = {
  locale: 'zh-CN',
  theme: 'light',
  soundEnabled: true,
  compactMode: false,
}

function getStorage(): Storage | null {
  try {
    return globalThis.localStorage
  } catch {
    return null
  }
}

function asLocale(value: unknown): Locale | null {
  return value === 'en' || value === 'zh-CN' ? value : null
}

function asTheme(value: unknown): ThemeMode | null {
  return value === 'light' || value === 'dark' ? value : null
}

function asBoolean(value: unknown, fallback: boolean): boolean {
  return typeof value === 'boolean' ? value : fallback
}

export function loadPreferencesState(): PersistedPreferencesState {
  const storage = getStorage()
  if (!storage) {
    return { ...defaultPreferencesState }
  }

  try {
    const raw = storage.getItem(PREFERENCES_STORAGE_KEY)
    if (!raw) {
      return { ...defaultPreferencesState }
    }

    const parsed = JSON.parse(raw) as Partial<PersistedPreferencesState>
    return {
      locale: asLocale(parsed.locale) ?? defaultPreferencesState.locale,
      theme: asTheme(parsed.theme) ?? defaultPreferencesState.theme,
      soundEnabled: asBoolean(parsed.soundEnabled, defaultPreferencesState.soundEnabled),
      compactMode: asBoolean(parsed.compactMode, defaultPreferencesState.compactMode),
    }
  } catch {
    return { ...defaultPreferencesState }
  }
}

export function savePreferencesState(state: PersistedPreferencesState): void {
  const storage = getStorage()
  if (!storage) {
    return
  }

  storage.setItem(PREFERENCES_STORAGE_KEY, JSON.stringify(state))
}
