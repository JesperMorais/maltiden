import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type {
  LandingPageData,
  HeroContent,
  FeaturesContent,
  CtaContent
} from '@/api/types/landing.types'
import { getLandingPageData } from '@/api/landing.api'

/**
 * Landing Page Store
 *
 * Manages the state for the landing page content.
 * Handles loading, error states, and caching.
 */
export const useLandingStore = defineStore('landing', () => {
  // State
  const landingData = ref<LandingPageData | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Computed getters for each section
  const hero = computed<HeroContent | null>(() => landingData.value?.hero ?? null)
  const features = computed<FeaturesContent | null>(() => landingData.value?.features ?? null)
  const cta = computed<CtaContent | null>(() => landingData.value?.cta ?? null)

  // Computed: sorted features by order
  const sortedFeatures = computed(() => {
    if (!features.value) return []
    return [...features.value.features].sort((a, b) => a.order - b.order)
  })

  // Actions
  async function fetchLandingData(forceRefresh = false): Promise<void> {
    // Skip if already loaded and not forcing refresh
    if (landingData.value && !forceRefresh) {
      return
    }

    isLoading.value = true
    error.value = null

    try {
      landingData.value = await getLandingPageData()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load landing page'
      console.error('Landing store error:', e)
    } finally {
      isLoading.value = false
    }
  }

  function clearError(): void {
    error.value = null
  }

  return {
    // State
    landingData,
    isLoading,
    error,

    // Getters
    hero,
    features,
    cta,
    sortedFeatures,

    // Actions
    fetchLandingData,
    clearError
  }
})
