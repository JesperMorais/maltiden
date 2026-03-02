/**
 * Axios API Client
 * Centralized HTTP client with JWT authentication and error handling
 */

import axios from 'axios'
import type { AxiosError } from 'axios'
import { tokenUtils } from '@/utils/token'

export const API_BASE_URL =
  import.meta.env.VITE_API_URL || (import.meta.env.PROD ? '' : 'http://localhost:8080')

/**
 * Track when a login/register succeeded so we can avoid redirect loops.
 * If /households/me 401s right after login, redirecting back to /login
 * creates an inescapable loop.
 */
let lastAuthSuccessTime = 0
export function markAuthSuccess() {
  lastAuthSuccessTime = Date.now()
}

/**
 * Configured Axios instance with interceptors
 */
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 5000,
  headers: {
    'Content-Type': 'application/json',
  },
})

/**
 * Request interceptor - adds JWT token to all requests
 */
apiClient.interceptors.request.use(
  (config) => {
    const token = tokenUtils.get()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  },
)

/**
 * Response interceptor - handles common errors
 */
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    const requestUrl = error.config?.url ?? ''
    const isAuthEndpoint = requestUrl.startsWith('/auth/')

    if (error.response?.status === 401 && !isAuthEndpoint) {
      const errorCode = (error.response?.data as { error?: string } | undefined)?.error
      console.warn('[Auth] 401 from', requestUrl, '— code:', errorCode)

      tokenUtils.remove()

      // Guard against redirect loops: if we authenticated less than 10s ago,
      // don't redirect — let the calling code handle the error gracefully.
      const timeSinceAuth = Date.now() - lastAuthSuccessTime
      if (timeSinceAuth < 10_000) {
        console.warn('[Auth] 401 shortly after login — skipping redirect to prevent loop')
      } else {
        const path = window.location.pathname
        if (!path.includes('/login') && !path.includes('/register')) {
          sessionStorage.setItem('session_expired', 'true')
          window.location.href = '/login'
        }
      }
    }

    if (!error.response) {
      console.warn('[Auth] Network error — API may be unavailable')
    }

    return Promise.reject(error)
  },
)

/**
 * Health check function (backwards compatible)
 */
export async function checkHealth(): Promise<boolean> {
  try {
    const res = await apiClient.get('/health')
    return res.status === 200
  } catch {
    return false
  }
}

export default apiClient
