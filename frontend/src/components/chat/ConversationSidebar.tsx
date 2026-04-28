import type { Conversation } from '../../types/chat'

interface ConversationSidebarProps {
  conversations: Conversation[]
  activeConversationID: number | null
  onCreateConversation: () => void
  onSelectConversation: (conversation: Conversation) => void
}

function getConversationTitle(conversation: Conversation): string {
  if (conversation.type === 2) {
    return conversation.name || 'Group chat'
  }

  return conversation.name || `Direct chat #${conversation.id}`
}

function getConversationMeta(conversation: Conversation): string {
  return conversation.type === 2 ? 'Group conversation' : 'Direct conversation'
}

export default function ConversationSidebar({
  conversations,
  activeConversationID,
  onCreateConversation,
  onSelectConversation,
}: ConversationSidebarProps) {
  return (
    <aside
      style={{
        width: 280,
        background: '#ffffff',
        borderRight: '1px solid #e5e7eb',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <div style={{ padding: 16, borderBottom: '1px solid #f3f4f6' }}>
        <button
          type="button"
          onClick={onCreateConversation}
          style={{
            width: '100%',
            padding: '10px 14px',
            borderRadius: 8,
            background: '#2563eb',
            color: '#ffffff',
            fontSize: 14,
            fontWeight: 600,
          }}
        >
          New conversation
        </button>
      </div>

      <div style={{ padding: '12px 16px', fontSize: 12, color: '#6b7280', borderBottom: '1px solid #f9fafb' }}>
        {conversations.length} conversation{conversations.length === 1 ? '' : 's'}
      </div>

      <div style={{ flex: 1, overflowY: 'auto' }}>
        {conversations.length === 0 ? (
          <div style={{ padding: 24, color: '#6b7280', fontSize: 14, textAlign: 'center' }}>
            No conversations yet.
          </div>
        ) : (
          conversations.map((conversation) => {
            const selected = conversation.id === activeConversationID

            return (
              <button
                key={conversation.id}
                type="button"
                onClick={() => onSelectConversation(conversation)}
                style={{
                  width: '100%',
                  padding: '14px 16px',
                  borderBottom: '1px solid #f9fafb',
                  background: selected ? '#eff6ff' : 'transparent',
                  textAlign: 'left',
                  display: 'flex',
                  alignItems: 'flex-start',
                  gap: 12,
                }}
              >
                <div
                  aria-hidden="true"
                  style={{
                    width: 40,
                    height: 40,
                    borderRadius: '50%',
                    background: selected ? '#bfdbfe' : '#e5e7eb',
                    color: '#1f2937',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: 16,
                    fontWeight: 700,
                    flexShrink: 0,
                  }}
                >
                  {conversation.type === 2 ? 'G' : 'D'}
                </div>

                <div style={{ minWidth: 0 }}>
                  <div style={{ fontSize: 14, fontWeight: 600, color: '#111827' }}>
                    {getConversationTitle(conversation)}
                  </div>
                  <div style={{ fontSize: 12, color: '#6b7280', marginTop: 2 }}>
                    {getConversationMeta(conversation)}
                  </div>
                </div>
              </button>
            )
          })
        )}
      </div>
    </aside>
  )
}
