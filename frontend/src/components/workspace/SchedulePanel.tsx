import type { AppTranslations } from '../../i18n'
import type { Locale, ScheduleEvent } from '../../types/workspace'

interface SchedulePanelProps {
  events: ScheduleEvent[]
  locale: Locale
  text: AppTranslations
  getDisplayName: (username: string) => string
}

function formatStart(locale: Locale, value: string): string {
  const date = new Date(value)
  return new Intl.DateTimeFormat(locale === 'zh-CN' ? 'zh-CN' : 'en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

export default function SchedulePanel({ events, locale, text, getDisplayName }: SchedulePanelProps) {
  return (
    <section className="module-stack">
      <div className="module-header">
        <div>
          <h2>{text.schedule.title}</h2>
          <p>{text.schedule.subtitle}</p>
        </div>
      </div>

      <div className="list-grid">
        {events.map((event) => (
          <article key={event.id} className="data-card">
            <div className="panel-heading">
              <h3>{event.title}</h3>
              <p>{event.detail}</p>
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
                <dt>{text.schedule.attendees}</dt>
                <dd>{event.attendees.map(getDisplayName).join(', ')}</dd>
              </div>
            </dl>
          </article>
        ))}
      </div>
    </section>
  )
}
