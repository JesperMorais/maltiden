<script setup lang="ts">
import { onMounted, computed, ref, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useShoppingList } from '@/composables/useShoppingList'
import { useDashboardStore } from '@/stores/dashboard'
import { getCurrentMenu } from '@/api/menu.api'
import { useSkeleton } from '@/composables/useSkeleton'
import ErrorState from '@/components/common/ErrorState.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import BackLink from '@/components/common/BackLink.vue'
import {
  ClipboardList,
  ShoppingCart,
  Search,
  Plus,
  X,
  ChevronRight,
  Check,
  Trash2,
  Refrigerator,
} from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()
const dashboardStore = useDashboardStore()
const {
  isLoading,
  error,
  totalItems,
  checkedItemCount,
  remainingItems,
  progress,
  fetchList,
  toggle,

  // Search
  searchQuery,
  isSearchOpen,
  toggleSearch,

  // Collapse
  collapsedCategories,
  toggleCategory,

  // Separated lists
  uncheckedCategories,
  checkedItems,
  isCheckedCollapsed,

  // Custom items
  addCustomItem,
  removeCustomItem,
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

// Add item UI
const isAddOpen = ref(false)
const newItemName = ref('')
const newItemInput = ref<HTMLInputElement | null>(null)

function toggleAddItem() {
  isAddOpen.value = !isAddOpen.value
  if (isAddOpen.value) {
    nextTick(() => newItemInput.value?.focus())
  } else {
    newItemName.value = ''
  }
}

async function submitNewItem() {
  const name = newItemName.value.trim()
  if (!name) return
  newItemName.value = ''
  await addCustomItem(name)
}

// Prep-mode note: true when the current menu batch-cooks (any leftovers day).
// Leftovers days contribute nothing to the shopping list (folded into the cook
// day's quantities by the backend), so we surface a short explanatory note.
const usedPrepMode = ref(false)

// Search input ref
const searchInput = ref<HTMLInputElement | null>(null)

function handleToggleSearch() {
  toggleSearch()
  if (isSearchOpen.value) {
    nextTick(() => searchInput.value?.focus())
  }
}

onMounted(async () => {
  if (!dashboardStore.dashboardData) {
    await dashboardStore.fetchDashboard()
  }
  if (menuId.value) {
    fetchList(menuId.value)
  }
  try {
    const menu = await getCurrentMenu()
    usedPrepMode.value = menu?.days.some((d) => d.leftover) ?? false
  } catch {
    usedPrepMode.value = false
  }
})
</script>

<template>
  <div class="shopping-list-view">
    <!-- Sticky header -->
    <header class="sticky-header">
      <div class="header-inner">
        <BackLink :to="{ name: 'dashboard' }" label="Dashboard" />
        <div class="header-row">
          <div class="header-info">
            <h1 class="title">Inköpslista</h1>
            <span v-if="totalItems > 0" class="progress-text">
              {{ checkedItemCount }} av {{ totalItems }}
              <span class="progress-separator">&middot;</span>
              {{ remainingItems }} kvar
            </span>
          </div>
          <div class="header-actions">
            <button
              class="icon-btn"
              :class="{ active: isSearchOpen }"
              @click="handleToggleSearch"
              aria-label="Sök"
            >
              <Search :size="20" :stroke-width="2" />
            </button>
            <button
              class="icon-btn"
              :class="{ active: isAddOpen }"
              @click="toggleAddItem"
              aria-label="Lägg till vara"
            >
              <Plus :size="20" :stroke-width="2" />
            </button>
          </div>
        </div>

        <!-- Progress bar -->
        <div v-if="totalItems > 0" class="progress-bar-track">
          <div class="progress-bar-fill" :style="{ width: `${progress * 100}%` }" />
        </div>

        <!-- Search bar -->
        <Transition name="search-slide">
          <div v-if="isSearchOpen" class="search-bar">
            <Search :size="16" :stroke-width="2" class="search-icon" />
            <input
              ref="searchInput"
              v-model="searchQuery"
              type="text"
              placeholder="Sök vara..."
              class="search-input"
            />
            <button
              v-if="searchQuery"
              class="search-clear"
              @click="searchQuery = ''"
              aria-label="Rensa sökning"
            >
              <X :size="16" :stroke-width="2" />
            </button>
          </div>
        </Transition>
      </div>
    </header>

    <!-- Add item section -->
    <Transition name="add-slide">
      <div v-if="isAddOpen" class="add-item-section">
        <div class="add-item-inner">
          <input
            ref="newItemInput"
            v-model="newItemName"
            type="text"
            placeholder="T.ex. Hushållspapper"
            class="add-item-input"
            @keydown.enter="submitNewItem"
          />
          <button
            class="add-item-btn"
            :disabled="!newItemName.trim()"
            @click="submitNewItem"
          >
            Lägg till
          </button>
        </div>
      </div>
    </Transition>

    <!-- Main content -->
    <main class="list-content">
      <div class="list-container">
        <!-- Skeleton loading -->
        <div v-if="showSkeleton" class="skeleton-wrapper">
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
          <template #icon>
            <ClipboardList :size="48" color="var(--text-muted)" />
          </template>
        </EmptyState>

        <!-- Error state -->
        <ErrorState v-else-if="error" :description="error" @retry="fetchList(menuId!)" />

        <!-- Empty state -->
        <EmptyState
          v-else-if="totalItems === 0 && !isLoading"
          title="Ingen inköpslista"
          description="Generera en meny först så skapas din inköpslista automatiskt."
          action-label="Generera meny"
          @action="router.push({ name: 'generate-menu' })"
        >
          <template #icon>
            <ShoppingCart :size="48" color="var(--text-muted)" />
          </template>
        </EmptyState>

        <!-- Shopping list -->
        <template v-else>
          <!-- Prep-mode note -->
          <p v-if="usedPrepMode" class="prep-note">
            <Refrigerator :size="16" :stroke-width="2" class="prep-note-icon" />
            <span>Rester-dagar ingår i lagningsdagens mängder.</span>
          </p>

          <!-- Category sections -->
          <section
            v-for="category in uncheckedCategories"
            :key="category.name"
            class="category-section"
          >
            <button class="category-header" @click="toggleCategory(category.name)">
              <ChevronRight
                :size="18"
                :stroke-width="2.5"
                class="chevron-icon"
                :class="{ expanded: !collapsedCategories.has(category.name) }"
              />
              <span class="category-name">{{ category.name }}</span>
              <span class="category-count">{{ category.items.length }}</span>
            </button>

            <Transition name="collapse">
              <ul
                v-show="!collapsedCategories.has(category.name)"
                class="item-list"
              >
                <li
                  v-for="item in category.items"
                  :key="item.id"
                  class="item-row"
                  role="button"
                  tabindex="0"
                  @click="toggle(item.id, true)"
                  @keydown.enter="toggle(item.id, true)"
                  @keydown.space.prevent="toggle(item.id, true)"
                >
                  <span class="custom-checkbox">
                    <Check :size="14" :stroke-width="3" class="check-icon" />
                  </span>
                  <span class="item-name">{{ item.name }}</span>
                  <span class="item-amount">{{ item.amount }} {{ item.unit }}</span>
                  <button
                    v-if="item.isCustom"
                    class="delete-btn"
                    @click.stop="removeCustomItem(item.id)"
                    aria-label="Ta bort"
                  >
                    <Trash2 :size="16" :stroke-width="2" />
                  </button>
                </li>
              </ul>
            </Transition>
          </section>

          <!-- Checked "Klart" section -->
          <section v-if="checkedItems.length > 0" class="checked-section">
            <button
              class="category-header checked-header"
              @click="isCheckedCollapsed = !isCheckedCollapsed"
            >
              <ChevronRight
                :size="18"
                :stroke-width="2.5"
                class="chevron-icon"
                :class="{ expanded: !isCheckedCollapsed }"
              />
              <span class="category-name">Klart</span>
              <span class="category-count">{{ checkedItems.length }}</span>
            </button>

            <Transition name="collapse">
              <ul v-show="!isCheckedCollapsed" class="item-list checked-list">
                <li
                  v-for="item in checkedItems"
                  :key="item.id"
                  class="item-row checked"
                  role="button"
                  tabindex="0"
                  @click="toggle(item.id, false)"
                  @keydown.enter="toggle(item.id, false)"
                  @keydown.space.prevent="toggle(item.id, false)"
                >
                  <span class="custom-checkbox is-checked">
                    <Check :size="14" :stroke-width="3" class="check-icon" />
                  </span>
                  <span class="item-name">{{ item.name }}</span>
                  <span class="item-amount">{{ item.amount }} {{ item.unit }}</span>
                </li>
              </ul>
            </Transition>
          </section>
        </template>
      </div>
    </main>

    <!-- Bottom spacer for MobileBottomNav -->
    <div class="bottom-spacer" />
  </div>
</template>

<style scoped>
.shopping-list-view {
  min-height: 100vh;
  min-height: 100dvh;
  background: var(--bg-primary);
  display: flex;
  flex-direction: column;
}

/* Sticky header */
.sticky-header {
  position: sticky;
  top: 0;
  z-index: 10;
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border-color);
}

