/**
 * Integration test: Dashboard store shopping list refresh
 *
 * Verifies the full flow:
 *  toggleMemberDay → updateDayServings → saveMenu → refreshShoppingList → widget updates
 *
 * Uses mocked API calls to simulate real backend responses.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useDashboardStore } from '../dashboard'
import type { DashboardData } from '@/api/types/dashboard.types'

// Track call order
const callOrder: string[] = []

vi.mock('@/api/client', () => ({
  default: { get: vi.fn() },
}))

vi.mock('@/api/menu.api', () => ({
  getCurrentMenu: vi.fn(),
  saveMenu: vi.fn().mockImplementation(() => {
    callOrder.push('saveMenu')
    return Promise.resolve({
      id: 'menu_real_abc',
      days: [{ date: '2026-03-03', recipeId: 'rec_1', servings: 3 }],
    })
  }),
}))

vi.mock('@/api/shopping.api', () => ({
  getShoppingList: vi.fn().mockImplementation(() => {
    callOrder.push('getShoppingList')
    return Promise.resolve({
      menuId: 'menu_real_abc',
      categories: [
        {
          name: 'Mejeri',
          items: [
            { id: 'i1', name: 'Mjölk', amount: 3, unit: 'dl', checked: false },
            { id: 'i2', name: 'Ost', amount: 150, unit: 'g', checked: true },
          ],
        },
        {
          name: 'Grönsaker',
          items: [
            { id: 'i3', name: 'Tomat', amount: 2, unit: 'st', checked: false },
            { id: 'i4', name: 'Lök', amount: 1, unit: 'st', checked: false },
          ],
        },
      ],
    })
  }),
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

function seedRealDataStore(): ReturnType<typeof useDashboardStore> {
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
        { id: 'usr_3', name: 'Lisa', role: 'member', isEatingToday: true, wantsLunchBox: false },
      ],
    },
    todaysMeal: { id: 'rec_1', name: 'Pasta', emoji: '🍝', portions: 4 },
    weeklyMenu: [
      {
        date: '2026-03-03',
        dayName: 'Tisdag',
        dayShort: 'Tis',
        meal: { id: 'rec_1', name: 'Pasta', emoji: '🍝', portions: 4 },
        isToday: true,
        isSkipped: false,
      },
    ],
    shoppingList: { totalItems: 0, checkedItems: 0, categories: [] },
  }
  store.dashboardData = data

  // Set a real menu ID so isUsingRealData returns true
  store.currentMenuId = 'menu_real_abc'

  return store
}

describe('Dashboard store — shopping list integration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    callOrder.length = 0
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('isUsingRealData is true when currentMenuId is a real ID', () => {
    const store = seedRealDataStore()
    expect(store.isUsingRealData).toBe(true)
    expect(store.menuId).toBe('menu_real_abc')
  })

  it('refreshShoppingList fetches from API and updates widget in real-data mode', async () => {
    const store = seedRealDataStore()
    const { getShoppingList } = await import('@/api/shopping.api')

    // Initially empty
    expect(store.shoppingList?.totalItems).toBe(0)

    await store.refreshShoppingList()

    // Should have called the API
    expect(getShoppingList).toHaveBeenCalledWith('menu_real_abc')

    // Widget should now reflect the mock API response
    expect(store.shoppingList?.totalItems).toBe(4)
    expect(store.shoppingList?.checkedItems).toBe(1)
    expect(store.shoppingItemsRemaining).toBe(3)
    expect(store.shoppingList?.categories).toHaveLength(2)
    expect(store.shoppingList?.categories[0]?.name).toBe('Mejeri')
    expect(store.shoppingList?.categories[0]?.count).toBe(2)
    expect(store.shoppingList?.categories[1]?.name).toBe('Grönsaker')
    expect(store.shoppingList?.categories[1]?.count).toBe(2)
  })

  it('toggleMemberDay triggers saveMenu then refreshShoppingList in real-data mode', async () => {
    const store = seedRealDataStore()
    const { saveMenu } = await import('@/api/menu.api')
    const { getShoppingList } = await import('@/api/shopping.api')

    // 3 members eating, 0 lunchboxes
    expect(store.getMembersEatingDay('2026-03-03')).toHaveLength(3)

    // Toggle one member off
    store.toggleMemberDay('2026-03-03', 'usr_3')

    // Now 2 members eating
    expect(store.getMembersEatingDay('2026-03-03')).toHaveLength(2)

    // saveMenu should have been called with updated servings
    expect(saveMenu).toHaveBeenCalledTimes(1)
    const savedDays = vi.mocked(saveMenu).mock.calls[0]![0]
    const day = savedDays.find((d) => d.date === '2026-03-03')
    expect(day?.servings).toBe(2) // 2 eating + 0 lunchboxes

    // Wait for async refreshShoppingList
    await vi.waitFor(() => {
      expect(getShoppingList).toHaveBeenCalledTimes(1)
    })

    // Verify call order: saveMenu first, then getShoppingList
    expect(callOrder).toEqual(['saveMenu', 'getShoppingList'])

    // Shopping list widget should be updated
    expect(store.shoppingList?.totalItems).toBe(4)
    expect(store.shoppingItemsRemaining).toBe(3)
  })

  it('lunchbox change triggers saveMenu then refreshShoppingList', async () => {
    const store = seedRealDataStore()
    const { saveMenu } = await import('@/api/menu.api')
    const { getShoppingList } = await import('@/api/shopping.api')

    // Add 2 lunchboxes
    store.setDayLunchBoxCount('2026-03-03', 2)

    // Call updateDayServings (this is what WeeklyMenuGrid does on lunchbox change)
    await store.updateDayServings('2026-03-03', 3, 2)

    // saveMenu called with 3 eating + 2 lunchboxes = 5 servings
    expect(saveMenu).toHaveBeenCalledTimes(1)
    const savedDays = vi.mocked(saveMenu).mock.calls[0]![0]
    const day = savedDays.find((d) => d.date === '2026-03-03')
    expect(day?.servings).toBe(5)

    // Shopping list refreshed after save
    await vi.waitFor(() => {
      expect(getShoppingList).toHaveBeenCalledTimes(1)
    })
  })

  it('multiple rapid toggles each trigger their own save', async () => {
    const store = seedRealDataStore()
    const { saveMenu } = await import('@/api/menu.api')

    // Toggle two members rapidly
    store.toggleMemberDay('2026-03-03', 'usr_2')
    store.toggleMemberDay('2026-03-03', 'usr_3')

    // Both should trigger saves
    expect(saveMenu).toHaveBeenCalledTimes(2)

    // First save: 2 eating (usr_2 removed)
    const firstDays = vi.mocked(saveMenu).mock.calls[0]![0]
    expect(firstDays.find((d) => d.date === '2026-03-03')?.servings).toBe(2)

    // Second save: 1 eating (usr_3 also removed)
    const secondDays = vi.mocked(saveMenu).mock.calls[1]![0]
    expect(secondDays.find((d) => d.date === '2026-03-03')?.servings).toBe(1)
  })
})
