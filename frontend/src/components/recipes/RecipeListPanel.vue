<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import type { RecipeSummary, Recipe } from '@/api/recipes.api'
import { getRecipes } from '@/api/recipes.api'
import RecipeCard from '@/components/recipes/RecipeCard.vue'
import RecipeDetailModal from '@/components/recipes/RecipeDetailModal.vue'
import SkeletonSwitch from '@/components/skeleton/SkeletonSwitch.vue'
import RecipesSkeleton from '@/components/skeleton/layouts/RecipesSkeleton.vue'
import { useSkeleton } from '@/composables/useSkeleton'

const emit = defineEmits<{
  (e: 'navigate-to-add'): void
}>()

const recipes = ref<RecipeSummary[]>([])
const isLoading = ref(true)
const { showSkeleton } = useSkeleton(isLoading, { minDuration: 300 })
const error = ref('')
const searchTerm = ref('')
const selectedTags = ref<Set<string>>(new Set())
const selectedRecipeId = ref<string | null>(null)

// All unique tags across recipes
const allTags = computed(() => {
  const tags = new Set<string>()
  for (const r of recipes.value) {
    for (const t of r.tags) {
      tags.add(t)
    }
  }
  return Array.from(tags).sort()
})

// Filtered recipes based on search + tag filters
const filteredRecipes = computed(() => {
  let result = recipes.value

  const term = searchTerm.value.toLowerCase().trim()
  if (term) {
    result = result.filter(r =>
      r.name.toLowerCase().includes(term) ||
      r.tags.some(t => t.toLowerCase().includes(term))
    )
  }

  if (selectedTags.value.size > 0) {
    result = result.filter(r =>
      r.tags.some(t => selectedTags.value.has(t))
    )
  }

  return result
})

function toggleTag(tag: string) {
  const next = new Set(selectedTags.value)
  if (next.has(tag)) {
    next.delete(tag)
  } else {
    next.add(tag)
  }
  selectedTags.value = next
}

async function fetchRecipes() {
  isLoading.value = true
  error.value = ''
  try {
    const data = await getRecipes()
    recipes.value = data.recipes
  } catch {
    error.value = 'Kunde inte ladda recept. Försök igen.'
  } finally {
    isLoading.value = false
  }
}

function openRecipe(id: string) {
  selectedRecipeId.value = id
}

function closeDetail() {
  selectedRecipeId.value = null
}

function handleRecipeUpdated(updated: Recipe) {
  const index = recipes.value.findIndex((r) => r.id === updated.id)
  if (index !== -1) {
    recipes.value[index] = {
      id: updated.id,
      name: updated.name,
      servings: updated.servings,
      tags: updated.tags,
      emoji: updated.emoji,
    }
  }
}

function handleRecipeDeleted(recipeId: string) {
  recipes.value = recipes.value.filter((r) => r.id !== recipeId)
  selectedRecipeId.value = null
}

function refresh() {
  fetchRecipes()
}

defineExpose({ refresh })

onMounted(fetchRecipes)
</script>

<template>
  <div class="recipe-list-panel">
    <!-- Error state -->
    <div v-if="error" class="error-state">
      <div class="error-icon">⚠️</div>
      <p class="error-message">{{ error }}</p>
      <button class="retry-button" @click="fetchRecipes">
        Försök igen
      </button>
    </div>

    <!-- Skeleton / Content switch -->
    <SkeletonSwitch v-else :loading="showSkeleton">
      <template #skeleton>
        <RecipesSkeleton />
      </template>

      <!-- Search & filters -->
      <div v-if="recipes.length" class="toolbar">
        <input
          v-model="searchTerm"
          type="text"
          placeholder="Sök recept..."
          class="search-input"
        />
        <div v-if="allTags.length" class="tag-filters">
          <button
            v-for="tag in allTags"
            :key="tag"
            class="filter-chip"
            :class="{ active: selectedTags.has(tag) }"
            @click="toggleTag(tag)"
          >
            {{ tag }}
          </button>
        </div>
      </div>

      <!-- Recipe grid -->
      <div v-if="filteredRecipes.length" class="recipe-grid">
        <RecipeCard
          v-for="recipe in filteredRecipes"
          :key="recipe.id"
          :recipe="recipe"
          @click="openRecipe(recipe.id)"
        />
      </div>

      <!-- Empty state: no recipes at all -->
      <div v-else-if="!recipes.length" class="empty-state">
        <div class="empty-emoji">📖</div>
        <h2 class="empty-title">Inga recept ännu</h2>
        <p class="empty-text">
          Börja med att lägga till ditt första recept.
        </p>
        <button class="cta-button" @click="emit('navigate-to-add')">
          Lägg till recept
        </button>
      </div>

      <!-- Empty state: no search results -->
      <div v-else class="empty-state">
        <div class="empty-emoji">🔍</div>
        <h2 class="empty-title">Inga träffar</h2>
        <p class="empty-text">
          Försök med andra sökord eller ta bort filter.
        </p>
      </div>
    </SkeletonSwitch>

    <!-- Detail modal -->
    <RecipeDetailModal
      :recipe-id="selectedRecipeId"
      @close="closeDetail"
      @updated="handleRecipeUpdated"
      @deleted="handleRecipeDeleted"
    />
  </div>
</template>

<style scoped>
.recipe-list-panel {
  position: relative;
}

/* Toolbar */
.toolbar {
  margin-bottom: 2rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.search-input {
  width: 100%;
  padding: 0.9rem 1.25rem;
  border: 2px solid var(--border-color);
  border-radius: 14px;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1rem;
  color: var(--text-primary);
  background: var(--bg-card);
  transition: all 0.3s ease;
  box-sizing: border-box;
}

.search-input:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 4px rgba(255, 107, 91, 0.1);
}

.search-input::placeholder {
  color: var(--text-secondary);
  opacity: 0.6;
}

.tag-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.filter-chip {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.8rem;
  padding: 0.35rem 0.9rem;
  border-radius: 100px;
  background: var(--bg-secondary);
  color: var(--text-secondary);
  border: 1.5px solid var(--border-color);
  cursor: pointer;
  transition: all 0.2s ease;
}

.filter-chip:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.filter-chip.active {
  background: var(--accent);
  color: white;
  border-color: var(--accent);
}

/* Recipe grid */
.recipe-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 1.5rem;
}

/* Error state */
.error-state {
  text-align: center;
  padding: 3rem 2rem;
}

.error-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.error-message {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1.1rem;
  color: var(--text-primary);
  margin: 0 0 1.5rem;
}

.retry-button {
  padding: 0.875rem 2rem;
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

.retry-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(255, 107, 91, 0.4);
}

/* Empty state */
.empty-state {
  text-align: center;
  padding: 4rem 2rem;
}

.empty-emoji {
  font-size: 4rem;
  margin-bottom: 1rem;
}

.empty-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.75rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
}

.empty-text {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1.05rem;
  color: var(--text-secondary);
  margin: 0 0 1.5rem;
  line-height: 1.6;
}

.cta-button {
  padding: 0.875rem 2rem;
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

.cta-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(255, 107, 91, 0.4);
}

/* Responsive */
@media (max-width: 768px) {
  .recipe-grid {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 1rem;
  }
}
</style>
