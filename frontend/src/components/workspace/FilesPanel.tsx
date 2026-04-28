import { useRef } from 'react'
import type { ChangeEvent } from 'react'
import type { AppTranslations } from '../../i18n'
import type { SharedDocument } from '../../types/workspace'

interface FilesPanelProps {
  documents: SharedDocument[]
  text: AppTranslations
  getDisplayName: (username: string) => string
  uploading: boolean
  onUpload: (file: File) => Promise<void>
  onDelete: (document: SharedDocument) => Promise<void> | void
  onShare: (document: SharedDocument) => Promise<void> | void
}

function formatUpdatedAt(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString()
}

export default function FilesPanel({ documents, text, getDisplayName, uploading, onUpload, onDelete, onShare }: FilesPanelProps) {
  const fileInputRef = useRef<HTMLInputElement | null>(null)

  const handleFileChange = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    event.target.value = ''
    if (!file) {
      return
    }

    await onUpload(file)
  }

  return (
    <section className="module-stack">
      <div className="module-header">
        <div>
          <h2>{text.files.title}</h2>
          <p>{text.files.subtitle}</p>
        </div>
        <button type="button" className="primary-button" onClick={() => fileInputRef.current?.click()} disabled={uploading}>
          {uploading ? text.common.loading : text.files.uploadAction}
        </button>
      </div>

      <input ref={fileInputRef} type="file" className="visually-hidden" onChange={(event) => void handleFileChange(event)} />

      <div className="banner-info">{text.files.uploadHint}</div>

      {documents.length === 0 ? (
        <div className="empty-panel">{text.files.empty}</div>
      ) : (
        <div className="list-grid">
          {documents.map((document) => (
            <article key={document.id} className="data-card">
              <div className="panel-heading">
                <div>
                  <h3>{document.title}</h3>
                  <p>{document.summary}</p>
                </div>
              </div>

              <div className="status-row">
                <span className="chip">{document.category}</span>
                <span className="chip subtle">{text.files.categories[document.status]}</span>
                <span className="chip neutral">{document.source === 'seed' ? text.files.sourceSeed : text.files.sourceUpload}</span>
              </div>

              <dl className="meta-pairs">
                <div>
                  <dt>{text.files.owner}</dt>
                  <dd>{getDisplayName(document.ownerUsername)}</dd>
                </div>
                <div>
                  <dt>{text.files.updated}</dt>
                  <dd>{formatUpdatedAt(document.updatedAt)}</dd>
                </div>
                <div>
                  <dt>{text.files.sharedCount}</dt>
                  <dd>{document.sharedCount}</dd>
                </div>
                <div>
                  <dt>{text.files.attachmentLabel}</dt>
                  <dd>{document.attachmentName || text.common.unavailable}</dd>
                </div>
              </dl>

              <div className="task-actions">
                {document.attachmentUrl ? (
                  <a className="secondary-button button-link" href={document.attachmentUrl} target="_blank" rel="noreferrer">
                    {text.files.openAction}
                  </a>
                ) : null}
                <button type="button" className="secondary-button" onClick={() => void onShare(document)}>
                  {text.files.shareAction}
                </button>
                <button type="button" className="secondary-button" onClick={() => void onDelete(document)}>
                  {text.files.deleteAction}
                </button>
              </div>
            </article>
          ))}
        </div>
      )}
    </section>
  )
}
