<script setup lang="ts">
import type { RecipeSummary } from '@/api/recipes.api'
import BaseCard from '@/components/common/BaseCard.vue'
import SpotlightCard from '@/components/vue-bits/SpotlightCard.vue'

interface Props {
  recipe: RecipeSummary
}

defineProps<Props>()

defineEmits<{
  click: []
}>()
</script>

<template>
  <SpotlightCard
    spotlight-color="rgba(255, 107, 91, 0.12)"
    class-name="recipe-spotlight"
  >
    <BaseCard class="recipe-card" padding="md" @click="$emit('click')">
      <div class="recipe-emoji">{{ recipe.emoji || '🍽️' }}</div>
      <h3 class="recipe-name">{{ recipe.name }}</h3>
      <p class="recipe-servings">{{ recipe.servings }} portioner</p>
      <div v-if="recipe.tags.length" class="recipe-tags">
        <span v-for="tag in recipe.tags" :key="tag" class="tag-chip">
          {{ tag }}
        </span>
      </div>
    </BaseCard>
  </SpotlightCard>
</template>

<style scoped>
:deep(.recipe-spotlight) {
  border-radius: 20px;
}

.recipe-card {
  cursor: pointer;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  min-height: 220px;
  justify-content: center;
}

.recipe-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.recipe-emoji {
  font-size: 3.5rem;
  line-height: 1;
  margin-bottom: 0.25rem;
}

.recipe-name {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.1rem;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.3;
}

.recipe-servings {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin: 0;
}

.recipe-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
  justify-content: center;
  margin-top: 0.25rem;
}

.tag-chip {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.7rem;
  padding: 0.2rem 0.6rem;
  border-radius: 100px;
  background: var(--bg-secondary);
  color: var(--text-secondary);
  border: 1px solid var(--border-color);
}

@media (max-width: 768px) {
  .recipe-card {
    min-height: 180px;
  }

  .recipe-emoji {
    font-size: 2.5rem;
  }

  .recipe-name {
    font-size: 1rem;
  }
}
</style>
