/**
 * Mock Data Index
 *
 * Centralized export of all mock data.
 * Set VITE_USE_REAL_API=true in environment to use real backend.
 */

export const USE_MOCKS = import.meta.env.DEV && !import.meta.env.VITE_USE_REAL_API

export { mockLandingData } from './landing.mock'
export { mockDashboardData, mockGuestDashboardData } from './dashboard.mock'
