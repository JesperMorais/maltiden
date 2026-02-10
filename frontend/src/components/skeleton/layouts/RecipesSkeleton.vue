<script setup lang="ts">
/**
 * Skeleton layout mirroring the recipe list
 *
 * Matches the recipe grid with cards showing emoji circle,
 * title block, servings text, and tag chips.
 */
import SkeletonBlock from '../SkeletonBlock.vue'
import SkeletonCircle from '../SkeletonCircle.vue'
</script>

<template>
  <div class="recipes-skeleton" aria-hidden="true">
    <!-- Search bar skeleton -->
    <div class="toolbar-skeleton">
      <SkeletonBlock width="100%" height="52px" radius="14px" />
      <div class="tag-filters-skeleton">
        <SkeletonBlock v-for="i in 5" :key="i" :width="`${50 + i * 12}px`" height="30px" radius="100px" />
      </div>
    </div>

    <!-- Recipe grid skeleton -->
    <div class="recipe-grid-skeleton">
      <div v-for="i in 6" :key="i" class="recipe-card-skeleton">
        <SkeletonCircle size="56px" />
        <SkeletonBlock width="75%" height="1.1rem" radius="8px" />
        <SkeletonBlock width="40%" height="0.85rem" radius="6px" />
        <div class="tags-skeleton">
          <SkeletonBlock width="48px" height="20px" radius="100px" />
          <SkeletonBlock width="56px" height="20px" radius="100px" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.recipes-skeleton {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.toolbar-skeleton {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.tag-filters-skeleton {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.recipe-grid-skeleton {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 1.5rem;
}

.recipe-card-skeleton {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 24px;
  padding: 1.5rem;
  min-height: 220px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
}

.tags-skeleton {
  display: flex;
  gap: 0.375rem;
  margin-top: 0.25rem;
}

@media (max-width: 768px) {
  .recipe-grid-skeleton {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 1rem;
  }

  .recipe-card-skeleton {
    min-height: 180px;
    padding: 1.25rem;
  }
}
</style>
