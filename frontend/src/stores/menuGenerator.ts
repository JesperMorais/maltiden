import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { generateMenu, saveMenu } from '@/api/menu.api'
import type { SaveMenuDay, MenuEconomy } from '@/api/menu.api'
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
  /** True when this day reuses leftovers from a prior cook day (prep mode). */
  leftover?: boolean
  /** The date of the cook day this leftovers day draws from. */
  cookDate?: string
  /** True when this day is a cook day (some other day's cookDate === its date). */
  isCookDay?: boolean
}

export interface DraftMenu {
  days: DraftMenuDay[]
}

const PREP_MODE_STORAGE_KEY = 'maltiden_prep_mode'

function loadPrepMode(): boolean {
  try {
    return localStorage.getItem(PREP_MODE_STORAGE_KEY) === 'true'
  } catch {
    return false
  }
}

function persistPrepMode(value: boolean): void {
  try {
    localStorage.setItem(PREP_MODE_STORAGE_KEY, value ? 'true' : 'false')
  } catch {
    // ignore storage failures (e.g. private mode)
  }
}

// ============================================
// DATE UTILITIES
// ============================================

/**
 * Get a 7-day window starting at `startDate` (default: today), today-anchored.
 *
 * IMPORTANT: the backend builds menu day dates from `time.Now()` (today +0..+6),
 * NOT from the Monday of the week. The frontend must use the same anchor so the
 * placeholder dates we roll on match the dates the server returns — otherwise
 * date-keyed merges and locked-day maps never line up except when today is
 * Monday.
 */
function getWeekDates(startDate: Date = new Date()): Date[] {
  const start = new Date(startDate)
  start.setHours(0, 0, 0, 0)

  const dates: Date[] = []
  for (let i = 0; i < 7; i++) {
    const date = new Date(start)
    date.setDate(start.getDate() + i)
    dates.push(date)
  }

  return dates
}

/**
 * Format date as ISO string (YYYY-MM-DD) using the LOCAL calendar date.
 *
 * NOTE: deliberately not `toISOString()` — that converts to UTC and, for
 * Sweden (UTC+1/+2), rolls a late-evening local date to the next/previous
 * day. Frontend-generated date strings (skipDays, week labels) must match the
 * user's local calendar day and the backend's local day.
 */
