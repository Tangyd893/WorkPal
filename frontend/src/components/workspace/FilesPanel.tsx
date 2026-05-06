import { useRef, useState } from 'react'
import type { ChangeEvent } from 'react'
import type { AppTranslations } from '../../i18n'
import type { SharedDocument } from '../../types/workspace'
import ProgressBar from '../ProgressBar'

interface FilesPanelProps {
  documents: SharedDocument[]
  text: AppTranslations
  getDisplayName: (username: string) => string
  uploading: boolean
  uploadProgress: number
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

function getPreviewKind(document: SharedDocument): 'image' | 'pdf' | 'other' {
  const value = `${document.attachmentName ?? ''} ${document.attachmentUrl ?? ''}`.toLocaleLowerCase()
  if (/\.(png|jpe?g|gif|webp|bmp|svg)(\?|$|\s)/.test(value)) {
    return 'image'
  }
  if (/\.pdf(\?|$|\s)/.test(value)) {
    return 'pdf'
  }
  return 'other'
}

export default function FilesPanel({
  documents,
  text,
  getDisplayName,
  uploading,
  uploadProgress,
  onUpload,
  onDelete,
  onShare,
}: FilesPanelProps) {
  const fileInputRef = useRef<HTMLInputElement | null>(null)
  const [previewDocument, setPreviewDocument] = useState<SharedDocument | null>(null)

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
        <button
          type="button"
          className="primary-button"
          onClick={() => fileInputRef.current?.click()}
          disabled={uploading}
          aria-busy={uploading}
        >
          {uploading ? text.common.loading : text.files.uploadAction}
        </button>
      </div>

      <input
        ref={fileInputRef}
        type="file"
        className="visually-hidden"
        aria-label={text.files.uploadAction}
        onChange={(event) => void handleFileChange(event)}
      />

      <div className="banner-info" role="status">{text.files.uploadHint}</div>
      {uploading ? <ProgressBar value={uploadProgress} label={text.files.uploadProgress} /> : null}

      {documents.length === 0 ? (
        <div className="empty-panel" role="status">{text.files.empty}</div>
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
                  <button type="button" className="secondary-button" onClick={() => setPreviewDocument(document)}>
                    {text.files.previewAction}
                  </button>
                ) : null}
                {document.attachmentUrl ? (
                  <a className="secondary-button button-link" href={document.attachmentUrl} target="_blank" rel="noreferrer noopener">
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

      {previewDocument?.attachmentUrl ? (
        <dialog className="preview-dialog" open aria-labelledby="file-preview-title">
          <div className="dialog-header">
            <div>
              <h3 id="file-preview-title">{text.files.previewTitle}</h3>
              <p>{previewDocument.title}</p>
            </div>
            <button type="button" className="secondary-button" onClick={() => setPreviewDocument(null)}>
              {text.common.close}
            </button>
          </div>
          <div className="preview-body">
            {getPreviewKind(previewDocument) === 'image' ? (
              <img src={previewDocument.attachmentUrl} alt={previewDocument.title} />
            ) : getPreviewKind(previewDocument) === 'pdf' ? (
              <iframe title={previewDocument.title} src={previewDocument.attachmentUrl} />
            ) : (
              <a className="primary-button button-link" href={previewDocument.attachmentUrl} target="_blank" rel="noreferrer noopener">
                {text.files.openAction}
              </a>
            )}
          </div>
        </dialog>
      ) : null}
    </section>
  )
}
