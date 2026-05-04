import { useRef } from 'react'
import type { ChangeEvent } from 'react'
import type { AppTranslations } from '../../i18n'
import type { Conversation } from '../../types/chat'
import type { ConversationFile } from '../../types/workspace'

interface GroupDetailsPanelProps {
  conversation: Conversation
  labels: AppTranslations['chat']
  common: AppTranslations['common']
  announcement: string
  announcementSaving: boolean
  files: ConversationFile[]
  filesLoading: boolean
  uploading: boolean
  onAnnouncementChange: (value: string) => void
  onSaveAnnouncement: () => Promise<void>
  onUploadFile: (file: File) => Promise<void>
  onDeleteFile: (fileID: number) => Promise<void>
  onShareFile: (fileID: number) => Promise<void>
}

function formatFileTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString()
}

export default function GroupDetailsPanel({
  conversation,
  labels,
  common,
  announcement,
  announcementSaving,
  files,
  filesLoading,
  uploading,
  onAnnouncementChange,
  onSaveAnnouncement,
  onUploadFile,
  onDeleteFile,
  onShareFile,
}: GroupDetailsPanelProps) {
  const fileInputRef = useRef<HTMLInputElement | null>(null)

  const handleFileChange = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    event.target.value = ''
    if (!file) {
      return
    }

    await onUploadFile(file)
  }

  return (
    <aside className="group-details-panel">
      <section className="data-card">
        <div className="panel-heading">
          <div>
            <h3>{conversation.name || labels.groupChat}</h3>
            <p>{labels.groupConversation}</p>
          </div>
        </div>
        <div className="form-item">
          <label htmlFor="group-announcement">{labels.announcementTitle}</label>
          <textarea
            id="group-announcement"
            rows={5}
            value={announcement}
            onChange={(event) => onAnnouncementChange(event.target.value)}
            placeholder={labels.announcementPlaceholder}
          />
        </div>
        <div className="task-actions">
          <button type="button" className="primary-button" onClick={() => void onSaveAnnouncement()} disabled={announcementSaving}>
            {announcementSaving ? common.loading : labels.announcementSave}
          </button>
        </div>
      </section>

      <section className="data-card">
        <div className="panel-heading">
          <div>
            <h3>{labels.groupFilesTitle}</h3>
            <p>{labels.groupConversation}</p>
          </div>
          <button type="button" className="secondary-button" onClick={() => fileInputRef.current?.click()} disabled={uploading}>
            {uploading ? common.loading : labels.uploadFile}
          </button>
        </div>

        <input ref={fileInputRef} type="file" className="visually-hidden" onChange={(event) => void handleFileChange(event)} />

        {filesLoading ? <div className="empty-panel compact-empty">{common.loading}</div> : null}

        {!filesLoading && files.length === 0 ? <div className="empty-panel compact-empty">{labels.noFiles}</div> : null}

        {!filesLoading && files.length > 0 ? (
          <div className="stack-list">
            {files.map((file) => (
              <article key={file.id} className="stack-row">
                <div>
                  <strong>{file.name}</strong>
                  <p>{formatFileTime(file.created_at)}</p>
                </div>
                <div className="stack-row-actions">
                  <a className="secondary-button button-link" href={file.download_path} target="_blank" rel="noreferrer">
                    {common.open}
                  </a>
                  <button type="button" className="secondary-button" onClick={() => void onShareFile(file.id)}>
                    {labels.shareFile}
                  </button>
                  <button type="button" className="secondary-button" onClick={() => void onDeleteFile(file.id)}>
                    {labels.deleteFile}
                  </button>
                </div>
              </article>
            ))}
          </div>
        ) : null}
      </section>
    </aside>
  )
}
