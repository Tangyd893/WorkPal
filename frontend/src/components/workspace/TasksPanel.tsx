import { useMemo, useState } from 'react'
import type { AppTranslations } from '../../i18n'
import type { CreateTaskInput, TaskPriority, TaskStatus, WorkspaceTask, WorkspaceUser } from '../../types/workspace'

interface TasksPanelProps {
  users: WorkspaceUser[]
  tasks: WorkspaceTask[]
  text: AppTranslations
  getDisplayName: (username: string) => string
  onAdvanceTask: (taskId: string) => void
  onResetTask: (taskId: string) => void
  onAddTask: (draft: CreateTaskInput) => void
  onDeleteTask: (taskId: string) => void
  onShareTask: (taskId: string) => void
}

interface TaskDraftState {
  title: string
  summary: string
  project: string
  ownerUsername: string
  dueDate: string
  priority: TaskPriority
  teammates: string[]
}

const boardOrder: TaskStatus[] = ['planned', 'in_progress', 'review', 'done']

function buildInitialDraft(users: WorkspaceUser[]): TaskDraftState {
  return {
    title: '',
    summary: '',
    project: '',
    ownerUsername: users[0]?.username ?? '',
    dueDate: new Date().toISOString().slice(0, 10),
    priority: 'medium',
    teammates: [],
  }
}

export default function TasksPanel({
  users,
  tasks,
  text,
  getDisplayName,
  onAdvanceTask,
  onResetTask,
  onAddTask,
  onDeleteTask,
  onShareTask,
}: TasksPanelProps) {
  const [formOpen, setFormOpen] = useState(false)
  const [draft, setDraft] = useState<TaskDraftState>(() => buildInitialDraft(users))

  const sortedUsers = useMemo(
    () => [...users].sort((left, right) => (left.nickname || left.username).localeCompare(right.nickname || right.username)),
    [users],
  )

  const toggleTeammate = (username: string) => {
    setDraft((current) => ({
      ...current,
      teammates: current.teammates.includes(username)
        ? current.teammates.filter((item) => item !== username)
        : [...current.teammates, username],
    }))
  }

  const resetDraft = () => {
    setDraft(buildInitialDraft(sortedUsers))
    setFormOpen(false)
  }

  const handleCreate = () => {
    if (!draft.title.trim() || !draft.ownerUsername) {
      return
    }

    onAddTask({
      title: draft.title.trim(),
      summary: draft.summary.trim(),
      project: draft.project.trim() || 'General',
      ownerUsername: draft.ownerUsername,
      teammates: draft.teammates.filter((item) => item !== draft.ownerUsername),
      dueDate: draft.dueDate,
      priority: draft.priority,
    })
    resetDraft()
  }

  return (
    <section className="module-stack">
      <div className="module-header">
        <div>
          <h2>{text.tasks.title}</h2>
          <p>{text.tasks.subtitle}</p>
        </div>
        <button type="button" className="primary-button" onClick={() => setFormOpen((current) => !current)}>
          {text.tasks.addTask}
        </button>
      </div>

      <div className="banner-info">{text.tasks.addHint}</div>

      {formOpen ? (
        <section className="data-card">
          <div className="form-grid two-column">
            <div className="form-item">
              <label htmlFor="task-title">{text.tasks.titleLabel}</label>
              <input
                id="task-title"
                type="text"
                value={draft.title}
                onChange={(event) => setDraft((current) => ({ ...current, title: event.target.value }))}
              />
            </div>
            <div className="form-item">
              <label htmlFor="task-project">{text.tasks.projectLabel}</label>
              <input
                id="task-project"
                type="text"
                value={draft.project}
                onChange={(event) => setDraft((current) => ({ ...current, project: event.target.value }))}
              />
            </div>
            <div className="form-item">
              <label htmlFor="task-owner">{text.tasks.owner}</label>
              <select
                id="task-owner"
                value={draft.ownerUsername}
                onChange={(event) => setDraft((current) => ({ ...current, ownerUsername: event.target.value }))}
              >
                {sortedUsers.map((user) => (
                  <option key={user.id} value={user.username}>
                    {user.nickname || user.username}
                  </option>
                ))}
              </select>
            </div>
            <div className="form-item">
              <label htmlFor="task-due">{text.tasks.due}</label>
              <input
                id="task-due"
                type="date"
                value={draft.dueDate}
                onChange={(event) => setDraft((current) => ({ ...current, dueDate: event.target.value }))}
              />
            </div>
            <div className="form-item">
              <label htmlFor="task-priority">{text.tasks.priorityLabel}</label>
              <select
                id="task-priority"
                value={draft.priority}
                onChange={(event) =>
                  setDraft((current) => ({
                    ...current,
                    priority: event.target.value as TaskPriority,
                  }))
                }
              >
                {(['high', 'medium', 'low'] as const).map((priority) => (
                  <option key={priority} value={priority}>
                    {text.tasks.priorities[priority]}
                  </option>
                ))}
              </select>
            </div>
            <div className="form-item two-column-span">
              <label htmlFor="task-summary">{text.tasks.summaryLabel}</label>
              <textarea
                id="task-summary"
                rows={3}
                value={draft.summary}
                onChange={(event) => setDraft((current) => ({ ...current, summary: event.target.value }))}
              />
            </div>
            <div className="form-item two-column-span">
              <label>{text.tasks.teammates}</label>
              <div className="checkbox-grid">
                {sortedUsers.map((user) => (
                  <label key={user.id} className="checkbox-pill">
                    <input
                      type="checkbox"
                      checked={draft.teammates.includes(user.username)}
                      onChange={() => toggleTeammate(user.username)}
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
            <button type="button" className="primary-button" onClick={handleCreate}>
              {text.tasks.createAction}
            </button>
          </div>
        </section>
      ) : null}

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
                {columnTasks.length === 0 ? <div className="empty-panel compact-empty">{text.tasks.empty}</div> : null}

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
                        <dd>{task.teammates.map(getDisplayName).join(', ') || text.common.unavailable}</dd>
                      </div>
                      <div>
                        <dt>{text.tasks.sharedCount}</dt>
                        <dd>{task.sharedCount}</dd>
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
                      <button type="button" className="secondary-button" onClick={() => onShareTask(task.id)}>
                        {text.tasks.shareAction}
                      </button>
                      <button type="button" className="secondary-button" onClick={() => onDeleteTask(task.id)}>
                        {text.tasks.deleteAction}
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
