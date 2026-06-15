import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useMenuGeneratorStore } from '../menuGenerator'
import type { Menu, GenerateMenuRequest } from '@/api/menu.api'

vi.mock('@/api/client', () => ({
  default: { get: vi.fn(), post: vi.fn(), put: vi.fn() },
}))

vi.mock('@/api/menu.api', () => ({
  generateMenu: vi.fn(),
  saveMenu: vi.fn(),
}))

const economyA: Menu['economy'] = {
  sharedIngredients: [{ canonicalName: 'lok', name: 'Gul lök', recipeCount: 3 }],
  distinctItemsToBuy: 8,
  totalIngredientRefs: 12,
}

const economyB: Menu['economy'] = {
  sharedIngredients: [{ canonicalName: 'pasta', name: 'Pasta', recipeCount: 2 }],
  distinctItemsToBuy: 6,
  totalIngredientRefs: 10,
}

/** Build a full-week Menu response keyed off the store's draft dates. */
function buildWeek(dates: string[], prefix: string, economy: Menu['economy']): Menu {
  return {
    id: 'menu_test',
    days: dates.map((date, i) => ({
      date,
      recipeId: `${prefix}_rec_${i}`,
      recipeName: `${prefix} Recept ${i}`,
      emoji: '🍽️',
      servings: 4,
    })),
    economy,
  }
}

describe('menuGenerator store — locks + economy', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('generateInitialMenu passes empty lockedDays and stores economy', async () => {
    const { generateMenu } = await import('@/api/menu.api')
    const store = useMenuGeneratorStore()
    store.initializeWeek()
    const dates = store.orderedDays.map((d) => d.date)

    ;(generateMenu as ReturnType<typeof vi.fn>).mockResolvedValue(
      buildWeek(dates, 'init', economyA),
    )

    await store.generateInitialMenu()

    const req = (generateMenu as ReturnType<typeof vi.fn>).mock
      .calls[0]![0] as GenerateMenuRequest
    expect(req.lockedDays).toEqual({})
    expect(store.economy).toEqual(economyA)
  })

  it('regenerateUnlockedDays derives lockedDays from the locked Set and maps by date', async () => {
    const { generateMenu } = await import('@/api/menu.api')
    const store = useMenuGeneratorStore()
    store.initializeWeek()
    const dates = store.orderedDays.map((d) => d.date)

    // Initial generate
    ;(generateMenu as ReturnType<typeof vi.fn>).mockResolvedValueOnce(
      buildWeek(dates, 'init', economyA),
    )
    await store.generateInitialMenu()

    // Lock the first two days.
    store.toggleDayLock(dates[0]!)
    store.toggleDayLock(dates[1]!)

    const lockedRecipe0 = store.orderedDays[0]!.recipeId
    const lockedRecipe1 = store.orderedDays[1]!.recipeId

    // Server returns a full new week (locked days come back unchanged).
    const serverDays: Menu['days'] = dates.map((date, i) => {
      if (i === 0) {
        return { date, recipeId: lockedRecipe0, recipeName: 'init Recept 0', emoji: '🍽️', servings: 4 }
      }
      if (i === 1) {
        return { date, recipeId: lockedRecipe1, recipeName: 'init Recept 1', emoji: '🍽️', servings: 4 }
      }
      return { date, recipeId: `regen_rec_${i}`, recipeName: `regen Recept ${i}`, emoji: '🔄', servings: 4 }
    })
    ;(generateMenu as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      id: 'menu_test',
      days: serverDays,
      economy: economyB,
    })

    await store.regenerateUnlockedDays()

    // Request carried the locked map derived from the Set.
    const req = (generateMenu as ReturnType<typeof vi.fn>).mock
      .calls[1]![0] as GenerateMenuRequest
    expect(req.lockedDays).toEqual({
      [dates[0]!]: lockedRecipe0,
      [dates[1]!]: lockedRecipe1,
    })

    // Locked days preserved, unlocked replaced — mapped by date.
    expect(store.orderedDays[0]!.recipeId).toBe(lockedRecipe0)
    expect(store.orderedDays[1]!.recipeId).toBe(lockedRecipe1)
    expect(store.orderedDays[2]!.recipeId).toBe('regen_rec_2')
    expect(store.orderedDays[2]!.emoji).toBe('🔄')

    // New economy stored.
    expect(store.economy).toEqual(economyB)
  })

  it('clearDraft clears economy', async () => {
    const { generateMenu } = await import('@/api/menu.api')
    const store = useMenuGeneratorStore()
    store.initializeWeek()
    const dates = store.orderedDays.map((d) => d.date)

    ;(generateMenu as ReturnType<typeof vi.fn>).mockResolvedValue(
      buildWeek(dates, 'init', economyA),
    )
    await store.generateInitialMenu()
    expect(store.economy).toEqual(economyA)

    store.clearDraft()
    expect(store.economy).toBeNull()
  })
})
