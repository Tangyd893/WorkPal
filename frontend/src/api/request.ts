import axios from 'axios'

const STORAGE_KEY = 'workpal-auth'

// Read token directly from localStorage to avoid Zustand hydration timing issues
const getStoredToken = (): string | null => {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) {
      const parsed = JSON.parse(raw)
      return parsed.token || null
    }
  } catch {}
  return null
}

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// 请求拦截器：注入 Token
request.interceptors.request.use((config) => {
  const token = getStoredToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器：统一错误处理
request.interceptors.response.use(
  (res) => res.data,
  (err) => {
    const msg = err.response?.data?.message || err.message || '网络错误'
    console.error('API Error:', msg)
    throw new Error(msg)
  }
)

export default request
