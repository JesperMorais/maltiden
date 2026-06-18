/**
 * Menu API Service
 * Handles menu generation and retrieval
 */

import apiClient from './client'
import { USE_MOCKS } from '@/mocks'
import { mockGenerateMenu, mockGetCurrentMenu, mockSaveMenu } from '@/mocks/menu.mock'
import type { Macros } from './recipes.api'

export interface MenuDay {
  date: string
  recipeId?: string
  recipeName?: string
  emoji?: string
  servings: number
  skip?: boolean
  /** True when this day reuses leftovers from a prior cook day (prep mode). */
  leftover?: boolean
  /** The date ("YYYY-MM-DD") of the cook day a leftovers day draws from. */
  cookDate?: string
  /** Per-serving nutrition for the day's recipe, present only when enriched. */
  macros?: Macros
}

export interface SharedIngredient {
  canonicalName: string
  name: string
  recipeCount: number
}

export interface MenuEconomy {
  sharedIngredients: SharedIngredient[]
  distinctItemsToBuy: number
  totalIngredientRefs: number
}

export interface Menu {
  id: string
  days: MenuDay[]
  economy?: MenuEconomy
}

export interface GenerateMenuRequest {
  days: number
  skipDays?: string[]
  servings: number
  extraPortions?: Record<string, number>
  /** Locked days kept and scored server-side: date "YYYY-MM-DD" → recipeId */
  lockedDays?: Record<string, string>
  /** Batch-cook mode: pair cook days (2× servings) with leftovers days. */
  prepMode?: boolean
  /** Diet/nutrition profile to bias recipe selection, e.g. 'high-protein'. */
  nutritionProfile?: string
  /** Desired protein per serving per day (grams). 0 = no target. */
  proteinTargetPerDay?: number
}

/**
 * Planning preferences persisted server-side and used as generation defaults.
 */
export interface MenuPreferences {
  /** Default prep-mode (batch cooking) state for new menus. */
  prepModeDefault?: boolean
  /** Diet/nutrition profile applied during generation, e.g. 'high-protein' or ''. */
  nutritionProfile?: string
  /** Desired protein per serving per day (grams). 0 = no target. */
  proteinTargetPerDay?: number
}

/** Payload accepted by the update-preferences endpoint. */
export interface UpdateMenuPreferencesRequest {
  prepModeDefault?: boolean
  nutritionProfile?: string
  proteinTargetPerDay?: number
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

export interface SaveMenuDay {
  date: string
  recipeId?: string
  servings: number
  skip?: boolean
  /** True when this day reuses leftovers from a prior cook day (prep mode). */
  leftover?: boolean
  /** The date ("YYYY-MM-DD") of the cook day a leftovers day draws from. */
  cookDate?: string
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
