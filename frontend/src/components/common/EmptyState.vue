<script setup lang="ts">
interface Props {
  icon?: string
  title: string
  description?: string
  actionLabel?: string
}

defineProps<Props>()

const emit = defineEmits<{
  action: []
}>()
</script>

<template>
  <div class="empty-state">
    <div class="empty-state-icon">
      <slot name="icon">
        <span v-if="icon" class="icon-text" aria-hidden="true">{{ icon }}</span>
      </slot>
    </div>
    <h2 class="empty-state-title">{{ title }}</h2>
    <p v-if="description" class="empty-state-description">{{ description }}</p>
    <button v-if="actionLabel" class="empty-state-action" @click="emit('action')">
      {{ actionLabel }}
    </button>
  </div>
</template>

<style scoped>
.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  animation: empty-state-enter 0.4s ease-out;
}

@keyframes empty-state-enter {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.empty-state-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.icon-text {
  font-size: 4rem;
  line-height: 1;
}

.empty-state-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.75rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
}

.empty-state-description {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1.05rem;
  color: var(--text-secondary);
  margin: 0 0 1.5rem;
  line-height: 1.6;
}

.empty-state-action {
  padding: 0.875rem 2rem;
  background: var(--btn-primary-bg);
  color: var(--btn-primary-text);
  border: none;
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.25, 1, 0.5, 1);
}

.empty-state-action:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-accent);
}
</style>
