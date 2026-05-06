import { create } from 'zustand'
import {
  clearAuthState,
  emptyAuthState,
  loadAuthState,
  saveAuthState,
  type PersistedAuthState,
} from '../utils/authStorage'

interface AuthState extends PersistedAuthState {
  setAuth: (token: string, userId: number, username: string) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()((set) => ({
  ...loadAuthState(),
  setAuth: (token, userId, username) => {
    const nextState = { token, userId, username }
    saveAuthState(nextState)
    set(nextState)
  },
  logout: () => {
    clearAuthState()
    set({ ...emptyAuthState })
  },
}))
