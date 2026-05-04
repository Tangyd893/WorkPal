import { act, type ReactElement } from 'react'
import { createRoot, type Root } from 'react-dom/client'

export function render(ui: ReactElement) {
  const container = document.createElement('div')
  document.body.appendChild(container)

  let root: Root | null = null
  act(() => {
    root = createRoot(container)
    root.render(ui)
  })

  return {
    container,
    unmount: () => {
      if (root) {
        act(() => {
          root?.unmount()
        })
      }
      container.remove()
    },
  }
}

export function click(element: Element): void {
  act(() => {
    element.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true }))
  })
}

export async function changeFile(input: HTMLInputElement, file: File): Promise<void> {
  Object.defineProperty(input, 'files', {
    value: [file],
    configurable: true,
  })

  await act(async () => {
    input.dispatchEvent(new Event('change', { bubbles: true }))
    await Promise.resolve()
  })
}
