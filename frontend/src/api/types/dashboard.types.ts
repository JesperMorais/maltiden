/**
 * Dashboard API Types
 *
 * This file defines the contract between frontend and backend
 * for the dashboard data.
 *
 * API Endpoint: GET /api/dashboard
 * Returns: DashboardResponse
 */

// ============================================
// USER & AUTHENTICATION
// ============================================

export type UserRole = 'owner' | 'member' | 'guest'

export interface User {
  /** Unique user identifier */
  id: string

  /** Display name */
  name: string

  /** Email address (only for members) */
  email?: string

  /** User role - determines access level */
  role: UserRole

  /** Profile picture URL */
  avatarUrl?: string
}

// ============================================
// MEALS & RECIPES
// ============================================

export interface Meal {
  /** Unique meal identifier */
  id: string

  /** Meal/recipe name */
  name: string

  /** Image URL for the meal */
  imageUrl?: string

  /** Emoji fallback if no image */
  emoji?: string

  /** Number of portions */
  portions: number
}

// ============================================
// WEEKLY MENU
// ============================================

export interface MenuDay {
  /** ISO date string (YYYY-MM-DD) */
  date: string

  /** Localized day name (e.g., "Måndag") */
  dayName: string

  /** Short day name (e.g., "Mån") */
  dayShort: string

  /** Planned meal for this day, null if none */
  meal: Meal | null

  /** Whether this is today */
  isToday: boolean

  /** Whether this day is skipped (no cooking) */
  isSkipped: boolean
}

// ============================================
// HOUSEHOLD
// ============================================

export interface HouseholdMember {
  /** Unique member identifier */
  id: string

  /** Display name */
  name: string

  /** Member or guest */
  role: UserRole

  /** Profile picture URL */
  avatarUrl?: string

  /** Whether they're eating today */
  isEatingToday: boolean

  /** Whether they want a lunch box (matlåda) for tomorrow */
  wantsLunchBox: boolean
}

export interface Household {
  /** Unique household identifier */
  id: string

  /** Household display name */
  name: string

  /** Invite code for sharing */
  inviteCode: string

  /** List of household members */
  members: HouseholdMember[]
}

// ============================================
// SHOPPING LIST
// ============================================

export interface ShoppingCategory {
  /** Category name (e.g., "Mejeri") */
  name: string

  /** Number of items in category */
  count: number
}

export interface ShoppingListSummary {
  /** Total number of items */
  totalItems: number

  /** Number of checked/bought items */
  checkedItems: number

  /** Items grouped by category */
  categories: ShoppingCategory[]
}

// ============================================
// DASHBOARD DATA
// ============================================

export interface DashboardData {
  /** Current logged in user */
  user: User

  /** User's household */
  household: Household

  /** Today's planned meal */
  todaysMeal: Meal | null

  /** Full week menu (7 days) */
  weeklyMenu: MenuDay[]

  /** Shopping list summary */
  shoppingList: ShoppingListSummary
}

// ============================================
// API RESPONSE
// ============================================

export interface ApiResponse<T> {
  success: boolean
  data: T
  error?: {
    code: string
    message: string
  }
}

export type DashboardResponse = ApiResponse<DashboardData>
