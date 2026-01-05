<script setup lang="ts">
import type { MenuDay } from '@/api/types/dashboard.types'

interface Props {
  weeklyMenu: MenuDay[]
}

defineProps<Props>()

const emit = defineEmits<{
  'day-click': [day: MenuDay]
}>()
</script>

<template>
  <section class="weekly-menu">
    <header class="menu-header">
      <h3 class="menu-title">Veckans meny</h3>
      <span class="menu-subtitle">7 dagar</span>
    </header>

    <div class="days-grid">
      <button
        v-for="day in weeklyMenu"
        :key="day.date"
        class="day-card"
        :class="{
          today: day.isToday,
          skipped: day.isSkipped,
          'no-meal': !day.meal && !day.isSkipped
        }"
        @click="emit('day-click', day)"
      >
        <span class="day-name">{{ day.dayShort }}</span>
        <div class="day-meal">
          <span v-if="day.meal" class="meal-emoji">{{ day.meal.emoji || '🍽️' }}</span>
          <span v-else-if="day.isSkipped" class="skipped-icon">✕</span>
          <span v-else class="empty-icon">+</span>
        </div>
        <div v-if="day.isToday" class="today-indicator"></div>
      </button>
    </div>
  </section>
</template>

<style scoped>
.weekly-menu {
  background: var(--bg-primary);
  border-radius: 24px;
  padding: 1.5rem;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
}

.menu-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.25rem;
}

.menu-title {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 1rem;
  color: var(--text-primary);
  margin: 0;
}

.menu-subtitle {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.8rem;
  color: var(--text-secondary);
  background: var(--bg-hover);
  padding: 0.25rem 0.75rem;
  border-radius: 100px;
}

.days-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 0.5rem;
}

.day-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 0.5rem;
  background: var(--bg-card);
  border: 2px solid transparent;
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  position: relative;
}

.day-card:hover {
  transform: translateY(-4px);
  border-color: var(--border-color-hover);
  box-shadow: var(--shadow-sm);
}

.day-card.today {
  background: var(--bg-hover);
  border-color: var(--accent);
}

.day-card.today:hover {
  border-color: var(--accent);
  box-shadow: var(--shadow-md);
}

.day-card.skipped {
  opacity: 0.5;
}

.day-card.skipped:hover {
  opacity: 0.7;
}

.day-card.no-meal {
  border-style: dashed;
  border-color: var(--border-color);
}

.day-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.75rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.day-card.today .day-name {
  color: var(--accent);
}

.day-meal {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.meal-emoji {
  font-size: 1.75rem;
  line-height: 1;
  transition: transform 0.3s ease;
}

.day-card:hover .meal-emoji {
  transform: scale(1.15) rotate(5deg);
}

.skipped-icon {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 1rem;
  color: var(--text-secondary);
  opacity: 0.4;
}

.empty-icon {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 1.25rem;
  color: var(--accent);
  opacity: 0.4;
  transition: all 0.3s ease;
}

.day-card:hover .empty-icon {
  opacity: 1;
  transform: scale(1.2);
}

.today-indicator {
  position: absolute;
  bottom: 4px;
  left: 50%;
  transform: translateX(-50%);
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--accent);
}

/* Responsive */
@media (max-width: 640px) {
  .days-grid {
    gap: 0.35rem;
  }

  .day-card {
    padding: 0.5rem 0.25rem;
    border-radius: 12px;
  }

  .day-name {
    font-size: 0.65rem;
  }

  .day-meal {
    width: 32px;
    height: 32px;
  }

  .meal-emoji {
    font-size: 1.35rem;
  }
}
</style>
