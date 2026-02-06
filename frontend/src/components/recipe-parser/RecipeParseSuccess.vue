<script setup lang="ts">
import BaseButton from '@/components/common/BaseButton.vue'

interface Props {
  recipeName: string
  emoji: string
}

defineProps<Props>()

const emit = defineEmits<{
  (e: 'parse-another'): void
  (e: 'go-dashboard'): void
}>()
</script>

<template>
  <div class="parse-success">
    <div class="success-icon">{{ emoji }}</div>
    <h2 class="success-title">Recept sparat!</h2>
    <p class="success-name">{{ recipeName }}</p>

    <div class="success-actions">
      <BaseButton variant="primary" size="lg" @click="emit('parse-another')">
        Tolka nytt recept
      </BaseButton>
      <BaseButton variant="outline" size="md" @click="emit('go-dashboard')">
        Till dashboard
      </BaseButton>
    </div>
  </div>
</template>

<style scoped>
.parse-success {
  text-align: center;
  padding: 3rem 2rem;
  max-width: 500px;
  margin: 0 auto;
}

.success-icon {
  font-size: 5rem;
  margin-bottom: 1.5rem;
  animation: celebrate 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes celebrate {
  0% {
    transform: scale(0);
    opacity: 0;
  }
  100% {
    transform: scale(1);
    opacity: 1;
  }
}

.success-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 2rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
}

.success-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1.25rem;
  color: var(--text-secondary);
  margin: 0 0 2.5rem;
}

.success-actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}
</style>
