<script setup lang="ts">
import { computed } from 'vue'
import type { ShoppingListSummary } from '@/api/types/dashboard.types'
import CountUp from '@/components/vue-bits/CountUp.vue'
import SpotlightCard from '@/components/vue-bits/SpotlightCard.vue'
import { ShoppingCart, ArrowRight, Package } from 'lucide-vue-next'

interface Props {
  shoppingList: ShoppingListSummary | null
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'view-list': []
}>()

const remainingItems = computed(() =>
  props.shoppingList ? props.shoppingList.totalItems - props.shoppingList.checkedItems : 0,
)

const progressPercent = computed(() => {
  if (!props.shoppingList || props.shoppingList.totalItems === 0) return 0
  return Math.round((props.shoppingList.checkedItems / props.shoppingList.totalItems) * 100)
})

const categoryColors = [
  'var(--success)',
  'var(--accent)',
  'var(--warning)',
  'var(--category-purple)',
  'var(--category-pink)',
  'var(--category-blue)',
]
</script>

<template>
  <SpotlightCard spotlight-color="rgba(104, 211, 145, 0.15)" class-name="shopping-spotlight">
    <section class="shopping-widget" @click="emit('view-list')">
      <!-- Header -->
      <div class="widget-header">
        <div class="header-icon">
          <ShoppingCart :size="20" :stroke-width="2" />
        </div>
        <h3 class="widget-title">Inköpslista</h3>
      </div>

      <!-- Stats row -->
      <div v-if="shoppingList" class="stats-row">
        <div class="stat-block">
          <span class="stat-number">
            <CountUp :to="remainingItems" :duration="1.5" />
          </span>
          <span class="stat-label">kvar</span>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-block">
          <span class="stat-number">
            <CountUp :to="shoppingList.checkedItems" :duration="1.5" :delay="0.2" />
          </span>
          <span class="stat-label">klart</span>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-block">
          <span class="stat-number">
            <CountUp :to="shoppingList.totalItems" :duration="1.5" :delay="0.3" />
          </span>
          <span class="stat-label">totalt</span>
        </div>
      </div>

      <!-- Progress -->
      <div v-if="shoppingList" class="progress-section">
        <div class="progress-track">
          <div
            class="progress-fill"
            :style="{ width: `${progressPercent}%` }"
          ></div>
        </div>
        <span class="progress-label">{{ progressPercent }}%</span>
      </div>

      <!-- Categories -->
      <div v-if="shoppingList && shoppingList.categories.length" class="categories-section">
        <span class="categories-heading">Kategorier</span>
        <ul class="category-list">
          <li
            v-for="(cat, i) in shoppingList.categories"
            :key="cat.name"
            class="category-item"
          >
            <span
              class="category-badge"
              :style="{ background: categoryColors[i % categoryColors.length] }"
            >{{ cat.count }}</span>
            <span class="category-name">{{ cat.name }}</span>
          </li>
        </ul>
      </div>

      <!-- Empty state -->
      <div v-else-if="!shoppingList" class="empty-state">
        <Package :size="28" :stroke-width="1.5" class="empty-icon" />
        <p class="empty-text">Generera en meny för att skapa en inköpslista</p>
      </div>

      <!-- CTA -->
      <button class="view-list-cta">
        <span>Visa hela listan</span>
        <ArrowRight :size="16" :stroke-width="2.5" class="cta-arrow" />
      </button>
    </section>
  </SpotlightCard>
</template>

<style scoped>
.shopping-widget {
  background: var(--bg-primary);
  border-radius: 20px;
  padding: 1.25rem;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
}

.shopping-widget:hover {
  box-shadow: var(--shadow-md);
  border-color: var(--border-color-hover);
}

/* Header */
.widget-header {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  margin-bottom: 1rem;
}

.header-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  background: var(--bg-hover);
  border-radius: 9px;
  color: var(--success);
}

.widget-title {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 0.95rem;
  color: var(--text-primary);
  margin: 0;
}

/* Stats row */
.stats-row {
  display: flex;
  align-items: center;
  justify-content: space-around;
  background: var(--bg-card);
  border-radius: 12px;
  padding: 0.65rem 0.5rem;
  margin-bottom: 0.85rem;
}

.stat-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.1rem;
}

.stat-number {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.25rem;
  color: var(--text-primary);
  line-height: 1.2;
}

.stat-label {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.65rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.stat-divider {
  width: 1px;
  height: 28px;
  background: var(--border-color);
}

/* Progress */
.progress-section {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  margin-bottom: 1rem;
}

.progress-track {
  flex: 1;
  height: 6px;
  background: var(--border-color);
  border-radius: 100px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--success-dark) 0%, var(--success) 100%);
  border-radius: 100px;
  transition: width 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.progress-label {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.7rem;
  color: var(--text-muted);
  min-width: 2.5em;
  text-align: right;
}

/* Categories */
.categories-section {
  margin-bottom: 0.75rem;
}

.categories-heading {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.7rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  display: block;
  margin-bottom: 0.5rem;
}

.category-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.category-item {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.45rem 0.5rem;
  border-radius: 10px;
  transition: background 0.2s ease;
}

.category-item:hover {
  background: var(--bg-hover);
}

.category-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border-radius: 8px;
  flex-shrink: 0;
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 0.72rem;
  color: var(--text-on-accent);
}

.category-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-primary);
  flex: 1;
}

/* Empty state */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 1.25rem 0;
}

.empty-icon {
  color: var(--text-muted);
  opacity: 0.4;
}

.empty-text {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.8rem;
  color: var(--text-muted);
  text-align: center;
  margin: 0;
  max-width: 180px;
  line-height: 1.4;
}

/* CTA Button */
.view-list-cta {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.65rem 1rem;
  background: var(--bg-hover);
  border: 1.5px solid var(--border-color);
  border-radius: 12px;
  cursor: pointer;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.82rem;
  color: var(--text-secondary);
  transition: all 0.25s ease;
}

.view-list-cta:hover {
  color: var(--success);
  border-color: var(--success);
  background: var(--bg-card);
}

.cta-arrow {
  transition: transform 0.25s ease;
}

.view-list-cta:hover .cta-arrow {
  transform: translateX(3px);
}

/* Mobile: compact row */
@media (max-width: 768px) {
  .shopping-widget {
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: 0.75rem;
    padding: 1rem;
  }

  .widget-header {
    flex: 1;
    margin-bottom: 0;
    min-width: 0;
  }

  .stats-row,
  .progress-section,
  .categories-section,
  .empty-state {
    display: none;
  }

  .view-list-cta {
    width: auto;
    padding: 0.5rem 0.75rem;
    flex-shrink: 0;
  }
}
</style>