function formatDateISO(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

/**
 * Parse a local "YYYY-MM-DD" date string into a local Date (midnight local).
 * Using `new Date('YYYY-MM-DD')` would parse as UTC, so we build it explicitly.
 */
function parseDateISO(iso: string): Date {
  const [year, month, day] = iso.split('-').map(Number)
  return new Date(year!, (month ?? 1) - 1, day ?? 1)
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

/**
 * Derive `isCookDay` for each day over the FINAL set of days.
 *
 * A cook day is one whose date is referenced as another day's `cookDate`. This
 * MUST run over the complete merged week (not a regenerated subset) so the
 * "Lagas (2 dagar)" badge always lands on the day that actually owns the batch
 * — even after a regenerate relocates the pair or a locked partner is absent
 * from the response. Mutates each day's `isCookDay` in place.
 */
function deriveCookDays(days: DraftMenuDay[]): void {
  for (const day of days) {
    day.isCookDay = days.some((d) => d.cookDate === day.date)
  }
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
  const economy = ref<MenuEconomy | null>(null)

  const isGenerating = ref(false)
  const isRegenerating = ref(false)
  const isSaving = ref(false)
  const isSlotAnimating = ref(false)
  const error = ref<string | null>(null)

  const weekStart = ref<string>('') // Monday ISO date
  const servings = ref(4) // Default servings

  // Prep mode (batch cooking): per-week toggle, persisted like other planning
  // prefs so it survives reloads and defaults the next generation.
  const prepMode = ref(loadPrepMode())

  /** Set prep mode and persist it. */
  function setPrepMode(value: boolean): void {
    prepMode.value = value
    persistPrepMode(value)
  }

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
   * Initialize the 7-day window with placeholder dates.
   *
   * Today-anchored to match the backend (which builds dates from time.Now()).
   * generateInitialMenu later adopts the exact dates the server returns, but
   * starting from the same anchor keeps the slot-machine animation (which is
   * keyed by date string) consistent across the API call.
   */
  function initializeWeek(startDate?: Date): void {
    const dates = getWeekDates(startDate)

    weekStart.value = formatDateISO(dates[0]!)

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
        skipDays: [],
        lockedDays: {},
        prepMode: prepMode.value
      })

      // Adopt the backend's dates as the single source of truth.
      //
      // Build draftMenu.days directly from the response so that
      // draftMenu.days[].date == the backend's date strings. This keeps locked
      // maps (date→recipeId) and the date-keyed regenerate merge consistent
      // end-to-end, regardless of which weekday "today" is. Weekday labels are
      // derived from the backend date strings so the display stays correct.
      if (draftMenu.value && menu.days) {
        const apiDays = menu.days
        draftMenu.value = {
          days: apiDays.map((apiDay) => {
            const date = parseDateISO(apiDay.date)
            return {
              date: apiDay.date,
              dayName: getDayName(date),
              dayShort: getDayShort(date),
              recipeId: apiDay.recipeId,
              recipeName: apiDay.recipeName,
              emoji: apiDay.emoji,
              servings: apiDay.servings,
              leftover: apiDay.leftover,
              cookDate: apiDay.cookDate
            }
          })
        }
        // Flag cook days over the full week so the "Lagas" badge lands right.
        deriveCookDays(draftMenu.value.days)
        weekStart.value = apiDays[0]?.date ?? weekStart.value
      }

      // Store ingredient economy returned by the server
      economy.value = menu.economy ?? null
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
      // Build locked map (date → recipeId) from the locked Set so the server
      // keeps and scores those days.
      const lockedMap: Record<string, string> = {}
      draftMenu.value.days.forEach((day) => {
        if (lockedDays.value.has(day.date) && day.recipeId) {
          lockedMap[day.date] = day.recipeId
        }
      })

      // Server returns the full week respecting locks.
      const newMenu = await generateMenu({
        days: 7,
        servings: servings.value,
        skipDays: [],
        lockedDays: lockedMap,
        prepMode: prepMode.value
      })

      // Map the response by date: locked days come back unchanged,
      // unlocked days are replaced.
      if (draftMenu.value && newMenu.days) {
        const newDays = newMenu.days
        const newDaysByDate = new Map(newDays.map((d) => [d.date, d]))

        draftMenu.value.days.forEach((day, index) => {
          const newDay = newDaysByDate.get(day.date)
          if (newDay) {
            draftMenu.value!.days[index] = {
              ...day,
              recipeId: newDay.recipeId,
              recipeName: newDay.recipeName,
              emoji: newDay.emoji,
              servings: newDay.servings,
              leftover: newDay.leftover,
              cookDate: newDay.cookDate
            }
          }
        })
        // Re-derive cook days over the FINAL merged week — not just the
        // regenerated subset — so a relocated batch pair flags the new cook day
        // and clears the old one.
        deriveCookDays(draftMenu.value.days)
      }

      // Store the new ingredient economy returned by the server
      economy.value = newMenu.economy ?? null
    } catch (e: unknown) {
      console.error('Failed to regenerate menu:', e)
      error.value = getSwedishMenuError(e, 'Kunde inte generera nya recept. Försök igen.')
      throw e
    } finally {
      isRegenerating.value = false
    }
  }

  /**
   * Toggle lock state for a day.
   *
   * In prep mode, lock/skip acts on the COOK day; a leftovers day follows its
   * cook day and is not independently lockable. We guard here so that even if a
   * leftovers date is passed, it is ignored (its cook day owns the lock).
   */
  function toggleDayLock(date: string): void {
    const day = draftMenu.value?.days.find((d) => d.date === date)
    if (day?.leftover) {
      return
    }
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
        leftover: day.leftover,
        cookDate: day.cookDate,
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
    economy.value = null
    // prepMode is a persisted planning preference (like activeDays) and
    // deliberately survives clearDraft so it defaults the next generation.
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
    economy,
    isGenerating,
    isRegenerating,
    isSaving,
    isSlotAnimating,
    isLoading,
    error,
    weekStart,
    servings,
    prepMode,

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
    setPrepMode,
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
