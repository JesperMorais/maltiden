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
  '#8b7cf6',
  '#f472b6',
  '#38bdf8',
  '#a78bfa',
  '#fb923c',
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
        <div class="header-text">
          <h3 class="widget-title">Inköpslista</h3>
          <p v-if="shoppingList" class="widget-subtitle">
            <CountUp :to="remainingItems" :duration="1.5" /> varor kvar av
            <CountUp :to="shoppingList.totalItems" :duration="1.5" :delay="0.2" />
          </p>
          <p v-else class="widget-subtitle muted">Ingen lista ännu</p>
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
        <div class="categories-divider"></div>
        <ul class="category-list">
          <li
            v-for="(cat, i) in shoppingList.categories"
            :key="cat.name"
            class="category-item"
          >
            <span
              class="category-dot"
              :style="{ background: categoryColors[i % categoryColors.length] }"
            ></span>
            <span class="category-name">{{ cat.name }}</span>
            <span class="category-count">{{ cat.count }}</span>
          </li>
        </ul>
      </div>

      <!-- Empty categories placeholder -->
      <div v-else-if="!shoppingList" class="empty-categories">
        <Package :size="32" :stroke-width="1.5" class="empty-icon" />
        <p class="empty-text">Generera en meny för att skapa en inköpslista</p>
      </div>

      <!-- Spacer pushes CTA to bottom -->
      <div class="spacer"></div>

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
  display: flex;
  flex-direction: column;
  min-height: 280px;
}

.shopping-widget:hover {
  box-shadow: var(--shadow-md);
  border-color: var(--border-color-hover);
}

/* Header */
.widget-header {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.header-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  background: var(--bg-hover);
  border-radius: 10px;
  color: var(--success);
}

.header-text {
  flex: 1;
  min-width: 0;
}

.widget-title {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 0.95rem;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.3;
}

.widget-subtitle {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.8rem;
  color: var(--text-secondary);
  margin: 0.15rem 0 0;
}

.widget-subtitle.muted {
  color: var(--text-muted);
}

/* Progress */
.progress-section {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  margin-bottom: 0.75rem;
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
  background: linear-gradient(90deg, var(--success) 0%, #68d391 100%);
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
  flex: 1;
}

.categories-divider {
  height: 1px;
  background: var(--border-color);
  margin-bottom: 0.75rem;
}

.category-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.category-item {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.35rem 0.5rem;
  border-radius: 8px;
  transition: background 0.2s ease;
}

.category-item:hover {
  background: var(--bg-hover);
}

.category-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.category-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.82rem;
  color: var(--text-primary);
  flex: 1;
}

.category-count {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 0.85rem;
  color: var(--text-secondary);
}

/* Empty state */
.empty-categories {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 1.5rem 0;
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

/* Spacer */
.spacer {
  flex: 1;
  min-height: 0.5rem;
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
  margin-top: 0.75rem;
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

/* Mobile: revert to compact row */
@media (max-width: 768px) {
  .shopping-widget {
    min-height: 0;
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

  .progress-section,
  .categories-section,
  .empty-categories,
  .spacer {
    display: none;
  }

  .view-list-cta {
    margin-top: 0;
    width: auto;
    padding: 0.5rem 0.75rem;
    flex-shrink: 0;
  }
}
</style>
