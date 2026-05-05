import { beforeEach, describe, expect, it } from 'vitest'
import { AxiosHeaders } from 'axios'
import { useAuthStore, useConvStore, useWSStore } from '../hooks/useAuthStore'
import request, { createTraceID, TRACE_ID_HEADER, TRACE_PARENT_HEADER } from './request'

beforeEach(() => {
  localStorage.clear()
  useAuthStore.getState().logout()
  useConvStore.setState({ conversations: [], activeConvID: null })
  useWSStore.setState({ connected: false, ws: null, messages: {} })
})

describe('useAuthStore (Zustand)', () => {
  it('should export auth store', () => {
    expect(typeof useAuthStore).toBe('function')
  })

  it('should have expected auth state keys', () => {
    const state = useAuthStore.getState()
    expect(state).toHaveProperty('token')
    expect(state).toHaveProperty('userId')
    expect(state).toHaveProperty('username')
    expect(state).toHaveProperty('logout')
    expect(state).toHaveProperty('setAuth')
  })

  it('should initialize with null values', () => {
    const state = useAuthStore.getState()
    expect(state.token).toBeNull()
    expect(state.userId).toBeNull()
    expect(state.username).toBeNull()
  })

  it('should be able to set auth data', () => {
    const { setAuth } = useAuthStore.getState()
    setAuth('test-token', 123, 'testuser')
    const state = useAuthStore.getState()
    expect(state.token).toBe('test-token')
    expect(state.userId).toBe(123)
    expect(state.username).toBe('testuser')
  })

  it('should be able to logout', () => {
    const { logout } = useAuthStore.getState()
    logout()
    const state = useAuthStore.getState()
    expect(state.token).toBeNull()
    expect(state.userId).toBeNull()
    expect(state.username).toBeNull()
  })
})

describe('useConvStore (Zustand)', () => {
  it('should export conv store', () => {
    expect(typeof useConvStore).toBe('function')
  })

  it('should have expected conv state keys', () => {
    const state = useConvStore.getState()
    expect(state).toHaveProperty('conversations')
    expect(state).toHaveProperty('setConversations')
    expect(state).toHaveProperty('activeConvID')
    expect(state).toHaveProperty('setActiveConvID')
  })

  it('should initialize with empty conversations', () => {
    const state = useConvStore.getState()
    expect(state.conversations).toEqual([])
    expect(state.activeConvID).toBeNull()
  })

  it('should be able to set active conv', () => {
    const { setActiveConvID } = useConvStore.getState()
    setActiveConvID(99)
    const state = useConvStore.getState()
    expect(state.activeConvID).toBe(99)
  })
})

describe('useWSStore (Zustand)', () => {
  it('should export ws store', () => {
    expect(typeof useWSStore).toBe('function')
  })

  it('should have expected ws state keys', () => {
    const state = useWSStore.getState()
    expect(state).toHaveProperty('connected')
    expect(state).toHaveProperty('setConnected')
    expect(state).toHaveProperty('addMessage')
    expect(state).toHaveProperty('messages')
    expect(state).toHaveProperty('setMessages')
  })

  it('should initialize with disconnected state', () => {
    const state = useWSStore.getState()
    expect(state.connected).toBe(false)
    expect(state.messages).toEqual({})
  })

  it('should be able to set connected state', () => {
    const { setConnected } = useWSStore.getState()
    setConnected(true)
    const state = useWSStore.getState()
    expect(state.connected).toBe(true)
  })

  it('should be able to add message to conversation', () => {
    const { addMessage } = useWSStore.getState()
    const msg = { id: 1, conv_id: 100, sender_id: 10, type: 1, content: 'hello', created_at: new Date().toISOString() }
    addMessage(100, msg)
    const state = useWSStore.getState()
    expect(state.messages[100]).toBeDefined()
    expect(state.messages[100].length).toBe(1)
    expect(state.messages[100][0].content).toBe('hello')
  })

  it('should ignore duplicate messages', () => {
    const { addMessage } = useWSStore.getState()
    const msg = { id: 1, conv_id: 100, sender_id: 10, type: 1, content: 'hello', created_at: new Date().toISOString() }
    addMessage(100, msg)
    addMessage(100, msg)
    const state = useWSStore.getState()
    expect(state.messages[100].length).toBe(1)
  })

  it('should be able to set messages for conversation', () => {
    const { setMessages } = useWSStore.getState()
    const msgs = [
      { id: 2, conv_id: 200, sender_id: 20, type: 1, content: 'msg1', created_at: new Date().toISOString() },
      { id: 3, conv_id: 200, sender_id: 30, type: 1, content: 'msg2', created_at: new Date().toISOString() },
    ]
    setMessages(200, msgs)
    const state = useWSStore.getState()
    expect(state.messages[200].length).toBe(2)
  })
})

describe('request client observability', () => {
  it('should create non-empty trace ids', () => {
    expect(createTraceID()).toMatch(/^[a-f0-9]{32}$/i)
  })

  it('should add trace headers to outgoing requests', async () => {
    const response = await request.get<{ traceID: string | null; traceParent: string | null }>('/probe', {
      adapter: async (config) => {
        const headers = AxiosHeaders.from(config.headers)
        const traceID = headers.get(TRACE_ID_HEADER)
        const traceParent = headers.get(TRACE_PARENT_HEADER)
        return {
          data: {
            traceID: typeof traceID === 'string' ? traceID : null,
            traceParent: typeof traceParent === 'string' ? traceParent : null,
          },
          status: 200,
          statusText: 'OK',
          headers: {},
          config,
        }
      },
    })

    expect(response.data.traceID).toBeTruthy()
    expect(response.data.traceParent).toContain(response.data.traceID)
  })

  it('should preserve an explicit trace header', async () => {
    const response = await request.get<{ traceID: string | null }>('/probe', {
      headers: { [TRACE_ID_HEADER]: 'trace-from-caller' },
      adapter: async (config) => {
        const headers = AxiosHeaders.from(config.headers)
        const traceID = headers.get(TRACE_ID_HEADER)
        return {
          data: { traceID: typeof traceID === 'string' ? traceID : null },
          status: 200,
          statusText: 'OK',
          headers: {},
          config,
        }
      },
    })

    expect(response.data.traceID).toBe('trace-from-caller')
  })

  it('should derive traceparent from an explicit W3C trace id', async () => {
    const traceID = '4bf92f3577b34da6a3ce929d0e0e4736'
    const response = await request.get<{ traceParent: string | null }>('/probe', {
      headers: { [TRACE_ID_HEADER]: traceID },
      adapter: async (config) => {
        const headers = AxiosHeaders.from(config.headers)
        const traceParent = headers.get(TRACE_PARENT_HEADER)
        return {
          data: { traceParent: typeof traceParent === 'string' ? traceParent : null },
          status: 200,
          statusText: 'OK',
          headers: {},
          config,
        }
      },
    })

    expect(response.data.traceParent).toContain(traceID)
  })
})