.header-inner {
  max-width: 640px;
  margin: 0 auto;
  padding: 0.75rem 1rem 0;
}

.header-inner :deep(.back-link) {
  margin-bottom: 0.25rem;
  margin-left: -0.5rem;
}

.header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.header-info {
  display: flex;
  align-items: baseline;
  gap: 0.75rem;
  min-width: 0;
}

.title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.5rem;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.3;
  white-space: nowrap;
}

.progress-text {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-secondary);
  white-space: nowrap;
}

.progress-separator {
  color: var(--text-muted);
  margin: 0 0.1rem;
}

.header-actions {
  display: flex;
  gap: 0.25rem;
  flex-shrink: 0;
}

.icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--duration-fast) ease;
  -webkit-tap-highlight-color: transparent;
}

.icon-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.icon-btn.active {
  color: var(--accent);
  background: var(--bg-hover);
}

/* Progress bar */
.progress-bar-track {
  height: 4px;
  background: var(--bg-secondary);
  border-radius: var(--radius-full);
  overflow: hidden;
  margin: 0.5rem 0 0;
}

.progress-bar-fill {
  height: 100%;
  background: var(--success);
  border-radius: var(--radius-full);
  transition: width 0.4s ease;
}

/* Search bar */
.search-bar {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.5rem;
  padding: 0.5rem 0.75rem;
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
}

