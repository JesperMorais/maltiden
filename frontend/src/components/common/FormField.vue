<script setup lang="ts">
interface Props {
  label: string
  error?: string
  helpText?: string
  required?: boolean
  htmlFor?: string
}

withDefaults(defineProps<Props>(), {
  required: false,
})
</script>

<template>
  <div class="form-field" :class="{ 'has-error': error }">
    <label v-if="label" class="field-label" :for="htmlFor">
      {{ label }}
      <span v-if="required" class="required-mark">*</span>
    </label>
    <slot />
    <p v-if="error" class="field-error">{{ error }}</p>
    <p v-if="helpText && !error" class="field-help">{{ helpText }}</p>
  </div>
</template>

<style scoped>
.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.field-label {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--text-primary);
}

.required-mark {
  color: var(--error);
  margin-left: 0.15rem;
}

.field-error {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.8rem;
  color: var(--error);
  margin: 0;
}

.field-help {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.8rem;
  color: var(--text-muted, var(--text-secondary));
  margin: 0;
}
</style>
