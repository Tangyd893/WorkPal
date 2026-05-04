import { afterEach, describe, expect, it, vi } from 'vitest'
import { translations } from '../../i18n'
import { render, click } from '../../test/render'
import type { ScheduleEvent } from '../../types/workspace'
import SchedulePanel from './SchedulePanel'

let view: ReturnType<typeof render> | null = null

afterEach(() => {
  view?.unmount()
  view = null
})

const mockEvent: ScheduleEvent = {
  id: 'ws-1',
  title: 'Daily Standup',
  detail: 'Sprint sync meeting.',
  ownerUsername: 'admin',
  startsAt: '2026-06-01T09:00:00Z',
  durationMinutes: 15,
  attendees: ['emma.chen'],
  room: 'Room A',
  sharedCount: 1,
  source: 'custom',
}

const mockUsers = [
  { id: 1, username: 'admin', nickname: 'Admin', email: '', phone: '', status: 1, department_id: 1, department_name: '', employee_id: 1, employee_no: '001', job_title: '', office_location: '', bio: '', created_at: '', updated_at: '' },
  { id: 2, username: 'emma.chen', nickname: 'Emma Chen', email: '', phone: '', status: 1, department_id: 1, department_name: '', employee_id: 2, employee_no: '002', job_title: '', office_location: '', bio: '', created_at: '', updated_at: '' },
]

function renderPanel(events: ScheduleEvent[] = [], onAddEvent = vi.fn()) {
  view = render(
    <SchedulePanel
      users={mockUsers}
      events={events}
      locale="en"
      text={translations.en}
      getDisplayName={(u) => u}
      onAddEvent={onAddEvent}
      onDeleteEvent={vi.fn()}
      onShareEvent={vi.fn()}
    />,
  )
  return { onAddEvent }
}

describe('SchedulePanel', () => {
  it('renders the schedule title and add event button', () => {
    renderPanel()
    expect(view?.container.textContent).toContain('Schedule')
    expect(view?.container.querySelector('.primary-button')?.textContent).toBe('Add event')
  })

  it('opens the create form when add event is clicked', () => {
    renderPanel()
    click(view!.container.querySelector('.primary-button')!)
    expect(view?.container.querySelector('#event-title')).not.toBeNull()
    expect(view?.container.querySelector('#event-start')).not.toBeNull()
    expect(view?.container.querySelector('#event-detail')).not.toBeNull()
  })

  it('renders an event card with title and detail', () => {
    renderPanel([mockEvent])
    expect(view?.container.textContent).toContain('Daily Standup')
    expect(view?.container.textContent).toContain('Sprint sync meeting.')
  })

  it('shows empty state when no events exist', () => {
    renderPanel()
    expect(view?.container.querySelector('.empty-panel')).not.toBeNull()
  })

  it('renders event metadata fields', () => {
    renderPanel([mockEvent])
    expect(view?.container.textContent).toContain('Room A')
    expect(view?.container.textContent).toContain('admin')
  })

  it('renders share and delete buttons for each event', () => {
    renderPanel([mockEvent])
    const buttons = view?.container.querySelectorAll('.task-actions .secondary-button')
    expect(buttons?.length).toBe(2)
    expect(buttons?.[0].textContent).toBe('Share')
    expect(buttons?.[1].textContent).toBe('Delete')
  })
})
