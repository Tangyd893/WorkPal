import { apiGet, apiPost } from './request'
import type {
  ChatMessage,
  Conversation,
  CreateConversationDraft,
  LoginResponse,
  PageData,
  SearchResult,
} from '../types/chat'
import type { WorkspaceUser } from '../types/workspace'

interface LoginPayload {
  username: string
  password: string
}

interface SendMessagePayload {
  type: number
  content: string
}

type CreateConversationPayload =
  | {
      type: 1
      target_uid: number
    }
  | {
      type: 2
      name: string
      member_ids: number[]
    }

function unwrapPageData<T>(value: T[] | PageData<T>): T[] {
  return Array.isArray(value) ? value : value.items
}

export const workpalApi = {
  login(payload: LoginPayload) {
    return apiPost<LoginResponse, LoginPayload>('/auth/login', payload)
  },

  getMe() {
    return apiGet<WorkspaceUser>('/users/me')
  },

  async listUsers(pageSize = 100): Promise<WorkspaceUser[]> {
    const data = await apiGet<WorkspaceUser[] | PageData<WorkspaceUser>>('/users', {
      params: {
        page: 1,
        page_size: pageSize,
      },
    })
    return unwrapPageData(data)
  },

  async listConversations(): Promise<Conversation[]> {
    const data = await apiGet<Conversation[] | PageData<Conversation>>('/conversations')
    return unwrapPageData(data)
  },

  getConversationMessages(convID: number) {
    return apiGet<ChatMessage[]>(`/conversations/${convID}/messages`)
  },

  sendMessage(convID: number, content: string) {
    const payload: SendMessagePayload = { type: 1, content }
    return apiPost<ChatMessage, SendMessagePayload>(`/conversations/${convID}/messages`, payload)
  },

  searchMessages(query: string, convID?: number, page = 1, pageSize = 20) {
    return apiGet<SearchResult>('/search/messages', {
      params: {
        q: query,
        conv_id: convID,
        page,
        page_size: pageSize,
      },
    })
  },

  createConversation(draft: CreateConversationDraft) {
    const payload: CreateConversationPayload =
      draft.mode === 'private'
        ? {
            type: 1,
            target_uid: draft.targetUserId,
          }
        : {
            type: 2,
            name: draft.name || 'Group chat',
            member_ids: draft.memberIds,
          }

    return apiPost<Conversation, CreateConversationPayload>('/conversations', payload)
  },
}
