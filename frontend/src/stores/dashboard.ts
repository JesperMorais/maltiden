import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { DashboardData, Meal, MenuDay, Household, HouseholdMember, ShoppingListSummary } from '@/api/types/dashboard.types'
import { mockDashboardData } from '@/mocks/dashboard.mock'
import { USE_MOCKS } from '@/mocks'
import { useUserStore } from './user'
import apiClient from '@/api/client'
import { getCurrentMenu, saveMenu } from '@/api/menu.api'
import type { Menu, SaveMenuDay } from '@/api/menu.api'
import { getShoppingList } from '@/api/shopping.api'
import type { ShoppingList } from '@/api/shopping.api'
import { updateHousehold as apiUpdateHousehold } from '@/api/household.api'

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
  const selectedDate = ref<string | null>(null)

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

  function setSelectedDate(date: string | null) {
    selectedDate.value = date
  }

  const displayDate = computed(() => {
    if (selectedDate.value) return selectedDate.value
    return new Date().toISOString().split('T')[0]!
  })

  const membersForDisplayDate = computed(() =>
    householdMembers.value.filter((m) => isMemberEatingDay(displayDate.value, m.id))
  )

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

  function toShoppingSummary(list: ShoppingList): ShoppingListSummary {
    let totalItems = 0
    let checkedItems = 0
    const categories: { name: string; count: number }[] = []
    for (const cat of list.categories) {
      totalItems += cat.items.length
      checkedItems += cat.items.filter((i) => i.checked).length
      categories.push({ name: cat.name, count: cat.items.length })
    }
    return { totalItems, checkedItems, categories }
  }

  async function refreshShoppingList() {
    if (!currentMenuId.value) return
    try {
      const list = await getShoppingList(currentMenuId.value)
      if (dashboardData.value) {
        dashboardData.value.shoppingList = toShoppingSummary(list)
      }
    } catch (e) {
      console.warn('Could not refresh shopping list:', e)
    }
  }

  // Actions
  async function fetchDashboard(forceRefresh = false): Promise<void> {
    if (dashboardData.value && !forceRefresh) {
      return
    }

    isLoading.value = true
    error.value = null

    if (USE_MOCKS) {
      // In mock mode, use mock data directly — no network requests
      try {
        const menu = await getCurrentMenu()
        if (menu) {
          currentMenuId.value = menu.id
          const transformed = transformMenuToDashboard(menu)
          dashboardData.value = {
            ...mockDashboardData,
            todaysMeal: transformed.todaysMeal,
            weeklyMenu: transformed.weeklyMenu,
          }
        } else {
          dashboardData.value = mockDashboardData
          currentMenuId.value = 'menu_current'
        }

        if (dashboardData.value.user) {
          userStore.setUser(dashboardData.value.user)
        }
      } finally {
        isLoading.value = false
      }
      return
    }

    try {
      // Fetch real household data from backend
      const { data: householdData } = await apiClient.get('/households/me', {
        skipAuthRedirect: true,
      })

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

      // Populate shopping list with real data if we have a menu
      if (currentMenuId.value) {
        refreshShoppingList()
      }
    } catch (e) {
      // If backend fails (including 401 with skipAuthRedirect), fall back to mock data.
      // This is intentional: a stale-token 401 is swallowed here to prevent the
      // post-login redirect loop. See client.ts markAuthSuccess() for context.
      const msg = e instanceof Error ? e.message : String(e)
      console.warn('Using mock dashboard data —', msg)
      dashboardData.value = mockDashboardData
      currentMenuId.value = 'menu_current'

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

  // Per-day lunchbox count state
  const dayLunchBox = ref<Record<string, number>>({})

  function getDayLunchBoxCount(date: string): number {
    const val = dayLunchBox.value[date]
    return typeof val === 'number' ? val : 0
  }

  function setDayLunchBoxCount(date: string, count: number) {
    dayLunchBox.value[date] = Math.max(0, count)
    localStorage.setItem('maltiden_day_lunchbox', JSON.stringify(dayLunchBox.value))
  }

  function initLunchBoxDays() {
    const stored = localStorage.getItem('maltiden_day_lunchbox')
    if (stored) {
      try {
        const parsed = JSON.parse(stored) as Record<string, unknown>
        const cleaned: Record<string, number> = {}
        for (const [key, val] of Object.entries(parsed)) {
          cleaned[key] = typeof val === 'number' ? val : 0
        }
        dayLunchBox.value = cleaned
      } catch {
        dayLunchBox.value = {}
      }
    }
  }

  // Per-day member exclusions (who is NOT eating on a given date)
  const dayMemberExclusions = ref<Record<string, string[]>>({})

  function initDayMemberExclusions() {
    const stored = localStorage.getItem('maltiden_day_member_exclusions')
    if (stored) {
      try {
        const parsed = JSON.parse(stored) as Record<string, unknown>
        const cleaned: Record<string, string[]> = {}
        for (const [key, val] of Object.entries(parsed)) {
          if (Array.isArray(val)) {
            cleaned[key] = val.filter((v): v is string => typeof v === 'string')
          }
        }
        dayMemberExclusions.value = cleaned
      } catch {
        dayMemberExclusions.value = {}
      }
    }
  }

  function isMemberEatingDay(date: string, memberId: string): boolean {
    const exclusions = dayMemberExclusions.value[date]
    if (!exclusions) return true
    return !exclusions.includes(memberId)
  }

  function toggleMemberDay(date: string, memberId: string) {
    const current = dayMemberExclusions.value[date] ?? []
    const idx = current.indexOf(memberId)
    if (idx >= 0) {
      current.splice(idx, 1)
    } else {
      current.push(memberId)
    }
    dayMemberExclusions.value[date] = current
    localStorage.setItem('maltiden_day_member_exclusions', JSON.stringify(dayMemberExclusions.value))

    // Clamp lunchbox count if it exceeds new member count
    const eatingCount = getMembersEatingDay(date).length
    const currentLunchBoxes = getDayLunchBoxCount(date)
    if (currentLunchBoxes > eatingCount) {
      setDayLunchBoxCount(date, eatingCount)
    }

    // Persist new servings and refresh shopping list
    updateDayServings(date, eatingCount, getDayLunchBoxCount(date))
  }

  function getMembersEatingDay(date: string): HouseholdMember[] {
    return householdMembers.value.filter((m) => isMemberEatingDay(date, m.id))
  }

  // Whether the dashboard loaded from the real API (not mock fallback)
  const isUsingRealData = computed(() =>
    currentMenuId.value !== null && currentMenuId.value !== 'menu_current',
  )

  async function updateDayServings(date: string, baseServings: number, lunchBoxCount: number) {
    if (!currentMenuId.value) return

    const totalServings = baseServings + lunchBoxCount
    const menu = weeklyMenu.value
    const days: SaveMenuDay[] = menu.map((day) => ({
      date: day.date,
      recipeId: day.meal?.id,
      servings: day.date === date ? totalServings : (day.meal?.portions ?? 4),
      skip: day.isSkipped,
    }))

    try {
      const savedMenu = await saveMenu(days)
      // Update store with saved servings to prevent stale data on next toggle
      const transformed = transformMenuToDashboard(savedMenu)
      if (dashboardData.value) {
        dashboardData.value.weeklyMenu = transformed.weeklyMenu
        dashboardData.value.todaysMeal = transformed.todaysMeal
      }
      await refreshShoppingList()
    } catch (e) {
      console.warn('Could not save menu servings:', e)
    }
  }

  function clearError() {
    error.value = null
  }

  async function updateHouseholdName(name: string): Promise<boolean> {
    const trimmed = name.trim()
    if (!trimmed || !dashboardData.value?.household) return false

    const prev = dashboardData.value.household.name
    dashboardData.value.household.name = trimmed
    try {
      await apiUpdateHousehold({ name: trimmed })
      return true
    } catch (e) {
      dashboardData.value.household.name = prev
      error.value = e instanceof Error ? e.message : 'Kunde inte uppdatera hushållsnamn'
      return false
    }
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
    currentMenuId,

    // Selected date
    selectedDate,
    setSelectedDate,
    displayDate,
    membersForDisplayDate,

    // Shopping
    shoppingList,
    shoppingItemsRemaining,
    refreshShoppingList,

    // Day Lunchbox
    dayLunchBox,
    getDayLunchBoxCount,
    setDayLunchBoxCount,
    initLunchBoxDays,
    updateDayServings,

    // Day Member Exclusions
    dayMemberExclusions,
    initDayMemberExclusions,
    isMemberEatingDay,
    toggleMemberDay,
    getMembersEatingDay,

    // Data source
    isUsingRealData,

    // Actions
    fetchDashboard,
    updateMemberLocally,
    updateHouseholdName,
    clearError
  }
})
