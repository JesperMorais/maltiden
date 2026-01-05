import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { User, UserRole } from '@/api/types/dashboard.types'

/**
 * User Store
 *
 * Manages current user state and role-based access control.
 * Used to determine what actions are available (member vs guest).
 */
export const useUserStore = defineStore('user', () => {
  // State
  const currentUser = ref<User | null>(null)
  const isAuthenticated = ref(false)

  // Computed
  const userRole = computed<UserRole | null>(() => currentUser.value?.role ?? null)
  const isOwner = computed(() => currentUser.value?.role === 'owner')
  const isMember = computed(() =>
    currentUser.value?.role === 'member' || currentUser.value?.role === 'owner'
  )
  const isGuest = computed(() => currentUser.value?.role === 'guest')
  const userName = computed(() => currentUser.value?.name ?? 'Gäst')

  // Actions
  function setUser(user: User) {
    currentUser.value = user
    isAuthenticated.value = true
  }

  function clearUser() {
    currentUser.value = null
    isAuthenticated.value = false
  }

  /**
   * Check if user can perform a member-only action
   */
  function canPerformAction(requiresMember = true): boolean {
    if (!requiresMember) return true
    return isMember.value
  }

  return {
    // State
    currentUser,
    isAuthenticated,

    // Computed
    userRole,
    isOwner,
    isMember,
    isGuest,
    userName,

    // Actions
    setUser,
    clearUser,
    canPerformAction
  }
})
