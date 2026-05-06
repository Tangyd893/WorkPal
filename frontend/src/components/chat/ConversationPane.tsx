import { useState } from 'react'
import type { KeyboardEvent, RefObject } from 'react'
import type { AppTranslations } from '../../i18n'
import type { ChatMessage, Conversation } from '../../types/chat'
import ConfirmDialog from '../ConfirmDialog'

interface ConversationPaneProps {
  conversation: Conversation | null
  userId: number | null
  labels: AppTranslations['chat']
  commonLabels: AppTranslations['common']
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
  onCommitEdit: (messageID: number, content: string) => Promise<void>
  onRecallMessage: (messageID: number) => Promise<void>
  messagesEndRef: RefObject<HTMLDivElement>
}

function getConversationTitle(conversation: Conversation, labels: AppTranslations['chat']): string {
  if (conversation.type === 2) {
    return conversation.name || labels.groupChat
  }

  return conversation.name || `${labels.directChatPrefix} #${conversation.id}`
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
  labels,
  commonLabels,
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
  onCommitEdit,
  onRecallMessage,
  messagesEndRef,
}: ConversationPaneProps) {
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editDraft, setEditDraft] = useState('')
  const [recallTargetID, setRecallTargetID] = useState<number | null>(null)

  const handleInputKeyDown = async (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key !== 'Enter') {
      return
    }

    event.preventDefault()
    await onSend()
  }

  const handleStartEdit = (messageID: number, content: string) => {
    setEditingId(messageID)
    setEditDraft(content)
  }

  const handleCancelEdit = () => {
    setEditingId(null)
    setEditDraft('')
  }

  const handleCommitEdit = async () => {
    if (editingId === null) {
      return
    }

    const trimmed = editDraft.trim()
    if (!trimmed) {
      handleCancelEdit()
      return
    }

    await onCommitEdit(editingId, trimmed)
    setEditingId(null)
    setEditDraft('')
  }

  const handleConfirmRecall = async () => {
    if (recallTargetID === null) {
      return
    }

    await onRecallMessage(recallTargetID)
    setRecallTargetID(null)
  }

  if (!conversation) {
    return <section className="conversation-pane empty-panel">{labels.selectConversation}</section>
  }

  return (
    <section className="conversation-pane">
      <div className="conversation-toolbar">
        <div className="conversation-heading">
          <h3>{getConversationTitle(conversation, labels)}</h3>
          <p>{conversation.type === 2 ? labels.groupConversation : labels.directConversation}</p>
        </div>

        <form
          className="inline-form"
          onSubmit={(event) => {
            event.preventDefault()
            void onSearch()
          }}
        >
          <input
            type="text"
            value={searchQuery}
            onChange={(event) => onSearchChange(event.target.value)}
            placeholder={labels.searchPlaceholder}
            aria-label={labels.searchPlaceholder}
          />
          <button type="submit" className="secondary-button">
            {labels.searchAction}
          </button>
          {(searchQuery.trim() || searchActive) && (
            <button type="button" className="secondary-button" onClick={onClearSearch}>
              {labels.clearAction}
            </button>
          )}
        </form>
      </div>

      <div className="message-stream" aria-live="polite">
        {searching ? <div className="empty-panel" role="status">{labels.searching}</div> : null}

        {!searching && messages.length === 0 ? (
          <div className="empty-panel" role="status">{searchActive ? labels.noSearchResults : labels.noMessages}</div>
        ) : null}

        {messages.map((message) => {
          const ownMessage = message.sender_id === userId
          const isEditing = editingId === message.id

          return (
            <div key={message.id} className={ownMessage ? 'message-row self' : 'message-row'}>
              <div className={ownMessage ? 'message-bubble self' : 'message-bubble'}>
                {isEditing ? (
                  <div className="edit-inline">
                    <input
                      type="text"
                      value={editDraft}
                      onChange={(event) => setEditDraft(event.target.value)}
                      onKeyDown={(event) => {
                        if (event.key === 'Enter') {
                          event.preventDefault()
                          void handleCommitEdit()
                        } else if (event.key === 'Escape') {
                          handleCancelEdit()
                        }
                      }}
                      className="edit-input"
                      aria-label="Edit message"
                    />
                    <div className="edit-actions">
                      <button type="button" className="secondary-button" onClick={() => void handleCommitEdit()}>
                        {commonLabels.save}
                      </button>
                      <button type="button" className="secondary-button" onClick={handleCancelEdit}>
                        {commonLabels.cancel}
                      </button>
                    </div>
                  </div>
                ) : (
                  <>
                    <div className="message-content">{message.content}</div>
                    <div className="message-time">{formatTimestamp(message.created_at)}</div>
                  </>
                )}
              </div>
              {ownMessage && !isEditing ? (
                <div className="message-actions">
                  <button
                    type="button"
                    className="action-link"
                    onClick={() => handleStartEdit(message.id, message.content)}
                    aria-label={labels.editMessage}
                  >
                    {labels.editMessage}
                  </button>
                  <button
                    type="button"
                    className="action-link"
                    onClick={() => setRecallTargetID(message.id)}
                    aria-label={labels.recallMessage}
                  >
                    {labels.recallMessage}
                  </button>
                </div>
              ) : null}
            </div>
          )
        })}

        <div ref={messagesEndRef} />
      </div>

      <div className="composer">
        <input
          type="text"
          value={input}
          onChange={(event) => onInputChange(event.target.value)}
          onKeyDown={(event) => {
            void handleInputKeyDown(event)
          }}
          placeholder={labels.writeMessage}
          aria-label={labels.writeMessage}
        />
        <button
          type="button"
          className="primary-button"
          onClick={() => {
            void onSend()
          }}
        >
          {labels.send}
        </button>
      </div>

      <ConfirmDialog
        open={recallTargetID !== null}
        title={labels.recallMessage}
        message={labels.recallConfirmMessage}
        confirmText={labels.recallMessage}
        cancelText={commonLabels.cancel}
        variant="danger"
        onConfirm={() => void handleConfirmRecall()}
        onCancel={() => setRecallTargetID(null)}
      />
    </section>
  )
}
