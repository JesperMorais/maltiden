<script setup lang="ts">
import { computed } from 'vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { Trash2 } from 'lucide-vue-next'
import type { Ingredient, CreateRecipeRequest } from '@/api/recipes.api'

type EditableRecipe = CreateRecipeRequest & { emoji?: string }

interface Props {
  recipe: EditableRecipe
  confidence: number
  warnings: string[]
  isSaving: boolean
  showConfidence?: boolean
  backLabel?: string
}

const props = withDefaults(defineProps<Props>(), {
  showConfidence: true,
  backLabel: 'Tolka igen'
})

const emit = defineEmits<{
  (e: 'save'): void
  (e: 'back'): void
  (e: 'update:recipe', value: EditableRecipe): void
}>()

const confidenceLabel = computed(() => {
  if (props.confidence >= 0.8) return { text: 'Hög', class: 'high' }
  if (props.confidence >= 0.5) return { text: 'Medel', class: 'medium' }
  return { text: 'Låg', class: 'low' }
})

const confidencePercent = computed(() => Math.round(props.confidence * 100))

function updateField<K extends keyof EditableRecipe>(field: K, value: EditableRecipe[K]) {
  emit('update:recipe', { ...props.recipe, [field]: value })
}

function updateIngredient(index: number, field: keyof Ingredient, value: string | number) {
  const ingredients = [...props.recipe.ingredients]
  ingredients[index] = { ...ingredients[index], [field]: value } as Ingredient
  updateField('ingredients', ingredients)
}

function addIngredient() {
  updateField('ingredients', [...props.recipe.ingredients, { name: '', amount: 0, unit: '' }])
}

function removeIngredient(index: number) {
  updateField('ingredients', props.recipe.ingredients.filter((_, i) => i !== index))
}

function updateInstruction(index: number, value: string) {
  const instructions = [...props.recipe.instructions]
  instructions[index] = value
  updateField('instructions', instructions)
}

function addInstruction() {
  updateField('instructions', [...props.recipe.instructions, ''])
}

function removeInstruction(index: number) {
  updateField('instructions', props.recipe.instructions.filter((_, i) => i !== index))
}

function updateTags(value: string) {
  updateField('tags', value.split(',').map(t => t.trim()).filter(Boolean))
}

const tagsString = computed(() => props.recipe.tags.join(', '))
</script>

<template>
  <div class="edit-form">
    <!-- Confidence badge -->
    <div v-if="showConfidence" class="confidence-bar">
      <span class="confidence-badge" :class="confidenceLabel.class">
        {{ confidenceLabel.text }} säkerhet — {{ confidencePercent }}%
      </span>
    </div>

    <!-- Warnings -->
    <div v-if="showConfidence && warnings.length" class="warnings">
      <div v-for="(warning, i) in warnings" :key="i" class="warning-item">
        ⚠️ {{ warning }}
      </div>
    </div>

    <!-- Basic info -->
    <div class="form-section">
      <div class="form-row">
        <label class="form-label flex-1">
          <span>Receptnamn</span>
          <input
            :value="recipe.name"
            @input="updateField('name', ($event.target as HTMLInputElement).value)"
            type="text"
            class="form-input"
            placeholder="Namn på receptet"
          />
        </label>
        <label class="form-label emoji-field">
          <span>Emoji</span>
          <input
            :value="recipe.emoji"
            @input="updateField('emoji', ($event.target as HTMLInputElement).value)"
            type="text"
            class="form-input emoji-input"
            placeholder="🍽️"
          />
        </label>
      </div>

      <div class="form-row">
        <label class="form-label servings-field">
          <span>Portioner</span>
          <input
            :value="recipe.servings"
            @input="updateField('servings', Number(($event.target as HTMLInputElement).value))"
            type="number"
            min="1"
            class="form-input"
          />
        </label>
        <label class="form-label flex-1">
          <span>Taggar</span>
          <input
            :value="tagsString"
            @input="updateTags(($event.target as HTMLInputElement).value)"
            type="text"
            class="form-input"
            placeholder="pasta, italienskt, snabb"
          />
        </label>
      </div>
    </div>

    <!-- Ingredients -->
    <div class="form-section">
      <h3 class="section-title">Ingredienser</h3>
      <div class="ingredient-list">
        <div v-for="(ing, i) in recipe.ingredients" :key="i" class="ingredient-row">
          <input
            :value="ing.name"
            @input="updateIngredient(i, 'name', ($event.target as HTMLInputElement).value)"
            type="text"
            class="form-input flex-1"
            placeholder="Ingrediens"
          />
          <input
            :value="ing.amount"
            @input="updateIngredient(i, 'amount', Number(($event.target as HTMLInputElement).value))"
            type="number"
            min="0"
            step="any"
            class="form-input amount-input"
            placeholder="Mängd"
          />
          <input
            :value="ing.unit"
            @input="updateIngredient(i, 'unit', ($event.target as HTMLInputElement).value)"
            type="text"
            class="form-input unit-input"
            placeholder="Enhet"
          />
          <button class="remove-btn" aria-label="Ta bort ingrediens" @click="removeIngredient(i)"><Trash2 :size="16" /></button>
        </div>
      </div>
      <button class="add-btn" @click="addIngredient">+ Lägg till ingrediens</button>
    </div>

    <!-- Instructions -->
    <div class="form-section">
      <h3 class="section-title">Instruktioner</h3>
      <div class="instruction-list">
        <div v-for="(step, i) in recipe.instructions" :key="i" class="instruction-row">
          <span class="step-number">{{ i + 1 }}</span>
          <input
            :value="step"
            @input="updateInstruction(i, ($event.target as HTMLInputElement).value)"
            type="text"
            class="form-input flex-1"
            :placeholder="'Steg ' + (i + 1)"
          />
          <button class="remove-btn" aria-label="Ta bort steg" @click="removeInstruction(i)"><Trash2 :size="16" /></button>
        </div>
      </div>
      <button class="add-btn" @click="addInstruction">+ Lägg till steg</button>
    </div>

    <!-- Actions -->
    <div class="form-actions">
      <BaseButton variant="outline" size="md" @click="emit('back')">
        {{ backLabel }}
      </BaseButton>
      <BaseButton variant="primary" size="lg" :loading="isSaving" @click="emit('save')">
        Spara recept
      </BaseButton>
    </div>
  </div>
