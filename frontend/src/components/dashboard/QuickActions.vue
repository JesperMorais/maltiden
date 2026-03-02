<script setup lang="ts">
import { type Component } from 'vue'
import { useUserStore } from '@/stores/user'
import LockedAction from '@/components/common/LockedAction.vue'
import SpotlightCard from '@/components/vue-bits/SpotlightCard.vue'
import ClickSpark from '@/components/vue-bits/ClickSpark.vue'
import { RefreshCw, BookOpen, Users, FileText } from 'lucide-vue-next'

const userStore = useUserStore()

const emit = defineEmits<{
  (e: 'generate-menu' | 'view-recipes' | 'invite-member' | 'parse-recipe'): void
}>()

const actions: {
  id: string
  icon: Component
  label: string
  description: string
  event: 'generate-menu' | 'view-recipes' | 'invite-member' | 'parse-recipe'
  requiresMember: boolean
}[] = [
  {
    id: 'generate',
    icon: RefreshCw,
    label: 'Generera meny',
    description: 'Skapa ny veckomeny',
    event: 'generate-menu',
    requiresMember: true
  },
  {
    id: 'recipes',
    icon: BookOpen,
    label: 'Recept',
    description: 'Hantera dina recept',
    event: 'view-recipes',
    requiresMember: true
  },
  {
    id: 'invite',
    icon: Users,
    label: 'Bjud in',
    description: 'Dela med familjen',
    event: 'invite-member',
    requiresMember: true
  },
  {
    id: 'parse',
    icon: FileText,
    label: 'Tolka recept',
    description: 'Klistra in & tolka',
    event: 'parse-recipe',
    requiresMember: true
  }
]

function handleAction(eventName: 'generate-menu' | 'view-recipes' | 'invite-member' | 'parse-recipe') {
  if (userStore.isMember) {
    emit(eventName)
  }
}
</script>

<template>
  <SpotlightCard
    spotlight-color="rgba(255, 107, 91, 0.12)"
    class-name="quick-actions-spotlight"
  >
    <section class="quick-actions">
      <h3 class="widget-title">Snabbåtgärder</h3>

      <div class="actions-list">
        <LockedAction
          v-for="action in actions"
          :key="action.id"
          :requires-member="action.requiresMember"
        >
          <ClickSpark
            spark-color="#ff6b5b"
            :spark-radius="25"
            :spark-count="6"
            :duration="500"
          >
            <button
              class="action-button"
              @click="handleAction(action.event)"
            >
              <span class="action-icon"><component :is="action.icon" :size="24" /></span>
              <div class="action-text">
                <span class="action-label">{{ action.label }}</span>
                <span class="action-desc">{{ action.description }}</span>
              </div>
            </button>
          </ClickSpark>
        </LockedAction>
      </div>
    </section>
  </SpotlightCard>
</template>

<style scoped>
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
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  flex-shrink: 0;
  color: var(--accent);
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

@media (max-width: 768px) {
  .actions-list {
    flex-direction: row;
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
    gap: 0.5rem;
    padding-bottom: 0.25rem;
  }

  .action-button {
    flex-direction: column;
    align-items: center;
    text-align: center;
    min-width: 72px;
    padding: 0.6rem 0.5rem;
    gap: 0.35rem;
  }

  .action-button:hover {
    transform: none;
  }

  .action-desc {
    display: none;
  }

  .action-label {
    font-size: 0.7rem;
    white-space: nowrap;
  }
}
</style>
