import { create } from 'zustand'
import type { Locale, ThemeMode } from '../types/workspace'
import {
  defaultPreferencesState,
  loadPreferencesState,
  savePreferencesState,
  type PersistedPreferencesState,
} from '../utils/preferencesStorage'

interface PreferencesState extends PersistedPreferencesState {
  setLocale: (locale: Locale) => void
  setTheme: (theme: ThemeMode) => void
  setSoundEnabled: (enabled: boolean) => void
  setCompactMode: (enabled: boolean) => void
  reset: () => void
}

function pickPersistedState(state: PersistedPreferencesState): PersistedPreferencesState
function pickPersistedState(state: PreferencesState): PersistedPreferencesState
function pickPersistedState(state: PersistedPreferencesState | PreferencesState): PersistedPreferencesState {
  return {
    locale: state.locale,
    theme: state.theme,
    soundEnabled: state.soundEnabled,
    compactMode: state.compactMode,
  }
}

function persistState(partial: PersistedPreferencesState, setState: (next: PersistedPreferencesState) => void): void {
  savePreferencesState(partial)
  setState(partial)
}

export const usePreferencesStore = create<PreferencesState>()((set, get) => ({
  ...loadPreferencesState(),
  setLocale: (locale) => {
    persistState({ ...pickPersistedState(get()), locale }, (next) => set(next))
  },
  setTheme: (theme) => {
    persistState({ ...pickPersistedState(get()), theme }, (next) => set(next))
  },
  setSoundEnabled: (soundEnabled) => {
    persistState({ ...pickPersistedState(get()), soundEnabled }, (next) => set(next))
  },
  setCompactMode: (compactMode) => {
    persistState({ ...pickPersistedState(get()), compactMode }, (next) => set(next))
  },
  reset: () => {
    persistState({ ...defaultPreferencesState }, (next) => set(next))
  },
}))
