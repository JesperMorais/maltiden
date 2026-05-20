import { ref } from 'vue'
import { defineStore } from 'pinia'
import { getTopDiscounts } from '@/api/offers.api'
import type { TjekOffer } from '@/api/types/offers.types'

const TTL_MS = 10 * 60 * 1000 // 10 minutes

export const useOffersStore = defineStore('offers', () => {
  const offers = ref<TjekOffer[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const lastFetched = ref<number | null>(null)

  // Precomputed lowercase haystacks: id → "heading description"
  const haystacks = new Map<string, string>()

  async function fetchOffers(force = false): Promise<void> {
    const now = Date.now()
    if (!force && lastFetched.value !== null && now - lastFetched.value < TTL_MS) {
      return
    }
    if (isLoading.value) return

    isLoading.value = true
    error.value = null

    try {
      const result = await getTopDiscounts()
      offers.value = result.offers
      haystacks.clear()
      for (const offer of result.offers) {
        haystacks.set(offer.id, `${offer.heading} ${offer.description}`.toLowerCase())
      }
      lastFetched.value = Date.now()
    } catch {
      error.value = 'Kunde inte hämta erbjudanden'
    } finally {
      isLoading.value = false
    }
  }

  function matchOffer(itemName: string): TjekOffer | null {
    if (offers.value.length === 0) return null
    // Split item name into words (min 3 chars to avoid noise)
    const words = itemName
      .toLowerCase()
      .split(/\s+/)
      .filter((w) => w.length >= 3)
    if (words.length === 0) return null

    for (const offer of offers.value) {
      const haystack = haystacks.get(offer.id) ?? ''
      if (words.some((word) => new RegExp(`\\b${word}`, 'i').test(haystack))) {
        return offer
      }
    }
    return null
  }

  return {
    offers,
    isLoading,
    error,
    lastFetched,
    fetchOffers,
    matchOffer,
  }
})
