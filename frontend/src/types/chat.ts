export type ConversationType = 1 | 2

export interface ChatMessage {
  id: number
  conv_id: number
  sender_id: number
  type: number
  content: string
  metadata?: string
  reply_to?: number
  created_at: string
  updated_at?: string
}

export interface Conversation {
  id: number
  type: ConversationType
  name: string
  avatar_url?: string
  owner_id: number
  created_at: string
  updated_at: string
}

export interface PageData<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export interface SearchResult {
  messages: ChatMessage[]
  total: number
}

export interface LoginResponse {
  token: string
  expires_at: number
  user_id: number
  username: string
  nickname?: string
}

export type CreateConversationDraft =
  | {
      mode: 'private'
      targetUserId: number
    }
  | {
      mode: 'group'
      name: string
      memberIds: number[]
    }
