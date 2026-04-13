<script setup lang="ts">
import { RouterLink } from 'vue-router'

interface Props {
  to: string | { name: string }
  label?: string
}

withDefaults(defineProps<Props>(), {
  label: 'Tillbaka',
})
</script>

<template>
  <RouterLink :to="to" class="back-link">
    <span class="back-arrow">&larr;</span>
    <span>{{ label }}</span>
  </RouterLink>
</template>

<style scoped>
.back-link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--text-secondary);
  text-decoration: none;
  padding: 0.5rem 1rem;
  border-radius: 100px;
  transition: all 0.3s ease;
}

.back-link:hover {
  color: var(--accent);
  background: var(--bg-hover);
}

.back-arrow {
  transition: transform 0.3s ease;
}

.back-link:hover .back-arrow {
  transform: translateX(-4px);
}
</style>
