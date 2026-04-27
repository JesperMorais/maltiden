import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { generateMenu, saveMenu } from '@/api/menu.api'
import type { SaveMenuDay } from '@/api/menu.api'
import { useDashboardStore } from './dashboard'
import { useToast } from '@/composables/useToast'
import { useRouter } from 'vue-router'
import { isAxiosError } from 'axios'

function getSwedishMenuError(e: unknown, fallback: string): string {
  if (isAxiosError(e)) {
    const code = e.response?.data?.error as string | undefined
    if (code === 'no_recipes_available') return 'Inga recept tillgängliga — lägg till recept först'
    if (code === 'invalid_days') return 'Ogiltigt antal dagar'
    if (code === 'invalid_servings') return 'Ogiltigt antal portioner'
  }
  return fallback
}

/**
 * Menu Generator Store
 *
 * Manages the menu generation workflow including:
 * - Draft menu state
 * - Locked days tracking
 * - Generation and regeneration logic
 * - Save functionality
 */

// ============================================
// TYPES
// ============================================

export interface DraftMenuDay {
  date: string // ISO date (YYYY-MM-DD)
  dayName: string // "Måndag", "Tisdag", etc.
  dayShort: string // "Mån", "Tis", etc.
  recipeId?: string
  recipeName?: string
  emoji?: string
  servings: number
}

export interface DraftMenu {
  days: DraftMenuDay[]
}

// ============================================
// DATE UTILITIES
// ============================================

/**
 * Get the Monday of the current week (or a specific week)
 */
function getMonday(date: Date = new Date()): Date {
  const d = new Date(date)
  const day = d.getDay()
  const diff = d.getDate() - day + (day === 0 ? -6 : 1) // Adjust when day is Sunday
  d.setDate(diff)
  d.setHours(0, 0, 0, 0)
  return d
}

/**
 * Get Monday–Sunday dates for a given week (full 7-day week).
 */
function getWeekDates(startDate?: Date): Date[] {
  const monday = getMonday(startDate)
  const dates: Date[] = []

  for (let i = 0; i < 7; i++) {
    const date = new Date(monday)
    date.setDate(monday.getDate() + i)
    dates.push(date)
  }

  return dates
}

/**
 * Format date as ISO string (YYYY-MM-DD)
 */
function formatDateISO(date: Date): string {
  return date.toISOString().split('T')[0]!
}

/**
 * Get Swedish day name
 */
function getDayName(date: Date): string {
  const dayNames = ['Söndag', 'Måndag', 'Tisdag', 'Onsdag', 'Torsdag', 'Fredag', 'Lördag']
  return dayNames[date.getDay()]!
}

/**
 * Get Swedish day short name
 */
function getDayShort(date: Date): string {
  const dayShorts = ['Sön', 'Mån', 'Tis', 'Ons', 'Tor', 'Fre', 'Lör']
  return dayShorts[date.getDay()]!
}

// ============================================
// STORE DEFINITION
// ============================================

