import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { DashboardData, Meal, MenuDay, Household, ShoppingListSummary } from '@/api/types/dashboard.types'
import { mockDashboardData } from '@/mocks/dashboard.mock'
import { useUserStore } from './user'
import apiClient from '@/api/client'
import { getCurrentMenu } from '@/api/menu.api'
import type { Menu } from '@/api/menu.api'

/**
 * Dashboard Store
 *
 * Manages dashboard data including today's meal, weekly menu,
 * household info, and shopping list summary.
 */
export const useDashboardStore = defineStore('dashboard', () => {
  const userStore = useUserStore()

  // State
  const dashboardData = ref<DashboardData | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const currentMenuId = ref<string | null>(null)

  // Computed - Today's Meal
  const todaysMeal = computed<Meal | null>(() => dashboardData.value?.todaysMeal ?? null)
  const hasTodaysMeal = computed(() => todaysMeal.value !== null)

  // Computed - Weekly Menu
  const weeklyMenu = computed<MenuDay[]>(() => dashboardData.value?.weeklyMenu ?? [])
  const todayFromMenu = computed(() => weeklyMenu.value.find(day => day.isToday) ?? null)

  // Computed - Household
  const household = computed<Household | null>(() => dashboardData.value?.household ?? null)
  const householdName = computed(() => household.value?.name ?? 'Mitt hushåll')
  const householdMembers = computed(() => household.value?.members ?? [])
  const membersEatingToday = computed(() =>
    householdMembers.value.filter(m => m.isEatingToday)
  )
  const inviteCode = computed(() => household.value?.inviteCode ?? '')

  // Computed - Menu ID (for shopping list)
  const menuId = computed(() => currentMenuId.value)

  // Computed - Shopping List
  const shoppingList = computed<ShoppingListSummary | null>(
    () => dashboardData.value?.shoppingList ?? null
  )
  const shoppingItemsRemaining = computed(() => {
    if (!shoppingList.value) return 0
    return shoppingList.value.totalItems - shoppingList.value.checkedItems
  })

  /**
   * Transform API menu into dashboard MenuDay[] format
   */
  function transformMenuToDashboard(menu: Menu): { weeklyMenu: MenuDay[]; todaysMeal: Meal | null } {
    const dayNames = ['Söndag', 'Måndag', 'Tisdag', 'Onsdag', 'Torsdag', 'Fredag', 'Lördag']
    const dayShorts = ['Sön', 'Mån', 'Tis', 'Ons', 'Tor', 'Fre', 'Lör']
    const todayStr = new Date().toISOString().split('T')[0]!

    let todaysMeal: Meal | null = null

    const weeklyMenu: MenuDay[] = menu.days.map((apiDay) => {
      const date = new Date(apiDay.date + 'T12:00:00')
      const dayIndex = date.getDay()
      const isToday = apiDay.date === todayStr
      const hasMeal = !!apiDay.recipeId && !apiDay.skip

      const meal: Meal | null = hasMeal
        ? {
            id: apiDay.recipeId!,
            name: apiDay.recipeName || 'Recept',
            emoji: apiDay.emoji,
            portions: apiDay.servings,
          }
        : null

      if (isToday && meal) {
        todaysMeal = meal
      }

      return {
        date: apiDay.date,
        dayName: dayNames[dayIndex]!,
        dayShort: dayShorts[dayIndex]!,
        meal,
        isToday,
        isSkipped: apiDay.skip ?? false,
      }
    })

    return { weeklyMenu, todaysMeal }
  }

  // Actions
  async function fetchDashboard(forceRefresh = false): Promise<void> {
    if (dashboardData.value && !forceRefresh) {
      return
    }

    isLoading.value = true
    error.value = null

    try {
      // Fetch real household data from backend
      const { data: householdData } = await apiClient.get('/households/me')

      // Fetch current menu from backend
      let weeklyMenu: MenuDay[] = []
      let todaysMeal: Meal | null = null

      try {
        const menu = await getCurrentMenu()
        if (menu) {
          currentMenuId.value = menu.id
          const transformed = transformMenuToDashboard(menu)
          weeklyMenu = transformed.weeklyMenu
          todaysMeal = transformed.todaysMeal
        } else {
          currentMenuId.value = null
        }
      } catch (menuErr) {
        console.warn('Could not fetch menu:', menuErr)
        currentMenuId.value = null
      }

      dashboardData.value = {
        user: userStore.currentUser!,
        household: {
          id: householdData.id,
          name: householdData.name,
          inviteCode: householdData.inviteCode || '',
          members: householdData.members.map((m: { id: string; name: string; role: string }) => ({
            id: m.id,
            name: m.name,
            role: m.role as 'owner' | 'member' | 'guest',
            isEatingToday: true,
            wantsLunchBox: false,
          })),
        },
        todaysMeal,
        weeklyMenu,
        shoppingList: {
          totalItems: 0,
          checkedItems: 0,
          categories: [],
        },
      }
    } catch (e) {
      // If backend fails, fall back to full mock data
      console.warn('Using mock dashboard data:', e)
      dashboardData.value = mockDashboardData

      if (dashboardData.value.user) {
        userStore.setUser(dashboardData.value.user)
      }
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Update a member's status locally (for optimistic UI).
   * Mutates dashboardData in place so computed refs react.
   */
  function updateMemberLocally(
    memberId: string,
    update: Partial<{ isEatingToday: boolean; wantsLunchBox: boolean }>,
  ) {
    const members = dashboardData.value?.household.members
    if (!members) {
      console.warn('updateMemberLocally: no household data loaded')
      return
    }
    const member = members.find((m) => m.id === memberId)
    if (!member) {
      console.warn(`updateMemberLocally: member ${memberId} not found`)
      return
    }
    Object.assign(member, update)
  }

  function clearError() {
    error.value = null
  }

  return {
    // State
    dashboardData,
    isLoading,
    error,

    // Today's Meal
    todaysMeal,
    hasTodaysMeal,

    // Weekly Menu
    weeklyMenu,
    todayFromMenu,

    // Household
    household,
    householdName,
    householdMembers,
    membersEatingToday,
    inviteCode,

    // Menu
    menuId,

    // Shopping
    shoppingList,
    shoppingItemsRemaining,

    // Actions
    fetchDashboard,
    updateMemberLocally,
    clearError
  }
})
