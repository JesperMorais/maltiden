/**
 * Regression: dashboard saves must NOT strip prep-mode markers.
 *
 * The shopping list honors `leftover` server-side (leftovers days contribute
 * nothing). If a dashboard interaction (updateDayServings / swapRecipeForDay)
 * re-PUTs the week with `leftover` dropped, the backend re-counts the batch
 * recipe and the shopping list double-buys. These tests pin the lossless
 * round-trip.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useDashboardStore } from '../dashboard'
import type { DashboardData } from '@/api/types/dashboard.types'

vi.mock('@/api/client', () => ({
  default: { get: vi.fn() },
}))

vi.mock('@/api/menu.api', () => ({
  getCurrentMenu: vi.fn(),
  // Echo the saved days back so the post-save transform keeps the markers.
  saveMenu: vi.fn().mockImplementation((days) =>
    Promise.resolve({ id: 'menu_real_abc', days }),
  ),
}))

vi.mock('@/api/shopping.api', () => ({
  getShoppingList: vi.fn().mockResolvedValue({ menuId: 'menu_real_abc', categories: [] }),
}))

vi.mock('@/utils/token', () => ({
  tokenUtils: { get: vi.fn(), set: vi.fn(), remove: vi.fn(), exists: vi.fn() },
}))

vi.mock('@/stores/user', () => ({
  useUserStore: () => ({
    currentUser: { id: 'usr_1', name: 'Test', role: 'owner' },
    setUser: vi.fn(),
  }),
}))

/** Seed a real-data store whose week contains a batch pair (cook + leftover). */
function seedPrepStore(): ReturnType<typeof useDashboardStore> {
  const store = useDashboardStore()
  store.initDayMemberExclusions()
  store.initLunchBoxDays()

  const data: DashboardData = {
    user: { id: 'usr_1', name: 'Test', role: 'owner' },
    household: {
      id: 'hh_1',
      name: 'Testfamiljen',
      inviteCode: 'ABC',
      members: [
        { id: 'usr_1', name: 'Anna', role: 'owner', isEatingToday: true, wantsLunchBox: false },
        { id: 'usr_2', name: 'Erik', role: 'member', isEatingToday: true, wantsLunchBox: false },
      ],
    },
    todaysMeal: null,
    weeklyMenu: [
      {
        date: '2026-03-02',
        dayName: 'Måndag',
        dayShort: 'Mån',
        meal: { id: 'rec_1', name: 'Pasta', emoji: '🍝', portions: 4 },
        isToday: false,
        isSkipped: false,
      },
      // Cook day (2× servings).
      {
        date: '2026-03-03',
        dayName: 'Tisdag',
        dayShort: 'Tis',
        meal: { id: 'rec_2', name: 'Köttfärssås', emoji: '🍖', portions: 8 },
        isToday: false,
        isSkipped: false,
      },
      // Leftovers day, drawing from the cook day above.
      {
        date: '2026-03-04',
        dayName: 'Onsdag',
        dayShort: 'Ons',
        meal: { id: 'rec_2', name: 'Köttfärssås', emoji: '🍖', portions: 4 },
        isToday: false,
        isSkipped: false,
        leftover: true,
        cookDate: '2026-03-03',
      },
    ],
    shoppingList: { totalItems: 0, checkedItems: 0, categories: [] },
  }
  store.dashboardData = data
  store.currentMenuId = 'menu_real_abc'
  return store
}

describe('Dashboard store — prep-mode round-trip (regression)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('updateDayServings on an unrelated day keeps leftover/cookDate on the leftovers day', async () => {
    const store = seedPrepStore()
    const { saveMenu } = await import('@/api/menu.api')

    // Change servings on the unrelated Monday.
    await store.updateDayServings('2026-03-02', 3, 0)

    const savedDays = vi.mocked(saveMenu).mock.calls[0]![0]
    const leftoverDay = savedDays.find((d) => d.date === '2026-03-04')
    expect(leftoverDay?.leftover).toBe(true)
    expect(leftoverDay?.cookDate).toBe('2026-03-03')

    // The cook day stays a plain (non-leftover) day.
    const cookDay = savedDays.find((d) => d.date === '2026-03-03')
    expect(cookDay?.leftover).toBeFalsy()

    // And the post-save transform preserved the markers in the store.
    const restInStore = store.weeklyMenu.find((d) => d.date === '2026-03-04')
    expect(restInStore?.leftover).toBe(true)
    expect(restInStore?.cookDate).toBe('2026-03-03')
  })

  it('swapRecipeForDay on an unrelated day keeps leftover/cookDate on the leftovers day', async () => {
    const store = seedPrepStore()
    const { saveMenu } = await import('@/api/menu.api')

    await store.swapRecipeForDay('2026-03-02', 'rec_7')

    const savedDays = vi.mocked(saveMenu).mock.calls[0]![0]
    const leftoverDay = savedDays.find((d) => d.date === '2026-03-04')
    expect(leftoverDay?.leftover).toBe(true)
    expect(leftoverDay?.cookDate).toBe('2026-03-03')
  })
})
