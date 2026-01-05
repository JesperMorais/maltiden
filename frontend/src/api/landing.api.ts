import type { LandingPageData } from './types/landing.types'
import { mockLandingData } from '@/mocks'

/**
 * Landing Page API
 *
 * Landing page content is static marketing data.
 * Always uses mock data - no backend endpoint needed.
 */

/**
 * Fetches landing page content
 *
 * @returns Promise<LandingPageData>
 */
export async function getLandingPageData(): Promise<LandingPageData> {
  // Landing page is static content - always use local data
  // No need for a backend endpoint for marketing copy
  await new Promise((resolve) => setTimeout(resolve, 100))
  return mockLandingData
}
