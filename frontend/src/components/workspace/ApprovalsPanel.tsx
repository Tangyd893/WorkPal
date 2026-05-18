import { useCallback, useEffect, useState } from 'react'
import type { AppTranslations } from '../../i18n'
import { workpalApi } from '../../api/workpal'
import type { ApprovalInstance, ApprovalTemplate } from '../../types/workspace'

interface ApprovalsPanelProps {
  text: AppTranslations
  getDisplayName: (accountUsername: string) => string
}

export default function ApprovalsPanel({ text, getDisplayName }: ApprovalsPanelProps) {
  const [instances, setInstances] = useState<ApprovalInstance[]>([])
  const [templates, setTemplates] = useState<ApprovalTemplate[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [templateId, setTemplateId] = useState(0)
  const [title, setTitle] = useState('')
  const [comment, setComment] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [ts, insts] = await Promise.all([
        workpalApi.listApprovalTemplates(),
        workpalApi.listApprovalInstances(),
      ])
      setTemplates(ts)
      setInstances(insts)
    } catch { /* ignore */ }
    setLoading(false)
  }, [])

  useEffect(() => { load() }, [load])

  const handleSubmit = async () => {
    if (!templateId || !title.trim()) return
    try {
      await workpalApi.createApprovalInstance({ template_id: templateId, title })
      setShowForm(false)
      setTitle('')
      load()
    } catch { /* ignore */ }
  }

  const handleAction = async (id: number, action: string) => {
    try {
      await workpalApi.processApprovalAction(id, { action, comment })
      setComment('')
      load()
    } catch { /* ignore */ }
  }

  if (loading) return <div className="module-surface"><p>{text.common.loading}</p></div>

  return (
    <div className="module-surface">
      <div className="module-header">
        <div>
          <h2>{text.approvals.title}</h2>
          <p>{text.approvals.subtitle}</p>
        </div>
        <button type="button" className="primary-button" onClick={() => setShowForm(true)}>
          {text.approvals.submit}
        </button>
      </div>

      {showForm && (
        <div className="card" style={{ marginBottom: 16 }}>
          <select className="input" value={templateId} onChange={e => setTemplateId(Number(e.target.value))}>
            <option value={0}>{text.approvals.templates}</option>
            {templates.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
          </select>
          <input className="input" style={{ marginTop: 8 }} placeholder={text.approvals.submit} value={title} onChange={e => setTitle(e.target.value)} />
          <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
            <button className="primary-button" onClick={handleSubmit}>{text.approvals.submit}</button>
            <button className="secondary-button" onClick={() => setShowForm(false)}>{text.common.cancel}</button>
          </div>
        </div>
      )}

      {templates.length === 0 && <p>{text.approvals.noTemplates}</p>}
      {instances.length === 0 && <p>{text.approvals.noInstances}</p>}

      {instances.map(inst => (
        <div key={inst.id} className="card" style={{ marginBottom: 8 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <strong>{inst.title}</strong>
            <span className={inst.status === 'pending' ? 'badge pending' : inst.status === 'approved' ? 'badge success' : 'badge danger'}>
              {inst.status === 'pending' ? text.approvals.pending : inst.status === 'approved' ? text.approvals.approve : text.approvals.reject}
            </span>
          </div>
          <small>{getDisplayName(String(inst.submitter_id))} | {new Date(inst.submitted_at).toLocaleString()}</small>
          {inst.status === 'pending' && (
            <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
              <input className="input" placeholder={text.chat.writeMessage} value={comment} onChange={e => setComment(e.target.value)} />
              <button className="primary-button" onClick={() => handleAction(inst.id, 'approve')}>{text.approvals.approve}</button>
              <button className="danger-button" onClick={() => handleAction(inst.id, 'reject')}>{text.approvals.reject}</button>
            </div>
          )}
          {inst.actions && inst.actions.length > 0 && (
            <div style={{ marginTop: 8, borderTop: '1px solid var(--border)', paddingTop: 8 }}>
              {inst.actions.map((a, i) => (
                <div key={i}><small>{a.action} — {getDisplayName(String(a.user_id))} {a.comment ? `: ${a.comment}` : ''}</small></div>
              ))}
            </div>
          )}
        </div>
      ))}
    </div>
  )
}