.search-icon {
  color: var(--text-muted);
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  border: none;
  background: transparent;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--text-primary);
  outline: none;
}

.search-input::placeholder {
  color: var(--text-muted);
}

.search-clear {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  background: var(--bg-hover);
  border-radius: var(--radius-full);
  color: var(--text-muted);
  cursor: pointer;
  flex-shrink: 0;
}

.search-slide-enter-active,
.search-slide-leave-active {
  transition: all var(--duration-fast) ease;
  overflow: hidden;
}

.search-slide-enter-from,
.search-slide-leave-to {
  opacity: 0;
  max-height: 0;
  margin-top: 0;
  padding-top: 0;
  padding-bottom: 0;
}

.search-slide-enter-to,
.search-slide-leave-from {
  opacity: 1;
  max-height: 60px;
}

/* Add item section */
.add-item-section {
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-secondary);
}

.add-item-inner {
  max-width: 640px;
  margin: 0 auto;
  padding: 0.75rem 1rem;
  display: flex;
  gap: 0.5rem;
}

.add-item-input {
  flex: 1;
  border: 1.5px solid var(--border-color);
  background: var(--bg-card);
  border-radius: var(--radius-md);
  padding: 0.6rem 0.75rem;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--text-primary);
  outline: none;
  transition: border-color var(--duration-fast) ease;
}

.add-item-input:focus {
  border-color: var(--accent);
}

.add-item-input::placeholder {
  color: var(--text-muted);
}

.add-item-btn {
  padding: 0.6rem 1rem;
  border: none;
  background: var(--accent);
  color: var(--text-on-accent);
  border-radius: var(--radius-md);
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  cursor: pointer;
  white-space: nowrap;
  transition: all var(--duration-fast) ease;
}

.add-item-btn:hover:not(:disabled) {
  background: var(--accent-dark);
}

.add-item-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.add-slide-enter-active,
.add-slide-leave-active {
  transition: all var(--duration-fast) ease;
  overflow: hidden;
}

.add-slide-enter-from,
.add-slide-leave-to {
  opacity: 0;
  max-height: 0;
}

.add-slide-enter-to,
.add-slide-leave-from {
  opacity: 1;
  max-height: 80px;
}

/* List content */
.list-content {
  flex: 1;
  padding: 1rem;
}

.list-container {
  max-width: 640px;
  margin: 0 auto;
}

/* Prep-mode note */
.prep-note {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0 0 1rem;
  padding: 0.65rem 0.85rem;
  background: var(--accent-bg);
  border-radius: var(--radius-md);
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-secondary);
  line-height: 1.4;
}

.prep-note-icon {
  color: var(--accent);
  flex-shrink: 0;
}

