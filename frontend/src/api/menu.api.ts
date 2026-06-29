/**
 * Menu API Service
 * Handles menu generation and retrieval
 */

import apiClient from './client'
import { USE_MOCKS } from '@/mocks'
import { mockGenerateMenu, mockGetCurrentMenu, mockSaveMenu } from '@/mocks/menu.mock'

/** Per-serving nutrition for a recipe (#248 Phase 3); fields absent when unknown. */
export interface Nutrition {
  calories?: number
  proteinG?: number
  carbsG?: number
  fatG?: number
}

/** Weekly aggregated macro total for a menu (#248 Phase 3). */
export interface MenuNutrition {
  calories: number
  proteinG: number
  carbsG: number
  fatG: number
  /** True when some cooked day lacked nutrition data, so the total is incomplete. */
  partial: boolean
}

export interface MenuDay {
  date: string
  recipeId?: string
  recipeName?: string
  emoji?: string
  servings: number
  skip?: boolean
  /** "batch" marks a cook-once-eat-twice cook-day (batch cooking, #248). */
  prepMode?: string
  /** The cook-day date this day reuses as leftovers, when set. */
  leftoverOf?: string
  /** The recipe's per-serving nutrition for this day, when known (#248 Phase 3). */
  nutrition?: Nutrition
}

/** An ingredient reused across two or more of the week's recipes. */
export interface SharedIngredient {
  name: string
  recipeCount: number
}

export interface Menu {
  id: string
  days: MenuDay[]
  /** Ingredients shared across the week's recipes, most-shared first. */
  sharedIngredients?: SharedIngredient[]
  /** Short Swedish explanation of the week from the AI experience layer (#248 Phase 4). */
  rationale?: string
  /** Aggregated weekly nutrition total, when any day carries macro data (#248). */
  nutrition?: MenuNutrition
}

export interface GenerateMenuRequest {
  days: number
  skipDays?: string[]
  servings: number
  extraPortions?: Record<string, number>
  /** Enable batch cooking for the week: batchable recipes occupy 2 slots (#248). */
  prepMode?: boolean
  /** Free-text Swedish wishes parsed into per-run constraints (#248 Phase 4). */
  wishes?: string
  /** Ask the AI layer to arrange recipes across weekdays + write a rationale (#248 Phase 4). */
  arrange?: boolean
}

export interface GenerateMenuResponse extends Menu {
  /** True when wishes were supplied but the AI layer was unavailable, so they had no effect. */
  wishesIgnored?: boolean
}

/**
 * Generate a new weekly menu
 */
export async function generateMenu(request: GenerateMenuRequest): Promise<GenerateMenuResponse> {
  if (USE_MOCKS) {
    return mockGenerateMenu(request)
  }

  const { data } = await apiClient.post<GenerateMenuResponse>('/menus/generate', request)
  return data
}

export interface SaveMenuDay {
  date: string
  recipeId?: string
  servings: number
  skip?: boolean
  /** Preserve batch-cooking markers on save so they persist to the menu (#248). */
  prepMode?: string
  leftoverOf?: string
}

/**
 * Save exact recipe-day selections to the current menu
 */
export async function saveMenu(days: SaveMenuDay[]): Promise<Menu> {
  if (USE_MOCKS) {
    return mockSaveMenu(days)
  }

  const { data } = await apiClient.put<Menu>('/menus/current', { days })
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
