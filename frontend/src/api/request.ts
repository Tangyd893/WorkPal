import axios, { AxiosHeaders, type AxiosRequestConfig } from 'axios'
import { getStoredToken } from '../utils/authStorage'

interface ApiResponse<T> {
  code: number
  message: string
  data?: T
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
  const token = getStoredToken()
  if (!token) {
    return config
  }

  const headers = AxiosHeaders.from(config.headers)
  headers.set('Authorization', `Bearer ${token}`)
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
