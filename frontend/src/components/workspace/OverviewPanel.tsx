import type { AppTranslations } from '../../i18n'
import type { ScheduleEvent, SharedDocument, WorkspaceTask, WorkspaceUser } from '../../types/workspace'

interface OverviewPanelProps {
  users: WorkspaceUser[]
  tasks: WorkspaceTask[]
  events: ScheduleEvent[]
  documents: SharedDocument[]
  text: AppTranslations
  getDisplayName: (username: string) => string
}

export default function OverviewPanel({ users, tasks, events, documents, text, getDisplayName }: OverviewPanelProps) {
  const activeTasks = tasks.filter((task) => task.status !== 'done')

  return (
    <section className="module-stack">
      <div className="module-header">
        <div>
          <h2>{text.overview.title}</h2>
          <p>{text.overview.subtitle}</p>
        </div>
      </div>

      <div className="metric-grid">
        <article className="data-card">
          <span className="metric-label">{text.overview.cards.teammates}</span>
          <strong>{users.length}</strong>
        </article>
        <article className="data-card">
          <span className="metric-label">{text.overview.cards.activeTasks}</span>
          <strong>{activeTasks.length}</strong>
        </article>
        <article className="data-card">
          <span className="metric-label">{text.overview.cards.todayMeetings}</span>
          <strong>{events.length}</strong>
        </article>
        <article className="data-card">
          <span className="metric-label">{text.overview.cards.sharedFiles}</span>
          <strong>{documents.length}</strong>
        </article>
      </div>

      <div className="overview-grid">
        <section className="data-card">
          <div className="panel-heading">
            <h3>{text.overview.sections.priorities}</h3>
            <p>{text.overview.sections.prioritiesHint}</p>
          </div>
          <div className="stack-list">
            {activeTasks.map((task) => (
              <article key={task.id} className="stack-row">
                <div>
                  <strong>{task.title}</strong>
                  <p>{task.summary}</p>
                </div>
                <span className="chip">{getDisplayName(task.ownerUsername)}</span>
              </article>
            ))}
          </div>
        </section>

        <section className="data-card">
          <div className="panel-heading">
            <h3>{text.overview.sections.agenda}</h3>
            <p>{text.overview.sections.agendaHint}</p>
          </div>
          <div className="stack-list">
            {events.map((event) => (
              <article key={event.id} className="stack-row">
                <div>
                  <strong>{event.title}</strong>
                  <p>{event.detail}</p>
                </div>
                <span className="chip">{getDisplayName(event.ownerUsername)}</span>
              </article>
            ))}
          </div>
        </section>

        <section className="data-card">
          <div className="panel-heading">
            <h3>{text.overview.sections.docs}</h3>
            <p>{text.overview.sections.docsHint}</p>
          </div>
          <div className="stack-list">
            {documents.map((document) => (
              <article key={document.id} className="stack-row">
                <div>
                  <strong>{document.title}</strong>
                  <p>{document.summary}</p>
                </div>
                <span className="chip">{document.category}</span>
              </article>
            ))}
          </div>
        </section>
      </div>
    </section>
  )
}
