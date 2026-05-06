import { create } from 'zustand'
import type { ChatMessage } from '../types/chat'

interface WSState {
  connected: boolean
  ws: WebSocket | null
  messages: Record<number, ChatMessage[]>
  setConnected: (connected: boolean) => void
  setWS: (ws: WebSocket | null) => void
  addMessage: (convID: number, message: ChatMessage) => void
  setMessages: (convID: number, messages: ChatMessage[]) => void
}

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
