<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { searchOffersFiltered, getTopDiscounts, getAvailableStores } from '@/api/offers.api'
import type { TjekOffer } from '@/api/types/offers.types'

const query = ref('')
const offers = ref<TjekOffer[]>([])
const isLoading = ref(false)
const error = ref<string | null>(null)
const hasSearched = ref(false)
const showingDiscounts = ref(false)

// Store filter
const availableStores = ref<string[]>([])
const excludedStores = ref<string[]>([])
const showFilterDropdown = ref(false)

onMounted(async () => {
  try {
    const result = await getAvailableStores()
    availableStores.value = result.stores || []
  } catch (e) {
    console.error('Failed to load stores:', e)
  }
})

function toggleStoreExclusion(store: string) {
  const index = excludedStores.value.indexOf(store)
  if (index === -1) {
    excludedStores.value.push(store)
  } else {
    excludedStores.value.splice(index, 1)
  }
}

function isStoreExcluded(store: string): boolean {
  return excludedStores.value.includes(store)
}

async function handleSearch() {
  const searchTerm = query.value.trim()
  if (!searchTerm || isLoading.value) return

  isLoading.value = true
  error.value = null
  hasSearched.value = true
  showingDiscounts.value = false

  try {
    const result = await searchOffersFiltered(searchTerm, excludedStores.value)
    offers.value = result.offers || []
  } catch (e) {
    error.value = 'Kunde inte hämta erbjudanden. Kontrollera att backend körs.'
    offers.value = []
    console.error('Offer search error:', e)
  } finally {
    isLoading.value = false
  }
}

async function handleShowDiscounts() {
  if (isLoading.value) return

  isLoading.value = true
  error.value = null
  hasSearched.value = true
  showingDiscounts.value = true
  query.value = ''

  try {
    const result = await getTopDiscounts(excludedStores.value)
    offers.value = result.offers || []
  } catch (e) {
    error.value = 'Kunde inte hämta rabatter. Kontrollera att backend körs.'
    offers.value = []
    console.error('Discounts error:', e)
  } finally {
    isLoading.value = false
  }
}

function formatPrice(offer: TjekOffer): string {
  const price = offer.price.toFixed(2)
  if (offer.prePrice && offer.prePrice > offer.price) {
    return `${price} kr (ord. ${offer.prePrice.toFixed(2)} kr)`
  }
  return `${price} kr`
}

function formatValidity(offer: TjekOffer): string {
  const from = new Date(offer.validFrom).toLocaleDateString('sv-SE')
  const to = new Date(offer.validTo).toLocaleDateString('sv-SE')
  return `${from} - ${to}`
}

function calculateDiscount(offer: TjekOffer): number | null {
  if (!offer.prePrice || offer.prePrice <= offer.price) return null
  return Math.round((1 - offer.price / offer.prePrice) * 100)
}
</script>

