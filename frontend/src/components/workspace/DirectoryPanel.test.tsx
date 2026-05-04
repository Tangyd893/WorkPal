import { afterEach, describe, expect, it } from 'vitest'
import { translations } from '../../i18n'
import { render } from '../../test/render'
import type { Department } from '../../types/workspace'
import DirectoryPanel from './DirectoryPanel'

let view: ReturnType<typeof render> | null = null

afterEach(() => {
  view?.unmount()
  view = null
})

const mockDepartments: Department[] = [
  { id: 1, code: 'ENG', name: 'Engineering', description: '', parent_id: 0, leader_id: 0, created_at: '', updated_at: '' },
  { id: 2, code: 'DES', name: 'Design', description: '', parent_id: 0, leader_id: 0, created_at: '', updated_at: '' },
]

const mockUsers = [
  { id: 1, username: 'admin', nickname: 'Admin', email: 'admin@example.com', phone: '1234', status: 1, department_id: 1, department_name: 'Engineering', employee_id: 1, employee_no: '001', job_title: 'Platform Engineer', office_location: 'Floor 1', bio: 'System admin.', created_at: '', updated_at: '' },
  { id: 2, username: 'emma.chen', nickname: 'Emma Chen', email: 'emma@example.com', phone: '5678', status: 1, department_id: 1, department_name: 'Engineering', employee_id: 2, employee_no: '002', job_title: 'Designer', office_location: 'Floor 2', bio: 'UI/UX.', created_at: '', updated_at: '' },
]

function renderPanel(users = mockUsers, loading = false) {
  view = render(
    <DirectoryPanel
      users={users}
      departments={mockDepartments}
      query=""
      selectedDepartmentId={0}
      currentUserId={1}
      text={translations.en}
      loading={loading}
      onQueryChange={() => {}}
      onDepartmentChange={() => {}}
    />,
  )
}

describe('DirectoryPanel', () => {
  it('renders the directory title and search input', () => {
    renderPanel()
    expect(view?.container.textContent).toContain('Directory')
    const searchInput = view?.container.querySelector('input[type="search"]')
    expect(searchInput).not.toBeNull()
  })

  it('renders a department filter dropdown', () => {
    renderPanel()
    const select = view?.container.querySelector('.search-shell select')
    expect(select).not.toBeNull()
    const options = select?.querySelectorAll('option')
    expect(options?.length).toBe(3) // "All departments" + 2 departments
  })

  it('renders user cards in the directory grid', () => {
    renderPanel()
    const cards = view?.container.querySelectorAll('.person-card')
    expect(cards.length).toBe(2)
  })

  it('renders user details including email and phone', () => {
    renderPanel([mockUsers[0]])
    expect(view?.container.textContent).toContain('admin@example.com')
    expect(view?.container.textContent).toContain('1234')
  })

  it('highlights the current user card', () => {
    renderPanel()
    const highlighted = view?.container.querySelector('.person-card.highlight')
    expect(highlighted).not.toBeNull()
    expect(highlighted?.textContent).toContain('Admin')
  })

  it('shows the current user chip', () => {
    renderPanel()
    expect(view?.container.textContent).toContain('You')
  })

  it('shows no results message when users array is empty', () => {
    renderPanel([])
    expect(view?.container.textContent).toContain('No staff profiles match the current filters.')
  })

  it('shows loading state when loading is true', () => {
    renderPanel([], true)
    expect(view?.container.textContent).toContain('Loading...')
  })
})
