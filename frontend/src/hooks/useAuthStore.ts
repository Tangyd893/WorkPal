import { create } from 'zustand'
import type { ChatMessage, Conversation } from '../types/chat'
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

interface WSState {
  connected: boolean
  ws: WebSocket | null
  messages: Record<number, ChatMessage[]>
  setConnected: (connected: boolean) => void
  setWS: (ws: WebSocket | null) => void
  addMessage: (convID: number, message: ChatMessage) => void
  setMessages: (convID: number, messages: ChatMessage[]) => void
}

interface ConvState {
  conversations: Conversation[]
  activeConvID: number | null
  setConversations: (conversations: Conversation[]) => void
  setActiveConvID: (id: number | null) => void
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

export const useWSStore = create<WSState>()((set) => ({
  connected: false,
  ws: null,
  messages: {},
  setConnected: (connected) => set({ connected }),
  setWS: (ws) => set({ ws }),
  addMessage: (convID, message) =>
    set((state) => {
      const existingMessages = state.messages[convID] ?? []
      const alreadyExists = message.id !== 0 && existingMessages.some((item) => item.id === message.id)
      if (alreadyExists) {
        return state
      }

      return {
        messages: {
          ...state.messages,
          [convID]: [...existingMessages, message],
        },
      }
    }),
  setMessages: (convID, messages) =>
    set((state) => ({
      messages: {
        ...state.messages,
        [convID]: messages,
      },
    })),
}))

export const useConvStore = create<ConvState>()((set) => ({
  conversations: [],
  activeConvID: null,
  setConversations: (conversations) => set({ conversations }),
  setActiveConvID: (activeConvID) => set({ activeConvID }),
}))
