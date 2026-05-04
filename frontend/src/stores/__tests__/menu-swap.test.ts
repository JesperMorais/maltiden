import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useDashboardStore } from '../dashboard'
import type { DashboardData } from '@/api/types/dashboard.types'

vi.mock('@/api/client', () => ({
  default: { get: vi.fn() },
}))

vi.mock('@/api/menu.api', () => ({
  getCurrentMenu: vi.fn(),
  saveMenu: vi.fn(),
}))

vi.mock('@/api/shopping.api', () => ({
  getShoppingList: vi.fn().mockResolvedValue({
    menuId: 'menu_real_123',
    categories: [
      {
        name: 'Mejeri',
        items: [{ id: 'i1', name: 'Mjölk', amount: 1, unit: 'l', checked: false }],
      },
    ],
  }),
}))

vi.mock('@/api/recipes.api', () => ({
  getRecipe: vi.fn().mockResolvedValue({
    id: 'rec_2',
    name: 'Tacos',
    emoji: '🌮',
    servings: 4,
    tags: ['vardag'],
    ingredients: [],
    instructions: [],
  }),
}))

vi.mock('@/api/household.api', () => ({
  updateHousehold: vi.fn(),
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

const TODAY = new Date().toISOString().split('T')[0]!

function makeDashboardData(): DashboardData {
  return {
    user: { id: 'usr_1', name: 'Test', role: 'owner' },
    household: {
      id: 'hh_1',
      name: 'Testfamiljen',
      inviteCode: 'ABC',
      members: [
        { id: 'usr_1', name: 'Anna', role: 'owner', isEatingToday: true, wantsLunchBox: false },
      ],
    },
    todaysMeal: { id: 'rec_1', name: 'Pasta', emoji: '🍝', portions: 4 },
    weeklyMenu: [
      {
        date: TODAY,
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

describe('Dashboard store — swapRecipeForDay', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('updates the weekly menu optimistically before saveMenu resolves', async () => {
    const { saveMenu } = await import('@/api/menu.api')
    let resolveSave!: (value: unknown) => void
    ;(saveMenu as ReturnType<typeof vi.fn>).mockReturnValue(
      new Promise((resolve) => {
        resolveSave = resolve
      }),
    )

    const store = useDashboardStore()
    store.dashboardData = makeDashboardData()
    store.currentMenuId = 'menu_real_123'

    const promise = store.swapRecipeForDay(TODAY, 'rec_2')

    // Optimistic: meal id has changed before the API resolves.
    expect(store.weeklyMenu[0]?.meal?.id).toBe('rec_2')
    expect(store.todaysMeal?.id).toBe('rec_2')

    resolveSave({
      id: 'menu_real_123',
      days: [
        {
          date: TODAY,
          recipeId: 'rec_2',
          recipeName: 'Tacos',
          emoji: '🌮',
          servings: 4,
        },
      ],
    })

    const ok = await promise
    expect(ok).toBe(true)
    expect(store.weeklyMenu[0]?.meal?.id).toBe('rec_2')
    expect(store.weeklyMenu[0]?.meal?.name).toBe('Tacos')
  })

  it('rolls back local state when saveMenu fails', async () => {
    const { saveMenu } = await import('@/api/menu.api')
    ;(saveMenu as ReturnType<typeof vi.fn>).mockRejectedValue(new Error('boom'))

    const store = useDashboardStore()
    store.dashboardData = makeDashboardData()
    store.currentMenuId = 'menu_real_123'

    const ok = await store.swapRecipeForDay(TODAY, 'rec_2')

    expect(ok).toBe(false)
    expect(store.weeklyMenu[0]?.meal?.id).toBe('rec_1')
    expect(store.weeklyMenu[0]?.meal?.name).toBe('Pasta')
    expect(store.todaysMeal?.id).toBe('rec_1')
    expect(store.error).toBe('Kunde inte byta recept. Försök igen.')
  })

  it('keeps the new recipe and refreshes shopping list on success', async () => {
    const { saveMenu } = await import('@/api/menu.api')
    const { getShoppingList } = await import('@/api/shopping.api')

    ;(saveMenu as ReturnType<typeof vi.fn>).mockResolvedValue({
      id: 'menu_real_123',
      days: [
        {
          date: TODAY,
          recipeId: 'rec_2',
          recipeName: 'Tacos',
          emoji: '🌮',
          servings: 4,
        },
      ],
    })

    const store = useDashboardStore()
    store.dashboardData = makeDashboardData()
    store.currentMenuId = 'menu_real_123'

    const ok = await store.swapRecipeForDay(TODAY, 'rec_2')

    expect(ok).toBe(true)
    expect(store.weeklyMenu[0]?.meal?.id).toBe('rec_2')
    expect(getShoppingList).toHaveBeenCalledWith('menu_real_123')
  })

  it('returns false when no menu is loaded', async () => {
    const store = useDashboardStore()
    // No dashboardData / currentMenuId set
    const ok = await store.swapRecipeForDay(TODAY, 'rec_2')
    expect(ok).toBe(false)
  })
})
