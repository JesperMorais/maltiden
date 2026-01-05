import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { DashboardData, Meal, MenuDay, Household, ShoppingListSummary } from '@/api/types/dashboard.types'
import { USE_MOCKS } from '@/mocks'
import { mockDashboardData } from '@/mocks/dashboard.mock'
import { useUserStore } from './user'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

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
      if (USE_MOCKS) {
        // Simulate network delay
        await new Promise(resolve => setTimeout(resolve, 500))
        dashboardData.value = mockDashboardData

        // Set current user from dashboard data
        if (dashboardData.value.user) {
          userStore.setUser(dashboardData.value.user)
        }
      } else {
        const response = await fetch(`${API_URL}/api/dashboard`)

        if (!response.ok) {
          throw new Error(`Failed to fetch dashboard: ${response.status}`)
        }

        const result = await response.json()

        if (!result.success) {
          throw new Error(result.error?.message || 'Unknown error')
        }

        dashboardData.value = result.data

        if (result.data.user) {
          userStore.setUser(result.data.user)
        }
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Kunde inte ladda dashboard'
      console.error('Dashboard store error:', e)
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