</template>

<style scoped>
.edit-form {
  max-width: 750px;
  margin: 0 auto;
}

/* Confidence */
.confidence-bar {
  margin-bottom: 1rem;
  display: flex;
  justify-content: center;
}

.confidence-badge {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.85rem;
  padding: 0.4rem 1rem;
  border-radius: 100px;
}

.confidence-badge.high {
  background: var(--success-bg);
  color: #2f855a;
}

.confidence-badge.medium {
  background: var(--warning-bg);
  color: #b7791f;
}

.confidence-badge.low {
  background: var(--error-bg);
  color: #c53030;
}

/* Warnings */
.warnings {
  margin-bottom: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.warning-item {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.9rem;
  color: #b7791f;
  background: var(--warning-bg);
  padding: 0.5rem 0.75rem;
  border-radius: 8px;
}

/* Form sections */
.form-section {
  margin-bottom: 2rem;
}

.section-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.25rem;
  color: var(--text-primary);
  margin: 0 0 1rem;
}

/* Form fields */
.form-label {
  display: block;
  margin-bottom: 1rem;
}

.form-label span {
  display: block;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
}

.form-input {
  width: 100%;
  padding: 0.75rem 1rem;
  border: 2px solid var(--border-color);
  border-radius: 14px;
  font-family: 'Nunito', sans-serif;
  font-size: 0.95rem;
  color: var(--text-primary);
  background: var(--bg-card);
  transition: all 0.3s ease;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 4px var(--accent-bg);
}

.form-input::placeholder {
  color: var(--text-secondary);
  opacity: 0.6;
}

/* Layout helpers */
.form-row {
  display: flex;
  gap: 1rem;
  align-items: flex-start;
}

.flex-1 {
  flex: 1;
  min-width: 0;
}

.emoji-field {
  width: 80px;
  flex-shrink: 0;
}

.emoji-input {
  text-align: center;
  font-size: 1.5rem;
  padding: 0.5rem;
}

.servings-field {
  width: 120px;
  flex-shrink: 0;
}

/* Ingredients */
.ingredient-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.ingredient-row {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.amount-input {
  width: 90px;
  flex-shrink: 0;
}

.unit-input {
  width: 90px;
  flex-shrink: 0;
}

/* Instructions */
.instruction-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.instruction-row {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.step-number {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1rem;
  color: var(--accent-text);
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--accent-bg);
  border-radius: 50%;
  flex-shrink: 0;
}

/* Shared buttons */
.remove-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1rem;
  padding: 0.5rem;
  border-radius: 8px;
  transition: all 0.2s ease;
  flex-shrink: 0;
  opacity: 0.5;
}

.remove-btn:hover {
  opacity: 1;
  background: var(--error-bg);
}

.add-btn {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--accent-text);
  background: none;
  border: 2px dashed var(--border-color);
  border-radius: 14px;
  padding: 0.6rem 1rem;
  cursor: pointer;
  width: 100%;
  margin-top: 0.5rem;
  transition: all 0.3s ease;
}

.add-btn:hover {
  border-color: var(--accent);
  background: var(--bg-hover);
}

/* Actions */
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border-color);
}

/* Responsive */
@media (max-width: 600px) {
  .form-row {
    flex-direction: column;
    gap: 0;
  }

  .emoji-field,
  .servings-field {
    width: 100%;
  }

  .ingredient-row {
    flex-wrap: wrap;
  }

  .amount-input,
  .unit-input {
    width: calc(50% - 1.25rem);
    flex: none;
  }
}

@media (max-width: 480px) {
  .ingredient-row {
    flex-wrap: wrap;
    gap: 0.35rem;
  }

  .ingredient-row .flex-1 {
    flex: 1 1 100%;
  }

  .amount-input,
  .unit-input {
    width: calc(50% - 1rem);
    flex: 1;
  }

  .form-actions {
    flex-direction: column;
  }

  .form-actions > * {
    width: 100%;
  }
}
</style>
