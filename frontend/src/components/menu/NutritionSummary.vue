<script setup lang="ts">
import { computed } from 'vue'
import { Motion } from 'motion-v'
import { Beef, Flame, Target } from 'lucide-vue-next'
import type { WeeklyNutrition } from '@/stores/menuGenerator'

interface Props {
  /** Weekly per-serving roll-up, or null when no cooked day has macros. */
  nutrition: WeeklyNutrition | null
  /** Per-day protein target (grams). 0 = no target set. */
  proteinTarget?: number
}

const props = withDefaults(defineProps<Props>(), {
  proteinTarget: 0,
})

/** Only render when at least one cooked day carries nutrition data. */
const hasNutrition = computed(() => props.nutrition !== null)

const avgProtein = computed(() => Math.round(props.nutrition?.avgProtein ?? 0))
const avgKcal = computed(() => Math.round(props.nutrition?.avgKcal ?? 0))

const hasTarget = computed(() => props.proteinTarget > 0)

/** True once the average per-day protein meets or beats the target. */
const targetReached = computed(
  () => hasTarget.value && avgProtein.value >= props.proteinTarget,
)
</script>

<template>
  <Motion
    v-if="hasNutrition"
    tag="section"
    class="nutrition-summary"
    aria-label="Näringssammanfattning denna vecka"
    :initial="{ opacity: 0, y: 16 }"
    :animate="{ opacity: 1, y: 0 }"
    :transition="{ duration: 0.4, ease: 'easeOut' }"
  >
    <div class="summary-header">
      <Beef :size="20" :stroke-width="2" class="header-icon" />
      <h2 class="summary-title">Näring denna vecka</h2>
    </div>

    <div class="stat-row">
      <div class="stat-chip" :class="{ reached: targetReached }">
        <Beef :size="16" :stroke-width="2" class="stat-icon" />
        <span class="stat-text">
          Protein: <strong>ca {{ avgProtein }} g</strong>/dag
          <template v-if="hasTarget">
            · mål {{ proteinTarget }} g
          </template>
        </span>
      </div>

      <div class="stat-chip subtle">
        <Flame :size="16" :stroke-width="2" class="stat-icon" />
        <span class="stat-text">
          <strong>ca {{ avgKcal }} kcal</strong>/dag
        </span>
      </div>

      <div v-if="targetReached" class="stat-chip reached-badge">
        <Target :size="16" :stroke-width="2" class="stat-icon" />
        <span class="stat-text">Proteinmål uppnått</span>
      </div>
    </div>
  </Motion>
</template>

<style scoped>
.nutrition-summary {
  margin-top: var(--space-lg);
  padding: var(--space-lg);
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-sm);
}

.summary-header {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  margin-bottom: var(--space-md);
}

.header-icon {
  color: var(--accent);
  flex-shrink: 0;
}

.summary-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.25rem;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.2;
}

.stat-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-sm);
}

.stat-chip {
  display: inline-flex;
  align-items: center;
  gap: var(--space-xs);
  padding: 0.375rem 0.875rem;
  border-radius: var(--radius-full);
  background: var(--accent-bg);
  border: 1px solid var(--border-color-hover);
  font-family: 'Nunito', sans-serif;
}

.stat-chip.subtle {
  background: var(--bg-secondary);
  border-color: var(--border-color);
}

.stat-chip.reached-badge {
  background: var(--accent);
  border-color: var(--accent);
}

.stat-chip.reached-badge .stat-text,
.stat-chip.reached-badge .stat-icon {
  color: var(--text-on-accent);
}

.stat-icon {
  color: var(--accent);
  flex-shrink: 0;
}

.stat-chip.subtle .stat-icon {
  color: var(--text-secondary);
}

.stat-text {
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--text-secondary);
  line-height: 1.2;
}

.stat-text strong {
  font-weight: 800;
  color: var(--text-primary);
}

@media (max-width: 768px) {
  .nutrition-summary {
    padding: var(--space-md);
  }

  .summary-title {
    font-size: 1.1rem;
  }
}
</style>
