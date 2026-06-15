import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
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

/**
 * Backend-style date builder: 7 days anchored at `today` (time.Now()), local
 * calendar dates. This is what the server actually returns — NOT Monday
 * anchored.
 */
function backendDates(today: Date): string[] {
  const dates: string[] = []
  for (let i = 0; i < 7; i++) {
    const d = new Date(today)
    d.setDate(today.getDate() + i)
    const y = d.getFullYear()
    const m = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    dates.push(`${y}-${m}-${day}`)
  }
  return dates
}

function buildWeek(dates: string[], prefix: string): Menu {
  return {
    id: 'menu_test',
    days: dates.map((date, i) => ({
      date,
      recipeId: `${prefix}_rec_${i}`,
      recipeName: `${prefix} Recept ${i}`,
      emoji: '🍽️',
      servings: 4,
    })),
    economy: {
      sharedIngredients: [{ canonicalName: 'lok', name: 'Gul lök', recipeCount: 2 }],
      distinctItemsToBuy: 7,
      totalIngredientRefs: 11,
    },
  }
}

describe('menuGenerator store — backend-date alignment (regression)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
    localStorage.clear()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('formatDateISO uses LOCAL date — late-evening does not roll to next/prev UTC day', () => {
    // 2026-06-17 23:30 local (a Wednesday). In UTC+2 this is 21:30 UTC same
    // day, but the original toISOString() bug could roll dates near midnight.
    // Pin a late local evening and assert the placeholder window starts on the
    // SAME local calendar day, not a UTC-shifted one.
    vi.useFakeTimers()
    vi.setSystemTime(new Date(2026, 5, 17, 23, 30, 0)) // month is 0-indexed → June

    const store = useMenuGeneratorStore()
    store.initializeWeek()

    // First placeholder date must be today's LOCAL calendar date.
    expect(store.orderedDays[0]!.date).toBe('2026-06-17')
    // And the window is today-anchored (7 consecutive local days).
    expect(store.orderedDays.map((d) => d.date)).toEqual([
      '2026-06-17',
      '2026-06-18',
      '2026-06-19',
      '2026-06-20',
      '2026-06-21',
      '2026-06-22',
      '2026-06-23',
    ])
  })

  it('adopts the backend dates on initial generate (today-anchored, not Monday)', async () => {
    const { generateMenu } = await import('@/api/menu.api')
    // 2026-06-17 is a Wednesday → NOT Monday.
    vi.useFakeTimers()
    vi.setSystemTime(new Date(2026, 5, 17, 9, 0, 0))

    const dates = backendDates(new Date(2026, 5, 17, 9, 0, 0))

    const store = useMenuGeneratorStore()
    store.initializeWeek()
    ;(generateMenu as ReturnType<typeof vi.fn>).mockResolvedValue(buildWeek(dates, 'init'))

    await store.generateInitialMenu()

    // Draft adopts the backend dates verbatim.
    expect(store.orderedDays.map((d) => d.date)).toEqual(dates)
    // Weekday labels derive from those dates (2026-06-17 is a Wednesday).
    expect(store.orderedDays[0]!.dayName).toBe('Onsdag')
  })

  it('regenerate on a non-Monday today: keeps locked day, replaces unlocked', async () => {
    const { generateMenu } = await import('@/api/menu.api')
    // Thursday → NOT Monday. This is the exact case the original bug broke.
    vi.useFakeTimers()
    vi.setSystemTime(new Date(2026, 5, 18, 9, 0, 0))
    const today = new Date(2026, 5, 18, 9, 0, 0)
    const dates = backendDates(today)

    const store = useMenuGeneratorStore()
    store.initializeWeek()

    ;(generateMenu as ReturnType<typeof vi.fn>).mockResolvedValueOnce(buildWeek(dates, 'init'))
    await store.generateInitialMenu()

    // Lock the third day.
    const lockedDate = dates[2]!
    store.toggleDayLock(lockedDate)
    const lockedRecipeId = store.orderedDays[2]!.recipeId
    expect(lockedRecipeId).toBe('init_rec_2')

    // Server returns the full week respecting the lock: locked day unchanged,
    // others get new recipes.
    const serverDays: Menu['days'] = dates.map((date, i) => {
      if (i === 2) {
        return { date, recipeId: lockedRecipeId, recipeName: 'init Recept 2', emoji: '🍽️', servings: 4 }
      }
      return { date, recipeId: `regen_rec_${i}`, recipeName: `regen Recept ${i}`, emoji: '🔄', servings: 4 }
    })
    ;(generateMenu as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      id: 'menu_test',
      days: serverDays,
      economy: {
        sharedIngredients: [{ canonicalName: 'pasta', name: 'Pasta', recipeCount: 2 }],
        distinctItemsToBuy: 6,
        totalIngredientRefs: 10,
      },
    })

    await store.regenerateUnlockedDays()

    // The locked map sent to the server is keyed by the BACKEND date.
    const req = (generateMenu as ReturnType<typeof vi.fn>).mock
      .calls[1]![0] as GenerateMenuRequest
    expect(req.lockedDays).toEqual({ [lockedDate]: lockedRecipeId })

    // Locked day preserved.
    expect(store.orderedDays[2]!.recipeId).toBe(lockedRecipeId)

    // At least one unlocked day actually changed (the date-keyed merge HIT —
    // this is what the original Monday-vs-today mismatch silently broke).
    const changed = store.orderedDays.filter((d) => d.recipeId?.startsWith('regen_rec_'))
    expect(changed.length).toBeGreaterThan(0)
    expect(store.orderedDays[0]!.recipeId).toBe('regen_rec_0')
    expect(store.orderedDays[0]!.emoji).toBe('🔄')
  })
})
