import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import LoginPage from './LoginPage'
import { useAuthStore } from '../hooks/useAuthStore'
import { usePreferencesStore } from '../hooks/usePreferencesStore'
import { click, render } from '../test/render'

let view: ReturnType<typeof render> | null = null

beforeEach(() => {
  localStorage.clear()
  useAuthStore.getState().logout()
  usePreferencesStore.getState().reset()
  usePreferencesStore.getState().setLocale('en')
})

afterEach(() => {
  view?.unmount()
  view = null
})

describe('LoginPage', () => {
  it('renders a password-manager friendly and accessible sign-in form', () => {
    view = render(
      <MemoryRouter future={{ v7_relativeSplatPath: true, v7_startTransition: true }}>
        <LoginPage />
      </MemoryRouter>,
    )

    const form = view.container.querySelector('form')
    const username = view.container.querySelector<HTMLInputElement>('#username')
    const password = view.container.querySelector<HTMLInputElement>('#password')
    const englishButton = [...view.container.querySelectorAll('button')].find((button) => button.textContent === 'English')
    const lightButton = [...view.container.querySelectorAll('button')].find((button) => button.textContent === 'Light')

    expect(form?.getAttribute('aria-busy')).toBe('false')
    expect(username?.getAttribute('autocomplete')).toBe('username')
    expect(username?.required).toBe(true)
    expect(password?.type).toBe('password')
    expect(password?.getAttribute('autocomplete')).toBe('current-password')
    expect(password?.required).toBe(true)
    expect(englishButton?.getAttribute('aria-pressed')).toBe('true')
    expect(lightButton?.getAttribute('aria-pressed')).toBe('true')
  })

  it('fills a seeded account through an explicitly labelled action', () => {
    view = render(
      <MemoryRouter future={{ v7_relativeSplatPath: true, v7_startTransition: true }}>
        <LoginPage />
      </MemoryRouter>,
    )

    const username = view.container.querySelector<HTMLInputElement>('#username')
    const password = view.container.querySelector<HTMLInputElement>('#password')
    const fillAdmin = view.container.querySelector<HTMLButtonElement>('button[aria-label="Fill account: admin"]')

    expect(fillAdmin).not.toBeNull()
    click(fillAdmin as HTMLButtonElement)

    expect(username?.value).toBe('admin')
    expect(password?.value).toBe('admin123')
  })
})
