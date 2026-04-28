import type { AppTranslations } from '../i18n'
import ConversationPane from '../components/chat/ConversationPane'
import ConversationSidebar from '../components/chat/ConversationSidebar'
import CreateConversationModal from '../components/chat/CreateConversationModal'
import { useChatController } from '../hooks/useChatController'
import type { WorkspaceUser } from '../types/workspace'

interface ChatPageProps {
  teamMembers: WorkspaceUser[]
  text: AppTranslations
}

export default function ChatPage({ teamMembers, text }: ChatPageProps) {
  const chat = useChatController()

  return (
    <section className="module-surface">
      <div className="module-header">
        <div>
          <h2>{text.chat.title}</h2>
          <p>{text.chat.subtitle}</p>
        </div>
        <div className="status-row">
          <span className={chat.connected ? 'status-badge positive' : 'status-badge critical'}>
            {chat.connected ? text.chat.connectionOn : text.chat.connectionOff}
          </span>
          <span className="subtle-label">{chat.username || text.common.unavailable}</span>
        </div>
      </div>

      <div className="chat-layout">
        <ConversationSidebar
          conversations={chat.conversations}
          activeConversationID={chat.activeConvID}
          labels={text.chat}
          onCreateConversation={chat.openCreateDialog}
          onSelectConversation={(conversation) => {
            void chat.handleSelectConversation(conversation)
          }}
        />

        <ConversationPane
          conversation={chat.currentConversation}
          userId={chat.userId}
          labels={text.chat}
          messages={chat.displayedMessages}
          input={chat.input}
          searchQuery={chat.searchQuery}
          searching={chat.searching}
          searchActive={chat.searchActive}
          onInputChange={chat.setInput}
          onSearchChange={chat.setSearchQuery}
          onSearch={chat.handleSearch}
          onClearSearch={chat.handleClearSearch}
          onSend={chat.handleSend}
          messagesEndRef={chat.messagesEndRef}
        />
      </div>

      <CreateConversationModal
        open={chat.createDialogOpen}
        users={teamMembers}
        currentUserId={chat.userId}
        labels={text.chat}
        common={text.common}
        onClose={chat.closeCreateDialog}
        onSubmit={chat.handleCreateConversation}
      />
    </section>
  )
}
