import { ref } from 'vue'
import { defineStore } from 'pinia'

const THEME_KEY = 'maltiden_theme'

/**
 * Theme Store
 *
 * Manages dark/light mode with localStorage persistence.
 * Theme is applied via data-theme attribute on <html> element.
 */
export const useThemeStore = defineStore('theme', () => {
  const isDarkMode = ref(false)

  /**
   * Initialize theme from localStorage or system preference
   */
  function initTheme() {
    const saved = localStorage.getItem(THEME_KEY)

    if (saved) {
      isDarkMode.value = saved === 'dark'
    } else {
      // Check system preference
      isDarkMode.value = window.matchMedia('(prefers-color-scheme: dark)').matches
    }

    applyTheme()
  }

  /**
   * Toggle between light and dark mode
   */
  function toggleDarkMode() {
    isDarkMode.value = !isDarkMode.value
    applyTheme()
  }

  /**
   * Set theme explicitly
   */
  function setDarkMode(dark: boolean) {
    isDarkMode.value = dark
    applyTheme()
  }

  /**
   * Apply current theme to DOM and persist to localStorage
   */
  function applyTheme() {
    const theme = isDarkMode.value ? 'dark' : 'light'
    document.documentElement.setAttribute('data-theme', theme)
    localStorage.setItem(THEME_KEY, theme)
  }

  return {
    isDarkMode,
    initTheme,
    toggleDarkMode,
    setDarkMode
  }
})
