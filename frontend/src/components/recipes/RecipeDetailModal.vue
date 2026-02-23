<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { Recipe, CreateRecipeRequest } from '@/api/recipes.api'
import { getRecipe, updateRecipe, deleteRecipe } from '@/api/recipes.api'
import { getCurrentMenu, saveMenu } from '@/api/menu.api'
import type { Menu, MenuDay as ApiMenuDay } from '@/api/menu.api'
import RecipeEditForm from '@/components/recipe-parser/RecipeEditForm.vue'
import ErrorState from '@/components/common/ErrorState.vue'
import { useToast } from '@/composables/useToast'
import { useFocusTrap } from '@/composables/useFocusTrap'

type EditableRecipe = CreateRecipeRequest & { emoji?: string }

interface Props {
  recipeId: string | null
}

const props = defineProps<Props>()

const emit = defineEmits<{
  close: []
  updated: [recipe: Recipe]
  deleted: [recipeId: string]
}>()

const toast = useToast()
const modalCardRef = ref<HTMLElement | null>(null)
const isModalActive = computed(() => !!props.recipeId)

useFocusTrap(modalCardRef, {
  isActive: isModalActive,
  onEscape: handleClose,
})

const recipe = ref<Recipe | null>(null)
const isLoading = ref(false)
const error = ref('')

// Edit mode state
const mode = ref<'detail' | 'edit' | 'confirm-delete'>('detail')
const editableRecipe = ref<EditableRecipe>({
  name: '',
  servings: 4,
  ingredients: [],
  instructions: [],
  tags: [],
  emoji: '',
})
const isSaving = ref(false)
const isDeleting = ref(false)

// Add to menu state
const showDayPicker = ref(false)
const addingToMenu = ref(false)
const currentMenu = ref<Menu | null>(null)
const dayNames = ['Sön', 'Mån', 'Tis', 'Ons', 'Tor', 'Fre', 'Lör']

watch(
  () => props.recipeId,
  async (id) => {
    mode.value = 'detail'
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
  },
  { immediate: true },
)

function startEdit() {
  if (!recipe.value) return
  editableRecipe.value = {
    name: recipe.value.name,
    servings: recipe.value.servings,
    ingredients: recipe.value.ingredients.map((i) => ({ ...i })),
    instructions: [...recipe.value.instructions],
    tags: [...recipe.value.tags],
    emoji: recipe.value.emoji,
  }
  mode.value = 'edit'
}

function cancelEdit() {
  mode.value = 'detail'
}

function handleUpdateRecipe(updated: EditableRecipe) {
  editableRecipe.value = updated
}

async function handleSaveEdit() {
  if (!recipe.value || !props.recipeId) return

  isSaving.value = true
  try {
    const { name, servings, ingredients, instructions, tags, emoji } = editableRecipe.value
    const updated = await updateRecipe(props.recipeId, {
      name,
      servings,
      ingredients,
      instructions,
      tags,
      emoji,
    })
    recipe.value = updated
    mode.value = 'detail'
    toast.success('Receptet har uppdaterats')
    emit('updated', updated)
  } catch {
    toast.error('Kunde inte uppdatera receptet. Försök igen.')
  } finally {
    isSaving.value = false
  }
}

function confirmDelete() {
  mode.value = 'confirm-delete'
}

function cancelDelete() {
  mode.value = 'detail'
}

async function handleDelete() {
  if (!props.recipeId) return

  isDeleting.value = true
  try {
    await deleteRecipe(props.recipeId)
    toast.success('Receptet har tagits bort')
    emit('deleted', props.recipeId)
    emit('close')
  } catch {
    toast.error('Kunde inte ta bort receptet. Försök igen.')
  } finally {
    isDeleting.value = false
  }
}

async function openDayPicker() {
  try {
    const menu = await getCurrentMenu()
    if (!menu) {
      toast.error('Generera en meny först')
      return
    }
    currentMenu.value = menu
    showDayPicker.value = true
  } catch {
    toast.error('Kunde inte hämta menyn')
  }
}

function getDayLabel(day: ApiMenuDay): string {
  const date = new Date(day.date + 'T12:00:00')
  return dayNames[date.getDay()]!
}

