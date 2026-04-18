<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useShoppingList } from '@/composables/useShoppingList'
import { useDashboardStore } from '@/stores/dashboard'
import { useSkeleton } from '@/composables/useSkeleton'
import ErrorState from '@/components/common/ErrorState.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import BackLink from '@/components/common/BackLink.vue'
import { ClipboardList, ShoppingCart } from 'lucide-vue-next'

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
  { minDuration: 300 }
)

const menuId = computed(() => {
  const queryId = route.query.menuId
  if (typeof queryId === 'string' && queryId) return queryId
  return dashboardStore.menuId
})

onMounted(async () => {
  // Ensure dashboard is loaded so we have menuId
  if (!dashboardStore.dashboardData) {
    await dashboardStore.fetchDashboard()
  }
  if (menuId.value) {
    fetchList(menuId.value)
  }
})

// Swedish locale: comma as decimal separator. Trim trailing zeros so "3.00"
// shows as "3" while "3.50" becomes "3,5". Backend already rounds the value;
// this is purely presentational.
function formatAmount(amount: number): string {
  return Number(amount.toFixed(2))
    .toLocaleString('sv-SE', { maximumFractionDigits: 2 })
}
</script>

<template>
  <div class="shopping-list-view">
    <!-- Header -->
    <header class="header">
      <div class="header-content">
        <BackLink :to="{ name: 'dashboard' }" label="Dashboard" />
        <h1 class="title">Inköpslista</h1>
        <p class="description">
          Alla ingredienser du behöver till veckans meny.
        </p>
      </div>
    </header>

    <!-- Main content -->
    <main class="content">
      <div class="content-container">
        <!-- Skeleton loading -->
        <div v-if="showSkeleton" class="skeleton-wrapper">
          <div class="skeleton-progress" />
          <div v-for="n in 3" :key="n" class="skeleton-category">
            <div class="skeleton-heading" />
            <div v-for="m in 4" :key="m" class="skeleton-item" />
          </div>
        </div>

        <!-- No menu state -->
        <EmptyState
          v-else-if="!menuId"
          title="Ingen aktiv meny"
          description="Generera en meny först så skapas din inköpslista automatiskt."
          action-label="Generera meny"
          @action="router.push({ name: 'generate-menu' })"
        >
          <template #icon><ClipboardList :size="48" color="var(--text-muted)" /></template>
        </EmptyState>

        <!-- Error state -->
        <ErrorState v-else-if="error" :description="error" @retry="fetchList(menuId!)" />

        <!-- Empty state -->
        <EmptyState
          v-else-if="totalItems === 0"
          title="Ingen inköpslista"
          description="Generera en meny först så skapas din inköpslista automatiskt."
          action-label="Generera meny"
          @action="router.push({ name: 'generate-menu' })"
        >
          <template #icon><ShoppingCart :size="48" color="var(--text-muted)" /></template>
        </EmptyState>

        <!-- Shopping list -->
        <template v-else>
          <!-- Progress summary -->
          <div class="progress-card">
            <div class="progress-text">
              <span class="progress-count">{{ checkedItems }} av {{ totalItems }} varor</span>
            </div>
            <div class="progress-bar-track">
              <div
                class="progress-bar-fill"
                :style="{ width: `${progress * 100}%` }"
              />
            </div>
          </div>

          <!-- Categories -->
          <div
            v-for="category in categories"
            :key="category.name"
            class="category-section"
          >
            <h2 class="category-heading">{{ category.name }}</h2>
            <ul class="item-list">
              <li
                v-for="item in category.items"
                :key="item.id"
                class="item-row"
                :class="{ checked: item.checked }"
              >
                <label class="item-label">
                  <input
                    type="checkbox"
                    class="item-checkbox"
                    :checked="item.checked"
                    @change="toggle(item.id, !item.checked)"
                  />
                  <span class="item-name">{{ item.name }}</span>
                  <span v-if="item.amount > 0" class="item-amount">
                    {{ formatAmount(item.amount) }} {{ item.unit }}
                  </span>
                </label>
              </li>
            </ul>
          </div>
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
  background: linear-gradient(
    180deg,
    var(--bg-card) 0%,
    var(--bg-primary) 100%
  );
  border-bottom: 1px solid var(--border-color);
}

.header-content {
  max-width: 720px;
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
  max-width: 720px;
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

/* Category sections */
.category-section {
  margin-bottom: 1.5rem;
}

.category-heading {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.2rem;
  color: var(--text-primary);
  margin: 0 0 0.75rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--border-color);
}

.item-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.item-row {
  transition: opacity 0.2s ease;
}

.item-row.checked {
  opacity: 0.65;
}

.item-label {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 0;
  cursor: pointer;
  border-bottom: 1px solid var(--border-color-light);
  min-height: 48px;
  -webkit-tap-highlight-color: transparent;
}

.item-label:active {
  background: var(--bg-hover);
  border-radius: 8px;
  margin: 0 -0.5rem;
  padding-left: 0.5rem;
  padding-right: 0.5rem;
}

.item-checkbox {
  width: 24px;
  height: 24px;
  min-width: 24px;
  accent-color: var(--accent);
  cursor: pointer;
  flex-shrink: 0;
  /* Expand touch target beyond visual size */
  padding: 10px;
  margin: -10px;
  box-sizing: content-box;
}

.item-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1rem;
  color: var(--text-primary);
  flex: 1;
  transition: all 0.2s ease;
}

.item-row.checked .item-name {
  text-decoration: line-through;
  color: var(--text-secondary);
}

.item-amount {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--text-secondary);
  white-space: nowrap;
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
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
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
}
</style>
