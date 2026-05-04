import { afterEach, describe, expect, it, vi } from 'vitest'
import { translations } from '../../i18n'
import { render, click } from '../../test/render'
import type { CreateTaskInput, WorkspaceTask } from '../../types/workspace'
import TasksPanel from './TasksPanel'

let view: ReturnType<typeof render> | null = null

afterEach(() => {
  view?.unmount()
  view = null
})

const mockTask: WorkspaceTask = {
  id: 'ws-1',
  title: 'Design API',
  summary: 'Create RESTful API design document.',
  project: 'Backend',
  ownerUsername: 'admin',
  teammates: ['emma.chen'],
  dueDate: '2026-06-01',
  priority: 'high',
  status: 'planned',
  sharedCount: 2,
  source: 'custom',
}

const mockUsers = [
  { id: 1, username: 'admin', nickname: 'Admin', email: '', phone: '', status: 1, department_id: 1, department_name: '', employee_id: 1, employee_no: '001', job_title: '', office_location: '', bio: '', created_at: '', updated_at: '' },
  { id: 2, username: 'emma.chen', nickname: 'Emma Chen', email: '', phone: '', status: 1, department_id: 1, department_name: '', employee_id: 2, employee_no: '002', job_title: '', office_location: '', bio: '', created_at: '', updated_at: '' },
]

function renderPanel(tasks: WorkspaceTask[] = [], onAddTask = vi.fn()) {
  view = render(
    <TasksPanel
      users={mockUsers}
      tasks={tasks}
      text={translations.en}
      getDisplayName={(u) => u}
      onAdvanceTask={vi.fn()}
      onResetTask={vi.fn()}
      onAddTask={onAddTask}
      onDeleteTask={vi.fn()}
      onShareTask={vi.fn()}
    />,
  )
  return { onAddTask }
}

describe('TasksPanel', () => {
  it('renders the board title and add task button', () => {
    renderPanel()
    expect(view?.container.textContent).toContain('Task board')
    expect(view?.container.querySelector('.primary-button')?.textContent).toBe('Add task')
  })

  it('renders four kanban columns', () => {
    renderPanel()
    const columns = view?.container.querySelectorAll('.board-column')
    expect(columns?.length).toBe(4)
  })

  it('displays tasks in the correct column by status', () => {
    renderPanel([mockTask])
    const plannedColumn = view?.container.querySelectorAll('.board-column')[0]
    expect(plannedColumn?.textContent).toContain('Design API')
  })

  it('shows empty message in columns without tasks', () => {
    renderPanel()
    const emptyMessages = view?.container.querySelectorAll('.compact-empty')
    expect(emptyMessages.length).toBe(4)
  })

  it('opens the create form when add task is clicked', () => {
    renderPanel()
    const addButton = view?.container.querySelector('.primary-button')
    click(addButton!)
    expect(view?.container.querySelector('#task-title')).not.toBeNull()
    expect(view?.container.querySelector('#task-summary')).not.toBeNull()
  })

  it('does not create task when title is empty', () => {
    const onAdd = vi.fn()
    renderPanel([], onAdd)
    click(view!.container.querySelector('.primary-button')!)
    click(view!.container.querySelector('#task-priority')!.parentElement!)
    const createButton = Array.from(view!.container.querySelectorAll('.primary-button')).find(
      (btn) => btn.textContent === 'Create task',
    )
    click(createButton!)
    expect(onAdd).not.toHaveBeenCalled()
  })

  it('renders task card with project and priority chips', () => {
    renderPanel([mockTask])
    expect(view?.container.textContent).toContain('Backend')
    expect(view?.container.textContent).toContain('High')
    expect(view?.container.textContent).toContain('Design API')
  })

  it('render advance button only for non-done tasks', () => {
    renderPanel([mockTask])
    const advanceBtn = Array.from(view!.container.querySelectorAll('button')).find(
      (btn) => btn.textContent === 'Move forward',
    )
    expect(advanceBtn).not.toBeUndefined()
  })
})
