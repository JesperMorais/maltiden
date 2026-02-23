<script setup lang="ts">
import { useUserStore } from '@/stores/user'

interface Props {
  /** Whether this action requires member role */
  requiresMember?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  requiresMember: true
})

const userStore = useUserStore()
const isLocked = props.requiresMember && userStore.isGuest
</script>

<template>
  <div class="locked-wrapper" :class="{ locked: isLocked }">
    <slot :is-locked="isLocked" />

    <div v-if="isLocked" class="lock-overlay">
      <div class="lock-badge">
        <span class="lock-icon">🔒</span>
        <span class="lock-text">Endast medlemmar</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.locked-wrapper {
  position: relative;
}

.locked-wrapper.locked {
  pointer-events: none;
}

.locked-wrapper.locked > :first-child:not(.lock-overlay) {
  opacity: 0.5;
  filter: grayscale(30%);
}

.lock-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: auto;
  cursor: not-allowed;
}

.lock-badge {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.35rem 0.75rem;
  background: rgba(61, 44, 41, 0.9);
  border-radius: 100px;
  opacity: 0;
  transform: scale(0.9);
  transition: all 0.2s ease;
}

.locked-wrapper:hover .lock-badge {
  opacity: 1;
  transform: scale(1);
}

.lock-icon {
  font-size: 0.85rem;
}

.lock-text {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.7rem;
  color: var(--text-on-accent);
  white-space: nowrap;
}
</style>