async function addToMenuDay(dayIndex: number) {
  if (!currentMenu.value || !recipe.value || !props.recipeId) return

  addingToMenu.value = true
  try {
    const updatedDays = currentMenu.value.days.map((d, i) => {
      if (i === dayIndex) {
        return {
          date: d.date,
          recipeId: props.recipeId!,
          servings: recipe.value!.servings,
        }
      }
      return {
        date: d.date,
        recipeId: d.recipeId,
        servings: d.servings,
        skip: d.skip,
      }
    })

    await saveMenu(updatedDays)
    const day = currentMenu.value.days[dayIndex]
    const label = day ? getDayLabel(day) : ''
    toast.success(`Recept tillagt för ${label}!`)
    showDayPicker.value = false
    emit('updated', recipe.value!)
  } catch {
    toast.error('Kunde inte uppdatera menyn')
  } finally {
    addingToMenu.value = false
  }
}

function handleClose() {
  mode.value = 'detail'
  showDayPicker.value = false
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <div v-if="recipeId" class="modal-overlay" @click="handleClose">
      <div ref="modalCardRef" class="modal-card" role="dialog" aria-modal="true" aria-labelledby="recipe-detail-title" @click.stop>
        <!-- Loading -->
        <div v-if="isLoading" class="modal-loading">
          <div class="spinner-emoji">🍳</div>
          <p class="loading-text">Laddar recept...</p>
        </div>

        <!-- Error -->
        <ErrorState v-else-if="error" icon="⚠️" title="Kunde inte ladda receptet" :show-retry="false" />

        <!-- Confirm delete -->
        <template v-else-if="mode === 'confirm-delete' && recipe">
          <div class="confirm-content">
            <div class="confirm-icon">⚠️</div>
            <h3 class="confirm-title">Ta bort recept?</h3>
            <p class="confirm-text">
              Är du säker på att du vill ta bort
              <strong>{{ recipe.name }}</strong>? Detta kan inte ångras.
            </p>
            <div class="confirm-actions">
              <button class="btn-cancel" @click="cancelDelete">Avbryt</button>
              <button class="btn-delete" :disabled="isDeleting" @click="handleDelete">
                {{ isDeleting ? 'Tar bort...' : 'Ta bort' }}
              </button>
            </div>
          </div>
        </template>

        <!-- Edit mode -->
        <template v-else-if="mode === 'edit'">
          <div class="edit-wrapper">
            <div class="edit-header">
              <h2>Redigera recept</h2>
            </div>
            <div class="edit-body">
              <RecipeEditForm
                :recipe="editableRecipe"
                :confidence="0"
                :warnings="[]"
                :is-saving="isSaving"
                :show-confidence="false"
                back-label="Avbryt"
                @save="handleSaveEdit"
                @back="cancelEdit"
                @update:recipe="handleUpdateRecipe"
              />
            </div>
          </div>
        </template>

        <!-- Detail view -->
        <template v-else-if="recipe">
          <div class="modal-header">
            <div class="recipe-emoji">{{ recipe.emoji || '🍽️' }}</div>
            <h2 id="recipe-detail-title" class="recipe-title">{{ recipe.name }}</h2>
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

          <!-- Day picker for add to menu -->
          <div v-if="showDayPicker && currentMenu" class="day-picker-section">
            <div class="day-picker-header">
              <span class="day-picker-title">Välj dag</span>
              <button class="day-picker-close" @click="showDayPicker = false">&times;</button>
            </div>
            <div class="day-picker-grid">
              <button
                v-for="(day, i) in currentMenu.days"
                :key="day.date"
                class="day-picker-btn"
                :disabled="addingToMenu"
                @click="addToMenuDay(i)"
              >
                <span class="day-picker-label">{{ getDayLabel(day) }}</span>
                <span class="day-picker-current">{{ day.emoji || (day.recipeId ? '🍽️' : '—') }}</span>
              </button>
            </div>
          </div>

          <div class="modal-footer">
            <button class="delete-btn" @click="confirmDelete">Ta bort</button>
            <button class="edit-btn" @click="openDayPicker">Lägg till i meny</button>
            <button class="edit-btn" @click="startEdit">Redigera</button>
            <button class="close-btn" @click="handleClose">Stäng</button>
          </div>
        </template>

        <!-- Fallback footer for loading/error states -->
        <div v-if="isLoading || error" class="modal-footer">
          <button class="close-btn" @click="handleClose">Stäng</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: var(--overlay-bg);
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

.modal-loading {
  text-align: center;
  padding: 3rem 2rem;
}

.spinner-emoji {
  font-size: 3rem;
  animation: spin 1.5s ease-in-out infinite;
}

@keyframes spin {
  0% {
    transform: rotate(0deg) scale(1);
  }
  50% {
    transform: rotate(180deg) scale(1.2);
  }
  100% {
    transform: rotate(360deg) scale(1);
  }
}

.loading-text {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1.1rem;
  color: var(--text-primary);
  margin: 1rem 0 0;
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
  gap: 0.75rem;
}

.close-btn {
  padding: 0.875rem 2.5rem;
  background: var(--accent);
  color: var(--text-on-accent);
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

.edit-btn {
  padding: 0.875rem 2rem;
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.3s ease;
}

.edit-btn:hover {
  border-color: var(--accent);
  color: var(--accent);
  transform: translateY(-2px);
}

.delete-btn {
  padding: 0.875rem 1.5rem;
  background: transparent;
  color: var(--error);
  border: none;
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s ease;
  opacity: 0.7;
}

.delete-btn:hover {
  opacity: 1;
  background: var(--error-bg);
}

/* Confirm delete */
.confirm-content {
  text-align: center;
  padding: 2.5rem 2rem;
}

.confirm-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.confirm-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.5rem;
  color: var(--text-primary);
  margin: 0 0 0.75rem;
}

.confirm-text {
  font-family: 'Nunito', sans-serif;
  font-size: 0.95rem;
  color: var(--text-secondary);
  line-height: 1.5;
  margin: 0 0 1.5rem;
}

.confirm-text strong {
  color: var(--text-primary);
}

.confirm-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: center;
}

