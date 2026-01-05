/**
 * Menu API Service
 * Handles menu generation and retrieval
 */

import apiClient from './client'
import { USE_MOCKS } from '@/mocks'
import { mockGenerateMenu, mockGetCurrentMenu } from '@/mocks/menu.mock'

export interface MenuDay {
  date: string
  recipeId?: string
  recipeName?: string
  emoji?: string
  servings: number
  skip?: boolean
}

export interface Menu {
  id: string
  days: MenuDay[]
}

export interface GenerateMenuRequest {
  days: number
  skipDays?: string[]
  servings: number
  extraPortions?: Record<string, number>
}

/**
 * Generate a new weekly menu
 */
export async function generateMenu(request: GenerateMenuRequest): Promise<Menu> {
  if (USE_MOCKS) {
    return mockGenerateMenu(request)
  }

  const { data } = await apiClient.post<Menu>('/menus/generate', request)
  return data
}

/**
 * Get the current active menu
 */
export async function getCurrentMenu(): Promise<Menu | null> {
  if (USE_MOCKS) {
    return mockGetCurrentMenu()
  }

  try {
    const { data } = await apiClient.get<Menu>('/menus/current')
    return data
  } catch (error: unknown) {
    // 404 means no active menu
    if (error && typeof error === 'object' && 'response' in error) {
      const axiosError = error as { response?: { status?: number } }
      if (axiosError.response?.status === 404) {
        return null
      }
    }
    throw error
  }
}
