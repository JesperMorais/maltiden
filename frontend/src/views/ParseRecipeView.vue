<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { parseRecipe, createRecipe } from '@/api/recipes.api'
import type { ParseRecipeResponse, CreateRecipeRequest } from '@/api/recipes.api'
import RecipeParseInput from '@/components/recipe-parser/RecipeParseInput.vue'
import RecipeEditForm from '@/components/recipe-parser/RecipeEditForm.vue'
import RecipeParseSuccess from '@/components/recipe-parser/RecipeParseSuccess.vue'

type EditableRecipe = CreateRecipeRequest & { emoji?: string }

const router = useRouter()

// Step state machine
const step = ref<'input' | 'edit' | 'success'>('input')

// Input step
const rawText = ref('')
const isParsing = ref(false)
const parseError = ref('')

// Edit step
const editableRecipe = ref<EditableRecipe>({
  name: '',
  servings: 4,
  ingredients: [],
  instructions: [],
  tags: [],
  emoji: ''
})
const confidence = ref(0)
const warnings = ref<string[]>([])
const isSaving = ref(false)

// Success step
const savedName = ref('')
const savedEmoji = ref('')

async function handleParse() {
  if (!rawText.value.trim()) return

  isParsing.value = true
  parseError.value = ''

  try {
    const result: ParseRecipeResponse = await parseRecipe({ rawText: rawText.value })

    editableRecipe.value = { ...result.recipe }
    confidence.value = result.confidence
    warnings.value = result.warnings ?? []
    step.value = 'edit'
  } catch (err: unknown) {
    const e = err as { response?: { data?: { error?: string } } }
    parseError.value = e?.response?.data?.error || 'Kunde inte tolka receptet. Försök igen.'
  } finally {
    isParsing.value = false
  }
}

function handleBack() {
  step.value = 'input'
}

function handleUpdateRecipe(updated: EditableRecipe) {
  editableRecipe.value = updated
}

async function handleSave() {
  isSaving.value = true

  try {
    const { name, servings, ingredients, instructions, tags } = editableRecipe.value
    await createRecipe({ name, servings, ingredients, instructions, tags })

    savedName.value = editableRecipe.value.name
    savedEmoji.value = editableRecipe.value.emoji || '🍽️'
    step.value = 'success'
  } catch (err: unknown) {
    console.error('Save failed:', err)
  } finally {
    isSaving.value = false
  }
}

function handleParseAnother() {
  rawText.value = ''
  parseError.value = ''
  step.value = 'input'
}

function handleGoDashboard() {
  router.push({ name: 'dashboard' })
}
</script>

<template>
  <div class="parse-recipe-view">
    <!-- Header -->
    <header class="header">
      <div class="header-content">
        <h1 class="title">Tolka recept</h1>
        <p class="description">
          Klistra in en recepttext så tolkar vi det automatiskt åt dig.
        </p>
      </div>
    </header>

    <!-- Main content -->
    <main class="content">
      <div class="content-container">
        <!-- Step 1: Input -->
        <RecipeParseInput
          v-if="step === 'input'"
          v-model="rawText"
          :is-loading="isParsing"
          :error="parseError"
          @parse="handleParse"
        />

        <!-- Step 2: Edit -->
        <RecipeEditForm
          v-if="step === 'edit'"
          :recipe="editableRecipe"
          :confidence="confidence"
          :warnings="warnings"
          :is-saving="isSaving"
          @update:recipe="handleUpdateRecipe"
          @save="handleSave"
          @back="handleBack"
        />

        <!-- Step 3: Success -->
        <RecipeParseSuccess
          v-if="step === 'success'"
          :recipe-name="savedName"
          :emoji="savedEmoji"
          @parse-another="handleParseAnother"
          @go-dashboard="handleGoDashboard"
        />

        <!-- Loading overlay -->
        <div v-if="isParsing" class="loading-overlay">
          <div class="loading-spinner">
            <div class="spinner-emoji">🧑‍🍳</div>
            <p class="loading-text">Claude tolkar ditt recept...</p>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
.parse-recipe-view {
  min-height: 100vh;
  background: var(--bg-primary);
  display: flex;
  flex-direction: column;
}

/* Header */
.header {
  padding: 3rem 2rem 2rem;
  background: linear-gradient(
    180deg,
    var(--bg-card) 0%,
    var(--bg-primary) 100%
  );
  border-bottom: 1px solid var(--border-color);
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  text-align: center;
}

.title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 2.5rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem 0;
  line-height: 1.2;
}

.description {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1.1rem;
  color: var(--text-secondary);
  margin: 0;
  line-height: 1.6;
  max-width: 700px;
  margin: 0 auto;
}

/* Content */
.content {
  flex: 1;
  padding: 3rem 2rem;
  position: relative;
}

.content-container {
  max-width: 1200px;
  margin: 0 auto;
  position: relative;
}

/* Loading overlay */
.loading-overlay {
  position: fixed;
  inset: 0;
  background: rgba(255, 252, 247, 0.9);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.loading-spinner {
  text-align: center;
}

.spinner-emoji {
  font-size: 4rem;
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
  font-size: 1.25rem;
  color: var(--text-primary);
  margin: 1rem 0 0 0;
}

/* Responsive */
@media (max-width: 768px) {
  .header {
    padding: 2rem 1rem 1.5rem;
  }

  .content {
    padding: 2rem 1rem;
  }

  .title {
    font-size: 1.75rem;
  }

  .description {
    font-size: 0.95rem;
  }
}
</style>
