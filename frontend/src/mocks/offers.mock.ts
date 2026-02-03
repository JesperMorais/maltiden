/**
 * Offers API Mock Data
 *
 * Mock data for testing the Tjek API grocery offers integration without backend.
 */

import type { OfferSearchResponse, TjekOffer } from '@/api/types/offers.types'

const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

// Sample mock offers for common Swedish grocery items
const mockOffers: TjekOffer[] = [
  {
    id: 'offer_1',
    heading: 'Arla Mellanmjölk',
    description: 'Mellanmjölk 1.5% 1L',
    price: 12.90,
    prePrice: 15.90,
    currency: 'SEK',
    validFrom: new Date().toISOString(),
    validTo: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
    storeName: 'ICA Maxi',
    storeLogo: 'https://example.com/ica-logo.png',
    storeAddress: 'Haninge Centrum 1',
    storeCity: 'Haninge',
    imageUrl: 'https://example.com/milk.jpg'
  },
  {
    id: 'offer_2',
    heading: 'Polarbröd',
    description: 'Polarbröd Tunnbröd Original',
    price: 19.90,
    prePrice: 24.90,
    currency: 'SEK',
    validFrom: new Date().toISOString(),
    validTo: new Date(Date.now() + 5 * 24 * 60 * 60 * 1000).toISOString(),
    storeName: 'Coop',
    storeLogo: 'https://example.com/coop-logo.png',
    storeAddress: 'Vega Allé 12',
    storeCity: 'Haninge',
    imageUrl: 'https://example.com/bread.jpg'
  },
  {
    id: 'offer_3',
    heading: 'Kycklingfilé',
    description: 'Svensk Kycklingfilé 500g',
    price: 49.90,
    prePrice: 69.90,
    currency: 'SEK',
    validFrom: new Date().toISOString(),
    validTo: new Date(Date.now() + 3 * 24 * 60 * 60 * 1000).toISOString(),
    storeName: 'Willys',
    storeLogo: 'https://example.com/willys-logo.png',
    storeAddress: 'Haninge Stortorg 8',
    storeCity: 'Haninge',
    imageUrl: 'https://example.com/chicken.jpg'
  },
  {
    id: 'offer_4',
    heading: 'Bananer',
    description: 'Bananer EKO Fairtrade',
    price: 14.90,
    prePrice: 19.90,
    currency: 'SEK',
    validFrom: new Date().toISOString(),
    validTo: new Date(Date.now() + 4 * 24 * 60 * 60 * 1000).toISOString(),
    storeName: 'ICA Kvantum',
    storeLogo: 'https://example.com/ica-logo.png',
    storeAddress: 'Haninge Torg 2',
    storeCity: 'Haninge',
    imageUrl: 'https://example.com/banana.jpg'
  },
  {
    id: 'offer_5',
    heading: 'Tomater',
    description: 'Körsbärstomater 250g',
    price: 17.90,
    prePrice: 24.90,
    currency: 'SEK',
    validFrom: new Date().toISOString(),
    validTo: new Date(Date.now() + 6 * 24 * 60 * 60 * 1000).toISOString(),
    storeName: 'Hemköp',
    storeLogo: 'https://example.com/hemkop-logo.png',
    storeAddress: 'Haninge Allé 5',
    storeCity: 'Haninge',
    imageUrl: 'https://example.com/tomatoes.jpg'
  },
  {
    id: 'offer_6',
    heading: 'Kaffe',
    description: 'Gevalia Mellanrost 500g',
    price: 39.90,
    prePrice: 54.90,
    currency: 'SEK',
    validFrom: new Date().toISOString(),
    validTo: new Date(Date.now() + 10 * 24 * 60 * 60 * 1000).toISOString(),
    storeName: 'ICA Maxi',
    storeLogo: 'https://example.com/ica-logo.png',
    storeAddress: 'Haninge Centrum 1',
    storeCity: 'Haninge',
    imageUrl: 'https://example.com/coffee.jpg'
  },
  {
    id: 'offer_7',
    heading: 'Pasta',
    description: 'Barilla Spaghetti 500g',
    price: 12.90,
    prePrice: 17.90,
    currency: 'SEK',
    validFrom: new Date().toISOString(),
    validTo: new Date(Date.now() + 5 * 24 * 60 * 60 * 1000).toISOString(),
    storeName: 'Willys',
    storeLogo: 'https://example.com/willys-logo.png',
    storeAddress: 'Haninge Stortorg 8',
    storeCity: 'Haninge',
    imageUrl: 'https://example.com/pasta.jpg'
  },
  {
    id: 'offer_8',
    heading: 'Ost',
    description: 'Präst Lagrad 400g',
    price: 49.90,
    prePrice: 64.90,
    currency: 'SEK',
    validFrom: new Date().toISOString(),
    validTo: new Date(Date.now() + 8 * 24 * 60 * 60 * 1000).toISOString(),
    storeName: 'Coop',
    storeLogo: 'https://example.com/coop-logo.png',
    storeAddress: 'Vega Allé 12',
    storeCity: 'Haninge',
    imageUrl: 'https://example.com/cheese.jpg'
  }
]

/**
 * Mock search for grocery offers
 */
export async function mockSearchOffers(query: string): Promise<OfferSearchResponse> {
  await delay(500) // Simulate network delay

  // Filter offers by query (case-insensitive partial match)
  const filtered = mockOffers.filter(
    (offer) =>
      offer.heading.toLowerCase().includes(query.toLowerCase()) ||
      offer.description.toLowerCase().includes(query.toLowerCase())
  )

  return {
    offers: filtered,
    count: filtered.length
  }
}

/**
 * Mock search with location (ignores location for simplicity)
 */
export async function mockSearchOffersAtLocation(
  query: string,
  _lat: number,
  _lng: number,
  _radius: number
): Promise<OfferSearchResponse> {
  return mockSearchOffers(query)
}

/**
 * Mock get top discounts
 */
export async function mockGetTopDiscounts(
  excludeStores: string[] = []
): Promise<OfferSearchResponse> {
  await delay(700)

  // Filter by excluded stores and sort by discount percentage
  const filtered = mockOffers
    .filter((offer) => !excludeStores.includes(offer.storeName))
    .sort((a, b) => {
      const discountA = a.prePrice ? ((a.prePrice - a.price) / a.prePrice) * 100 : 0
      const discountB = b.prePrice ? ((b.prePrice - b.price) / b.prePrice) * 100 : 0
      return discountB - discountA
    })

  return {
    offers: filtered,
    count: filtered.length
  }
}

/**
 * Mock get available stores
 */
export async function mockGetAvailableStores(): Promise<{ stores: string[] }> {
  await delay(300)

  const stores = [...new Set(mockOffers.map((offer) => offer.storeName))]
  return { stores }
}

/**
 * Mock search with filter
 */
export async function mockSearchOffersFiltered(
  query: string,
  excludeStores: string[] = []
): Promise<OfferSearchResponse> {
  const result = await mockSearchOffers(query)

  const filtered = result.offers.filter((offer) => !excludeStores.includes(offer.storeName))

  return {
    offers: filtered,
    count: filtered.length
  }
}
