import type { LandingPageData, LandingPageResponse } from './types/landing.types'
import { USE_MOCKS, mockLandingData } from '@/mocks'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

/**
 * Landing Page API
 *
 * Provides functions to fetch landing page content.
 * Automatically uses mock data in development unless VITE_USE_REAL_API is set.
 *
 * Backend endpoint required:
 * GET /api/landing -> returns LandingPageResponse
 */

/**
 * Fetches landing page content
 *
 * @returns Promise<LandingPageData>
 * @throws Error if API call fails
 */
export async function getLandingPageData(): Promise<LandingPageData> {
  // Use mock data in development
  if (USE_MOCKS) {
    // Simulate network delay for realistic testing
    await new Promise((resolve) => setTimeout(resolve, 300))
    return mockLandingData
  }

  // Real API call
  const response = await fetch(`${API_URL}/api/landing`)

  if (!response.ok) {
    throw new Error(`Failed to fetch landing data: ${response.status}`)
  }

  const result: LandingPageResponse = await response.json()

  if (!result.success) {
    throw new Error(result.error?.message || 'Unknown error')
  }

  return result.data
}
