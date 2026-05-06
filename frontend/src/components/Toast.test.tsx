import { afterEach, describe, expect, it } from 'vitest'
import { act } from 'react'
import { render } from '../test/render'
import { useToastStore } from '../hooks/useToastStore'
import ToastViewport from './Toast'

let view: ReturnType<typeof render> | null = null

afterEach(() => {
  view?.unmount()
  view = null
  useToastStore.getState().clearToasts()
})

describe('ToastViewport', () => {
  it('渲染全局通知并支持关闭', () => {
    act(() => {
      useToastStore.getState().addToast({ type: 'success', message: '保存成功', durationMs: 10000 })
    })

    view = render(<ToastViewport />)

    expect(view.container.textContent).toContain('保存成功')
    act(() => {
      view?.container.querySelector<HTMLButtonElement>('.toast-close')?.click()
    })
    expect(useToastStore.getState().toasts).toHaveLength(0)
  })
})
