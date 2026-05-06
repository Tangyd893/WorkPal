import { useMemo, useState } from 'react'
import type { AppTranslations } from '../../i18n'
import type { CreateScheduleInput, Locale, ScheduleEvent, WorkspaceUser } from '../../types/workspace'

interface SchedulePanelProps {
  users: WorkspaceUser[]
  events: ScheduleEvent[]
  locale: Locale
  text: AppTranslations
  getDisplayName: (username: string) => string
  onAddEvent: (draft: CreateScheduleInput) => Promise<void> | void
  onDeleteEvent: (eventId: string) => void
  onShareEvent: (eventId: string) => void
}

interface ScheduleDraftState {
  title: string
  detail: string
  ownerUsername: string
  startsAt: string
  durationMinutes: number
  attendees: string[]
  room: string
}

type ScheduleViewMode = 'list' | 'calendar'

function formatStart(locale: Locale, value: string): string {
  const date = new Date(value)
  return new Intl.DateTimeFormat(locale === 'zh-CN' ? 'zh-CN' : 'en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

function defaultDateTimeValue(): string {
  const now = new Date()
  const local = new Date(now.getTime() - now.getTimezoneOffset() * 60000)
  return local.toISOString().slice(0, 16)
}

function buildInitialDraft(users: WorkspaceUser[]): ScheduleDraftState {
  return {
    title: '',
    detail: '',
    ownerUsername: users[0]?.username ?? '',
    startsAt: defaultDateTimeValue(),
    durationMinutes: 30,
    attendees: [],
    room: '',
  }
}

export default function SchedulePanel({
  users,
  events,
  locale,
  text,
  getDisplayName,
  onAddEvent,
  onDeleteEvent,
  onShareEvent,
}: SchedulePanelProps) {
  const [formOpen, setFormOpen] = useState(false)
  const [draft, setDraft] = useState<ScheduleDraftState>(() => buildInitialDraft(users))
  const [formErrors, setFormErrors] = useState<Record<string, string>>({})
  const [submitting, setSubmitting] = useState(false)
  const [viewMode, setViewMode] = useState<ScheduleViewMode>('list')

  const sortedUsers = useMemo(
    () => [...users].sort((left, right) => (left.nickname || left.username).localeCompare(right.nickname || right.username)),
    [users],
  )

  const toggleAttendee = (username: string) => {
    setDraft((current) => ({
      ...current,
      attendees: current.attendees.includes(username)
        ? current.attendees.filter((item) => item !== username)
        : [...current.attendees, username],
    }))
  }

  const resetDraft = () => {
    setDraft(buildInitialDraft(sortedUsers))
    setFormErrors({})
    setSubmitting(false)
    setFormOpen(false)
  }

  const validateDraft = () => {
    const nextErrors: Record<string, string> = {}
    if (!draft.title.trim()) {
      nextErrors.title = text.validation.titleRequired
    }
    if (!draft.ownerUsername) {
      nextErrors.ownerUsername = text.validation.ownerRequired
    }
    if (new Date(draft.startsAt).getTime() <= Date.now()) {
      nextErrors.startsAt = text.validation.futureStartRequired
    }
    setFormErrors(nextErrors)
    return Object.keys(nextErrors).length === 0
  }

  const handleCreate = async () => {
    if (!validateDraft()) {
      return
    }

    setSubmitting(true)
    try {
      await onAddEvent({
        title: draft.title.trim(),
        detail: draft.detail.trim(),
        ownerUsername: draft.ownerUsername,
        startsAt: new Date(draft.startsAt).toISOString(),
        durationMinutes: Number(draft.durationMinutes) || 30,
        attendees: Array.from(new Set([draft.ownerUsername, ...draft.attendees])),
        room: draft.room.trim() || text.common.unavailable,
      })
      resetDraft()
    } finally {
      setSubmitting(false)
    }
  }

  const calendarDays = useMemo(() => {
    const now = new Date()
    const firstDay = new Date(now.getFullYear(), now.getMonth(), 1)
    const lastDay = new Date(now.getFullYear(), now.getMonth() + 1, 0)
    const leadingDays = firstDay.getDay()
    const totalCells = Math.ceil((leadingDays + lastDay.getDate()) / 7) * 7

    return Array.from({ length: totalCells }, (_, index) => {
      const date = new Date(firstDay)
      date.setDate(index - leadingDays + 1)
      return date
    })
  }, [])

  const eventsByDay = useMemo(() => {
    const grouped = new Map<string, ScheduleEvent[]>()
    events.forEach((event) => {
      const key = new Date(event.startsAt).toISOString().slice(0, 10)
      grouped.set(key, [...(grouped.get(key) ?? []), event])
    })
    return grouped
  }, [events])

  return (
    <section className="module-stack">
      <div className="module-header">
        <div>
          <h2>{text.schedule.title}</h2>
          <p>{text.schedule.subtitle}</p>
        </div>
        <div className="status-row">
          <div className="segmented-control">
            <button
              type="button"
              className={viewMode === 'list' ? 'segment-button active' : 'segment-button'}
              aria-pressed={viewMode === 'list'}
              onClick={() => setViewMode('list')}
            >
              {text.schedule.listView}
            </button>
            <button
              type="button"
              className={viewMode === 'calendar' ? 'segment-button active' : 'segment-button'}
              aria-pressed={viewMode === 'calendar'}
              onClick={() => setViewMode('calendar')}
            >
              {text.schedule.calendarView}
            </button>
          </div>
          <button type="button" className="primary-button" onClick={() => setFormOpen((current) => !current)}>
            {text.schedule.addEvent}
          </button>
        </div>
      </div>

      <div className="banner-info">{text.schedule.addHint}</div>

      {formOpen ? (
        <section className="data-card">
          <div className="form-grid two-column">
            <div className="form-item">
              <label htmlFor="event-title">{text.schedule.titleLabel}</label>
              <input
                id="event-title"
                type="text"
                value={draft.title}
                onChange={(event) => setDraft((current) => ({ ...current, title: event.target.value }))}
                aria-invalid={Boolean(formErrors.title)}
              />
              {formErrors.title ? <span className="error-text">{formErrors.title}</span> : null}
            </div>
            <div className="form-item">
              <label htmlFor="event-owner">{text.schedule.ownerLabel}</label>
              <select
                id="event-owner"
                value={draft.ownerUsername}
                onChange={(event) => setDraft((current) => ({ ...current, ownerUsername: event.target.value }))}
                aria-invalid={Boolean(formErrors.ownerUsername)}
              >
                {sortedUsers.map((user) => (
                  <option key={user.id} value={user.username}>
                    {user.nickname || user.username}
                  </option>
                ))}
              </select>
              {formErrors.ownerUsername ? <span className="error-text">{formErrors.ownerUsername}</span> : null}
            </div>
            <div className="form-item">
              <label htmlFor="event-start">{text.schedule.starts}</label>
              <input
                id="event-start"
                type="datetime-local"
                value={draft.startsAt}
                onChange={(event) => setDraft((current) => ({ ...current, startsAt: event.target.value }))}
                aria-invalid={Boolean(formErrors.startsAt)}
              />
              {formErrors.startsAt ? <span className="error-text">{formErrors.startsAt}</span> : null}
            </div>
            <div className="form-item">
              <label htmlFor="event-duration">{text.schedule.duration}</label>
              <input
                id="event-duration"
                type="number"
                min={15}
                step={15}
                value={draft.durationMinutes}
                onChange={(event) =>
                  setDraft((current) => ({
                    ...current,
                    durationMinutes: Number(event.target.value) || 30,
                  }))
                }
              />
            </div>
            <div className="form-item">
              <label htmlFor="event-room">{text.schedule.room}</label>
              <input
                id="event-room"
                type="text"
                value={draft.room}
                onChange={(event) => setDraft((current) => ({ ...current, room: event.target.value }))}
              />
            </div>
            <div className="form-item two-column-span">
              <label htmlFor="event-detail">{text.schedule.detailLabel}</label>
              <textarea
                id="event-detail"
                rows={3}
                value={draft.detail}
                onChange={(event) => setDraft((current) => ({ ...current, detail: event.target.value }))}
              />
            </div>
            <div className="form-item two-column-span">
              <label>{text.schedule.attendees}</label>
              <div className="checkbox-grid">
                {sortedUsers.map((user) => (
                  <label key={user.id} className="checkbox-pill">
                    <input
                      type="checkbox"
                      checked={draft.attendees.includes(user.username)}
                      onChange={() => toggleAttendee(user.username)}
                    />
                    <span>{user.nickname || user.username}</span>
                  </label>
                ))}
              </div>
            </div>
          </div>
          <div className="task-actions">
            <button type="button" className="secondary-button" onClick={resetDraft}>
              {text.common.cancel}
            </button>
            <button type="button" className="primary-button" onClick={() => void handleCreate()} disabled={submitting}>
              {submitting ? text.common.loading : text.schedule.createAction}
            </button>
          </div>
        </section>
      ) : null}

      {viewMode === 'calendar' ? (
        <div className="calendar-grid">
          {calendarDays.map((day) => {
            const key = day.toISOString().slice(0, 10)
            const dayEvents = eventsByDay.get(key) ?? []
            const inCurrentMonth = day.getMonth() === new Date().getMonth()

            return (
              <section key={key} className={inCurrentMonth ? 'calendar-cell' : 'calendar-cell muted'}>
                <strong>{day.getDate()}</strong>
                <div className="calendar-events">
                  {dayEvents.slice(0, 3).map((event) => (
                    <span key={event.id} className="calendar-event">
                      {event.title}
                    </span>
                  ))}
                  {dayEvents.length > 3 ? <span className="subtle-label">+{dayEvents.length - 3}</span> : null}
                </div>
              </section>
            )
          })}
        </div>
      ) : (
        <div className="list-grid">
        {events.length === 0 ? <div className="empty-panel">{text.schedule.empty}</div> : null}

        {events.map((event) => (
          <article key={event.id} className="data-card">
            <div className="panel-heading">
              <div>
                <h3>{event.title}</h3>
                <p>{event.detail}</p>
              </div>
            </div>
            <dl className="meta-pairs">
              <div>
                <dt>{text.schedule.starts}</dt>
                <dd>{formatStart(locale, event.startsAt)}</dd>
              </div>
              <div>
                <dt>{text.schedule.duration}</dt>
                <dd>
                  {event.durationMinutes} {text.schedule.minutes}
                </dd>
              </div>
              <div>
                <dt>{text.schedule.room}</dt>
                <dd>{event.room}</dd>
              </div>
              <div>
                <dt>{text.schedule.ownerLabel}</dt>
                <dd>{getDisplayName(event.ownerUsername)}</dd>
              </div>
              <div>
                <dt>{text.schedule.attendees}</dt>
                <dd>{event.attendees.map(getDisplayName).join(', ')}</dd>
              </div>
              <div>
                <dt>{text.schedule.sharedCount}</dt>
                <dd>{event.sharedCount}</dd>
              </div>
            </dl>
            <div className="task-actions">
              <button type="button" className="secondary-button" onClick={() => onShareEvent(event.id)}>
                {text.schedule.shareAction}
              </button>
              <button type="button" className="secondary-button" onClick={() => onDeleteEvent(event.id)}>
                {text.schedule.deleteAction}
              </button>
            </div>
          </article>
        ))}
        </div>
      )}
    </section>
  )
}
