<script setup lang="ts">
import { computed } from 'vue'
import type { ShoppingListSummary } from '@/api/types/dashboard.types'
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

// SVG circle constants
const RING_SIZE = 48
const RING_STROKE = 4
const RING_RADIUS = (RING_SIZE - RING_STROKE) / 2
const RING_CIRCUMFERENCE = 2 * Math.PI * RING_RADIUS

const ringOffset = computed(() => {
  const pct = progressPercent.value / 100
  return RING_CIRCUMFERENCE * (1 - pct)
})

const topCategories = computed(() => {
  if (!props.shoppingList) return []
  return props.shoppingList.categories.slice(0, 3)
})
</script>

<template>
  <SpotlightCard spotlight-color="var(--success-bg)" class-name="shopping-spotlight">
    <section class="shopping-widget" @click="emit('view-list')">
      <!-- Header -->
      <div class="widget-header">
        <div class="header-icon">
          <ShoppingCart :size="20" :stroke-width="2" />
        </div>
        <h3 class="widget-title">Inköpslista</h3>
      </div>

      <!-- Content with ring + info -->
      <div v-if="shoppingList" class="widget-body">
        <!-- Circular progress ring -->
        <div class="ring-container">
          <svg
            :width="RING_SIZE"
            :height="RING_SIZE"
            class="progress-ring"
            role="progressbar"
            :aria-valuenow="progressPercent"
            aria-valuemin="0"
            aria-valuemax="100"
            :aria-label="`Inköpslista ${progressPercent}% klart`"
          >
            <circle
              class="ring-bg"
              :cx="RING_SIZE / 2"
              :cy="RING_SIZE / 2"
              :r="RING_RADIUS"
              fill="none"
              :stroke-width="RING_STROKE"
            />
            <circle
              class="ring-fill"
              :cx="RING_SIZE / 2"
              :cy="RING_SIZE / 2"
              :r="RING_RADIUS"
              fill="none"
              :stroke-width="RING_STROKE"
              stroke-linecap="round"
              :stroke-dasharray="RING_CIRCUMFERENCE"
              :stroke-dashoffset="ringOffset"
            />
          </svg>
          <span class="ring-label">{{ progressPercent }}%</span>
        </div>

        <!-- Stats + preview -->
        <div class="info-column">
          <div class="stat-row">
            <span class="stat-remaining">{{ remainingItems }} kvar</span>
            <span class="stat-total">av {{ shoppingList.totalItems }}</span>
          </div>

          <!-- Top categories preview -->
          <ul v-if="topCategories.length" class="category-preview">
            <li v-for="cat in topCategories" :key="cat.name" class="preview-item">
              <span class="preview-dot" />
              <span class="preview-name">{{ cat.name }}</span>
              <span class="preview-count">{{ cat.count }}</span>
            </li>
          </ul>
        </div>
      </div>

      <!-- Empty state -->
      <div v-else class="empty-state">
        <Package :size="28" :stroke-width="1.5" class="empty-icon" />
        <p class="empty-text">Generera en meny för att skapa en inköpslista</p>
      </div>

      <!-- CTA -->
      <button type="button" class="view-list-cta" @click.stop="emit('view-list')">
        <span>Visa hela listan</span>
        <ArrowRight :size="16" :stroke-width="2.5" class="cta-arrow" />
      </button>
    </section>
  </SpotlightCard>
</template>

<style scoped>
.shopping-widget {
  background: var(--bg-primary);
  border-radius: var(--radius-xl);
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
  border-radius: var(--radius-md);
  color: var(--success);
}

.widget-title {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 0.95rem;
  color: var(--text-primary);
  margin: 0;
}

/* Body: ring + info */
.widget-body {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 1rem;
}

/* Circular progress ring */
.ring-container {
  position: relative;
  flex-shrink: 0;
}

.progress-ring {
  transform: rotate(-90deg);
}

.ring-bg {
  stroke: var(--border-color);
}

.ring-fill {
  stroke: var(--success);
  transition: stroke-dashoffset 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.ring-label {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 0.65rem;
  color: var(--text-secondary);
}

/* Info column */
.info-column {
  flex: 1;
  min-width: 0;
}

.stat-row {
  display: flex;
  align-items: baseline;
  gap: 0.4rem;
  margin-bottom: 0.4rem;
}

.stat-remaining {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.1rem;
  color: var(--text-primary);
  line-height: 1.2;
}

.stat-total {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.75rem;
  color: var(--text-muted);
}

/* Category preview */
.category-preview {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.preview-item {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.preview-dot {
  width: 6px;
  height: 6px;
  border-radius: var(--radius-full);
  background: var(--success);
  flex-shrink: 0;
}

.preview-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.78rem;
  color: var(--text-secondary);
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.preview-count {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.72rem;
  color: var(--text-muted);
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
  border-radius: var(--radius-lg);
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
    margin-bottom: 0;
    flex-shrink: 0;
  }

  .widget-body {
    flex: 1;
    margin-bottom: 0;
    align-items: center;
  }

  .info-column .category-preview {
    display: none;
  }

  .empty-state {
    display: none;
  }

  .view-list-cta {
    width: auto;
    padding: 0.5rem 0.75rem;
    flex-shrink: 0;
  }
}

/* Reduced motion */
@media (prefers-reduced-motion: reduce) {
  .ring-fill,
  .shopping-widget,
  .view-list-cta,
  .cta-arrow {
    transition: none;
  }
}
</style>
