import { afterEach, describe, expect, it, vi } from 'vitest'
import { translations } from '../i18n'
import { click, render } from '../test/render'
import type { WorkspaceUser } from '../types/workspace'
import Sidebar from './Sidebar'

let view: ReturnType<typeof render> | null = null

afterEach(() => {
  view?.unmount()
  view = null
})

const user: WorkspaceUser = {
  id: 1,
  username: 'admin',
  nickname: '管理员',
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

describe('Sidebar', () => {
  it('自动展开当前模块所在分组', () => {
    view = render(
      <Sidebar
        activeSection="tasks"
        userInfo={user}
        username="admin"
        labels={translations['zh-CN']}
        onNavigate={vi.fn()}
        onOpenSettings={vi.fn()}
      />,
    )

    const activeGroup = Array.from(view.container.querySelectorAll('.nav-group-button')).find((button) =>
      button.textContent?.includes('工作'),
    )

    expect(activeGroup?.getAttribute('aria-expanded')).toBe('true')
    expect(view.container.textContent).toContain('任务')
  })

  it('点击导航项触发模块切换', () => {
    const onNavigate = vi.fn()
    view = render(
      <Sidebar
        activeSection="overview"
        userInfo={user}
        username="admin"
        labels={translations['zh-CN']}
        onNavigate={onNavigate}
        onOpenSettings={vi.fn()}
      />,
    )

    const chatButton = Array.from(view.container.querySelectorAll('.nav-button')).find((button) => button.textContent === '沟通')
    click(chatButton!)

    expect(onNavigate).toHaveBeenCalledWith('chat')
  })
})
