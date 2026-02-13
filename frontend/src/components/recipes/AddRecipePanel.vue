<script setup lang="ts">
import { ref } from 'vue'
import { parseRecipe, createRecipe } from '@/api/recipes.api'
import type { ParseRecipeResponse, CreateRecipeRequest } from '@/api/recipes.api'
import RecipeParseInput from '@/components/recipe-parser/RecipeParseInput.vue'
import RecipeEditForm from '@/components/recipe-parser/RecipeEditForm.vue'
import RecipeParseSuccess from '@/components/recipe-parser/RecipeParseSuccess.vue'
import ClickSpark from '@/components/vue-bits/ClickSpark.vue'
import FadeContent from '@/components/vue-bits/FadeContent.vue'
import ProgressBar from '@/components/common/ProgressBar.vue'
import { useProgressBar } from '@/composables/useProgressBar'

type EditableRecipe = CreateRecipeRequest & { emoji?: string }

// Progress bar for AI parsing (Claude API can take up to 60s)
const { progress: parseProgress, isActive: parseActive, start: parseStart, finish: parseFinish } =
  useProgressBar({ duration: 20000 })

const emit = defineEmits<{
  (e: 'navigate-to-list'): void
}>()

// State machine
const step = ref<'choose' | 'ai-input' | 'ai-edit' | 'manual-edit' | 'success'>('choose')

// AI input step
const rawText = ref('')
const isParsing = ref(false)
const parseError = ref('')

// Edit step (shared by AI and manual)
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

function chooseAI() {
  step.value = 'ai-input'
}

function chooseManual() {
  editableRecipe.value = {
    name: '',
    servings: 4,
    ingredients: [{ name: '', amount: 0, unit: '' }],
    instructions: [''],
    tags: [],
    emoji: ''
  }
  confidence.value = 0
  warnings.value = []
  step.value = 'manual-edit'
}

async function handleParse() {
  if (!rawText.value.trim()) return

  isParsing.value = true
  parseError.value = ''
  parseStart()

  try {
    const result: ParseRecipeResponse = await parseRecipe({ rawText: rawText.value })

    editableRecipe.value = { ...result.recipe }
    confidence.value = result.confidence
    warnings.value = result.warnings ?? []
    step.value = 'ai-edit'
  } catch (err: unknown) {
    const e = err as { response?: { data?: { error?: string } } }
    parseError.value = e?.response?.data?.error || 'Kunde inte tolka receptet. Försök igen.'
  } finally {
    parseFinish()
    isParsing.value = false
  }
}

function handleBackToInput() {
  step.value = 'ai-input'
}

function handleBackToChoose() {
  step.value = 'choose'
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

function handleAddMore() {
  rawText.value = ''
  parseError.value = ''
  step.value = 'choose'
}

function handleViewRecipes() {
  emit('navigate-to-list')
}
</script>

<template>
  <div class="add-recipe-panel">
    <!-- Step: Choose method -->
    <div v-if="step === 'choose'" class="choose-step">
      <FadeContent :duration="600" :blur="true">
        <p class="choose-subtitle">Hur vill du lägga till ditt recept?</p>
      </FadeContent>

      <div class="choose-cards">
        <FadeContent :duration="500" :delay="100">
          <ClickSpark spark-color="#ff6b5b" :spark-radius="30" :spark-count="10" :duration="500">
            <button class="choose-card" @click="chooseAI">
              <span class="choose-card-icon">🤖</span>
              <span class="choose-card-title">Tolka med AI</span>
              <span class="choose-card-desc">
                Klistra in en recepttext så tolkar vi det automatiskt
              </span>
            </button>
          </ClickSpark>
        </FadeContent>

        <FadeContent :duration="500" :delay="200">
          <ClickSpark spark-color="#ff6b5b" :spark-radius="30" :spark-count="10" :duration="500">
            <button class="choose-card" @click="chooseManual">
              <span class="choose-card-icon">✏️</span>
              <span class="choose-card-title">Fyll i själv</span>
              <span class="choose-card-desc">
                Skriv in receptet manuellt steg för steg
              </span>
            </button>
          </ClickSpark>
        </FadeContent>
      </div>
    </div>

    <!-- Step: AI input -->
    <div v-if="step === 'ai-input'" class="ai-input-step">
      <button class="back-link" @click="handleBackToChoose">
        <span class="back-arrow">&larr;</span>
        <span>Tillbaka</span>
      </button>

      <RecipeParseInput
        v-model="rawText"
        :is-loading="isParsing"
        :error="parseError"
        @parse="handleParse"
      />
    </div>

    <!-- Step: AI edit -->
    <RecipeEditForm
      v-if="step === 'ai-edit'"
      :recipe="editableRecipe"
      :confidence="confidence"
      :warnings="warnings"
      :is-saving="isSaving"
      back-label="Tolka igen"
      @update:recipe="handleUpdateRecipe"
      @save="handleSave"
      @back="handleBackToInput"
    />

    <!-- Step: Manual edit -->
    <RecipeEditForm
      v-if="step === 'manual-edit'"
      :recipe="editableRecipe"
      :confidence="0"
      :warnings="[]"
      :is-saving="isSaving"
      :show-confidence="false"
      back-label="Avbryt"
      @update:recipe="handleUpdateRecipe"
      @save="handleSave"
      @back="handleBackToChoose"
    />

    <!-- Step: Success -->
    <RecipeParseSuccess
      v-if="step === 'success'"
      :recipe-name="savedName"
      :emoji="savedEmoji"
      primary-action-label="Lägg till fler"
      secondary-action-label="Visa mina recept"
      @parse-another="handleAddMore"
      @go-dashboard="handleViewRecipes"
    />

    <!-- Loading overlay -->
    <div v-if="isParsing" class="loading-overlay">
      <div class="loading-spinner">
        <div class="spinner-emoji">🧑‍🍳</div>
        <p class="loading-text">Claude tolkar ditt recept...</p>
        <ProgressBar :progress="parseProgress" :active="parseActive" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.add-recipe-panel {
  position: relative;
}

/* Choose step */
.choose-step {
  max-width: 600px;
  margin: 0 auto;
  text-align: center;
}

.choose-subtitle {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1.1rem;
  color: var(--text-secondary);
  margin: 0 0 2rem;
  line-height: 1.6;
}

.choose-cards {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.25rem;
}

.choose-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 2rem 1.5rem;
  background: var(--bg-card);
  border: 2px solid var(--border-color);
  border-radius: 20px;
  cursor: pointer;
  text-align: center;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.choose-card:hover {
  border-color: var(--accent);
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(255, 107, 91, 0.15);
}

.choose-card-icon {
  font-size: 2.5rem;
}

.choose-card-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.2rem;
  color: var(--text-primary);
}

.choose-card-desc {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--text-secondary);
  line-height: 1.5;
}

/* AI input step */
.ai-input-step {
  max-width: 700px;
  margin: 0 auto;
}

/* Back link */
.back-link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: none;
  border: none;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 0;
  margin-bottom: 1.5rem;
  transition: color 0.2s ease;
}

.back-link:hover {
  color: var(--accent);
}

.back-arrow {
  font-size: 1.1rem;
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
  width: 280px;
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
@media (max-width: 500px) {
  .choose-cards {
    grid-template-columns: 1fr;
  }
}
</style>
