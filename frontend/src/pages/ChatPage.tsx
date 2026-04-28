import ChatHeader from '../components/chat/ChatHeader'
import ConversationPane from '../components/chat/ConversationPane'
import ConversationSidebar from '../components/chat/ConversationSidebar'
import CreateConversationModal from '../components/chat/CreateConversationModal'
import { useChatController } from '../hooks/useChatController'

export default function ChatPage() {
  const chat = useChatController()

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column', background: '#f1f5f9' }}>
      <ChatHeader connected={chat.connected} username={chat.username} onLogout={chat.handleLogout} />

      <div style={{ flex: 1, display: 'flex', minHeight: 0 }}>
        <ConversationSidebar
          conversations={chat.conversations}
          activeConversationID={chat.activeConvID}
          onCreateConversation={chat.openCreateDialog}
          onSelectConversation={(conversation) => {
            void chat.handleSelectConversation(conversation)
          }}
        />

        <ConversationPane
          conversation={chat.currentConversation}
          userId={chat.userId}
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
        onClose={chat.closeCreateDialog}
        onSubmit={chat.handleCreateConversation}
      />
    </div>
  )
}
