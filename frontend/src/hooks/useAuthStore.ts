import { create } from 'zustand'

interface AuthState {
  token: string | null
  userId: number | null
  username: string | null
  setAuth: (token: string, userId: number, username: string) => void
  logout: () => void
}

const STORAGE_KEY = 'workpal-auth'

// Helper to read persisted state
const loadPersistedState: () => AuthState = () => {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) return JSON.parse(raw)
  } catch {}
  return { token: null, userId: null, username: null }
}

export const useAuthStore = create<AuthState>()((set) => ({
  ...loadPersistedState(),
  setAuth: (token, userId, username) => {
    const state = { token, userId, username }
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
    set(state)
  },
  logout: () => {
    localStorage.removeItem(STORAGE_KEY)
    set({ token: null, userId: null, username: null })
  },
}))

// WebSocket store
interface WSState {
  connected: boolean
  ws: WebSocket | null
  messages: Record<number, ChatMessage[]> // convID -> messages
  setConnected: (v: boolean) => void
  setWS: (ws: WebSocket | null) => void
  addMessage: (convID: number, msg: ChatMessage) => void
  setMessages: (convID: number, msgs: ChatMessage[]) => void
}

export interface ChatMessage {
  id: number
  conv_id: number
  sender_id: number
  type: number
  content: string
  metadata?: string
  reply_to?: number
  created_at: string
}

export const useWSStore = create<WSState>()((set) => ({
  connected: false,
  ws: null,
  messages: {},
  setConnected: (v) => set({ connected: v }),
  setWS: (ws) => set({ ws }),
  addMessage: (convID, msg) =>
    set((state) => ({
      messages: {
        ...state.messages,
        [convID]: [...(state.messages[convID] || []), msg],
      },
    })),
  setMessages: (convID, msgs) =>
    set((state) => ({
      messages: { ...state.messages, [convID]: msgs },
    })),
}))

// Conversation store
interface ConvState {
  conversations: Conversation[]
  setConversations: (convs: Conversation[]) => void
  activeConvID: number | null
  setActiveConvID: (id: number | null) => void
}

export interface Conversation {
  id: number
  type: number // 1=private, 2=group
  name: string
  avatar_url?: string
  owner_id: number
  created_at: string
  updated_at: string
}

export const useConvStore = create<ConvState>()((set) => ({
  conversations: [],
  setConversations: (convs) => set({ conversations: convs }),
  activeConvID: null,
  setActiveConvID: (id) => set({ activeConvID: id }),
}))
