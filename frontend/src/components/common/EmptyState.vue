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
        <span v-if="icon" class="icon-text">{{ icon }}</span>
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

.empty-state-action:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-accent);
}
</style>
