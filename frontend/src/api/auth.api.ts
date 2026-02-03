/**
 * Auth API Service
 * Handles login, register, and authentication endpoints
 */

import apiClient from './client'
import { USE_MOCKS } from '@/mocks'
import { mockLogin, mockRegister } from '@/mocks/auth.mock'

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  name: string
}

export interface AuthResponse {
  token: string
  user: {
    id: string
    email: string
    name: string
    householdId?: string
  }
}

export interface AuthError {
  error: 'invalid_credentials' | 'email_taken' | 'unauthorized'
  message?: string
}

/**
 * Login with email and password
 */
export async function login(email: string, password: string): Promise<AuthResponse> {
  if (USE_MOCKS) {
    return mockLogin(email, password)
  }

  const { data } = await apiClient.post<AuthResponse>('/auth/login', { email, password })
  return data
}

/**
 * Register a new user
 */
export async function register(name: string, email: string, password: string): Promise<AuthResponse> {
  if (USE_MOCKS) {
    return mockRegister(name, email, password)
  }

  const { data } = await apiClient.post<AuthResponse>('/auth/register', { name, email, password })
  return data
}
