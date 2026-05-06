import type { AppTranslations } from '../i18n'
import ConversationPane from '../components/chat/ConversationPane'
import ConversationSidebar from '../components/chat/ConversationSidebar'
import CreateConversationModal from '../components/chat/CreateConversationModal'
import GroupDetailsPanel from '../components/chat/GroupDetailsPanel'
import { useChatController } from '../hooks/useChatController'
import type { WorkspaceUser } from '../types/workspace'
import { useNavigate, useParams } from 'react-router-dom'

interface ChatPageProps {
  teamMembers: WorkspaceUser[]
  text: AppTranslations
}

export default function ChatPage({ teamMembers, text }: ChatPageProps) {
  const { conversationId } = useParams<{ conversationId: string }>()
  const navigate = useNavigate()
  const requestedConversationID = conversationId ? Number(conversationId) : undefined
  const chat = useChatController(
    Number.isFinite(requestedConversationID) ? requestedConversationID : undefined,
    (nextConversationID) => navigate(`/workspace/chat/${nextConversationID}`),
  )
  const showGroupDetails = chat.currentConversation?.type === 2

  return (
    <section className="module-surface">
      <div className="module-header">
        <div>
          <h2>{text.chat.title}</h2>
          <p>{text.chat.subtitle}</p>
        </div>
        <div className="status-row">
          <span className={chat.connected ? 'status-badge positive' : 'status-badge critical'} role="status">
            {chat.connected ? text.chat.connectionOn : text.chat.connectionOff}
          </span>
          <span className="subtle-label">{chat.username || text.common.unavailable}</span>
        </div>
      </div>

      {chat.error ? <div className="banner-error" role="alert">{chat.error}</div> : null}
      {chat.notice ? <div className="banner-info" role="status">{chat.notice}</div> : null}

      <div className={showGroupDetails ? 'chat-layout with-details' : 'chat-layout'}>
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
          commonLabels={text.common}
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
          onCommitEdit={chat.handleCommitEdit}
          onRecallMessage={chat.handleRecallMessage}
          messagesEndRef={chat.messagesEndRef}
        />

        {showGroupDetails && chat.currentConversation ? (
          <GroupDetailsPanel
            conversation={chat.currentConversation}
            labels={text.chat}
            common={text.common}
            announcement={chat.announcementDraft}
            announcementSaving={chat.announcementSaving}
            files={chat.groupFiles}
            filesLoading={chat.groupFilesLoading}
            uploading={chat.groupFileUploading}
            uploadProgress={chat.groupUploadProgress}
            onAnnouncementChange={chat.setAnnouncementDraft}
            onSaveAnnouncement={chat.handleSaveAnnouncement}
            onUploadFile={chat.handleUploadGroupFile}
            onDeleteFile={chat.handleDeleteGroupFile}
            onShareFile={chat.handleShareGroupFile}
          />
        ) : null}
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
