import { create } from 'zustand'
import type { Conversation } from '../types/chat'

interface ConvState {
  conversations: Conversation[]
  activeConvID: number | null
  setConversations: (conversations: Conversation[]) => void
  setActiveConvID: (id: number | null) => void
}

export const useConvStore = create<ConvState>()((set) => ({
  conversations: [],
  activeConvID: null,
  setConversations: (conversations) => set({ conversations }),
  setActiveConvID: (activeConvID) => set({ activeConvID }),
}))
