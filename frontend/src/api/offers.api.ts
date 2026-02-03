/**
 * Offers API Service
 * Handles grocery offer search via Tjek API
 */

import apiClient from './client'
import { USE_MOCKS } from '@/mocks'
import {
  mockSearchOffers,
  mockSearchOffersAtLocation,
  mockGetTopDiscounts,
  mockGetAvailableStores,
  mockSearchOffersFiltered
} from '@/mocks/offers.mock'
import type { OfferSearchResponse } from './types/offers.types'

/**
 * Search for grocery offers near Haninge
 * @param query - Product name to search for (e.g., "mjölk", "bröd")
 */
export async function searchOffers(query: string): Promise<OfferSearchResponse> {
  if (USE_MOCKS) {
    return mockSearchOffers(query)
  }

  const { data } = await apiClient.get<OfferSearchResponse>('/offers/search', {
    params: { q: query },
    timeout: 30000 // 30 seconds timeout for catalog fetching
  })
  return data
}

/**
 * Search for grocery offers with custom location
 * @param query - Product name to search for
 * @param lat - Latitude
 * @param lng - Longitude
 * @param radius - Search radius in meters (default 10000)
 */
export async function searchOffersAtLocation(
  query: string,
  lat: number,
  lng: number,
  radius: number = 10000
): Promise<OfferSearchResponse> {
  if (USE_MOCKS) {
    return mockSearchOffersAtLocation(query, lat, lng, radius)
  }

  const { data } = await apiClient.get<OfferSearchResponse>('/offers/search', {
    params: { q: query, lat, lng, radius }
  })
  return data
}

/**
 * Get top discounted offers sorted by discount percentage
 * @param excludeStores - Store names to exclude
 */
export async function getTopDiscounts(excludeStores: string[] = []): Promise<OfferSearchResponse> {
  if (USE_MOCKS) {
    return mockGetTopDiscounts(excludeStores)
  }

  const { data } = await apiClient.get<OfferSearchResponse>('/offers/discounts', {
    params: excludeStores.length > 0 ? { exclude: excludeStores.join(',') } : {},
    timeout: 30000
  })
  return data
}

/**
 * Get available stores in the area
 */
export async function getAvailableStores(): Promise<{ stores: string[] }> {
  if (USE_MOCKS) {
    return mockGetAvailableStores()
  }

  const { data } = await apiClient.get<{ stores: string[] }>('/offers/stores', {
    timeout: 10000
  })
  return data
}

/**
 * Search for grocery offers with exclusion filter
 * @param query - Product name to search for
 * @param excludeStores - Store names to exclude
 */
export async function searchOffersFiltered(
  query: string,
  excludeStores: string[] = []
): Promise<OfferSearchResponse> {
  if (USE_MOCKS) {
    return mockSearchOffersFiltered(query, excludeStores)
  }

  const params: Record<string, string> = { q: query }
  if (excludeStores.length > 0) {
    params.exclude = excludeStores.join(',')
  }
  const { data } = await apiClient.get<OfferSearchResponse>('/offers/search', {
    params,
    timeout: 30000
  })
  return data
}
