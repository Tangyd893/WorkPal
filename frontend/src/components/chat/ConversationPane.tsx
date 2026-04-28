import type { KeyboardEvent, RefObject } from 'react'
import type { ChatMessage, Conversation } from '../../types/chat'

interface ConversationPaneProps {
  conversation: Conversation | null
  userId: number | null
  messages: ChatMessage[]
  input: string
  searchQuery: string
  searching: boolean
  searchActive: boolean
  onInputChange: (value: string) => void
  onSearchChange: (value: string) => void
  onSearch: () => Promise<void>
  onClearSearch: () => void
  onSend: () => Promise<void>
  messagesEndRef: RefObject<HTMLDivElement>
}

function getConversationTitle(conversation: Conversation): string {
  if (conversation.type === 2) {
    return conversation.name || 'Group chat'
  }

  return conversation.name || `Direct chat #${conversation.id}`
}

function formatTimestamp(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ''
  }

  return date.toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
  })
}

export default function ConversationPane({
  conversation,
  userId,
  messages,
  input,
  searchQuery,
  searching,
  searchActive,
  onInputChange,
  onSearchChange,
  onSearch,
  onClearSearch,
  onSend,
  messagesEndRef,
}: ConversationPaneProps) {
  const handleInputKeyDown = async (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key !== 'Enter') {
      return
    }

    event.preventDefault()
    await onSend()
  }

  if (!conversation) {
    return (
      <main
        style={{
          flex: 1,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: '#6b7280',
          background: '#f8fafc',
        }}
      >
        Select a conversation to start chatting.
      </main>
    )
  }

  return (
    <main style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
      <div
        style={{
          padding: '14px 20px',
          background: '#ffffff',
          borderBottom: '1px solid #e5e7eb',
          display: 'flex',
          gap: 12,
          alignItems: 'center',
        }}
      >
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ fontSize: 16, fontWeight: 700, color: '#111827' }}>{getConversationTitle(conversation)}</div>
          <div style={{ fontSize: 12, color: '#6b7280', marginTop: 2 }}>
            {conversation.type === 2 ? 'Group conversation' : 'Direct conversation'}
          </div>
        </div>

        <form
          onSubmit={(event) => {
            event.preventDefault()
            void onSearch()
          }}
          style={{ display: 'flex', gap: 8 }}
        >
          <input
            type="text"
            value={searchQuery}
            onChange={(event) => onSearchChange(event.target.value)}
            placeholder="Search messages"
            style={{
              width: 220,
              padding: '8px 12px',
              borderRadius: 8,
              border: '1px solid #d1d5db',
              fontSize: 14,
            }}
          />
          <button
            type="submit"
            style={{
              padding: '8px 12px',
              borderRadius: 8,
              border: '1px solid #d1d5db',
              background: '#ffffff',
              color: '#111827',
            }}
          >
            Search
          </button>
          {searchQuery.trim() || searchActive ? (
            <button
              type="button"
              onClick={onClearSearch}
              style={{
                padding: '8px 12px',
                borderRadius: 8,
                border: '1px solid #d1d5db',
                background: '#ffffff',
                color: '#111827',
              }}
            >
              Clear
            </button>
          ) : null}
        </form>
      </div>

      <div
        style={{
          flex: 1,
          overflowY: 'auto',
          padding: '20px 24px',
          background: '#f8fafc',
          display: 'flex',
          flexDirection: 'column',
          gap: 12,
        }}
      >
        {searching ? (
          <div style={{ textAlign: 'center', color: '#6b7280', fontSize: 14 }}>Searching messages...</div>
        ) : null}

        {!searching && messages.length === 0 ? (
          <div style={{ textAlign: 'center', color: '#6b7280', fontSize: 14 }}>
            {searchActive ? 'No matching messages found.' : 'No messages yet.'}
          </div>
        ) : null}

        {messages.map((message) => {
          const ownMessage = message.sender_id === userId

          return (
            <div
              key={message.id}
              style={{
                display: 'flex',
                justifyContent: ownMessage ? 'flex-end' : 'flex-start',
              }}
            >
              <div
                style={{
                  maxWidth: '70%',
                  padding: '10px 14px',
                  borderRadius: 12,
                  background: ownMessage ? '#2563eb' : '#ffffff',
                  color: ownMessage ? '#ffffff' : '#111827',
                  boxShadow: '0 1px 3px rgba(15, 23, 42, 0.08)',
                  wordBreak: 'break-word',
                }}
              >
                <div style={{ fontSize: 14, lineHeight: 1.5 }}>{message.content}</div>
                <div
                  style={{
                    marginTop: 6,
                    fontSize: 11,
                    opacity: 0.75,
                    textAlign: 'right',
                  }}
                >
                  {formatTimestamp(message.created_at)}
                </div>
              </div>
            </div>
          )
        })}

        <div ref={messagesEndRef} />
      </div>

      <div style={{ padding: 16, background: '#ffffff', borderTop: '1px solid #e5e7eb' }}>
        <div style={{ display: 'flex', gap: 8 }}>
          <input
            type="text"
            value={input}
            onChange={(event) => onInputChange(event.target.value)}
            onKeyDown={(event) => {
              void handleInputKeyDown(event)
            }}
            placeholder="Write a message"
            style={{
              flex: 1,
              padding: '12px 14px',
              borderRadius: 8,
              border: '1px solid #d1d5db',
              fontSize: 14,
            }}
          />
          <button
            type="button"
            onClick={() => {
              void onSend()
            }}
            style={{
              padding: '12px 18px',
              borderRadius: 8,
              background: '#2563eb',
              color: '#ffffff',
              fontWeight: 600,
            }}
          >
            Send
          </button>
        </div>
      </div>
    </main>
  )
}
