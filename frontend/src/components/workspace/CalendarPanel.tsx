import { useCallback, useEffect, useState } from 'react'
import type { AppTranslations } from '../../i18n'
import { workpalApi } from '../../api/workpal'
import type { CalendarEvent } from '../../types/workspace'

interface CalendarPanelProps {
  text: AppTranslations
  getDisplayName: (accountUsername: string) => string
}

export default function CalendarPanel({ text, getDisplayName }: CalendarPanelProps) {
  const [events, setEvents] = useState<CalendarEvent[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [title, setTitle] = useState('')
  const [startsAt, setStartsAt] = useState('')
  const [endsAt, setEndsAt] = useState('')
  const [location, setLocation] = useState('')
  const [description, setDescription] = useState('')

  const loadEvents = useCallback(async () => {
    setLoading(true)
    try {
      const data = await workpalApi.listCalendarEvents()
      setEvents(data)
    } catch { /* ignore */ }
    setLoading(false)
  }, [])

  useEffect(() => { loadEvents() }, [loadEvents])

  const handleCreate = async () => {
    if (!title.trim() || !startsAt) return
    try {
      const ev = await workpalApi.createCalendarEvent({
        title,
        starts_at: new Date(startsAt).toISOString(),
        ends_at: endsAt ? new Date(endsAt).toISOString() : new Date(startsAt).toISOString(),
        location,
        description,
      })
      setEvents(prev => [...prev, ev])
      setShowForm(false)
      setTitle(''); setStartsAt(''); setEndsAt(''); setLocation(''); setDescription('')
    } catch { /* ignore */ }
  }

  const handleDelete = async (id: number) => {
    try {
      await workpalApi.deleteCalendarEvent(id)
      setEvents(prev => prev.filter(e => e.id !== id))
    } catch { /* ignore */ }
  }

  if (loading) return <div className="module-surface"><p>{text.common.loading}</p></div>

  return (
    <div className="module-surface">
      <div className="module-header">
        <div>
          <h2>{text.calendar.title}</h2>
          <p>{text.calendar.subtitle}</p>
        </div>
        <button type="button" className="primary-button" onClick={() => setShowForm(true)}>
          {text.calendar.addEvent}
        </button>
      </div>

      {showForm && (
        <div className="card" style={{ marginBottom: 16 }}>
          <input className="input" placeholder={text.calendar.addEvent} value={title} onChange={e => setTitle(e.target.value)} />
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8, marginTop: 8 }}>
            <label>
              <small>{text.calendar.startsAt}</small>
              <input className="input" type="datetime-local" value={startsAt} onChange={e => setStartsAt(e.target.value)} />
            </label>
            <label>
              <small>{text.calendar.endsAt}</small>
              <input className="input" type="datetime-local" value={endsAt} onChange={e => setEndsAt(e.target.value)} />
            </label>
          </div>
          <input className="input" style={{ marginTop: 8 }} placeholder={text.calendar.location} value={location} onChange={e => setLocation(e.target.value)} />
          <textarea className="input" style={{ marginTop: 8 }} rows={3} placeholder={text.schedule.detailLabel} value={description} onChange={e => setDescription(e.target.value)} />
          <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
            <button className="primary-button" onClick={handleCreate}>{text.common.create}</button>
            <button className="secondary-button" onClick={() => setShowForm(false)}>{text.common.cancel}</button>
          </div>
        </div>
      )}

      {events.length === 0 && <p>{text.calendar.noEvents}</p>}
      {events.map(ev => (
        <div key={ev.id} className="card" style={{ marginBottom: 8 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <strong>{ev.title}</strong>
            <button className="danger-button" onClick={() => handleDelete(ev.id)}>{text.common.delete}</button>
          </div>
          <small>
            {new Date(ev.starts_at).toLocaleString()} → {new Date(ev.ends_at).toLocaleString()}
            {ev.location && ` | ${ev.location}`}
            {` | ${getDisplayName(String(ev.organizer_id))}`}
          </small>
          {ev.description && <p style={{ marginTop: 4 }}>{ev.description}</p>}
        </div>
      ))}
    </div>
  )
}
