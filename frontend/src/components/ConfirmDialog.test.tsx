import { afterEach, describe, expect, it, vi } from 'vitest'
import { click, render } from '../test/render'
import ConfirmDialog from './ConfirmDialog'

let view: ReturnType<typeof render> | null = null

afterEach(() => {
  view?.unmount()
  view = null
})

describe('ConfirmDialog', () => {
  it('确认危险操作前调用确认回调', () => {
    const onConfirm = vi.fn()
    view = render(
      <ConfirmDialog
        open
        title="删除任务"
        message="该任务将被删除。"
        confirmText="删除"
        cancelText="取消"
        variant="danger"
        onConfirm={onConfirm}
        onCancel={vi.fn()}
      />,
    )

    expect(view.container.querySelector('[role="dialog"]')?.textContent).toContain('删除任务')
    const confirmButton = Array.from(view.container.querySelectorAll('button')).find((button) => button.textContent === '删除')
    click(confirmButton!)
    expect(onConfirm).toHaveBeenCalledTimes(1)
  })
})
