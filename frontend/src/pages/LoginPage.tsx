import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { workpalApi } from '../api/workpal'
import { getSeededAccounts } from '../data/workspace'
import { useAuthStore } from '../hooks/useAuthStore'
import { usePreferencesStore } from '../hooks/usePreferencesStore'
import { useI18n } from '../i18n'

interface LoginForm {
  username: string
  password: string
}

function getErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Unable to sign in.'
}

export default function LoginPage() {
  const [form, setForm] = useState<LoginForm>({ username: '', password: '' })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const navigate = useNavigate()
  const setAuth = useAuthStore((state) => state.setAuth)
  const { locale, t } = useI18n()
  const theme = usePreferencesStore((state) => state.theme)
  const setLocale = usePreferencesStore((state) => state.setLocale)
  const setTheme = usePreferencesStore((state) => state.setTheme)

  const seededAccounts = getSeededAccounts(locale)

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setError('')
    setLoading(true)

    try {
      const result = await workpalApi.login(form)
      setAuth(result.token, result.user_id, result.username)
      navigate('/workspace/overview', { replace: true })
    } catch (submitError) {
      setError(getErrorMessage(submitError))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-shell">
      <div className="login-topbar">
        <div className="segmented-control">
          <button
            type="button"
            className={locale === 'en' ? 'segment-button active' : 'segment-button'}
            aria-pressed={locale === 'en'}
            onClick={() => setLocale('en')}
          >
            English
          </button>
          <button
            type="button"
            className={locale === 'zh-CN' ? 'segment-button active' : 'segment-button'}
            aria-pressed={locale === 'zh-CN'}
            onClick={() => setLocale('zh-CN')}
          >
            简体中文
          </button>
        </div>

        <div className="segmented-control">
          <button
            type="button"
            className={theme === 'light' ? 'segment-button active' : 'segment-button'}
            aria-pressed={theme === 'light'}
            onClick={() => setTheme('light')}
          >
            {t.settings.light}
          </button>
          <button
            type="button"
            className={theme === 'dark' ? 'segment-button active' : 'segment-button'}
            aria-pressed={theme === 'dark'}
            onClick={() => setTheme('dark')}
          >
            {t.settings.dark}
          </button>
        </div>
      </div>

      <div className="login-grid">
        <section className="card login-card">
          <div className="panel-heading">
            <h2>{t.login.title}</h2>
            <p>{t.login.subtitle}</p>
          </div>

          <form onSubmit={handleSubmit} className="form-stack" aria-busy={loading}>
            <div className="form-item">
              <label htmlFor="username">{t.login.username}</label>
              <input
                id="username"
                type="text"
                autoComplete="username"
                value={form.username}
                onChange={(event) => setForm((current) => ({ ...current, username: event.target.value }))}
                placeholder={t.login.usernamePlaceholder}
                required
              />
            </div>

            <div className="form-item">
              <label htmlFor="password">{t.login.password}</label>
              <input
                id="password"
                type="password"
                autoComplete="current-password"
                value={form.password}
                onChange={(event) => setForm((current) => ({ ...current, password: event.target.value }))}
                placeholder={t.login.passwordPlaceholder}
                required
              />
            </div>

            {error ? <div className="error-msg" role="alert">{error}</div> : null}

            <button type="submit" className="primary-button block-button" disabled={loading}>
              {loading ? t.login.signingIn : t.login.signIn}
            </button>
          </form>
        </section>

        <section className="card helper-card">
          <div className="panel-heading">
            <h3>{t.login.seededAccounts}</h3>
            <p>{t.login.seededAccountsHint}</p>
          </div>

          <div className="stack-list">
            {seededAccounts.map((account) => (
              <article key={account.username} className="stack-row account-row">
                <div>
                  <strong>{account.nickname}</strong>
                  <p>
                    {account.username} / {account.password}
                  </p>
                  <span>{account.note}</span>
                </div>
                <button
                  type="button"
                  className="secondary-button"
                  aria-label={`${t.login.useAccount}: ${account.username}`}
                  onClick={() => setForm({ username: account.username, password: account.password })}
                >
                  {t.login.useAccount}
                </button>
              </article>
            ))}
          </div>

          <div className="panel-heading helper-list-heading">
            <h3>{t.login.helperTitle}</h3>
          </div>
          <ul className="helper-list">
            {t.login.helperItems.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ul>
        </section>
      </div>
    </div>
  )
}
