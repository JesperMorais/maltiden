import { ref, computed, type Ref } from 'vue'

// ============================================
// RECIPE DISPLAY POOL
// ============================================

const RECIPE_DISPLAY_POOL = [
  { recipeName: 'Pasta Carbonara', emoji: '🍝' },
  { recipeName: 'Kycklingwok', emoji: '🥘' },
  { recipeName: 'Tacos', emoji: '🌮' },
  { recipeName: 'Laxfilé med potatis', emoji: '🐟' },
  { recipeName: 'Köttfärssås', emoji: '🍖' },
  { recipeName: 'Vegetarisk curry', emoji: '🥗' },
  { recipeName: 'Pizza', emoji: '🍕' },
  { recipeName: 'Kycklingsallad', emoji: '🥙' },
  { recipeName: 'Pannkakor', emoji: '🥞' },
  { recipeName: 'Soppa', emoji: '🍲' },
]

// ============================================
// TIMING CONSTANTS
// ============================================

const CYCLE_INTERVAL = 120 // ms between recipe swaps while rolling
const MIN_ROLL_TIME = 800 // minimum rolling before first card can land
const STAGGER_DELAY = 500 // ms between each day landing
const LAND_BOUNCE_MS = 350 // css bounce animation duration

// Reduced motion overrides
const REDUCED_STAGGER = 50

// ============================================
// TYPES
// ============================================

export interface DisplayRecipe {
  recipeName: string
  emoji: string
}

interface SlotState {
  isRolling: boolean
  hasLanded: boolean
  currentDisplayRecipe: DisplayRecipe
}

type AnimationPhase = 'idle' | 'rolling' | 'landing' | 'complete'

// ============================================
// COMPOSABLE
// ============================================

export function useSlotMachine() {
  const slotStates = ref<Map<string, SlotState>>(new Map())
  const animationPhase = ref<AnimationPhase>('idle')
  const cycleIntervals = new Map<string, ReturnType<typeof setInterval>>()
  let rollingStartTime = 0

  // Check reduced motion preference
  const prefersReducedMotion = ref(false)
  if (typeof window !== 'undefined') {
    const mql = window.matchMedia('(prefers-reduced-motion: reduce)')
    prefersReducedMotion.value = mql.matches
    mql.addEventListener('change', (e) => {
      prefersReducedMotion.value = e.matches
    })
  }

  // ============================================
  // HELPERS
  // ============================================

  function getRandomRecipe(excludeEmoji?: string): DisplayRecipe {
    let recipe: DisplayRecipe
    do {
      recipe = RECIPE_DISPLAY_POOL[Math.floor(Math.random() * RECIPE_DISPLAY_POOL.length)]!
    } while (recipe.emoji === excludeEmoji && RECIPE_DISPLAY_POOL.length > 1)
    return recipe
  }

  // ============================================
  // ACTIONS
  // ============================================

  /**
   * Start the rolling animation for given dates.
   * Locked dates are skipped.
   */
  function startRolling(dates: string[], lockedDates: Set<string>): void {
    // Clean up any previous animation
    reset()

    animationPhase.value = 'rolling'
    rollingStartTime = Date.now()

    for (const date of dates) {
      if (lockedDates.has(date)) continue

      const initialRecipe = getRandomRecipe()
      const state: SlotState = {
        isRolling: true,
        hasLanded: false,
        currentDisplayRecipe: { ...initialRecipe },
      }
      slotStates.value.set(date, state)

      // Don't cycle in reduced motion mode
      if (!prefersReducedMotion.value) {
        const interval = setInterval(() => {
          const s = slotStates.value.get(date)
          if (s && s.isRolling) {
            const newRecipe = getRandomRecipe(s.currentDisplayRecipe.emoji)
            s.currentDisplayRecipe = { ...newRecipe }
          }
        }, CYCLE_INTERVAL)
        cycleIntervals.set(date, interval)
      }
    }
  }

  /**
   * Land each day sequentially (left-to-right in the dates array order).
   * Returns a promise that resolves when all landings are complete.
   */
  async function landSequentially(
    finalRecipes: Map<string, DisplayRecipe>,
    lockedDates: Set<string>
  ): Promise<void> {
    animationPhase.value = 'landing'

    // Ensure minimum roll time has elapsed
    const elapsed = Date.now() - rollingStartTime
    const minTime = prefersReducedMotion.value ? 200 : MIN_ROLL_TIME
    if (elapsed < minTime) {
      await new Promise((r) => setTimeout(r, minTime - elapsed))
    }

    // Get dates that are rolling (in order)
    const datesToLand = [...slotStates.value.keys()].filter(
      (date) => !lockedDates.has(date) && slotStates.value.get(date)?.isRolling
    )

    const stagger = prefersReducedMotion.value ? REDUCED_STAGGER : STAGGER_DELAY

    for (let i = 0; i < datesToLand.length; i++) {
      const date = datesToLand[i]!
      const final = finalRecipes.get(date)

      // Stop the cycling interval
      const interval = cycleIntervals.get(date)
      if (interval) {
        clearInterval(interval)
        cycleIntervals.delete(date)
      }

      // Set final recipe and mark as landed
      const state = slotStates.value.get(date)
      if (state && final) {
        state.currentDisplayRecipe = { ...final }
        state.isRolling = false
        state.hasLanded = true
      }

      // Wait stagger delay before landing next card
      if (i < datesToLand.length - 1) {
        await new Promise((r) => setTimeout(r, stagger))
      }
    }

    // Wait for last landing animation to finish
    const bounceDuration = prefersReducedMotion.value ? 50 : LAND_BOUNCE_MS
    await new Promise((r) => setTimeout(r, bounceDuration))

    animationPhase.value = 'complete'
  }

  /**
   * Reset all animation state and clean up intervals.
   */
  function reset(): void {
    // Clear all intervals
    for (const interval of cycleIntervals.values()) {
      clearInterval(interval)
    }
    cycleIntervals.clear()

    // Clear slot states
    slotStates.value.clear()

    animationPhase.value = 'idle'
    rollingStartTime = 0
  }

  // ============================================
  // GETTERS
  // ============================================

  function isSlotRolling(date: string): boolean {
    return slotStates.value.get(date)?.isRolling ?? false
  }

  function hasSlotLanded(date: string): boolean {
    return slotStates.value.get(date)?.hasLanded ?? false
  }

  function getDisplayRecipe(date: string): DisplayRecipe | undefined {
    return slotStates.value.get(date)?.currentDisplayRecipe
  }

  const isAnimating = computed(() => {
    return animationPhase.value === 'rolling' || animationPhase.value === 'landing'
  })

  return {
    animationPhase,
    isAnimating,
    startRolling,
    landSequentially,
    reset,
    isSlotRolling,
    hasSlotLanded,
    getDisplayRecipe,
  }
}
