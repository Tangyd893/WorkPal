import type { AppTranslations } from '../../i18n'
import type { Conversation } from '../../types/chat'

interface ConversationSidebarProps {
  conversations: Conversation[]
  activeConversationID: number | null
  labels: AppTranslations['chat']
  onCreateConversation: () => void
  onSelectConversation: (conversation: Conversation) => void
}

function getConversationTitle(conversation: Conversation, labels: AppTranslations['chat']): string {
  if (conversation.type === 2) {
    return conversation.name || labels.groupChat
  }

  return conversation.name || `${labels.directChatPrefix} #${conversation.id}`
}

function getConversationMeta(conversation: Conversation, labels: AppTranslations['chat']): string {
  return conversation.type === 2 ? labels.groupConversation : labels.directConversation
}

export default function ConversationSidebar({
  conversations,
  activeConversationID,
  labels,
  onCreateConversation,
  onSelectConversation,
}: ConversationSidebarProps) {
  return (
    <aside className="conversation-sidebar">
      <div className="sidebar-action">
        <button type="button" className="primary-button block-button" onClick={onCreateConversation}>
          {labels.newConversation}
        </button>
      </div>

      <div className="conversation-count">
        <span>{conversations.length}</span>
        <span>{labels.conversations}</span>
      </div>

      <div className="conversation-list">
        {conversations.length === 0 ? (
          <div className="empty-panel">{labels.noConversations}</div>
        ) : (
          conversations.map((conversation) => {
            const selected = conversation.id === activeConversationID

            return (
              <button
                key={conversation.id}
                type="button"
                className={selected ? 'conversation-item active' : 'conversation-item'}
                onClick={() => onSelectConversation(conversation)}
              >
                <div className="conversation-avatar">{conversation.type === 2 ? 'G' : 'D'}</div>
                <div className="conversation-copy">
                  <strong>{getConversationTitle(conversation, labels)}</strong>
                  <span>{getConversationMeta(conversation, labels)}</span>
                </div>
              </button>
            )
          })
        )}
      </div>
    </aside>
  )
}
