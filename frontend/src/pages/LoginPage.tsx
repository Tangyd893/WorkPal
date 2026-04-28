import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { workpalApi } from '../api/workpal'
import { useAuthStore } from '../hooks/useAuthStore'

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

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setError('')
    setLoading(true)

    try {
      const result = await workpalApi.login(form)
      setAuth(result.token, result.user_id, result.username)
      navigate('/chat', { replace: true })
    } catch (submitError) {
      setError(getErrorMessage(submitError))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="page-container">
      <div className="card">
        <h2 style={{ textAlign: 'center', marginBottom: 8 }}>Sign in to WorkPal</h2>
        <p style={{ textAlign: 'center', color: '#6b7280', marginBottom: 24 }}>
          Use your existing workspace credentials.
        </p>

        <form onSubmit={handleSubmit}>
          <div className="form-item">
            <label htmlFor="username">Username</label>
            <input
              id="username"
              type="text"
              value={form.username}
              onChange={(event) => setForm((current) => ({ ...current, username: event.target.value }))}
              placeholder="Enter your username"
              required
            />
          </div>

          <div className="form-item">
            <label htmlFor="password">Password</label>
            <input
              id="password"
              type="password"
              value={form.password}
              onChange={(event) => setForm((current) => ({ ...current, password: event.target.value }))}
              placeholder="Enter your password"
              required
            />
          </div>

          {error ? <div className="error-msg">{error}</div> : null}

          <div className="form-item" style={{ marginTop: 8 }}>
            <button type="submit" className="btn-primary" disabled={loading}>
              {loading ? 'Signing in...' : 'Sign in'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
