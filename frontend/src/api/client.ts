/**
 * Axios API Client
 * Centralized HTTP client with JWT authentication and error handling
 */

import axios from 'axios'
import type { AxiosError } from 'axios'
import { tokenUtils } from '@/utils/token'

export const API_BASE_URL = import.meta.env.VITE_API_URL || (import.meta.env.PROD ? '' : 'http://localhost:8080')

/**
 * Configured Axios instance with interceptors
 */
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 5000,
  headers: {
    'Content-Type': 'application/json'
  }
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
  }
)

/**
 * Response interceptor - handles common errors
 */
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    // Handle 401 Unauthorized - token expired or invalid
    if (error.response?.status === 401) {
      tokenUtils.remove()
      // Only redirect if not already on login page
      if (!window.location.pathname.includes('/login')) {
        sessionStorage.setItem('session_expired', 'true')
        window.location.href = '/login'
      }
    }

    // Handle network errors
    if (!error.response) {
      console.error('Network error - API may be unavailable')
    }

    return Promise.reject(error)
  }
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
