<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useShoppingList } from '@/composables/useShoppingList'
import { useSkeleton } from '@/composables/useSkeleton'

const router = useRouter()
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

onMounted(() => {
  fetchList()
})
</script>

<template>
  <div class="shopping-list-view">
    <!-- Header -->
    <header class="header">
      <div class="header-content">
        <button class="back-link" @click="router.push({ name: 'dashboard' })">
          <span class="back-arrow">&larr;</span>
          <span>Dashboard</span>
        </button>
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

        <!-- Error state -->
        <div v-else-if="error" class="error-state">
          <span class="error-icon">😅</span>
          <h2>Något gick fel</h2>
          <p>{{ error }}</p>
          <button class="retry-btn" @click="fetchList()">Försök igen</button>
        </div>

        <!-- Empty state -->
        <div v-else-if="totalItems === 0" class="empty-state">
          <span class="empty-icon">🛒</span>
          <h2>Ingen inköpslista</h2>
          <p>Generera en meny först så skapas din inköpslista automatiskt.</p>
          <button class="action-btn" @click="router.push({ name: 'generate-menu' })">
            Generera meny
          </button>
        </div>

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
                  <span class="item-amount">{{ item.amount }} {{ item.unit }}</span>
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

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: none;
  border: none;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 0;
  margin-bottom: 1rem;
  transition: color 0.2s ease;
}

.back-link:hover {
  color: var(--accent);
}

.back-arrow {
  font-size: 1.1rem;
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
  opacity: 0.5;
}

.item-label {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.65rem 0;
  cursor: pointer;
  border-bottom: 1px solid var(--border-color-light, rgba(0, 0, 0, 0.05));
}

.item-checkbox {
  width: 20px;
  height: 20px;
  accent-color: var(--accent);
  cursor: pointer;
  flex-shrink: 0;
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

/* Error state */
.error-state {
  text-align: center;
  padding: 4rem 2rem;
}

.error-icon {
  font-size: 3rem;
  display: block;
  margin-bottom: 1rem;
}

.error-state h2 {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.5rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
}

.error-state p {
  font-family: 'Nunito', sans-serif;
  color: var(--text-secondary);
  margin: 0 0 1.5rem;
}

.retry-btn {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1rem;
  color: white;
  background: var(--accent);
  border: none;
  border-radius: 100px;
  padding: 0.85em 2em;
  cursor: pointer;
  transition: all 0.3s ease;
}

.retry-btn:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-accent);
}

/* Empty state */
.empty-state {
  text-align: center;
  padding: 4rem 2rem;
}

.empty-icon {
  font-size: 3rem;
  display: block;
  margin-bottom: 1rem;
}

.empty-state h2 {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.5rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
}

.empty-state p {
  font-family: 'Nunito', sans-serif;
  color: var(--text-secondary);
  margin: 0 0 1.5rem;
}

.action-btn {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1rem;
  color: white;
  background: var(--accent);
  border: none;
  border-radius: 100px;
  padding: 0.85em 2em;
  cursor: pointer;
  transition: all 0.3s ease;
}

.action-btn:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-accent);
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
