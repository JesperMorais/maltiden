import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { DashboardData, Meal, MenuDay, Household, ShoppingListSummary } from '@/api/types/dashboard.types'
import { mockDashboardData } from '@/mocks/dashboard.mock'
import { useUserStore } from './user'
import apiClient from '@/api/client'

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

  // Computed - Shopping List
  const shoppingList = computed<ShoppingListSummary | null>(
    () => dashboardData.value?.shoppingList ?? null
  )
  const shoppingItemsRemaining = computed(() => {
    if (!shoppingList.value) return 0
    return shoppingList.value.totalItems - shoppingList.value.checkedItems
  })

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

      // Combine real household with mock data for features not yet implemented
      // (menu, shopping list, etc. will come from backend later)
      dashboardData.value = {
        ...mockDashboardData,
        user: userStore.currentUser!,
        household: {
          id: householdData.id,
          name: householdData.name,
          inviteCode: householdData.inviteCode || 'ABC123', // Backend may not have this yet
          members: householdData.members.map((m: { id: string; name: string; role: string }) => ({
            id: m.id,
            name: m.name,
            role: m.role as 'owner' | 'member' | 'guest',
            isEatingToday: true, // Default until backend supports this
            wantsLunchBox: false
          }))
        }
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

    // Shopping
    shoppingList,
    shoppingItemsRemaining,

    // Actions
    fetchDashboard,
    clearError
  }
})
