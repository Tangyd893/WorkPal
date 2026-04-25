import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../hooks/useAuthStore'
import request from '../api/request'

interface LoginForm {
  username: string
  password: string
}

export default function LoginPage() {
  const [form, setForm] = useState<LoginForm>({ username: '', password: '' })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const navigate = useNavigate()
  const setAuth = useAuthStore((s) => s.setAuth)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const raw = await request.post('/auth/login', form)
      // Backend wraps response in {code, message, data}, so token is at raw.data.token
      const res = (raw as any).data?.token ? (raw as any).data : (raw as any)
      if (!res.token) throw new Error('登录失败：未收到认证令牌')
      setAuth(res.token, res.user_id, res.username)
      navigate('/', { replace: true })
    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="page-container">
      <div className="card">
        <h2 style={{ textAlign: 'center', marginBottom: 24 }}>WorkPal 登录</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-item">
            <label>用户名</label>
            <input
              type="text"
              value={form.username}
              onChange={(e) => setForm({ ...form, username: e.target.value })}
              placeholder="请输入用户名"
              required
            />
          </div>
          <div className="form-item">
            <label>密码</label>
            <input
              type="password"
              value={form.password}
              onChange={(e) => setForm({ ...form, password: e.target.value })}
              placeholder="请输入密码"
              required
            />
          </div>
          {error && <div className="error-msg">{error}</div>}
          <div className="form-item" style={{ marginTop: 8 }}>
            <button type="submit" className="btn-primary" disabled={loading}>
              {loading ? '登录中...' : '登录'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
