<script setup lang="ts">
import { Lock, RefreshCw } from 'lucide-vue-next'

interface Props {
  hasMenu: boolean
  isLoading: boolean
  lockedCount: number
  totalDays?: number
}

withDefaults(defineProps<Props>(), {
  totalDays: 5
})

const emit = defineEmits<{
  regenerate: []
  save: []
  cancel: []
}>()
</script>

<template>
  <div class="menu-actions">
    <div class="actions-container">
      <!-- Left: Cancel button -->
      <div class="actions-left">
        <button class="btn btn-cancel" :disabled="isLoading" @click="emit('cancel')">
          <span class="btn-text">Avbryt</span>
        </button>
      </div>

      <!-- Center: Lock status -->
      <div v-if="hasMenu" class="actions-center">
        <div class="lock-status">
          <Lock :size="14" class="lock-icon" />
          <span class="lock-text">{{ lockedCount }} av {{ totalDays }} dagar låsta</span>
        </div>
      </div>

      <!-- Right: Action buttons -->
      <div class="actions-right">
        <button
          v-if="hasMenu"
          class="btn btn-regenerate"
          :disabled="isLoading || lockedCount === totalDays"
          :title="lockedCount === totalDays ? 'Lås upp minst en dag för att generera nya recept' : ''"
          @click="emit('regenerate')"
        >
          <RefreshCw :size="16" class="btn-icon" />
          <span class="btn-text">Generera nya</span>
        </button>

        <button
          v-if="hasMenu"
          class="btn btn-save"
          :disabled="isLoading"
          @click="emit('save')"
        >
          <span class="btn-icon">✓</span>
          <span class="btn-text">Spara meny</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.menu-actions {
  position: sticky;
  bottom: 0;
  left: 0;
  right: 0;
  border-top: 1px solid var(--border-color);
  padding: 1.5rem;
  z-index: 10;
  backdrop-filter: blur(10px);
  background: color-mix(in srgb, var(--bg-primary) 95%, transparent);
}

.actions-container {
  max-width: 1200px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  gap: 1.5rem;
}

.actions-left {
  display: flex;
  justify-content: flex-start;
}

.actions-center {
  display: flex;
  justify-content: center;
}

.actions-right {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
}

/* Lock status */
.lock-status {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--text-secondary);
}

.lock-icon {
  font-size: 1rem;
  line-height: 1;
}

.lock-text {
  line-height: 1;
}

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.875rem 1.75rem;
  border: 2px solid transparent;
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  white-space: nowrap;
}

.btn:hover:not(:disabled) {
  transform: translateY(-2px);
}

.btn:active:not(:disabled) {
  transform: translateY(0);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-icon {
  font-size: 1.2rem;
  line-height: 1;
}

.btn-text {
  line-height: 1;
}

/* Cancel button */
.btn-cancel {
  background: transparent;
  color: var(--text-secondary);
  border-color: var(--border-color);
}

.btn-cancel:hover:not(:disabled) {
  border-color: var(--text-secondary);
  color: var(--text-primary);
}

/* Regenerate button */
.btn-regenerate {
  background: var(--bg-card);
  color: var(--text-primary);
  border-color: var(--border-color);
}

.btn-regenerate:hover:not(:disabled) {
  border-color: var(--accent);
  background: var(--bg-hover);
}

/* Save button */
.btn-save {
  background: linear-gradient(135deg, var(--accent) 0%, var(--accent-dark) 100%);
  color: var(--text-on-accent);
  border-color: var(--accent);
  box-shadow: var(--shadow-accent-sm);
}

.btn-save:hover:not(:disabled) {
  box-shadow: var(--shadow-accent-hover);
}

/* Responsive */
@media (max-width: 1024px) {
  .actions-container {
    grid-template-columns: auto 1fr auto;
  }

  .actions-left {
    order: 1;
  }

  .actions-center {
    order: 3;
    grid-column: 1 / -1;
    margin-top: 0.5rem;
  }

  .actions-right {
    order: 2;
  }
}

@media (max-width: 768px) {
  .menu-actions {
    padding: 1rem;
  }

  .actions-container {
    grid-template-columns: 1fr;
    gap: 1rem;
  }

  .actions-left,
  .actions-center,
  .actions-right {
    justify-content: center;
  }

  .actions-center {
    order: 1;
    margin-top: 0;
  }

  .actions-left {
    order: 3;
  }

  .actions-right {
    order: 2;
    flex-wrap: wrap;
    justify-content: center;
  }

  .btn {
    flex: 1;
    min-width: 140px;
    justify-content: center;
  }

  .btn-cancel {
    width: 100%;
  }
}
</style>
