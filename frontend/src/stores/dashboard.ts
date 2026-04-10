import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type {
  DashboardData,
  Meal,
  MenuDay,
  Household,
  ShoppingListSummary,
} from '@/api/types/dashboard.types'
import { useUserStore } from './user'
import apiClient from '@/api/client'
import { getCurrentMenu } from '@/api/menu.api'
import { getMemberStatus } from '@/api/household.api'

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
  const currentMenuId = ref<string | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Computed - Today's Meal
  const todaysMeal = computed<Meal | null>(() => dashboardData.value?.todaysMeal ?? null)
  const hasTodaysMeal = computed(() => todaysMeal.value !== null)

  // Computed - Weekly Menu
  const weeklyMenu = computed<MenuDay[]>(() => dashboardData.value?.weeklyMenu ?? [])
  const todayFromMenu = computed(() => weeklyMenu.value.find((day) => day.isToday) ?? null)

  // Computed - Household
  const household = computed<Household | null>(() => dashboardData.value?.household ?? null)
  const householdName = computed(() => household.value?.name ?? 'Mitt hushåll')
  const householdMembers = computed(() => household.value?.members ?? [])
  const membersEatingToday = computed(() =>
    householdMembers.value.filter((m) => m.isEatingToday),
  )
  const inviteCode = computed(() => household.value?.inviteCode ?? '')

  // Computed - Shopping List
  const shoppingList = computed<ShoppingListSummary | null>(
    () => dashboardData.value?.shoppingList ?? null,
  )
  const shoppingItemsRemaining = computed(() => {
    if (!shoppingList.value) return 0
    return shoppingList.value.totalItems - shoppingList.value.checkedItems
  })

  // Swedish day names
  const dayNames: Record<string, { full: string; short: string }> = {
    '0': { full: 'Söndag', short: 'Sön' },
    '1': { full: 'Måndag', short: 'Mån' },
    '2': { full: 'Tisdag', short: 'Tis' },
    '3': { full: 'Onsdag', short: 'Ons' },
    '4': { full: 'Torsdag', short: 'Tor' },
    '5': { full: 'Fredag', short: 'Fre' },
    '6': { full: 'Lördag', short: 'Lör' },
  }

  // Actions
  async function fetchDashboard(forceRefresh = false): Promise<void> {
    if (dashboardData.value && !forceRefresh) {
      return
    }

    isLoading.value = true
    error.value = null

    try {
      // Fetch household data, menu, and member statuses in parallel
      const [householdRes, menu, memberStatusRes] = await Promise.all([
        apiClient.get('/households/me'),
        getCurrentMenu(),
        getMemberStatus().catch(() => ({
          members: [] as { id: string; isEatingToday: boolean; wantsLunchBox: boolean }[],
        })),
      ])

      const householdData = householdRes.data

      // Build member status lookup
      const statusMap = new Map<string, { isEatingToday: boolean; wantsLunchBox: boolean }>()
      for (const ms of memberStatusRes.members) {
        statusMap.set(ms.id, {
          isEatingToday: ms.isEatingToday,
          wantsLunchBox: ms.wantsLunchBox,
        })
      }

      // Build weekly menu from real data
      const todayStr = new Date().toISOString().slice(0, 10)
      let builtWeeklyMenu: MenuDay[] = []
      let builtTodaysMeal: Meal | null = null
      let builtShoppingList: ShoppingListSummary = {
        totalItems: 0,
        checkedItems: 0,
        categories: [],
      }

      if (menu) {
        currentMenuId.value = menu.id

        // Fetch recipe details for menu days
        const recipeIds = menu.days
          .filter((d) => d.recipeId && !d.skip)
          .map((d) => d.recipeId!)
        const uniqueIds = [...new Set(recipeIds)]
        const recipeLookup = new Map<
          string,
          { name: string; emoji?: string; servings: number }
        >()

        // Batch fetch recipes (they're public)
        await Promise.all(
          uniqueIds.map(async (id) => {
            try {
              const { data } = await apiClient.get(`/recipes/${id}`)
              recipeLookup.set(id, {
                name: data.name,
                emoji: data.emoji,
                servings: data.servings,
              })
            } catch {
              // Skip recipes that fail to load
            }
          }),
        )

        builtWeeklyMenu = menu.days.map((day) => {
          const date = new Date(day.date + 'T12:00:00')
          const dayOfWeek = date.getDay().toString()
          const isToday = day.date === todayStr
          const recipe = day.recipeId ? recipeLookup.get(day.recipeId) : undefined

          const menuDay: MenuDay = {
            date: day.date,
            dayName: dayNames[dayOfWeek]?.full ?? '',
            dayShort: dayNames[dayOfWeek]?.short ?? '',
            isToday,
            isSkipped: day.skip ?? false,
            meal: recipe
              ? {
                  id: day.recipeId!,
                  name: recipe.name,
                  emoji: recipe.emoji,
                  portions: day.servings,
                }
              : null,
          }

          if (isToday && menuDay.meal) {
            builtTodaysMeal = menuDay.meal
          }

          return menuDay
        })

        // Fetch shopping list summary
        try {
          const { data: shoppingData } = await apiClient.get('/shopping-list', {
            params: { menuId: menu.id },
          })
          const categories = shoppingData.categories ?? []
          const totalItems = categories.reduce(
            (sum: number, cat: { items: unknown[] }) => sum + cat.items.length,
            0,
          )
          const checkedItems = categories.reduce(
            (sum: number, cat: { items: { checked: boolean }[] }) =>
              sum + cat.items.filter((i: { checked: boolean }) => i.checked).length,
            0,
          )
          builtShoppingList = {
            totalItems,
            checkedItems,
            categories: categories.map((c: { name: string; items: unknown[] }) => ({
              name: c.name,
              count: c.items.length,
            })),
          }
        } catch {
          // No shopping list available
        }
      } else {
        currentMenuId.value = null
      }

      dashboardData.value = {
        user: userStore.currentUser!,
        household: {
          id: householdData.id,
          name: householdData.name,
          inviteCode: '',
          members: householdData.members.map(
            (m: { id: string; name: string; role: string }) => {
              const status = statusMap.get(m.id)
              return {
                id: m.id,
                name: m.name,
                role: m.role as 'owner' | 'member' | 'guest',
                isEatingToday: status?.isEatingToday ?? true,
                wantsLunchBox: status?.wantsLunchBox ?? false,
              }
            },
          ),
        },
        todaysMeal: builtTodaysMeal,
        weeklyMenu: builtWeeklyMenu,
        shoppingList: builtShoppingList,
      }
    } catch (e) {
      console.error('Failed to load dashboard:', e)
      error.value = 'Kunde inte ladda dashboarden. Kontrollera din anslutning.'
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
    currentMenuId,
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
    updateMemberLocally,
    clearError,
  }
})
