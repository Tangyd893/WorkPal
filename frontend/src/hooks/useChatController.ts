import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { workpalApi } from '../api/workpal'
import type { ChatMessage, Conversation, CreateConversationDraft } from '../types/chat'
import type { ConversationFile } from '../types/workspace'
import { copyText } from '../utils/clipboard'
import { playMessageTone } from '../utils/notifications'
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

function sortFiles(files: ConversationFile[]): ConversationFile[] {
  return [...files].sort(
    (left, right) => new Date(right.created_at).getTime() - new Date(left.created_at).getTime(),
  )
}

export function useChatController(requestedConversationID?: number, onConversationChange?: (conversationID: number) => void) {
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
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [groupFiles, setGroupFiles] = useState<ConversationFile[]>([])
  const [groupFilesLoading, setGroupFilesLoading] = useState(false)
  const [groupFileUploading, setGroupFileUploading] = useState(false)
  const [groupUploadProgress, setGroupUploadProgress] = useState(0)
  const [announcementDraft, setAnnouncementDraft] = useState('')
  const [announcementSaving, setAnnouncementSaving] = useState(false)

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

  const replaceConversation = useCallback(
    (updatedConversation: Conversation) => {
      setConversations(
        conversations.map((conversation) =>
          conversation.id === updatedConversation.id ? updatedConversation : conversation,
        ),
      )
    },
    [conversations, setConversations],
  )

  const loadMessages = useCallback(
    async (convID: number) => {
      const history = await workpalApi.getConversationMessages(convID)
      setMessages(convID, history)
    },
    [setMessages],
  )

  const loadConversations = useCallback(async () => {
    const nextConversations = await workpalApi.listConversations()
    setConversations(nextConversations)

    if (nextConversations.length === 0) {
      setActiveConvID(null)
      return
    }

    const routeConversation =
      typeof requestedConversationID === 'number'
        ? nextConversations.find((conversation) => conversation.id === requestedConversationID)
        : null
    const fallbackConversation =
      routeConversation ?? nextConversations.find((conversation) => conversation.id === activeConvID) ?? nextConversations[0] ?? null

    if (!fallbackConversation) {
      return
    }

    setActiveConvID(fallbackConversation.id)
    if (!messages[fallbackConversation.id]) {
      await loadMessages(fallbackConversation.id)
    }
  }, [activeConvID, loadMessages, messages, requestedConversationID, setActiveConvID, setConversations])

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
      } catch (socketError) {
        console.error('Unable to parse websocket message.', socketError)
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

    nextSocket.onerror = (socketError) => {
      console.error('WebSocket error.', socketError)
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

    void loadConversations().catch((loadError) => {
      setError(getErrorMessage(loadError))
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

  useEffect(() => {
    if (!currentConversation || currentConversation.type !== 2) {
      setAnnouncementDraft('')
      setGroupFiles([])
      setGroupFilesLoading(false)
      return
    }

    setAnnouncementDraft(currentConversation.announcement ?? '')
    let disposed = false

    const loadGroupFiles = async () => {
      setGroupFilesLoading(true)
      try {
        const files = await workpalApi.listConversationFiles(currentConversation.id)
        if (!disposed) {
          setGroupFiles(sortFiles(files))
        }
      } catch (loadError) {
        if (!disposed) {
          setError(getErrorMessage(loadError))
        }
      } finally {
        if (!disposed) {
          setGroupFilesLoading(false)
        }
      }
    }

    void loadGroupFiles()

    return () => {
      disposed = true
    }
  }, [currentConversation])

  const handleSelectConversation = useCallback(
    async (conversation: Conversation, syncRoute = true) => {
      setActiveConvID(conversation.id)
      if (syncRoute) {
        onConversationChange?.(conversation.id)
      }
      setSearchQuery('')
      setSearchResults([])
      setSearchActive(false)
      setError('')
      setNotice('')

      if (!messages[conversation.id]) {
        try {
          await loadMessages(conversation.id)
        } catch (loadError) {
          setError(getErrorMessage(loadError))
        }
      }
    },
    [loadMessages, messages, onConversationChange, setActiveConvID],
  )

  useEffect(() => {
    if (typeof requestedConversationID !== 'number' || requestedConversationID === activeConvID) {
      return
    }

    const requestedConversation = conversations.find((conversation) => conversation.id === requestedConversationID)
    if (requestedConversation) {
      void handleSelectConversation(requestedConversation, false)
    }
  }, [activeConvID, conversations, handleSelectConversation, requestedConversationID])

  const handleSend = useCallback(async () => {
    const trimmedInput = input.trim()
    if (!trimmedInput || !activeConvID) {
      return
    }

    try {
      const message = await workpalApi.sendMessage(activeConvID, trimmedInput)
      addMessage(activeConvID, message)
      setInput('')
      setError('')
    } catch (sendError) {
      setError(getErrorMessage(sendError))
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
      setError('')
    } catch {
      setSearchResults([])
      setSearchActive(true)
      setError('搜索暂不可用，消息收发不受影响。')
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
      onConversationChange?.(conversation.id)
      await loadMessages(conversation.id)
      setSearchQuery('')
      setSearchResults([])
      setSearchActive(false)
      setCreateDialogOpen(false)
      setError('')
      setNotice('')
    },
    [loadConversations, loadMessages, onConversationChange, setActiveConvID],
  )

  const handleSaveAnnouncement = useCallback(async () => {
    if (!currentConversation || currentConversation.type !== 2) {
      return
    }

    setAnnouncementSaving(true)
    try {
      const updated = await workpalApi.updateConversationAnnouncement(currentConversation.id, announcementDraft.trim())
      replaceConversation(updated)
      setNotice(updated.announcement || '')
      setError('')
    } catch (saveError) {
      setError(getErrorMessage(saveError))
    } finally {
      setAnnouncementSaving(false)
    }
  }, [announcementDraft, currentConversation, replaceConversation])

  const handleUploadGroupFile = useCallback(
    async (file: File) => {
      if (!currentConversation) {
        return
      }

      setGroupFileUploading(true)
      setGroupUploadProgress(0)
      try {
        const uploaded = await workpalApi.uploadConversationFile(currentConversation.id, file, setGroupUploadProgress)
        setGroupUploadProgress(100)
        setGroupFiles((current) => sortFiles([uploaded, ...current]))
        setNotice(uploaded.name)
        setError('')
      } catch (uploadError) {
        setError(getErrorMessage(uploadError))
      } finally {
        window.setTimeout(() => setGroupUploadProgress(0), 300)
        setGroupFileUploading(false)
      }
    },
    [currentConversation],
  )

  const handleDeleteGroupFile = useCallback(async (fileID: number) => {
    try {
      await workpalApi.deleteFile(fileID)
      setGroupFiles((current) => current.filter((file) => file.id !== fileID))
      setError('')
    } catch (deleteError) {
      setError(getErrorMessage(deleteError))
    }
  }, [])

  const handleShareGroupFile = useCallback(async (fileID: number) => {
    try {
      const shareInfo = await workpalApi.shareFile(fileID)
      const copied = await copyText(shareInfo.share_text)
      setNotice(copied ? shareInfo.share_text : shareInfo.download_path)
      setError('')
    } catch (shareError) {
      setError(getErrorMessage(shareError))
    }
  }, [])

  const handleCommitEdit = useCallback(
    async (messageID: number, content: string) => {
      if (!activeConvID) {
        return
      }

      try {
        const updated = await workpalApi.editMessage(messageID, content)
        const updatedMessage = updated as ChatMessage & { updated_at?: string }
        const current = messages[activeConvID] ?? []
        setMessages(
          activeConvID,
          current.map((msg) =>
            msg.id === messageID
              ? {
                  ...msg,
                  content: updated.content,
                  ...(updatedMessage.updated_at ? { updated_at: updatedMessage.updated_at } : {}),
                }
              : msg,
          ),
        )
        setError('')
      } catch (editError) {
        setError(getErrorMessage(editError))
      }
    },
    [activeConvID, messages, setMessages],
  )

  const handleRecallMessage = useCallback(
    async (messageID: number) => {
      if (!activeConvID) {
        return
      }

      try {
        await workpalApi.recallMessage(messageID)
        const current = messages[activeConvID] ?? []
        setMessages(
          activeConvID,
          current.filter((msg) => msg.id !== messageID),
        )
        setError('')
      } catch (recallError) {
        setError(getErrorMessage(recallError))
      }
    },
    [activeConvID, messages, setMessages],
  )

  return {
    activeConvID,
    announcementDraft,
    announcementSaving,
    connected,
    conversations,
    createDialogOpen,
    currentConversation,
    displayedMessages,
    error,
    groupFileUploading,
    groupFiles,
    groupFilesLoading,
    groupUploadProgress,
    input,
    messagesEndRef,
    notice,
    searchActive,
    searchQuery,
    searching,
    userId,
    username,
    closeCreateDialog: () => setCreateDialogOpen(false),
    handleClearSearch,
    handleCommitEdit,
    handleCreateConversation,
    handleDeleteGroupFile,
    handleRecallMessage,
    handleSaveAnnouncement,
    handleSearch,
    handleSelectConversation,
    handleSend,
    handleShareGroupFile,
    handleUploadGroupFile,
    openCreateDialog: () => setCreateDialogOpen(true),
    setAnnouncementDraft,
    setInput,
    setSearchQuery,
  }
}
