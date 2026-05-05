import axios, { AxiosHeaders, type AxiosRequestConfig } from 'axios'
import { getStoredToken } from '../utils/authStorage'

interface ApiResponse<T> {
  code: number
  message: string
  data?: T
}

export const TRACE_ID_HEADER = 'X-Trace-ID'
export const TRACE_PARENT_HEADER = 'traceparent'

export function createTraceID(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.getRandomValues === 'function') {
    const bytes = new Uint8Array(16)
    crypto.getRandomValues(bytes)
    return Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('')
  }

  return `${Date.now().toString(16).padStart(16, '0')}${Math.random().toString(16).slice(2, 18).padEnd(16, '0')}`.slice(0, 32)
}

function createTraceParent(traceID: string): string {
  return `00-${traceID}-0000000000000001-01`
}

function isApiResponse<T>(body: unknown): body is ApiResponse<T> {
  return typeof body === 'object' && body !== null && 'code' in body && typeof (body as { code: unknown }).code === 'number'
}

function getErrorMessage(error: unknown): string {
  if (axios.isAxiosError(error)) {
    const data = error.response?.data
    if (data && typeof data === 'object' && 'message' in data && typeof (data as { message: unknown }).message === 'string') {
      return (data as { message: string }).message
    }

    return error.message || 'Network request failed.'
  }

  if (error instanceof Error) {
    return error.message
  }

  return 'Network request failed.'
}

function unwrapApiBody<T>(body: unknown): T {
  if (!isApiResponse(body)) {
    return body as T
  }

  if (body.code !== 0) {
    throw new Error(body.message || 'Request failed.')
  }

  return (body.data ?? null) as T
}

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

request.interceptors.request.use((config) => {
  const headers = AxiosHeaders.from(config.headers)
  if (!headers.get(TRACE_ID_HEADER)) {
    const traceID = createTraceID()
    headers.set(TRACE_ID_HEADER, traceID)
    headers.set(TRACE_PARENT_HEADER, createTraceParent(traceID))
  } else if (!headers.get(TRACE_PARENT_HEADER)) {
    const traceID = String(headers.get(TRACE_ID_HEADER))
    if (/^[a-f0-9]{32}$/i.test(traceID)) {
      headers.set(TRACE_PARENT_HEADER, createTraceParent(traceID))
    }
  }

  const token = getStoredToken()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  config.headers = headers
  return config
})

request.interceptors.response.use(undefined, (error) => Promise.reject(new Error(getErrorMessage(error))))

export async function apiGet<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
  const response = await request.get<unknown>(url, config)
  return unwrapApiBody<T>(response.data)
}

export async function apiPost<TResponse, TBody = unknown>(
  url: string,
  body?: TBody,
  config?: AxiosRequestConfig<TBody>,
): Promise<TResponse> {
  const response = await request.post<unknown>(url, body, config)
  return unwrapApiBody<TResponse>(response.data)
}

export async function apiPut<TResponse, TBody = unknown>(
  url: string,
  body?: TBody,
  config?: AxiosRequestConfig<TBody>,
): Promise<TResponse> {
  const response = await request.put<unknown>(url, body, config)
  return unwrapApiBody<TResponse>(response.data)
}

export async function apiDelete<TResponse>(url: string, config?: AxiosRequestConfig): Promise<TResponse> {
  const response = await request.delete<unknown>(url, config)
  return unwrapApiBody<TResponse>(response.data)
}

export default request
