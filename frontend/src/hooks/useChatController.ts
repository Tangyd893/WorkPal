import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { workpalApi } from '../api/workpal'
import { playMessageTone } from '../utils/notifications'
import type { ChatMessage, Conversation, CreateConversationDraft } from '../types/chat'
import { useAuthStore, useConvStore, useWSStore } from './useAuthStore'
import { usePreferencesStore } from './usePreferencesStore'

interface WebSocketMessage {
  type: string
  id?: number
  from?: number
  conv_id?: number
  content?: unknown
  created_at?: string
}

function buildWebSocketURL(token: string): string {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}/ws?token=${encodeURIComponent(token)}`
}

function getErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Unexpected request failure.'
}

export function useChatController() {
  const { username, userId, token } = useAuthStore()
  const { conversations, setConversations, activeConvID, setActiveConvID } = useConvStore()
  const { connected, setConnected, setWS, addMessage, messages, setMessages } = useWSStore()
  const soundEnabled = usePreferencesStore((state) => state.soundEnabled)

  const [input, setInput] = useState('')
  const [searchQuery, setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<ChatMessage[]>([])
  const [searching, setSearching] = useState(false)
  const [searchActive, setSearchActive] = useState(false)
  const [createDialogOpen, setCreateDialogOpen] = useState(false)

  const messagesEndRef = useRef<HTMLDivElement>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimerRef = useRef<number | null>(null)
  const closingRef = useRef(false)

  const currentConversation = useMemo(
    () => conversations.find((conversation) => conversation.id === activeConvID) ?? null,
    [activeConvID, conversations],
  )

  const displayedMessages = useMemo(() => {
    if (searchActive) {
      return searchResults
    }

    if (!activeConvID) {
      return []
    }

    return messages[activeConvID] ?? []
  }, [activeConvID, messages, searchActive, searchResults])

  const loadMessages = useCallback(
    async (convID: number) => {
      try {
        const history = await workpalApi.getConversationMessages(convID)
        setMessages(convID, history)
      } catch (error) {
        console.error(`Unable to load messages for conversation ${convID}.`, error)
        throw error
      }
    },
    [setMessages],
  )

  const loadConversations = useCallback(async () => {
    try {
      const nextConversations = await workpalApi.listConversations()
      setConversations(nextConversations)

      if (nextConversations.length === 0) {
        setActiveConvID(null)
        return
      }

      const fallbackConversation =
        nextConversations.find((conversation) => conversation.id === activeConvID) ?? nextConversations[0] ?? null

      if (!fallbackConversation) {
        return
      }

      setActiveConvID(fallbackConversation.id)
      if (!messages[fallbackConversation.id]) {
        await loadMessages(fallbackConversation.id)
      }
    } catch (error) {
      console.error('Unable to load conversations.', error)
      throw error
    }
  }, [activeConvID, loadMessages, messages, setActiveConvID, setConversations])

  const handleSocketMessage = useCallback(
    (event: MessageEvent<string>) => {
      try {
        const payload = JSON.parse(event.data) as WebSocketMessage
        if (payload.type !== 'chat' || typeof payload.conv_id !== 'number' || typeof payload.from !== 'number') {
          return
        }

        const message: ChatMessage = {
          id: payload.id ?? Date.now(),
          conv_id: payload.conv_id,
          sender_id: payload.from,
          type: 1,
          content: typeof payload.content === 'string' ? payload.content : JSON.stringify(payload.content ?? ''),
          created_at: payload.created_at ?? new Date().toISOString(),
        }

        addMessage(payload.conv_id, message)

        if (payload.from !== userId && soundEnabled) {
          playMessageTone()
        }
      } catch (error) {
        console.error('Unable to parse websocket message.', error)
      }
    },
    [addMessage, soundEnabled, userId],
  )

  const connectWebSocket = useCallback(() => {
    if (!token) {
      return
    }

    const currentSocket = wsRef.current
    if (currentSocket && (currentSocket.readyState === WebSocket.OPEN || currentSocket.readyState === WebSocket.CONNECTING)) {
      return
    }

    closingRef.current = false
    const nextSocket = new WebSocket(buildWebSocketURL(token))

    nextSocket.onopen = () => {
      setConnected(true)
      setWS(nextSocket)
    }

    nextSocket.onmessage = handleSocketMessage

    nextSocket.onerror = (error) => {
      console.error('WebSocket error.', error)
    }

    nextSocket.onclose = () => {
      setConnected(false)
      setWS(null)
      wsRef.current = null

      if (!closingRef.current) {
        reconnectTimerRef.current = window.setTimeout(connectWebSocket, 3000)
      }
    }

    wsRef.current = nextSocket
    setWS(nextSocket)
  }, [handleSocketMessage, setConnected, setWS, token])

  useEffect(() => {
    if (!token) {
      return
    }

    void loadConversations().catch(() => {
      // Error already logged in loadConversations.
    })
    connectWebSocket()

    return () => {
      closingRef.current = true

      if (reconnectTimerRef.current !== null) {
        window.clearTimeout(reconnectTimerRef.current)
      }

      wsRef.current?.close()
      wsRef.current = null
      setWS(null)
      setConnected(false)
    }
  }, [connectWebSocket, loadConversations, setConnected, setWS, token])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [activeConvID, displayedMessages])

  const handleSelectConversation = useCallback(
    async (conversation: Conversation) => {
      setActiveConvID(conversation.id)
      setSearchQuery('')
      setSearchResults([])
      setSearchActive(false)

      if (!messages[conversation.id]) {
        try {
          await loadMessages(conversation.id)
        } catch {
          // Error already logged in loadMessages.
        }
      }
    },
    [loadMessages, messages, setActiveConvID],
  )

  const handleSend = useCallback(async () => {
    const trimmedInput = input.trim()
    if (!trimmedInput || !activeConvID) {
      return
    }

    try {
      const message = await workpalApi.sendMessage(activeConvID, trimmedInput)
      addMessage(activeConvID, message)
      setInput('')
    } catch (error) {
      console.error(`Unable to send message: ${getErrorMessage(error)}`, error)
    }
  }, [activeConvID, addMessage, input])

  const handleSearch = useCallback(async () => {
    const trimmedQuery = searchQuery.trim()
    if (!trimmedQuery) {
      setSearchResults([])
      setSearchActive(false)
      return
    }

    setSearching(true)
    try {
      const result = await workpalApi.searchMessages(trimmedQuery, activeConvID ?? undefined)
      setSearchResults(result.messages)
      setSearchActive(true)
    } catch (error) {
      setSearchResults([])
      setSearchActive(true)
      console.error(`Unable to search messages: ${getErrorMessage(error)}`, error)
    } finally {
      setSearching(false)
    }
  }, [activeConvID, searchQuery])

  const handleClearSearch = useCallback(() => {
    setSearchQuery('')
    setSearchResults([])
    setSearchActive(false)
  }, [])

  const handleCreateConversation = useCallback(
    async (draft: CreateConversationDraft) => {
      const conversation = await workpalApi.createConversation(draft)
      await loadConversations()
      setActiveConvID(conversation.id)
      await loadMessages(conversation.id)
      setSearchQuery('')
      setSearchResults([])
      setSearchActive(false)
      setCreateDialogOpen(false)
    },
    [loadConversations, loadMessages, setActiveConvID],
  )

  return {
    activeConvID,
    connected,
    conversations,
    createDialogOpen,
    currentConversation,
    displayedMessages,
    input,
    messagesEndRef,
    searchActive,
    searchQuery,
    searching,
    userId,
    username,
    handleClearSearch,
    handleCreateConversation,
    handleSearch,
    handleSelectConversation,
    handleSend,
    openCreateDialog: () => setCreateDialogOpen(true),
    closeCreateDialog: () => setCreateDialogOpen(false),
    setInput,
    setSearchQuery,
  }
}
