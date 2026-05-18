import { useCallback, useEffect, useState } from 'react'
import type { AppTranslations } from '../../i18n'
import { workpalApi } from '../../api/workpal'
import type { Document } from '../../types/workspace'

interface DocsPanelProps {
  text: AppTranslations
  getDisplayName: (accountUsername: string) => string
}

export default function DocsPanel({ text, getDisplayName }: DocsPanelProps) {
  const [docs, setDocs] = useState<Document[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [title, setTitle] = useState('')
  const [isFolder, setIsFolder] = useState(false)
  const [content, setContent] = useState('')
  const [selectedDoc, setSelectedDoc] = useState<Document | null>(null)

  const loadDocs = useCallback(async () => {
    setLoading(true)
    try {
      const data = await workpalApi.listDocuments()
      setDocs(data)
    } catch { /* ignore */ }
    setLoading(false)
  }, [])

  useEffect(() => { loadDocs() }, [loadDocs])

  const handleCreate = async () => {
    if (!title.trim()) return
    try {
      const doc = await workpalApi.createDocument({ title, is_folder: isFolder, content })
      setDocs(prev => [doc, ...prev])
      setShowForm(false)
      setTitle('')
      setContent('')
      setIsFolder(false)
    } catch { /* ignore */ }
  }

  const handleDelete = async (id: number) => {
    try {
      await workpalApi.deleteDocument(id)
      setDocs(prev => prev.filter(d => d.id !== id))
      if (selectedDoc?.id === id) setSelectedDoc(null)
    } catch { /* ignore */ }
  }

  const handleOpen = (doc: Document) => {
    setSelectedDoc(doc)
    if (!doc.is_folder && !doc.content && doc.id) {
      workpalApi.getDocument(doc.id).then(d => setSelectedDoc(d)).catch(() => {})
    }
  }

  if (loading) return <div className="module-surface"><p>{text.common.loading}</p></div>

  return (
    <div className="module-surface">
      <div className="module-header">
        <div>
          <h2>{text.docs.title}</h2>
          <p>{text.docs.subtitle}</p>
        </div>
        <button type="button" className="primary-button" onClick={() => setShowForm(true)}>
          {text.docs.addDoc}
        </button>
      </div>

      {showForm && (
        <div className="card" style={{ marginBottom: 16 }}>
          <input
            className="input"
            placeholder={text.docs.addDoc}
            value={title}
            onChange={e => setTitle(e.target.value)}
          />
          <label style={{ display: 'flex', gap: 8, margin: '8px 0', alignItems: 'center' }}>
            <input type="checkbox" checked={isFolder} onChange={e => setIsFolder(e.target.checked)} />
            {text.docs.addFolder}
          </label>
          {!isFolder && (
            <textarea
              className="input"
              rows={6}
              placeholder={text.docs.content}
              value={content}
              onChange={e => setContent(e.target.value)}
            />
          )}
          <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
            <button type="button" className="primary-button" onClick={handleCreate}>{text.common.create}</button>
            <button type="button" className="secondary-button" onClick={() => setShowForm(false)}>{text.common.cancel}</button>
          </div>
        </div>
      )}

      <div className="files-grid">
        {docs.length === 0 && <p>{text.docs.noDocs}</p>}
        {docs.map(doc => (
          <div key={doc.id} className="file-card">
            <div className="file-card-header">
              <strong>{doc.is_folder ? '📁' : '📄'} {doc.title}</strong>
              <span>{getDisplayName(String(doc.created_by))}</span>
            </div>
            <div className="file-card-actions">
              <button type="button" className="secondary-button" onClick={() => handleOpen(doc)}>{text.common.open}</button>
              <button type="button" className="danger-button" onClick={() => handleDelete(doc.id)}>{text.common.delete}</button>
            </div>
          </div>
        ))}
      </div>

      {selectedDoc && !selectedDoc.is_folder && (
        <div className="card" style={{ marginTop: 16 }}>
          <h3>{selectedDoc.title}</h3>
          <div className="document-content" style={{ whiteSpace: 'pre-wrap', marginTop: 8 }}>
            {typeof selectedDoc.content === 'string' ? selectedDoc.content : JSON.stringify(selectedDoc.content)}
          </div>
          <button type="button" className="secondary-button" style={{ marginTop: 8 }} onClick={() => setSelectedDoc(null)}>
            {text.common.close}
          </button>
        </div>
      )}
    </div>
  )
}
