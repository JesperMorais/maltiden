<script setup lang="ts">
/**
 * Skeleton layout mirroring GenerateMenuView.vue
 *
 * Matches the exact grid, spacing, and border-radius values from:
 * - MenuDayCard.vue (5-column grid, 20px radius, 280px min-height)
 * - Action bar (back button, status text, save/generate buttons)
 */
import SkeletonBlock from '../SkeletonBlock.vue'
import SkeletonCircle from '../SkeletonCircle.vue'

interface Props {
  dayCount?: number
}

withDefaults(defineProps<Props>(), {
  dayCount: 7,
})
</script>

<template>
  <div class="generate-menu-skeleton" aria-hidden="true">
    <div
      class="menu-grid-skeleton"
      :style="{ gridTemplateColumns: `repeat(${dayCount}, 1fr)` }"
    >
      <div v-for="i in dayCount" :key="i" class="day-card-skeleton">
        <SkeletonBlock width="48px" height="12px" radius="6px" />
        <SkeletonCircle size="64px" />
        <SkeletonBlock width="80%" height="18px" radius="8px" />
        <SkeletonBlock width="60%" height="12px" radius="6px" />
      </div>
    </div>
    <!-- Action bar skeleton -->
    <div class="actions-skeleton">
      <SkeletonBlock width="90px" height="40px" radius="100px" />
      <SkeletonBlock width="140px" height="14px" radius="8px" />
      <div class="actions-right-skeleton">
        <SkeletonBlock width="130px" height="40px" radius="100px" />
        <SkeletonBlock width="110px" height="40px" radius="100px" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.generate-menu-skeleton {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.menu-grid-skeleton {
  display: grid;
  /* grid-template-columns is bound via :style to match dayCount */
  gap: 1.5rem;
}

.day-card-skeleton {
  background: var(--bg-card);
  border: 2px solid var(--border-color);
  border-radius: 20px;
  padding: 1.5rem;
  min-height: 280px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
}

.actions-skeleton {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem 1.5rem;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  box-shadow: var(--shadow-sm);
}

.actions-right-skeleton {
  display: flex;
  gap: 0.75rem;
}

@media (max-width: 1024px) {
  .menu-grid-skeleton {
    grid-template-columns: repeat(3, 1fr);
    gap: 1rem;
  }
}

@media (max-width: 768px) {
  .menu-grid-skeleton {
    grid-template-columns: repeat(2, 1fr);
    gap: 0.75rem;
  }

  .day-card-skeleton {
    min-height: 220px;
    padding: 1rem;
  }

  .actions-skeleton {
    grid-template-columns: 1fr;
    justify-items: center;
    gap: 0.75rem;
  }

  .actions-right-skeleton {
    width: 100%;
    justify-content: center;
  }
}

@media (max-width: 480px) {
  .menu-grid-skeleton {
    grid-template-columns: 1fr;
  }
}
</style>
