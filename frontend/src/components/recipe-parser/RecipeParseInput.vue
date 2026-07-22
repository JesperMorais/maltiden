<script setup lang="ts">
import BaseButton from '@/components/common/BaseButton.vue'

interface Props {
  modelValue: string
  isLoading: boolean
  error: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'parse'): void
}>()

const MAX_CHARS = 10000
</script>

<template>
  <div class="parse-input">
    <label class="form-label">
      <span>Klistra in recepttext</span>
      <textarea
        aria-label="Klistra in recepttext"
        :value="props.modelValue"
        @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
        class="form-input textarea"
        :maxlength="MAX_CHARS"
        placeholder="Pasta Carbonara

4 portioner

400 g spaghetti
200 g bacon
4 äggulor
100 g parmesan
Svartpeppar efter smak

Koka pastan enligt förpackningen.
Stek baconet knaprigt.
Vispa ihop äggulor och parmesan.
Blanda het pasta med bacon.
Rör ner äggblandningen."
      />
    </label>

    <div class="input-footer">
      <span class="char-count" :class="{ warning: props.modelValue.length > MAX_CHARS * 0.9 }">
        {{ props.modelValue.length }} / {{ MAX_CHARS }}
      </span>
      <BaseButton
        variant="primary"
        size="lg"
        :disabled="!props.modelValue.trim()"
        :loading="props.isLoading"
        @click="emit('parse')"
      >
        Tolka recept
      </BaseButton>
    </div>

    <p v-if="props.error" class="form-error">{{ props.error }}</p>
  </div>
</template>

<style scoped>
.parse-input {
  max-width: 700px;
  margin: 0 auto;
}

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
  padding: 0.9rem 1.25rem;
  border: 2px solid var(--border-color);
  border-radius: 14px;
  font-family: 'Nunito', sans-serif;
  font-size: 1rem;
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

.textarea {
  min-height: 300px;
  resize: vertical;
  line-height: 1.6;
}

.input-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.char-count {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.char-count.warning {
  color: var(--accent-text);
}

@media (max-width: 480px) {
  .textarea {
    min-height: 180px;
  }
}

.form-error {
  font-family: 'Nunito', sans-serif;
  font-size: 0.9rem;
  color: var(--error);
  margin: 1rem 0 0;
  padding: 0.5rem 0.75rem;
  background: var(--error-bg);
  border-radius: 8px;
  text-align: center;
}
</style>
