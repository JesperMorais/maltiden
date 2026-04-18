<script setup lang="ts">
import { onMounted, computed, type Component } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useShoppingList } from '@/composables/useShoppingList'
import { useDashboardStore } from '@/stores/dashboard'
import { useSkeleton } from '@/composables/useSkeleton'
import ErrorState from '@/components/common/ErrorState.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import BackLink from '@/components/common/BackLink.vue'
import {
  Check,
  ClipboardList,
  ShoppingCart,
  ShoppingBasket,
  Beef,
  Milk,
  Apple,
  Wheat,
  Sparkles,
} from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()
const dashboardStore = useDashboardStore()
const {
  isLoading,
  error,
  categories,
  totalItems,
  checkedItems,
  progress,
  fetchList,
  toggle,
} = useShoppingList()

const { showSkeleton } = useSkeleton(
  computed(() => isLoading.value),
  { minDuration: 300 },
)

const menuId = computed(() => {
  const queryId = route.query.menuId
  if (typeof queryId === 'string' && queryId) return queryId
  return dashboardStore.menuId
})

onMounted(async () => {
  if (!dashboardStore.dashboardData) {
    await dashboardStore.fetchDashboard()
  }
  if (menuId.value) {
    fetchList(menuId.value)
  }
})

// Swedish locale: comma decimals, trim trailing zeros. Backend already rounds;
// this is purely presentational.
function formatAmount(amount: number): string {
  return Number(amount.toFixed(2)).toLocaleString('sv-SE', { maximumFractionDigits: 2 })
}

// Category → accent color (tinted headers) + icon.
const categoryAccent: Record<string, string> = {
  'Kött & Fisk': '#d6544d',
  Mejeri: '#e8a541',
  'Frukt & Grönt': '#6ba368',
  Grönsaker: '#6ba368',
  Skafferi: '#a67c4e',
  Kryddor: '#8b5a9f',
  Övrigt: '#6b7280',
}

const categoryIcons: Record<string, Component> = {
  'Kött & Fisk': Beef,
  Mejeri: Milk,
  'Frukt & Grönt': Apple,
  Grönsaker: Apple,
  Skafferi: Wheat,
  Kryddor: Sparkles,
}

function accentFor(category: string): string {
  return categoryAccent[category] ?? categoryAccent.Övrigt!
}

function iconFor(category: string): Component {
  return categoryIcons[category] ?? ShoppingBasket
}
</script>

<template>
  <div class="shopping-list-view">
    <header class="header">
      <div class="header-content">
        <BackLink :to="{ name: 'dashboard' }" label="Dashboard" />
        <h1 class="title">Inköpslista</h1>
        <p class="description">Alla ingredienser du behöver till veckans meny.</p>
      </div>
    </header>

    <main class="content">
      <div class="content-container">
        <!-- Skeleton -->
        <div v-if="showSkeleton" class="skeleton-wrapper">
          <div class="skeleton-progress" />
          <div v-for="n in 3" :key="n" class="skeleton-category">
            <div class="skeleton-heading" />
            <div v-for="m in 4" :key="m" class="skeleton-item" />
          </div>
        </div>

        <EmptyState
          v-else-if="!menuId"
          title="Ingen aktiv meny"
          description="Generera en meny först så skapas din inköpslista automatiskt."
          action-label="Generera meny"
          @action="router.push({ name: 'generate-menu' })"
        >
          <template #icon><ClipboardList :size="48" color="var(--text-muted)" /></template>
        </EmptyState>

        <ErrorState v-else-if="error" :description="error" @retry="fetchList(menuId!)" />

        <EmptyState
          v-else-if="totalItems === 0"
          title="Ingen inköpslista"
          description="Generera en meny först så skapas din inköpslista automatiskt."
          action-label="Generera meny"
          @action="router.push({ name: 'generate-menu' })"
        >
          <template #icon><ShoppingCart :size="48" color="var(--text-muted)" /></template>
        </EmptyState>

        <template v-else>
          <div class="progress-card">
            <div class="progress-text">
              <span class="progress-count">{{ checkedItems }} av {{ totalItems }} varor</span>
            </div>
            <div class="progress-bar-track">
              <div class="progress-bar-fill" :style="{ width: `${progress * 100}%` }" />
            </div>
          </div>

          <section
            v-for="category in categories"
            :key="category.name"
            class="category-card"
          >
            <header
              class="category-head"
              :style="{ '--accent-c': accentFor(category.name) }"
            >
              <component :is="iconFor(category.name)" :size="22" class="category-icon" />
              <h2 class="category-name">{{ category.name }}</h2>
              <span class="category-count">
                {{ category.items.filter((i) => i.checked).length }}/{{ category.items.length }}
              </span>
            </header>

            <div class="tile-grid">
              <button
                v-for="item in category.items"
                :key="item.id"
                type="button"
                class="tile"
                :class="{ checked: item.checked }"
                :aria-pressed="item.checked"
                @click="toggle(item.id, !item.checked)"
              >
                <span class="tile-check" :class="{ on: item.checked }">
                  <Check v-if="item.checked" :size="14" :stroke-width="3" />
                </span>
                <span class="tile-name">{{ item.name }}</span>
                <span v-if="item.amount > 0" class="tile-amount">
                  {{ formatAmount(item.amount) }} {{ item.unit }}
                </span>
              </button>
            </div>
          </section>
        </template>
      </div>
    </main>
  </div>
