<script setup lang="ts">
import { useUserStore } from '@/stores/user'
import LockedAction from '@/components/common/LockedAction.vue'

const userStore = useUserStore()

const emit = defineEmits<{
  'generate-menu': []
  'add-recipe': []
  'invite-member': []
}>()

const actions = [
  {
    id: 'generate',
    icon: '🔄',
    label: 'Generera meny',
    description: 'Skapa ny veckomeny',
    event: 'generate-menu' as const,
    requiresMember: true
  },
  {
    id: 'add',
    icon: '➕',
    label: 'Lägg till recept',
    description: 'Spara nytt recept',
    event: 'add-recipe' as const,
    requiresMember: true
  },
  {
    id: 'invite',
    icon: '👥',
    label: 'Bjud in',
    description: 'Dela med familjen',
    event: 'invite-member' as const,
    requiresMember: true
  }
]

function handleAction(eventName: 'generate-menu' | 'add-recipe' | 'invite-member') {
  if (userStore.isMember) {
    emit(eventName)
  }
}
</script>

<template>
  <section class="quick-actions">
    <h3 class="widget-title">Snabbåtgärder</h3>

    <div class="actions-list">
      <LockedAction
        v-for="action in actions"
        :key="action.id"
        :requires-member="action.requiresMember"
      >
        <button
          class="action-button"
          @click="handleAction(action.event)"
        >
          <span class="action-icon">{{ action.icon }}</span>
          <div class="action-text">
            <span class="action-label">{{ action.label }}</span>
            <span class="action-desc">{{ action.description }}</span>
          </div>
        </button>
      </LockedAction>
    </div>
  </section>
</template>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Nunito:wght@600;700;800&display=swap');

.quick-actions {
  background: var(--bg-primary);
  border-radius: 20px;
  padding: 1.25rem;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
}

.widget-title {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 0.9rem;
  color: var(--text-primary);
  margin: 0 0 1rem;
}

.actions-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.action-button {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
  padding: 0.75rem;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 14px;
  cursor: pointer;
  text-align: left;
  transition: all 0.3s ease;
}

.action-button:hover {
  border-color: var(--accent);
  box-shadow: var(--shadow-sm);
  transform: translateX(4px);
}

.action-icon {
  font-size: 1.5rem;
  flex-shrink: 0;
}

.action-text {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
}

.action-label {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--text-primary);
}

.action-desc {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.75rem;
  color: var(--text-secondary);
}
</style>
