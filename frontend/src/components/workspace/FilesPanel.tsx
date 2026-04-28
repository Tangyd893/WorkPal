import type { AppTranslations } from '../../i18n'
import type { SharedDocument } from '../../types/workspace'

interface FilesPanelProps {
  documents: SharedDocument[]
  text: AppTranslations
  getDisplayName: (username: string) => string
}

export default function FilesPanel({ documents, text, getDisplayName }: FilesPanelProps) {
  return (
    <section className="module-stack">
      <div className="module-header">
        <div>
          <h2>{text.files.title}</h2>
          <p>{text.files.subtitle}</p>
        </div>
      </div>

      <div className="list-grid">
        {documents.map((document) => (
          <article key={document.id} className="data-card">
            <div className="panel-heading">
              <h3>{document.title}</h3>
              <p>{document.summary}</p>
            </div>
            <div className="status-row">
              <span className="chip">{document.category}</span>
              <span className="chip subtle">{text.files.categories[document.status]}</span>
            </div>
            <dl className="meta-pairs">
              <div>
                <dt>{text.files.owner}</dt>
                <dd>{getDisplayName(document.ownerUsername)}</dd>
              </div>
              <div>
                <dt>{text.files.updated}</dt>
                <dd>{new Date(document.updatedAt).toLocaleString()}</dd>
              </div>
            </dl>
          </article>
        ))}
      </div>
    </section>
  )
}
