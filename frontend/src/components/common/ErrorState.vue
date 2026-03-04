<script setup lang="ts">
import { AlertTriangle } from 'lucide-vue-next'

interface Props {
  title?: string
  description?: string
  retryLabel?: string
  showRetry?: boolean
}

withDefaults(defineProps<Props>(), {
  title: 'Något gick fel',
  description: undefined,
  retryLabel: 'Försök igen',
  showRetry: true,
})

const emit = defineEmits<{
  retry: []
}>()
</script>

<template>
  <div class="error-state" role="alert">
    <div class="error-state-icon" aria-hidden="true">
      <slot name="icon"><AlertTriangle :size="48" color="var(--text-muted)" /></slot>
    </div>
    <h2 class="error-state-title">{{ title }}</h2>
    <p v-if="description" class="error-state-description">{{ description }}</p>
    <button v-if="showRetry" class="error-state-retry" @click="emit('retry')">
      {{ retryLabel }}
    </button>
  </div>
</template>

<style scoped>
.error-state {
  text-align: center;
  padding: 4rem 2rem;
}

.error-state-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
}

.error-state-title {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.5rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
}

.error-state-description {
  font-family: 'Nunito', sans-serif;
  color: var(--text-secondary);
  margin: 0 0 1.5rem;
}

.error-state-retry {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1rem;
  color: var(--text-on-accent);
  background: var(--accent);
  border: none;
  border-radius: 100px;
  padding: 0.85em 2em;
  cursor: pointer;
  transition: all 0.3s ease;
}

.error-state-retry:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-accent);
}
</style>