/* Category sections */
.category-section {
  margin-bottom: 0.5rem;
}

.category-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.75rem 0.25rem;
  border: none;
  background: transparent;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
}

.chevron-icon {
  color: var(--text-muted);
  transition: transform var(--duration-fast) ease;
  flex-shrink: 0;
}

.chevron-icon.expanded {
  transform: rotate(90deg);
}

.category-name {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.05rem;
  color: var(--text-primary);
  flex: 1;
  text-align: left;
}

.category-count {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.75rem;
  color: var(--text-secondary);
  background: var(--bg-secondary);
  border-radius: var(--radius-full);
  min-width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 0.4rem;
}

/* Collapse transition */
.collapse-enter-active,
.collapse-leave-active {
  transition:
    max-height var(--duration-normal) ease,
    opacity var(--duration-fast) ease;
  overflow: hidden;
}

.collapse-enter-from,
.collapse-leave-to {
  max-height: 0;
  opacity: 0;
}

.collapse-enter-to,
.collapse-leave-from {
  max-height: 2000px;
  opacity: 1;
}

/* Item list */
.item-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.item-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.65rem 0.5rem;
  min-height: 52px;
  cursor: pointer;
  border-radius: var(--radius-md);
  transition: all 150ms ease;
  -webkit-tap-highlight-color: transparent;
}

.item-row:hover {
  background: var(--bg-hover);
}

.item-row:active {
  transform: scale(0.98);
}

/* Custom checkbox */
.custom-checkbox {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  min-width: 22px;
  border: 2px solid var(--border-color-hover);
  border-radius: var(--radius-full);
  background: transparent;
  transition: all 200ms ease;
  flex-shrink: 0;
}

.custom-checkbox .check-icon {
  opacity: 0;
  color: var(--text-on-accent);
  transform: scale(0.5);
  transition: all 200ms ease;
}

.custom-checkbox.is-checked {
  background: var(--success);
  border-color: var(--success);
  animation: checkbox-pop 300ms ease;
}

.custom-checkbox.is-checked .check-icon {
  opacity: 1;
  transform: scale(1);
}

@keyframes checkbox-pop {
  0% { transform: scale(1); }
  50% { transform: scale(0.9); }
  100% { transform: scale(1); }
}

.item-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--text-primary);
  flex: 1;
  min-width: 0;
}

.item-amount {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-muted);
  white-space: nowrap;
  flex-shrink: 0;
}

.delete-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  border-radius: var(--radius-md);
  cursor: pointer;
  flex-shrink: 0;
  transition: all var(--duration-fast) ease;
}

.delete-btn:hover {
  color: var(--accent);
  background: var(--bg-hover);
}

/* Checked section */
.checked-section {
  margin-top: 1rem;
  border-top: 1px solid var(--border-color);
  padding-top: 0.25rem;
}

.checked-header .category-name {
  color: var(--text-secondary);
}

.checked-list .item-row {
  opacity: 0.6;
}

.checked-list .item-name {
  text-decoration: line-through;
  color: var(--text-secondary);
}

/* Skeleton */
.skeleton-wrapper {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.skeleton-category {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.skeleton-heading {
  height: 20px;
  width: 120px;
  background: var(--bg-secondary);
  border-radius: var(--radius-sm);
  margin-bottom: 0.25rem;
  animation: pulse 1.5s ease-in-out infinite;
}

.skeleton-item {
  height: 48px;
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
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

/* Bottom spacer */
.bottom-spacer {
  height: calc(var(--bottom-nav-height) + env(safe-area-inset-bottom, 0px) + 1rem);
}

/* Responsive — desktop */
@media (min-width: 769px) {
  .header-inner {
    padding: 1rem 1.5rem 0;
  }

  .list-content {
    padding: 1.5rem;
  }

  .title {
    font-size: 1.75rem;
  }
}

/* Reduced motion */
@media (prefers-reduced-motion: reduce) {
  .progress-bar-fill,
  .chevron-icon,
  .custom-checkbox,
  .custom-checkbox .check-icon,
  .item-row {
    transition: none;
  }

  .collapse-enter-active,
  .collapse-leave-active,
  .search-slide-enter-active,
  .search-slide-leave-active,
  .add-slide-enter-active,
  .add-slide-leave-active {
    transition: none;
  }

  @keyframes checkbox-pop {
    0%,
    50%,
    100% {
      transform: none;
    }
  }
}
</style>
