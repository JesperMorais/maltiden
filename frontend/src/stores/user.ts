import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { User, UserRole } from '@/api/types/dashboard.types'
import { tokenUtils } from '@/utils/token'
import * as authApi from '@/api/auth.api'
import { isAxiosError } from 'axios'

function getSwedishAuthError(e: unknown, fallback: string): string {
  if (isAxiosError(e)) {
    const status = e.response?.status
    const code = e.response?.data?.error as string | undefined
    if (status === 401 || code === 'invalid_credentials') return 'Fel e-post eller lösenord'
    if (status === 409 || code === 'email_taken' || code === 'email_already_exists')
      return 'E-postadressen är redan registrerad'
    if (code === 'weak_password')
      return 'Lösenordet måste innehålla minst 3 av: versaler, gemener, siffror, specialtecken'
    if (code === 'invalid_email') return 'Ogiltig e-postadress'
    if (status === 400) return 'Ogiltig förfrågan — kontrollera dina uppgifter'
    if (!e.response) return 'Kunde inte nå servern — kontrollera din internetanslutning'
  }
  return fallback
}

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
  const isLoading = ref(false)
  const error = ref<string | null>(null)

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
   * Logout - clear token and redirect to login
   */
  function logout() {
    tokenUtils.remove()
    clearUser()
    // Use window.location for hard redirect to clear any state
    window.location.href = '/login'
  }

  /**
   * Check if user can perform a member-only action
   */
  function canPerformAction(requiresMember = true): boolean {
    if (!requiresMember) return true
    return isMember.value
  }

  /**
   * Initialize user from token (if exists)
   */
  function initFromToken() {
    if (tokenUtils.exists()) {
      // In real app, we'd validate token with backend
      // For now, just mark as potentially authenticated
      isAuthenticated.value = true
    }
  }

  /**
   * Clear any error state
   */
  function clearError() {
    error.value = null
  }

  /**
   * Login with email and password
   */
  async function login(email: string, password: string): Promise<boolean> {
    isLoading.value = true
    error.value = null

    try {
      const response = await authApi.login(email, password)
      tokenUtils.set(response.token)
      setUser({
        id: response.user.id,
        name: response.user.name,
        email: response.user.email,
        role: 'member' // Default role, will be updated when fetching household
      })
      return true
    } catch (e) {
      error.value = getSwedishAuthError(e, 'Inloggningen misslyckades')
      return false
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Register a new user
   */
  async function register(name: string, email: string, password: string): Promise<boolean> {
    isLoading.value = true
    error.value = null

    try {
      const response = await authApi.register(name, email, password)
      tokenUtils.set(response.token)
      setUser({
        id: response.user.id,
        name: response.user.name,
        email: response.user.email,
        role: 'member'
      })
      return true
    } catch (e) {
      error.value = getSwedishAuthError(e, 'Registreringen misslyckades')
      return false
    } finally {
      isLoading.value = false
    }
  }

  return {
    // State
    currentUser,
    isAuthenticated,
    isLoading,
    error,

    // Computed
    userRole,
    isOwner,
    isMember,
    isGuest,
    userName,

    // Actions
    setUser,
    clearUser,
    logout,
    canPerformAction,
    initFromToken,
    clearError,
    login,
    register
  }
})
