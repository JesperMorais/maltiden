/**
 * Auth API Mock Data
 */

import type { AuthResponse } from '@/api/auth.api'

const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

/**
 * Mock login - accepts any valid-looking credentials
 */
export async function mockLogin(email: string, password: string): Promise<AuthResponse> {
  await delay(500)

  // Simulate invalid credentials
  if (password.length < 8) {
    throw { response: { status: 401, data: { error: 'invalid_credentials' } } }
  }

  const name = email.split('@')[0]

  return {
    token: 'mock_jwt_token_' + Date.now(),
    user: {
      id: 'usr_mock_' + Math.random().toString(36).substring(7),
      email,
      name: name.charAt(0).toUpperCase() + name.slice(1),
      householdId: 'hh_mock_123'
    }
  }
}

/**
 * Mock register - creates a new user
 */
export async function mockRegister(name: string, email: string, password: string): Promise<AuthResponse> {
  await delay(800)

  // Simulate email taken
  if (email === 'taken@example.com') {
    throw { response: { status: 400, data: { error: 'email_taken' } } }
  }

  // Simulate password too short
  if (password.length < 8) {
    throw { response: { status: 400, data: { error: 'invalid_password' } } }
  }

  return {
    token: 'mock_jwt_token_' + Date.now(),
    user: {
      id: 'usr_mock_' + Math.random().toString(36).substring(7),
      email,
      name
    }
  }
}
