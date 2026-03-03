import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useDashboardStore } from '../dashboard'
import type { DashboardData } from '@/api/types/dashboard.types'

// Mock API modules — vi.mock is hoisted, so no external references in factories
vi.mock('@/api/client', () => ({
  default: { get: vi.fn() },
}))

vi.mock('@/api/menu.api', () => ({
  getCurrentMenu: vi.fn(),
  saveMenu: vi.fn().mockResolvedValue(undefined),
}))

vi.mock('@/api/shopping.api', () => ({
  getShoppingList: vi.fn().mockResolvedValue({
    menuId: 'menu_real_123',
    categories: [
      {
        name: 'Mejeri',
        items: [
          { id: 'i1', name: 'Mjölk', amount: 1, unit: 'l', checked: false },
          { id: 'i2', name: 'Ost', amount: 200, unit: 'g', checked: true },
        ],
      },
      {
        name: 'Grönsaker',
        items: [
          { id: 'i3', name: 'Tomat', amount: 3, unit: 'st', checked: false },
        ],
      },
    ],
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

function makeDashboardData(): DashboardData {
  const today = new Date().toISOString().split('T')[0]!
  return {
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
    todaysMeal: { id: 'rec_1', name: 'Pasta', emoji: '🍝', portions: 4 },
    weeklyMenu: [
      {
        date: today,
        dayName: 'Måndag',
        dayShort: 'Mån',
        meal: { id: 'rec_1', name: 'Pasta', emoji: '🍝', portions: 4 },
        isToday: true,
        isSkipped: false,
      },
    ],
    shoppingList: { totalItems: 0, checkedItems: 0, categories: [] },
  }
}

describe('Dashboard store — shopping list sync', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('toggleMemberDay does NOT call saveMenu in mock/fallback mode', async () => {
    const store = useDashboardStore()
    const { saveMenu } = await import('@/api/menu.api')

    store.dashboardData = makeDashboardData()
    store.initDayMemberExclusions()
    store.initLunchBoxDays()

    // isUsingRealData is false (currentMenuId is null by default)
    expect(store.isUsingRealData).toBe(false)

    const today = new Date().toISOString().split('T')[0]!
    store.toggleMemberDay(today, 'usr_2')

    // saveMenu should NOT be called in fallback mode
    expect(saveMenu).not.toHaveBeenCalled()
  })

  it('toggleMemberDay updates exclusions and clamps lunchbox count', () => {
    const store = useDashboardStore()
    store.dashboardData = makeDashboardData()
    store.initDayMemberExclusions()
    store.initLunchBoxDays()

    const today = new Date().toISOString().split('T')[0]!

    // Both members eating initially
    expect(store.getMembersEatingDay(today)).toHaveLength(2)

    // Set lunchbox count to 2
    store.setDayLunchBoxCount(today, 2)
    expect(store.getDayLunchBoxCount(today)).toBe(2)

    // Toggle one member off — now only 1 eating
    store.toggleMemberDay(today, 'usr_2')
    expect(store.getMembersEatingDay(today)).toHaveLength(1)

    // Lunchbox count should be clamped to 1
    expect(store.getDayLunchBoxCount(today)).toBe(1)

    // Toggle back on — 2 eating again
    store.toggleMemberDay(today, 'usr_2')
    expect(store.getMembersEatingDay(today)).toHaveLength(2)
    // Lunchbox stays at 1 (not auto-incremented)
    expect(store.getDayLunchBoxCount(today)).toBe(1)
  })

  it('refreshShoppingList is a no-op when not using real data', async () => {
    const store = useDashboardStore()
    const { getShoppingList } = await import('@/api/shopping.api')

    store.dashboardData = makeDashboardData()

    // isUsingRealData is false (currentMenuId is null)
    expect(store.isUsingRealData).toBe(false)

    await store.refreshShoppingList()

    // Should NOT call the API
    expect(getShoppingList).not.toHaveBeenCalled()
  })

  it('shopping list summary computed values are correct', () => {
    const store = useDashboardStore()
    store.dashboardData = {
      ...makeDashboardData(),
      shoppingList: {
        totalItems: 12,
        checkedItems: 3,
        categories: [
          { name: 'Mejeri', count: 4 },
          { name: 'Kött', count: 3 },
          { name: 'Grönsaker', count: 5 },
        ],
      },
    }

    expect(store.shoppingList?.totalItems).toBe(12)
    expect(store.shoppingList?.checkedItems).toBe(3)
    expect(store.shoppingItemsRemaining).toBe(9)
    expect(store.shoppingList?.categories).toHaveLength(3)
  })

  it('selectedDate state is managed via setSelectedDate', () => {
    const store = useDashboardStore()

    expect(store.selectedDate).toBeNull()

    store.setSelectedDate('2026-03-03')
    expect(store.selectedDate).toBe('2026-03-03')
    expect(store.displayDate).toBe('2026-03-03')

    store.setSelectedDate(null)
    expect(store.selectedDate).toBeNull()
    // displayDate falls back to today
    const today = new Date().toISOString().split('T')[0]!
    expect(store.displayDate).toBe(today)
  })

  it('membersForDisplayDate reflects selected day exclusions', () => {
    const store = useDashboardStore()
    store.dashboardData = makeDashboardData()
    store.initDayMemberExclusions()

    // Default: all members eating
    expect(store.membersForDisplayDate).toHaveLength(2)

    // Select a specific date and exclude a member
    const today = new Date().toISOString().split('T')[0]!
    store.setSelectedDate(today)
    store.toggleMemberDay(today, 'usr_2')

    expect(store.membersForDisplayDate).toHaveLength(1)
    expect(store.membersForDisplayDate[0]!.id).toBe('usr_1')
  })
})
