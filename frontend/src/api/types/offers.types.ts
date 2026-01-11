/**
 * Offers API Types
 *
 * Types for the Tjek API grocery offers integration.
 *
 * API Endpoint: GET /offers/search?q={query}
 * Returns: OfferSearchResponse
 */

// ============================================
// OFFERS
// ============================================

export interface TjekOffer {
  /** Unique offer identifier */
  id: string

  /** Product heading/title */
  heading: string

  /** Product description */
  description: string

  /** Current sale price */
  price: number

  /** Original price before discount */
  prePrice?: number

  /** Currency code (e.g., "SEK") */
  currency: string

  /** Offer start date (ISO string) */
  validFrom: string

  /** Offer end date (ISO string) */
  validTo: string

  /** Store/chain name (e.g., "ICA", "Coop") */
  storeName: string

  /** Store logo URL */
  storeLogo?: string

  /** Nearest store street address */
  storeAddress?: string

  /** Nearest store city */
  storeCity?: string

  /** Product image URL */
  imageUrl?: string
}

// ============================================
// API RESPONSE
// ============================================

export interface OfferSearchResponse {
  /** List of matching offers */
  offers: TjekOffer[]

  /** Total count of offers returned */
  count: number
}
