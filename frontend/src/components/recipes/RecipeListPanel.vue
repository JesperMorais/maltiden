<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import type { RecipeSummary, Recipe } from '@/api/recipes.api'
import { getRecipes } from '@/api/recipes.api'
import RecipeCard from '@/components/recipes/RecipeCard.vue'
import RecipeDetailModal from '@/components/recipes/RecipeDetailModal.vue'
import SkeletonSwitch from '@/components/skeleton/SkeletonSwitch.vue'
import RecipesSkeleton from '@/components/skeleton/layouts/RecipesSkeleton.vue'
import ErrorState from '@/components/common/ErrorState.vue'
import EmptyState from '@/components/common/EmptyState.vue'
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
    <ErrorState v-if="error" icon="⚠️" :description="error" @retry="fetchRecipes" />

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
      <EmptyState
        v-else-if="!recipes.length"
        icon="📖"
        title="Inga recept ännu"
        description="Börja med att lägga till ditt första recept."
        action-label="Lägg till recept"
        @action="emit('navigate-to-add')"
      />

      <!-- Empty state: no search results -->
      <EmptyState
        v-else
        icon="🔍"
        title="Inga träffar"
        description="Försök med andra sökord eller ta bort filter."
      />
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
  box-shadow: 0 0 0 4px var(--accent-bg);
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
  color: var(--text-on-accent);
  border-color: var(--accent);
}

/* Recipe grid */
.recipe-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1.5rem;
}

/* Responsive */
@media (max-width: 768px) {
  .recipe-grid {
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
    gap: 1rem;
  }
}

@media (max-width: 480px) {
  .recipe-grid {
    grid-template-columns: 1fr;
  }
}
</style>
