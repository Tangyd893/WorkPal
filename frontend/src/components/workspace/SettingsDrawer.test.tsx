import { afterEach, describe, expect, it, vi } from 'vitest'
import { act } from 'react'
import { translations } from '../../i18n'
import { click, render } from '../../test/render'
import SettingsDrawer from './SettingsDrawer'

let view: ReturnType<typeof render> | null = null

afterEach(() => {
  view?.unmount()
  view = null
})

function renderDrawer(overrides: Partial<Parameters<typeof SettingsDrawer>[0]> = {}) {
  const props: Parameters<typeof SettingsDrawer>[0] = {
    open: true,
    locale: 'en',
    theme: 'light',
    soundEnabled: true,
    compactMode: false,
    text: translations.en,
    onClose: vi.fn(),
    onLocaleChange: vi.fn(),
    onThemeChange: vi.fn(),
    onSoundChange: vi.fn(),
    onCompactModeChange: vi.fn(),
    onReset: vi.fn(),
    ...overrides,
  }

  view = render(<SettingsDrawer {...props} />)
  return props
}

describe('SettingsDrawer', () => {
  it('does not render when closed', () => {
    renderDrawer({ open: false })
    expect(view?.container.querySelector('[role="dialog"]')).toBeNull()
  })

  it('renders modal semantics and selected segmented states', () => {
    renderDrawer()

    const dialog = view?.container.querySelector('[role="dialog"]')
    const title = view?.container.querySelector('#settings-drawer-title')
    const englishButton = [...(view?.container.querySelectorAll('button') ?? [])].find((button) => button.textContent === 'English')
    const comfortableButton = [...(view?.container.querySelectorAll('button') ?? [])].find(
      (button) => button.textContent === translations.en.settings.comfortable,
    )

    expect(dialog?.getAttribute('aria-modal')).toBe('true')
    expect(dialog?.getAttribute('aria-labelledby')).toBe('settings-drawer-title')
    expect(title?.textContent).toBe(translations.en.settings.title)
    expect(englishButton?.getAttribute('aria-pressed')).toBe('true')
    expect(comfortableButton?.getAttribute('aria-pressed')).toBe('true')
  })

  it('supports close button and Escape key dismissal', () => {
    const onClose = vi.fn()
    renderDrawer({ onClose })

    const closeButton = [...(view?.container.querySelectorAll('button') ?? [])].find(
      (button) => button.textContent === translations.en.settings.close,
    )

    click(closeButton as HTMLButtonElement)
    act(() => {
      window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    })

    expect(onClose).toHaveBeenCalledTimes(2)
  })
})
