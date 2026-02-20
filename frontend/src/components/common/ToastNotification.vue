<script setup lang="ts">
import { onUnmounted, watch } from 'vue'
import { AnimatePresence, Motion } from 'motion-v'
import { useToast, type Toast } from '@/composables/useToast'

const { toasts, remove } = useToast()

const timers = new Map<number, ReturnType<typeof setTimeout>>()

function startTimer(toast: Toast) {
  if (timers.has(toast.id)) return
  const timer = setTimeout(() => {
    timers.delete(toast.id)
    remove(toast.id)
  }, toast.duration)
  timers.set(toast.id, timer)
}

function dismiss(id: number) {
  const timer = timers.get(id)
  if (timer) {
    clearTimeout(timer)
    timers.delete(id)
  }
  remove(id)
}

watch(
  toasts,
  (current) => {
    for (const toast of current) {
      startTimer(toast)
    }
  },
  { immediate: true },
)

onUnmounted(() => {
  for (const timer of timers.values()) {
    clearTimeout(timer)
  }
  timers.clear()
})

const icons: Record<string, string> = {
  success: '\u2713',
  error: '\u2715',
  warning: '!',
  info: 'i',
}
</script>

<template>
  <Teleport to="body">
    <div class="toast-container" aria-live="polite">
      <AnimatePresence>
        <Motion
          v-for="toast in toasts"
          :key="toast.id"
          :initial="{ opacity: 0, y: -20, scale: 0.95 }"
          :animate="{ opacity: 1, y: 0, scale: 1 }"
          :exit="{ opacity: 0, y: -20, scale: 0.95 }"
          :transition="{ duration: 0.25, ease: 'easeOut' }"
          class="toast"
          :class="`toast-${toast.type}`"
          role="alert"
        >
          <span class="toast-icon" :class="`icon-${toast.type}`">
            {{ icons[toast.type] }}
          </span>
          <span class="toast-message">{{ toast.message }}</span>
          <button class="toast-dismiss" aria-label="Stäng" @click="dismiss(toast.id)">&times;</button>
        </Motion>
      </AnimatePresence>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-container {
  position: fixed;
  top: 1rem;
  right: 1rem;
  z-index: 10000;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-width: 380px;
  width: calc(100% - 2rem);
  pointer-events: none;
}

.toast {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.75rem 1rem;
  border-radius: 12px;
  background: var(--bg-card);
  box-shadow: var(--shadow-md);
  border: 1px solid var(--border-color);
  font-family: 'Nunito', sans-serif;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-primary);
  pointer-events: auto;
}

.toast-icon {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 800;
  flex-shrink: 0;
  color: white;
}

.icon-success {
  background: var(--success);
}

.icon-error {
  background: var(--error);
}

.icon-warning {
  background: var(--warning);
  color: #92400e;
}

.icon-info {
  background: var(--accent);
}

.toast-success {
  border-left: 3px solid var(--success);
}

.toast-error {
  border-left: 3px solid var(--error);
}

.toast-warning {
  border-left: 3px solid var(--warning);
}

.toast-info {
  border-left: 3px solid var(--accent);
}

.toast-message {
  flex: 1;
  line-height: 1.4;
}

.toast-dismiss {
  background: none;
  border: none;
  color: var(--text-muted);
  font-size: 1.25rem;
  cursor: pointer;
  padding: 0;
  line-height: 1;
  flex-shrink: 0;
  transition: color 0.2s ease;
}

.toast-dismiss:hover {
  color: var(--text-primary);
}

@media (max-width: 480px) {
  .toast-container {
    top: auto;
    bottom: 1rem;
    right: 0.5rem;
    left: 0.5rem;
    max-width: none;
    width: auto;
  }
}
</style>