</template>

<style scoped>
.shopping-list-view {
  min-height: 100vh;
  background: var(--bg-primary);
  display: flex;
  flex-direction: column;
}

/* Header */
.header {
  padding: 2rem 2rem 1.5rem;
  background: linear-gradient(180deg, var(--bg-card) 0%, var(--bg-primary) 100%);
  border-bottom: 1px solid var(--border-color);
}

.header-content {
  max-width: 820px;
  margin: 0 auto;
}

.header-content :deep(.back-link) {
  margin-bottom: 0.5rem;
  margin-left: -1rem;
}

.title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 2.5rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
  line-height: 1.2;
}

.description {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1.1rem;
  color: var(--text-secondary);
  margin: 0;
  line-height: 1.6;
}

/* Content */
.content {
  flex: 1;
  padding: 2rem;
}

.content-container {
  max-width: 820px;
  margin: 0 auto;
}

/* Progress card */
.progress-card {
  background: var(--bg-card);
  border-radius: 16px;
  padding: 1.25rem 1.5rem;
  margin-bottom: 2rem;
  border: 1px solid var(--border-color);
}

.progress-text {
  margin-bottom: 0.75rem;
}

.progress-count {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1rem;
  color: var(--text-primary);
}

.progress-bar-track {
  height: 8px;
  background: var(--bg-secondary);
  border-radius: 100px;
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  background: var(--accent);
  border-radius: 100px;
  transition: width 0.3s ease;
}

/* Category card — accent stripe + icon + count (V1 header) */
.category-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 18px;
  overflow: hidden;
  margin-bottom: 1.25rem;
}

.category-head {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.95rem 1.25rem;
  background: color-mix(in srgb, var(--accent-c) 12%, transparent);
  border-left: 6px solid var(--accent-c);
}

.category-icon {
  color: var(--accent-c);
  flex-shrink: 0;
}

.category-name {
  flex: 1;
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.2rem;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.2;
}

.category-count {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.85rem;
  color: var(--text-secondary);
  background: var(--bg-primary);
  padding: 0.2rem 0.6rem;
  border-radius: 999px;
}

/* Tile grid (V4 interior) */
.tile-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 0.75rem;
  padding: 1.25rem;
}

.tile {
  display: grid;
  grid-template-columns: auto 1fr;
  grid-template-rows: auto auto;
  column-gap: 0.65rem;
  row-gap: 0.2rem;
  align-items: center;
  padding: 0.9rem 1rem;
  background: var(--bg-primary);
  border: 2px solid var(--border-color);
  border-radius: 14px;
  cursor: pointer;
  text-align: left;
  transition: all 0.2s ease;
  font-family: inherit;
  -webkit-tap-highlight-color: transparent;
}

.tile:hover {
  border-color: var(--accent-c);
  transform: translateY(-2px);
}

.tile:focus-visible {
  outline: none;
  border-color: var(--accent-c);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent-c) 30%, transparent);
}

.tile.checked {
  background: var(--bg-hover);
  border-color: var(--accent-c);
  opacity: 0.75;
}

.tile-check {
  grid-row: 1;
  grid-column: 1;
  width: 22px;
  height: 22px;
  border: 2px solid var(--border-color);
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--text-on-accent);
  transition: all 0.2s ease;
}

.tile-check.on {
  background: var(--accent-c);
  border-color: var(--accent-c);
}

.tile-name {
  grid-row: 1;
  grid-column: 2;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.95rem;
  color: var(--text-primary);
  line-height: 1.2;
}

.tile.checked .tile-name {
  text-decoration: line-through;
  color: var(--text-secondary);
}

.tile-amount {
  grid-row: 2;
  grid-column: 2;
  font-family: 'Nunito', sans-serif;
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--text-secondary);
}

/* Skeleton */
.skeleton-wrapper {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.skeleton-progress {
  height: 64px;
  background: var(--bg-card);
  border-radius: 16px;
  animation: pulse 1.5s ease-in-out infinite;
}

.skeleton-category {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.skeleton-heading {
  height: 24px;
  width: 140px;
  background: var(--bg-secondary);
  border-radius: 8px;
  margin-bottom: 0.25rem;
  animation: pulse 1.5s ease-in-out infinite;
}

.skeleton-item {
  height: 44px;
  background: var(--bg-secondary);
  border-radius: 8px;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

/* Responsive */
@media (max-width: 768px) {
  .header {
    padding: 1.5rem 1rem 1rem;
  }

  .content {
    padding: 1.5rem 1rem;
  }

  .title {
    font-size: 1.75rem;
  }

  .description {
    font-size: 0.95rem;
  }

  .tile-grid {
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: 0.55rem;
    padding: 0.85rem;
  }

  .tile {
    padding: 0.75rem 0.85rem;
  }
}
</style>
