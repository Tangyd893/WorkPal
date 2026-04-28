import type { AppTranslations } from '../../i18n'
import type { TaskStatus, WorkspaceTask } from '../../types/workspace'

interface TasksPanelProps {
  tasks: WorkspaceTask[]
  text: AppTranslations
  getDisplayName: (username: string) => string
  onAdvanceTask: (taskId: string) => void
  onResetTask: (taskId: string) => void
}

const boardOrder: TaskStatus[] = ['planned', 'in_progress', 'review', 'done']

export default function TasksPanel({ tasks, text, getDisplayName, onAdvanceTask, onResetTask }: TasksPanelProps) {
  return (
    <section className="module-stack">
      <div className="module-header">
        <div>
          <h2>{text.tasks.title}</h2>
          <p>{text.tasks.subtitle}</p>
        </div>
      </div>

      <div className="board-grid">
        {boardOrder.map((status) => {
          const columnTasks = tasks.filter((task) => task.status === status)

          return (
            <section key={status} className="board-column">
              <div className="board-column-header">
                <strong>{text.tasks.statuses[status]}</strong>
                <span>{columnTasks.length}</span>
              </div>
              <div className="board-column-body">
                {columnTasks.map((task) => (
                  <article key={task.id} className="task-card">
                    <div className="task-card-top">
                      <span className="chip">{task.project}</span>
                      <span className="chip subtle">{text.tasks.priorities[task.priority]}</span>
                    </div>
                    <strong>{task.title}</strong>
                    <p>{task.summary}</p>
                    <dl className="meta-pairs">
                      <div>
                        <dt>{text.tasks.owner}</dt>
                        <dd>{getDisplayName(task.ownerUsername)}</dd>
                      </div>
                      <div>
                        <dt>{text.tasks.due}</dt>
                        <dd>{task.dueDate}</dd>
                      </div>
                      <div>
                        <dt>{text.tasks.teammates}</dt>
                        <dd>{task.teammates.map(getDisplayName).join(', ')}</dd>
                      </div>
                    </dl>
                    <div className="task-actions">
                      {status !== 'done' ? (
                        <button type="button" className="primary-button" onClick={() => onAdvanceTask(task.id)}>
                          {text.tasks.advance}
                        </button>
                      ) : null}
                      <button type="button" className="secondary-button" onClick={() => onResetTask(task.id)}>
                        {text.tasks.reset}
                      </button>
                    </div>
                  </article>
                ))}
              </div>
            </section>
          )
        })}
      </div>
    </section>
  )
}