export const useMenuGeneratorStore = defineStore('menuGenerator', () => {
  const dashboardStore = useDashboardStore()
  const toast = useToast()

  // ============================================
  // STATE
  // ============================================

  const draftMenu = ref<DraftMenu | null>(null)
  const lockedDays = ref<Set<string>>(new Set())

  const isGenerating = ref(false)
  const isRegenerating = ref(false)
  const isSaving = ref(false)
  const isSlotAnimating = ref(false)
  const error = ref<string | null>(null)

  const weekStart = ref<string>('') // Monday ISO date
  const servings = ref(4) // Default servings

  // ============================================
  // GETTERS
  // ============================================

  /**
   * Get days in order (Mon-Fri)
   */
  const orderedDays = computed<DraftMenuDay[]>(() => {
    return draftMenu.value?.days ?? []
  })

  /**
   * Check if a specific day is locked
   */
  const isDayLocked = computed(() => {
    return (date: string) => lockedDays.value.has(date)
  })

  /**
   * Count locked days
   */
  const lockedDaysCount = computed(() => {
    return lockedDays.value.size
  })

  /**
   * Count unlocked days
   */
  const unlockedDaysCount = computed(() => {
    if (!draftMenu.value) return 0
    return draftMenu.value.days.length - lockedDays.value.size
  })

  /**
   * Check if all days are locked
   */
  const allDaysLocked = computed(() => {
    if (!draftMenu.value) return false
    return lockedDays.value.size === draftMenu.value.days.length
  })

  /**
   * Check if menu exists
   */
  const hasMenu = computed(() => {
    return draftMenu.value !== null && draftMenu.value.days.some((day) => day.recipeId)
  })

  /**
   * Get array of unlocked dates for API
   */
  const unlockedDates = computed(() => {
    if (!draftMenu.value) return []
    return draftMenu.value.days
      .filter((day) => !lockedDays.value.has(day.date))
      .map((day) => day.date)
  })

  /**
   * Check if ready to save (at least one day with recipe)
   */
  const isReadyToSave = computed(() => {
    if (!draftMenu.value) return false
    return draftMenu.value.days.some((day) => day.recipeId)
  })

  /**
   * Check if any loading state is active
   */
  const isLoading = computed(() => {
    return isGenerating.value || isRegenerating.value || isSaving.value
  })

  // ============================================
  // ACTIONS
  // ============================================

  /**
   * Initialize week dates (Monday-Friday)
   */
  function initializeWeek(startDate?: Date): void {
    const monday = getMonday(startDate)
    weekStart.value = formatDateISO(monday)

    const dates = getWeekDates(startDate)

    // Create empty draft menu with dates
    draftMenu.value = {
      days: dates.map((date) => ({
        date: formatDateISO(date),
        dayName: getDayName(date),
        dayShort: getDayShort(date),
        servings: servings.value
      }))
    }

    // Clear locked days
    lockedDays.value.clear()
  }

  /**
   * Generate initial menu (full week, Mon–Sun)
   */
  async function generateInitialMenu(): Promise<void> {
    isGenerating.value = true
    error.value = null

    try {
      // Initialize week if not already done
      if (!draftMenu.value) {
        initializeWeek()
      }

      // Call API to generate menu
      const menu = await generateMenu({
        days: 7,
        servings: servings.value,
        skipDays: []
      })

      // Map API response to draft menu days
      if (draftMenu.value && menu.days) {
        menu.days.forEach((apiDay, index) => {
          if (draftMenu.value && draftMenu.value.days[index]) {
            draftMenu.value.days[index] = {
              ...draftMenu.value.days[index],
              recipeId: apiDay.recipeId,
              recipeName: apiDay.recipeName,
              emoji: apiDay.emoji,
              servings: apiDay.servings
            }
          }
        })
      }
    } catch (e: unknown) {
      console.error('Failed to generate menu:', e)
      error.value = getSwedishMenuError(e, 'Kunde inte generera meny. Försök igen.')
      throw e
    } finally {
      isGenerating.value = false
    }
  }

  /**
   * Regenerate only unlocked days
   */
  async function regenerateUnlockedDays(): Promise<void> {
    if (!draftMenu.value || allDaysLocked.value) {
      return
    }

    isRegenerating.value = true
    error.value = null

    try {
      // Generate new menu across the full week
      const newMenu = await generateMenu({
        days: 7,
        servings: servings.value,
        skipDays: []
      })

      // Merge: Keep locked days, replace unlocked days with new recipes
      if (draftMenu.value && newMenu.days) {
        let newDayIndex = 0

        draftMenu.value.days.forEach((day, index) => {
          if (!lockedDays.value.has(day.date)) {
            // Unlocked - replace with new recipe
            const newDay = newMenu.days[newDayIndex]
            if (newDay) {
              draftMenu.value!.days[index] = {
                ...day,
                recipeId: newDay.recipeId,
                recipeName: newDay.recipeName,
                emoji: newDay.emoji,
                servings: newDay.servings
              }
            }
            newDayIndex++
          }
          // Locked - keep unchanged
        })
      }
    } catch (e: unknown) {
      console.error('Failed to regenerate menu:', e)
      error.value = getSwedishMenuError(e, 'Kunde inte generera nya recept. Försök igen.')
      throw e
    } finally {
      isRegenerating.value = false
    }
  }

  /**
   * Toggle lock state for a day
   */
  function toggleDayLock(date: string): void {
    if (lockedDays.value.has(date)) {
      lockedDays.value.delete(date)
    } else {
      lockedDays.value.add(date)
    }
  }

  /**
   * Save menu to backend and navigate back
   */
  async function saveDraftMenu(router: ReturnType<typeof useRouter>): Promise<boolean> {
    if (!draftMenu.value || !isReadyToSave.value) {
      return false
    }

    isSaving.value = true
    error.value = null

    try {
      const days: SaveMenuDay[] = draftMenu.value.days.map((day) => ({
        date: day.date,
        recipeId: day.recipeId,
        servings: day.servings,
        skip: !day.recipeId,
      }))

      await saveMenu(days)

      // Refresh dashboard to show new menu
      await dashboardStore.fetchDashboard(true)

      toast.success('Menyn har sparats!')

      // Clear draft
      clearDraft()

      // Navigate to dashboard
      router.push({ name: 'dashboard' })

      return true
    } catch {
      error.value = 'Kunde inte spara menyn. Försök igen.'
      toast.error('Kunde inte spara menyn. Försök igen.')
      return false
    } finally {
      isSaving.value = false
    }
  }

  /**
   * Clear draft and reset state
   */
  function clearDraft(): void {
    draftMenu.value = null
    lockedDays.value.clear()
    weekStart.value = ''
    error.value = null
  }

  /**
   * Set error message
   */
  function setError(message: string): void {
    error.value = message
  }

  /**
   * Clear error
   */
  function clearError(): void {
    error.value = null
  }

  /**
   * Mark slot animation as started
   */
  function startSlotAnimation(): void {
    isSlotAnimating.value = true
  }

  /**
   * Called by view when slot animation fully completes
   */
  function onSlotAnimationComplete(): void {
    isSlotAnimating.value = false
  }

  // ============================================
  // RETURN
  // ============================================

  return {
    // State
    draftMenu,
    lockedDays,
    isGenerating,
    isRegenerating,
    isSaving,
    isSlotAnimating,
    isLoading,
    error,
    weekStart,
    servings,

    // Getters
    orderedDays,
    isDayLocked,
    lockedDaysCount,
    unlockedDaysCount,
    allDaysLocked,
    hasMenu,
    unlockedDates,
    isReadyToSave,

    // Actions
    initializeWeek,
    generateInitialMenu,
    regenerateUnlockedDays,
    toggleDayLock,
    saveDraftMenu,
    clearDraft,
    setError,
    clearError,
    startSlotAnimation,
    onSlotAnimationComplete
  }
})