<template>
  <div class="offer-search">
    <header class="search-header">
      <h1>Tjek API POC</h1>
      <p class="subtitle">Sök efter matvaror i butiker nära Stockholm</p>
    </header>

    <div class="search-form">
      <input
        v-model="query"
        type="text"
        placeholder="Sök produkt (t.ex. mjölk, bröd, ägg, kyckling)..."
        @keyup.enter="handleSearch"
        class="search-input"
      />
      <button @click="handleSearch" :disabled="isLoading" class="search-button">
        {{ isLoading && !showingDiscounts ? 'Söker...' : 'Sök' }}
      </button>
      <button @click="handleShowDiscounts" :disabled="isLoading" class="discount-button">
        {{ isLoading && showingDiscounts ? 'Laddar...' : 'Bästa rabatter' }}
      </button>
    </div>

    <div class="filter-section">
      <button @click="showFilterDropdown = !showFilterDropdown" class="filter-toggle">
        Filtrera butiker
        <span v-if="excludedStores.length > 0" class="filter-badge">{{ excludedStores.length }}</span>
      </button>

      <div v-if="showFilterDropdown" class="filter-dropdown">
        <p class="filter-hint">Avmarkera butiker du vill exkludera:</p>
        <div class="store-list">
          <label v-for="store in availableStores" :key="store" class="store-checkbox">
            <input
              type="checkbox"
              :checked="!isStoreExcluded(store)"
              @change="toggleStoreExclusion(store)"
            />
            {{ store }}
          </label>
        </div>
      </div>
    </div>

    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <div v-if="offers.length > 0" class="results">
      <p class="results-count">
        {{ showingDiscounts ? 'Topp ' + offers.length + ' rabatter just nu' : 'Hittade ' + offers.length + ' erbjudanden' }}
      </p>

      <div class="offers-grid">
        <div v-for="offer in offers" :key="offer.id" class="offer-card">
          <div class="offer-header">
            <img
              v-if="offer.storeLogo"
              :src="offer.storeLogo"
              :alt="offer.storeName"
              class="store-logo"
            />
            <div class="store-info">
              <span class="store-name">{{ offer.storeName }}</span>
              <span v-if="offer.storeAddress || offer.storeCity" class="store-location">
                {{ [offer.storeAddress, offer.storeCity].filter(Boolean).join(', ') }}
              </span>
            </div>
            <span v-if="calculateDiscount(offer)" class="discount-badge">
              -{{ calculateDiscount(offer) }}%
            </span>
          </div>

          <div v-if="offer.imageUrl" class="offer-image-container">
            <img :src="offer.imageUrl" :alt="offer.heading" class="offer-image" />
          </div>

          <h3 class="offer-heading">{{ offer.heading }}</h3>
          <p v-if="offer.description" class="offer-description">{{ offer.description }}</p>

          <div class="offer-price">{{ formatPrice(offer) }}</div>
          <div class="offer-validity">Gäller: {{ formatValidity(offer) }}</div>
        </div>
      </div>
    </div>

    <p v-else-if="hasSearched && !isLoading && offers.length === 0 && !error" class="no-results">
      Inga erbjudanden hittades för "{{ query }}"
    </p>

    <div v-if="!hasSearched && !isLoading" class="instructions">
      <h2>Så här testar du</h2>
      <ol>
        <li>Se till att backend körs: <code>go run cmd/server/main.go</code></li>
        <li>Sök efter svenska produktnamn: mjölk, bröd, ägg, ost, kyckling</li>
        <li>Resultat visar erbjudanden från butiker i Stockholmsområdet</li>
      </ol>
      <p class="note">Ingen API-nyckel behövs!</p>
    </div>
  </div>
</template>

