<script setup lang="ts">
interface Props {
  /** Progress value 0-100 */
  progress: number
  /** Whether the bar is visible */
  active?: boolean
}

withDefaults(defineProps<Props>(), {
  active: true,
})
</script>

<template>
  <Transition name="progress-bar">
    <div v-if="active" class="progress-bar-container">
      <div class="progress-bar-track">
        <div
          class="progress-bar-fill"
          :style="{ width: `${Math.min(progress, 100)}%` }"
        />
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.progress-bar-container {
  width: 100%;
  padding: 0 1rem;
}

.progress-bar-track {
  width: 100%;
  height: 6px;
  background: var(--border-color);
  border-radius: 100px;
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  background: var(--accent-gradient, var(--accent));
  border-radius: 100px;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
}

.progress-bar-fill::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(255, 255, 255, 0.3) 50%,
    transparent 100%
  );
  animation: shimmer 1.5s ease-in-out infinite;
}

@keyframes shimmer {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(100%);
  }
}

/* Transition */
.progress-bar-enter-active {
  transition: all 0.3s ease;
}

.progress-bar-leave-active {
  transition: all 0.2s ease;
}

.progress-bar-enter-from,
.progress-bar-leave-to {
  opacity: 0;
  transform: scaleY(0);
}
</style>
