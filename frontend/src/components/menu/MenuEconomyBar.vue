<script setup lang="ts">
import { computed } from 'vue'
import { Motion } from 'motion-v'
import { Leaf, ShoppingBasket, Sparkles } from 'lucide-vue-next'
import type { MenuEconomy } from '@/api/menu.api'

interface Props {
  economy: MenuEconomy | null
}

const props = defineProps<Props>()

/** Only render when we have an economy with shared ingredients to show. */
const hasEconomy = computed(
  () => props.economy !== null && props.economy.sharedIngredients.length > 0,
)

const sharedIngredients = computed(() => props.economy?.sharedIngredients ?? [])

/** Overlap saved = how many ingredient references are reused across recipes. */
const overlapSaved = computed(() => {
  if (!props.economy) return 0
  return Math.max(0, props.economy.totalIngredientRefs - props.economy.distinctItemsToBuy)
})

const distinctItems = computed(() => props.economy?.distinctItemsToBuy ?? 0)
</script>

<template>
  <Motion
    v-if="hasEconomy"
    tag="section"
    class="economy-bar"
    aria-label="Delade ingredienser denna vecka"
    :initial="{ opacity: 0, y: 16 }"
    :animate="{ opacity: 1, y: 0 }"
    :transition="{ duration: 0.4, ease: 'easeOut' }"
  >
    <div class="economy-header">
      <Leaf :size="20" :stroke-width="2" class="header-icon" />
      <h2 class="economy-title">Delade ingredienser denna vecka</h2>
    </div>

    <ul class="chip-row">
      <Motion
        v-for="(ingredient, index) in sharedIngredients"
        :key="ingredient.canonicalName"
        tag="li"
        class="ingredient-chip"
        :initial="{ opacity: 0, scale: 0.85, y: 8 }"
        :animate="{ opacity: 1, scale: 1, y: 0 }"
        :transition="{ duration: 0.3, delay: 0.1 + index * 0.05, ease: 'easeOut' }"
      >
        <span class="chip-name">{{ ingredient.name }}</span>
        <span class="chip-count">×{{ ingredient.recipeCount }}</span>
      </Motion>
    </ul>

    <Motion
      tag="div"
      class="savings-row"
      :initial="{ opacity: 0 }"
      :animate="{ opacity: 1 }"
      :transition="{ duration: 0.3, delay: 0.1 + sharedIngredients.length * 0.05 }"
    >
      <div class="savings-indicator">
        <ShoppingBasket :size="18" :stroke-width="2" class="savings-icon" />
        <span class="savings-text">
          <strong>{{ overlapSaved }}</strong> färre varor att köpa
        </span>
      </div>
      <div class="savings-indicator subtle">
        <Sparkles :size="16" :stroke-width="2" class="savings-icon" />
        <span class="savings-text">
          <strong>{{ distinctItems }}</strong> unika varor
        </span>
      </div>
    </Motion>
  </Motion>
</template>

<style scoped>
.economy-bar {
  margin-top: var(--space-xl);
  padding: var(--space-lg);
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-sm);
}

/* Header */
.economy-header {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  margin-bottom: var(--space-md);
}

.header-icon {
  color: var(--accent);
  flex-shrink: 0;
}

.economy-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.25rem;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.2;
}

/* Chip row */
.chip-row {
  list-style: none;
  margin: 0 0 var(--space-lg) 0;
  padding: 0;
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-sm);
}

.ingredient-chip {
  display: inline-flex;
  align-items: center;
  gap: var(--space-xs);
  padding: 0.375rem 0.875rem;
  border-radius: var(--radius-full);
  background: var(--accent-bg);
  border: 1px solid var(--border-color-hover);
  font-family: 'Nunito', sans-serif;
}

.chip-name {
  font-weight: 700;
  font-size: 0.875rem;
  color: var(--accent-text);
  line-height: 1;
}

.chip-count {
  font-weight: 800;
  font-size: 0.8rem;
  color: var(--accent-text);
  opacity: 0.85;
  line-height: 1;
}

/* Savings */
.savings-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-md);
}

.savings-indicator {
  display: inline-flex;
  align-items: center;
  gap: var(--space-sm);
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--text-secondary);
}

.savings-indicator strong {
  font-weight: 800;
  color: var(--text-primary);
}

.savings-indicator.subtle {
  color: var(--text-muted);
}

.savings-icon {
  color: var(--accent);
  flex-shrink: 0;
}

.savings-text {
  line-height: 1.2;
}

/* Responsive */
@media (max-width: 768px) {
  .economy-bar {
    padding: var(--space-md);
  }

  .economy-title {
    font-size: 1.1rem;
  }
}
</style>