<style scoped>
.offer-search {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.search-header {
  text-align: center;
  margin-bottom: 2rem;
}

.search-header h1 {
  font-family: 'Fraunces', serif;
  font-size: 2.5rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
}

.subtitle {
  color: var(--text-secondary);
  font-size: 1.1rem;
  margin: 0;
}

.search-form {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  margin-bottom: 2rem;
  max-width: 700px;
  margin-left: auto;
  margin-right: auto;
  justify-content: center;
}

.search-input {
  flex: 1;
  padding: 0.875rem 1.25rem;
  border: 2px solid var(--border);
  border-radius: 12px;
  font-size: 1rem;
  font-family: 'Nunito', sans-serif;
  background: var(--bg-primary);
  color: var(--text-primary);
  transition: border-color 0.2s, box-shadow 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 3px var(--accent-light);
}

.search-button {
  padding: 0.875rem 2rem;
  background: var(--accent);
  color: var(--text-on-accent);
  border: none;
  border-radius: 12px;
  cursor: pointer;
  font-weight: 600;
  font-size: 1rem;
  font-family: 'Nunito', sans-serif;
  transition: background 0.2s, transform 0.2s;
}

.search-button:hover:not(:disabled) {
  background: var(--accent-dark, #388E3C);
  transform: translateY(-1px);
}

.search-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.discount-button {
  padding: 0.875rem 1.5rem;
  background: #e53935;
  color: white;
  border: none;
  border-radius: 12px;
  cursor: pointer;
  font-weight: 600;
  font-size: 1rem;
  font-family: 'Nunito', sans-serif;
  transition: background 0.2s, transform 0.2s;
  white-space: nowrap;
}

.discount-button:hover:not(:disabled) {
  background: #c62828;
  transform: translateY(-1px);
}

.discount-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.filter-section {
  text-align: center;
  margin-bottom: 1.5rem;
  position: relative;
}

.filter-toggle {
  padding: 0.5rem 1rem;
  background: var(--bg-secondary, #f5f5f5);
  border: 1px solid var(--border, #ddd);
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.9rem;
  font-family: 'Nunito', sans-serif;
  color: var(--text-secondary);
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.filter-toggle:hover {
  background: var(--bg-primary);
  border-color: var(--accent);
}

.filter-badge {
  background: #e53935;
  color: white;
  font-size: 0.75rem;
  padding: 0.1rem 0.4rem;
  border-radius: 10px;
  font-weight: 600;
}

.filter-dropdown {
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  margin-top: 0.5rem;
  background: var(--bg-primary, white);
  border: 1px solid var(--border, #ddd);
  border-radius: 12px;
  padding: 1rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 100;
  min-width: 250px;
  max-width: 90vw;
}

.filter-hint {
  font-size: 0.85rem;
  color: var(--text-muted, #999);
  margin: 0 0 0.75rem;
}

.store-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-height: 300px;
  overflow-y: auto;
}

.store-checkbox {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 4px;
  font-size: 0.9rem;
}

.store-checkbox:hover {
  background: var(--bg-secondary, #f5f5f5);
}

.store-checkbox input {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.error-message {
  color: #d32f2f;
  padding: 1rem 1.5rem;
  background: #ffebee;
  border-radius: 12px;
  margin-bottom: 2rem;
  text-align: center;
}

.results-count {
  color: var(--text-secondary);
  margin-bottom: 1.5rem;
  text-align: center;
}

.offers-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1.5rem;
}

.offer-card {
  background: var(--bg-primary);
  border: 1px solid var(--border);
  border-radius: 16px;
  padding: 1.5rem;
  transition: box-shadow 0.3s, transform 0.3s;
}

.offer-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
  transform: translateY(-4px);
}

.offer-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.store-logo {
  width: 28px;
  height: 28px;
  object-fit: contain;
  border-radius: 4px;
}

.store-info {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
}

.store-name {
  font-weight: 600;
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.store-location {
  font-size: 0.75rem;
  color: var(--text-muted, #999);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.offer-image-container {
  margin: 0.75rem -1.5rem;
  background: #f8f8f8;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 120px;
}

.offer-image {
  max-width: 100%;
  max-height: 150px;
  object-fit: contain;
}

.discount-badge {
  margin-left: auto;
  background: #e53935;
  color: white;
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 700;
}

.offer-heading {
  font-family: 'Fraunces', serif;
  font-size: 1.2rem;
  margin: 0 0 0.5rem;
  color: var(--text-primary);
  line-height: 1.3;
}

.offer-description {
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin: 0 0 1rem;
  line-height: 1.5;
}

.offer-price {
  font-size: 1.4rem;
  font-weight: 700;
  color: var(--accent);
}

.offer-validity {
  font-size: 0.8rem;
  color: var(--text-muted, #999);
  margin-top: 0.5rem;
}

.no-results {
  text-align: center;
  color: var(--text-secondary);
  padding: 3rem;
  font-size: 1.1rem;
}

.instructions {
  max-width: 500px;
  margin: 2rem auto;
  padding: 2rem;
  background: var(--bg-secondary, #f5f5f5);
  border-radius: 16px;
}

.instructions h2 {
  font-family: 'Fraunces', serif;
  font-size: 1.3rem;
  margin: 0 0 1rem;
  color: var(--text-primary);
}

.instructions ol {
  margin: 0;
  padding-left: 1.5rem;
  color: var(--text-secondary);
  line-height: 1.8;
}

.instructions code {
  background: var(--bg-primary);
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.9rem;
}

.instructions .note {
  margin-top: 1rem;
  padding: 0.75rem;
  background: #e8f5e9;
  color: #2e7d32;
  border-radius: 8px;
  font-weight: 600;
  text-align: center;
}
</style>
