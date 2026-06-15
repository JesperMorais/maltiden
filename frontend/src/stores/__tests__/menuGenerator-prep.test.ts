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

/** Build a 7-day week starting at `start`, with an optional batch pair on the
 * first two days (cook day + leftovers day). */
function buildPrepWeek(start: Date): Menu {
  const dates: string[] = []
  for (let i = 0; i < 7; i++) {
    const d = new Date(start)
    d.setDate(start.getDate() + i)
    dates.push(
      `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`,
    )
  }
  return {
    id: 'menu_test',
    days: dates.map((date, i) => {
      if (i === 0) {
        return { date, recipeId: 'rec_cook', recipeName: 'Köttfärssås', emoji: '🍖', servings: 8 }
      }
      if (i === 1) {
        return {
          date,
          recipeId: 'rec_cook',
          recipeName: 'Köttfärssås',
          emoji: '🍖',
          servings: 4,
          leftover: true,
          cookDate: dates[0]!,
        }
      }
      return { date, recipeId: `rec_${i}`, recipeName: `Recept ${i}`, emoji: '🍽️', servings: 4 }
    }),
  }
}

describe('menuGenerator store — prep mode', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('sends prepMode in the generateInitialMenu request body', async () => {
    const { generateMenu } = await import('@/api/menu.api')
    const start = new Date(2026, 5, 17, 9, 0, 0)
    ;(generateMenu as ReturnType<typeof vi.fn>).mockResolvedValue(buildPrepWeek(start))

    const store = useMenuGeneratorStore()
    store.setPrepMode(true)
    store.initializeWeek(start)
    await store.generateInitialMenu()

    const req = (generateMenu as ReturnType<typeof vi.fn>).mock.calls[0]![0] as GenerateMenuRequest
    expect(req.prepMode).toBe(true)
  })

  it('sends prepMode in the regenerateUnlockedDays request body', async () => {
    const { generateMenu } = await import('@/api/menu.api')
    const start = new Date(2026, 5, 17, 9, 0, 0)
    ;(generateMenu as ReturnType<typeof vi.fn>).mockResolvedValue(buildPrepWeek(start))

    const store = useMenuGeneratorStore()
    store.setPrepMode(true)
    store.initializeWeek(start)
    await store.generateInitialMenu()
    await store.regenerateUnlockedDays()

    const req = (generateMenu as ReturnType<typeof vi.fn>).mock.calls[1]![0] as GenerateMenuRequest
    expect(req.prepMode).toBe(true)
  })

  it('maps leftover/cookDate and derives isCookDay from the response', async () => {
    const { generateMenu } = await import('@/api/menu.api')
    const start = new Date(2026, 5, 17, 9, 0, 0)
    const menu = buildPrepWeek(start)
    ;(generateMenu as ReturnType<typeof vi.fn>).mockResolvedValue(menu)

    const store = useMenuGeneratorStore()
    store.setPrepMode(true)
    store.initializeWeek(start)
    await store.generateInitialMenu()

    const cookDay = store.orderedDays[0]!
    const restDay = store.orderedDays[1]!

    // Cook day: referenced as another day's cookDate → isCookDay true, not leftover.
    expect(cookDay.isCookDay).toBe(true)
    expect(cookDay.leftover).toBeFalsy()
    expect(cookDay.servings).toBe(8)

    // Leftovers day: carries leftover + cookDate, is not a cook day.
    expect(restDay.leftover).toBe(true)
    expect(restDay.cookDate).toBe(cookDay.date)
    expect(restDay.isCookDay).toBe(false)
  })

  it('includes leftover/cookDate in the saveDraftMenu mapping (round-trip)', async () => {
    const { generateMenu, saveMenu } = await import('@/api/menu.api')
    const start = new Date(2026, 5, 17, 9, 0, 0)
    ;(generateMenu as ReturnType<typeof vi.fn>).mockResolvedValue(buildPrepWeek(start))
    ;(saveMenu as ReturnType<typeof vi.fn>).mockResolvedValue({ id: 'menu_saved', days: [] })

    const store = useMenuGeneratorStore()
    store.setPrepMode(true)
    store.initializeWeek(start)
    await store.generateInitialMenu()

    const router = { push: vi.fn() } as never
    await store.saveDraftMenu(router)

    const savedDays = (saveMenu as ReturnType<typeof vi.fn>).mock.calls[0]![0] as Array<{
      date: string
      leftover?: boolean
      cookDate?: string
    }>
    const rest = savedDays[1]!
    expect(rest.leftover).toBe(true)
    expect(rest.cookDate).toBe(savedDays[0]!.date)
  })

  it('does not let a leftovers day be independently locked', async () => {
    const { generateMenu } = await import('@/api/menu.api')
    const start = new Date(2026, 5, 17, 9, 0, 0)
    ;(generateMenu as ReturnType<typeof vi.fn>).mockResolvedValue(buildPrepWeek(start))

    const store = useMenuGeneratorStore()
    store.setPrepMode(true)
    store.initializeWeek(start)
    await store.generateInitialMenu()

    const restDate = store.orderedDays[1]!.date
    store.toggleDayLock(restDate)
    expect(store.isDayLocked(restDate)).toBe(false)

    // The cook day, by contrast, locks normally.
    const cookDate = store.orderedDays[0]!.date
    store.toggleDayLock(cookDate)
    expect(store.isDayLocked(cookDate)).toBe(true)
  })

  it('re-derives isCookDay over the merged week when a regenerate relocates the batch pair', async () => {
    const { generateMenu } = await import('@/api/menu.api')
    const start = new Date(2026, 5, 17, 9, 0, 0)
    const dates: string[] = []
    for (let i = 0; i < 7; i++) {
      const d = new Date(start)
      d.setDate(start.getDate() + i)
      dates.push(
        `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`,
      )
    }

    // Initial: batch pair on days 0 (cook) + 1 (leftover).
    ;(generateMenu as ReturnType<typeof vi.fn>).mockResolvedValueOnce(buildPrepWeek(start))

    const store = useMenuGeneratorStore()
    store.setPrepMode(true)
    store.initializeWeek(start)
    await store.generateInitialMenu()

    expect(store.orderedDays[0]!.isCookDay).toBe(true)
    expect(store.orderedDays[1]!.leftover).toBe(true)

    // Regenerate relocates the batch pair to days 3 (cook) + 4 (leftover).
    const relocated: Menu = {
      id: 'menu_test',
      days: dates.map((date, i) => {
        if (i === 3) {
          return { date, recipeId: 'rec_new_cook', recipeName: 'Lasagne', emoji: '🍝', servings: 8 }
        }
        if (i === 4) {
          return {
            date,
            recipeId: 'rec_new_cook',
            recipeName: 'Lasagne',
            emoji: '🍝',
            servings: 4,
            leftover: true,
            cookDate: dates[3]!,
          }
        }
        return { date, recipeId: `rec_r_${i}`, recipeName: `Recept ${i}`, emoji: '🍽️', servings: 4 }
      }),
    }
    ;(generateMenu as ReturnType<typeof vi.fn>).mockResolvedValueOnce(relocated)

    await store.regenerateUnlockedDays()

    // Old cook day (index 0) is no longer flagged; the relocated one (index 3) is.
    expect(store.orderedDays[0]!.isCookDay).toBe(false)
    expect(store.orderedDays[1]!.leftover).toBeFalsy()
    expect(store.orderedDays[3]!.isCookDay).toBe(true)
    expect(store.orderedDays[4]!.leftover).toBe(true)
    expect(store.orderedDays[4]!.cookDate).toBe(store.orderedDays[3]!.date)
  })

  it('persists prepMode to localStorage and survives clearDraft', () => {
    const store = useMenuGeneratorStore()
    store.setPrepMode(true)
    expect(localStorage.getItem('maltiden_prep_mode')).toBe('true')

    store.clearDraft()
    expect(store.prepMode).toBe(true)
  })
})
