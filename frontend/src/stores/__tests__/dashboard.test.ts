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

// Run the store as production does: real API, no mock mode.
vi.mock('@/mocks', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/mocks')>()),
  USE_MOCKS: false,
}))

vi.mock('@/utils/token', () => ({
  tokenUtils: { get: vi.fn(), set: vi.fn(), remove: vi.fn(), exists: vi.fn() },
}))

const { logout } = vi.hoisted(() => ({ logout: vi.fn() }))

vi.mock('@/stores/user', () => ({
  useUserStore: () => ({
    currentUser: { id: 'usr_1', name: 'Test', role: 'owner' },
    setUser: vi.fn(),
    logout,
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

describe('Dashboard store — failed load', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
    sessionStorage.clear()
  })

  function http401() {
    return Object.assign(new Error('Request failed with status code 401'), {
      isAxiosError: true,
      response: { status: 401, data: { error: 'unauthorized' } },
    })
  }

  it('shows an error, not mock data, when the backend is down', async () => {
    const { default: apiClient } = await import('@/api/client')
    vi.mocked(apiClient.get).mockRejectedValueOnce(new Error('Network Error'))

    const store = useDashboardStore()
    await store.fetchDashboard()

    expect(store.dashboardData).toBeNull()
    expect(store.error).toBe('Kunde inte ladda översikten. Försök igen.')
    expect(store.isUsingRealData).toBe(false)
    expect(logout).not.toHaveBeenCalled()
  })

  it('sends an expired session to login', async () => {
    const { default: apiClient } = await import('@/api/client')
    const { tokenUtils } = await import('@/utils/token')
    vi.mocked(apiClient.get).mockRejectedValueOnce(http401())
    vi.mocked(tokenUtils.get).mockReturnValue(null) // interceptor dropped it

    const store = useDashboardStore()
    await store.fetchDashboard()

    expect(logout).toHaveBeenCalledOnce()
    expect(sessionStorage.getItem('session_expired')).toBe('true')
    expect(store.dashboardData).toBeNull()
    expect(store.isLoading).toBe(false)
  })

  it('shows an error instead of redirecting on a 401 just after login', async () => {
    const { default: apiClient } = await import('@/api/client')
    const { tokenUtils } = await import('@/utils/token')
    vi.mocked(apiClient.get).mockRejectedValueOnce(http401())
    vi.mocked(tokenUtils.get).mockReturnValue('still-here') // grace window kept it

    const store = useDashboardStore()
    await store.fetchDashboard()

    expect(logout).not.toHaveBeenCalled()
    expect(store.dashboardData).toBeNull()
    expect(store.error).not.toBeNull()
  })
})
