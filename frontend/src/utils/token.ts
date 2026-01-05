/**
 * JWT Token utilities for authentication
 * Handles storing and retrieving tokens from localStorage
 */

const TOKEN_KEY = 'maltiden_token'

export const tokenUtils = {
  /**
   * Get the stored JWT token
   */
  get: (): string | null => {
    return localStorage.getItem(TOKEN_KEY)
  },

  /**
   * Store a JWT token
   */
  set: (token: string): void => {
    localStorage.setItem(TOKEN_KEY, token)
  },

  /**
   * Remove the stored token (logout)
   */
  remove: (): void => {
    localStorage.removeItem(TOKEN_KEY)
  },

  /**
   * Check if a token exists
   */
  exists: (): boolean => {
    return !!localStorage.getItem(TOKEN_KEY)
  }
}