.btn-cancel {
  flex: 1;
  max-width: 150px;
  padding: 0.75rem 1rem;
  border-radius: 12px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s ease;
  background: var(--bg-hover);
  border: none;
  color: var(--text-primary);
}

.btn-cancel:hover {
  background: var(--border-color);
}

.btn-delete {
  flex: 1;
  max-width: 150px;
  padding: 0.75rem 1rem;
  border-radius: 12px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s ease;
  background: #e53e3e;
  border: none;
  color: var(--text-on-accent);
}

.btn-delete:hover:not(:disabled) {
  background: #c53030;
  transform: translateY(-1px);
}

.btn-delete:disabled {
  opacity: 0.6;
  cursor: wait;
}

/* Day picker */
.day-picker-section {
  padding: 1rem 2rem;
  border-top: 1px solid var(--border-color);
}

.day-picker-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.75rem;
}

.day-picker-title {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.day-picker-close {
  background: none;
  border: none;
  font-size: 1.25rem;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 0.25rem;
  line-height: 1;
}

.day-picker-close:hover {
  color: var(--text-primary);
}

.day-picker-grid {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.day-picker-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.25rem;
  padding: 0.5rem 0.75rem;
  background: var(--bg-secondary);
  border: 1.5px solid var(--border-color);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  min-width: 52px;
}

.day-picker-btn:hover:not(:disabled) {
  border-color: var(--accent);
  background: var(--bg-hover);
}

.day-picker-btn:disabled {
  opacity: 0.5;
  cursor: wait;
}

.day-picker-label {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.75rem;
  color: var(--text-secondary);
  text-transform: uppercase;
}

.day-picker-current {
  font-size: 1.25rem;
  line-height: 1;
}

/* Edit mode */
.edit-wrapper {
  display: flex;
  flex-direction: column;
  max-height: 85vh;
}

.edit-header {
  padding: 1.5rem 2rem;
  border-bottom: 1px solid var(--border-color);
}

.edit-header h2 {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.5rem;
  color: var(--text-primary);
  margin: 0;
}

.edit-body {
  padding: 1.5rem 2rem;
  overflow-y: auto;
  flex: 1;
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

  .modal-footer {
    flex-wrap: wrap;
  }
}
</style>
