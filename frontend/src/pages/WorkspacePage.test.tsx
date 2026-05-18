import { afterEach, describe, expect, it, vi } from 'vitest'
import { act } from 'react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { workpalApi } from '../api/workpal'
import { render } from '../test/render'
import type { WorkspaceUser } from '../types/workspace'
import WorkspacePage from './WorkspacePage'

vi.mock('../api/workpal', () => ({
  workpalApi: {
    getMe: vi.fn(),
    listUsers: vi.fn(),
    listDepartments: vi.fn(),
    listUserFiles: vi.fn(),
    listTasks: vi.fn(),
    listSchedule: vi.fn(),
    listNotifications: vi.fn(),
    getUnreadNotificationCount: vi.fn(),
    markNotificationRead: vi.fn(),
    markAllNotificationsRead: vi.fn(),
  },
}))

let view: ReturnType<typeof render> | null = null

afterEach(() => {
  view?.unmount()
  view = null
  vi.clearAllMocks()
})

const user: WorkspaceUser = {
  id: 1,
  username: 'admin',
  nickname: 'Admin',
  email: '',
  phone: '',
  status: 1,
  department_id: 1,
  department_name: '',
  employee_id: 1,
  employee_no: '001',
  job_title: '',
  office_location: '',
  bio: '',
  created_at: '',
  updated_at: '',
}

function mockWorkspaceApi() {
  vi.mocked(workpalApi.getMe).mockResolvedValue(user)
  vi.mocked(workpalApi.listUsers).mockResolvedValue([user])
  vi.mocked(workpalApi.listDepartments).mockResolvedValue([])
  vi.mocked(workpalApi.listUserFiles).mockResolvedValue([])
  vi.mocked(workpalApi.listTasks).mockResolvedValue([])
  vi.mocked(workpalApi.listSchedule).mockResolvedValue([])
  vi.mocked(workpalApi.listNotifications).mockResolvedValue([])
  vi.mocked(workpalApi.getUnreadNotificationCount).mockResolvedValue({ count: 0 })
}

async function flushAsyncWork() {
  await act(async () => {
    await Promise.resolve()
    await Promise.resolve()
  })
}

describe('WorkspacePage', () => {
  it('按路由 section 渲染工作台模块并按需请求任务数据', async () => {
    mockWorkspaceApi()
    view = render(
      <MemoryRouter initialEntries={['/workspace/tasks']}>
        <Routes>
          <Route path="/workspace/:section" element={<WorkspacePage />} />
        </Routes>
      </MemoryRouter>,
    )

    expect(view.container.textContent).toContain('加载中...')
    await flushAsyncWork()

    expect(workpalApi.listTasks).toHaveBeenCalledTimes(1)
    expect(workpalApi.listSchedule).not.toHaveBeenCalled()
    expect(view.container.textContent).toContain('任务')
  })
})
