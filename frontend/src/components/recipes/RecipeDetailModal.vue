<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Recipe } from '@/api/recipes.api'
import { getRecipe } from '@/api/recipes.api'

interface Props {
  recipeId: string | null
}

const props = defineProps<Props>()

defineEmits<{
  close: []
}>()

const recipe = ref<Recipe | null>(null)
const isLoading = ref(false)
const error = ref('')

watch(() => props.recipeId, async (id) => {
  if (!id) {
    recipe.value = null
    return
  }

  isLoading.value = true
  error.value = ''
  try {
    recipe.value = await getRecipe(id)
  } catch {
    error.value = 'Kunde inte ladda receptet'
  } finally {
    isLoading.value = false
  }
}, { immediate: true })
</script>

<template>
  <Teleport to="body">
    <div v-if="recipeId" class="modal-overlay" @click="$emit('close')">
      <div class="modal-card" @click.stop>
        <!-- Loading -->
        <div v-if="isLoading" class="modal-loading">
          <div class="spinner-emoji">🍳</div>
          <p class="loading-text">Laddar recept...</p>
        </div>

        <!-- Error -->
        <div v-else-if="error" class="modal-error">
          <div class="error-icon">⚠️</div>
          <p class="error-text">{{ error }}</p>
        </div>

        <!-- Recipe content -->
        <template v-else-if="recipe">
          <div class="modal-header">
            <div class="recipe-emoji">{{ recipe.emoji || '🍽️' }}</div>
            <h2 class="recipe-title">{{ recipe.name }}</h2>
            <p class="recipe-meta">{{ recipe.servings }} portioner</p>
            <div v-if="recipe.tags.length" class="recipe-tags">
              <span v-for="tag in recipe.tags" :key="tag" class="tag-chip">
                {{ tag }}
              </span>
            </div>
          </div>

          <div class="modal-body">
            <section class="recipe-section">
              <h3 class="section-title">Ingredienser</h3>
              <ul class="ingredients-list">
                <li v-for="(ing, i) in recipe.ingredients" :key="i" class="ingredient-item">
                  <span class="ingredient-amount">{{ ing.amount }} {{ ing.unit }}</span>
                  <span class="ingredient-name">{{ ing.name }}</span>
                </li>
              </ul>
            </section>

            <section class="recipe-section">
              <h3 class="section-title">Instruktioner</h3>
              <ol class="instructions-list">
                <li v-for="(step, i) in recipe.instructions" :key="i" class="instruction-item">
                  {{ step }}
                </li>
              </ol>
            </section>
          </div>
        </template>

        <div class="modal-footer">
          <button class="close-btn" @click="$emit('close')">Stäng</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 2rem;
}

.modal-card {
  background: var(--bg-card);
  border-radius: 24px;
  box-shadow: var(--shadow-lg);
  max-width: 600px;
  width: 100%;
  max-height: 85vh;
  overflow-y: auto;
}

.modal-loading,
.modal-error {
  text-align: center;
  padding: 3rem 2rem;
}

.spinner-emoji {
  font-size: 3rem;
  animation: spin 1.5s ease-in-out infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg) scale(1); }
  50% { transform: rotate(180deg) scale(1.2); }
  100% { transform: rotate(360deg) scale(1); }
}

.loading-text {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1.1rem;
  color: var(--text-primary);
  margin: 1rem 0 0;
}

.error-icon {
  font-size: 3rem;
  margin-bottom: 0.5rem;
}

.error-text {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  color: var(--text-secondary);
  margin: 0;
}

.modal-header {
  padding: 2rem 2rem 1rem;
  text-align: center;
  border-bottom: 1px solid var(--border-color);
}

.recipe-emoji {
  font-size: 3.5rem;
  line-height: 1;
  margin-bottom: 0.75rem;
}

.recipe-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.75rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
}

.recipe-meta {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1rem;
  color: var(--text-secondary);
  margin: 0 0 0.75rem;
}

.recipe-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
  justify-content: center;
  margin-bottom: 0.5rem;
}

.tag-chip {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.75rem;
  padding: 0.25rem 0.75rem;
  border-radius: 100px;
  background: var(--bg-secondary);
  color: var(--text-secondary);
  border: 1px solid var(--border-color);
}

.modal-body {
  padding: 1.5rem 2rem;
}

.recipe-section {
  margin-bottom: 1.5rem;
}

.recipe-section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.2rem;
  color: var(--text-primary);
  margin: 0 0 0.75rem;
}

.ingredients-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.ingredient-item {
  font-family: 'Nunito', sans-serif;
  font-size: 0.95rem;
  color: var(--text-primary);
  display: flex;
  gap: 0.5rem;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--border-color);
}

.ingredient-amount {
  font-weight: 700;
  white-space: nowrap;
  min-width: 80px;
}

.ingredient-name {
  font-weight: 600;
}

.instructions-list {
  padding: 0 0 0 1.5rem;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.instruction-item {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--text-primary);
  line-height: 1.6;
}

.modal-footer {
  padding: 1.5rem 2rem;
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: center;
}

.close-btn {
  padding: 0.875rem 2.5rem;
  background: var(--accent);
  color: white;
  border: none;
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.close-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(255, 107, 91, 0.4);
}

@media (max-width: 768px) {
  .modal-overlay {
    padding: 1rem;
  }

  .modal-card {
    max-height: 90vh;
  }

  .modal-header {
    padding: 1.5rem 1.5rem 1rem;
  }

  .modal-body {
    padding: 1.25rem 1.5rem;
  }

  .recipe-title {
    font-size: 1.5rem;
  }

  .recipe-emoji {
    font-size: 2.5rem;
  }
}
</style>
