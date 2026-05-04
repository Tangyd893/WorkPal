import type { AppTranslations } from '../../i18n'
import type { ScheduleEvent, SharedDocument, WorkspaceSection, WorkspaceTask, WorkspaceUser } from '../../types/workspace'

interface OverviewPanelProps {
  users: WorkspaceUser[]
  tasks: WorkspaceTask[]
  events: ScheduleEvent[]
  documents: SharedDocument[]
  text: AppTranslations
  getDisplayName: (username: string) => string
  onOpenSection: (section: WorkspaceSection) => void
}

export default function OverviewPanel({
  users,
  tasks,
  events,
  documents,
  text,
  getDisplayName,
  onOpenSection,
}: OverviewPanelProps) {
  const activeTasks = tasks.filter((task) => task.status !== 'done')
  const upcomingEvents = events.slice(0, 3)
  const recentDocuments = documents.slice(0, 3)

  return (
    <section className="module-stack">
      <div className="module-header">
        <div>
          <h2>{text.overview.title}</h2>
          <p>{text.overview.subtitle}</p>
        </div>
      </div>

      <div className="metric-grid">
        <button type="button" className="data-card metric-button" onClick={() => onOpenSection('directory')}>
          <span className="metric-label">{text.overview.cards.teammates}</span>
          <strong>{users.length}</strong>
          <span>{text.overview.openSection}</span>
        </button>
        <button type="button" className="data-card metric-button" onClick={() => onOpenSection('tasks')}>
          <span className="metric-label">{text.overview.cards.activeTasks}</span>
          <strong>{activeTasks.length}</strong>
          <span>{text.overview.openSection}</span>
        </button>
        <button type="button" className="data-card metric-button" onClick={() => onOpenSection('schedule')}>
          <span className="metric-label">{text.overview.cards.todayMeetings}</span>
          <strong>{events.length}</strong>
          <span>{text.overview.openSection}</span>
        </button>
        <button type="button" className="data-card metric-button" onClick={() => onOpenSection('files')}>
          <span className="metric-label">{text.overview.cards.sharedFiles}</span>
          <strong>{documents.length}</strong>
          <span>{text.overview.openSection}</span>
        </button>
      </div>

      <section className="data-card">
        <div className="panel-heading">
          <div>
            <h3>{text.overview.quickActions}</h3>
            <p>{text.overview.quickActionsHint}</p>
          </div>
        </div>
        <div className="quick-action-row">
          <button type="button" className="secondary-button" onClick={() => onOpenSection('chat')}>
            {text.navigation.chat}
          </button>
          <button type="button" className="secondary-button" onClick={() => onOpenSection('tasks')}>
            {text.navigation.tasks}
          </button>
          <button type="button" className="secondary-button" onClick={() => onOpenSection('schedule')}>
            {text.navigation.schedule}
          </button>
          <button type="button" className="secondary-button" onClick={() => onOpenSection('files')}>
            {text.navigation.files}
          </button>
          <button type="button" className="secondary-button" onClick={() => onOpenSection('directory')}>
            {text.navigation.directory}
          </button>
        </div>
      </section>

      <div className="overview-grid">
        <section className="data-card">
          <div className="panel-heading">
            <div>
              <h3>{text.overview.sections.priorities}</h3>
              <p>{text.overview.sections.prioritiesHint}</p>
            </div>
            <button type="button" className="secondary-button" onClick={() => onOpenSection('tasks')}>
              {text.overview.openSection}
            </button>
          </div>
          <div className="stack-list">
            {activeTasks.length === 0 ? <div className="empty-panel compact-empty">{text.tasks.empty}</div> : null}
            {activeTasks.slice(0, 4).map((task) => (
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
            <div>
              <h3>{text.overview.sections.agenda}</h3>
              <p>{text.overview.sections.agendaHint}</p>
            </div>
            <button type="button" className="secondary-button" onClick={() => onOpenSection('schedule')}>
              {text.overview.openSection}
            </button>
          </div>
          <div className="stack-list">
            {upcomingEvents.length === 0 ? <div className="empty-panel compact-empty">{text.schedule.empty}</div> : null}
            {upcomingEvents.map((event) => (
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
            <div>
              <h3>{text.overview.sections.docs}</h3>
              <p>{text.overview.sections.docsHint}</p>
            </div>
            <button type="button" className="secondary-button" onClick={() => onOpenSection('files')}>
              {text.overview.openSection}
            </button>
          </div>
          <div className="stack-list">
            {recentDocuments.length === 0 ? <div className="empty-panel compact-empty">{text.files.empty}</div> : null}
            {recentDocuments.map((document) => (
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
