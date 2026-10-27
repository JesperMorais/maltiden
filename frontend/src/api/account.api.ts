/**
 * Account API — GDPR self-service: export my data, delete my account.
 */

import apiClient from './client'
import { USE_MOCKS } from '@/mocks'

/**
 * Download everything Måltiden stores about the signed-in user as a JSON file.
 */
export async function exportMyData(): Promise<Blob> {
  if (USE_MOCKS) {
    const mock = { exportedAt: new Date().toISOString(), user: { name: 'Mock' }, household: null }
    return new Blob([JSON.stringify(mock, null, 2)], { type: 'application/json' })
  }

  const { data } = await apiClient.get<Blob>('/account/export', {
    responseType: 'blob',
    timeout: 20000,
  })
  return data
}

/**
 * Permanently delete the signed-in user's account. The backend re-checks the
 * password and answers 403 `invalid_password` when it is wrong.
 */
export async function deleteMyAccount(password: string): Promise<void> {
  if (USE_MOCKS) {
    return
  }

  await apiClient.delete('/account', { data: { password }, skipAuthRedirect: true })
}
