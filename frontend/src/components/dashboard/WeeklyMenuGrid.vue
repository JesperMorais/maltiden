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
@import url('https://fonts.googleapis.com/css2?family=Nunito:wght@600;700;800&display=swap');

.weekly-menu {
  --coral: #ff6b5b;
  --coral-light: #ff8a7d;
  --peach: #ffb599;
  --cream: #fff8f0;
  --warm-white: #fffcf7;
  --text-dark: #3d2c29;
  --text-muted: #6b5a56;

  background: var(--warm-white);
  border-radius: 24px;
  padding: 1.5rem;
  box-shadow:
    0 4px 20px rgba(61, 44, 41, 0.05),
    0 0 0 1px rgba(255, 107, 91, 0.06);
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
  color: var(--text-dark);
  margin: 0;
}

.menu-subtitle {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.8rem;
  color: var(--text-muted);
  background: rgba(61, 44, 41, 0.05);
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
  background: white;
  border: 2px solid transparent;
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  position: relative;
}

.day-card:hover {
  transform: translateY(-4px);
  border-color: rgba(255, 107, 91, 0.2);
  box-shadow: 0 4px 12px rgba(61, 44, 41, 0.08);
}

.day-card.today {
  background: linear-gradient(165deg, rgba(255, 107, 91, 0.08) 0%, rgba(255, 181, 153, 0.08) 100%);
  border-color: var(--coral);
}

.day-card.today:hover {
  border-color: var(--coral);
  box-shadow: 0 4px 16px rgba(255, 107, 91, 0.2);
}

.day-card.skipped {
  opacity: 0.5;
}

.day-card.skipped:hover {
  opacity: 0.7;
}

.day-card.no-meal {
  border-style: dashed;
  border-color: rgba(61, 44, 41, 0.15);
}

.day-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.75rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.day-card.today .day-name {
  color: var(--coral);
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
  color: var(--text-muted);
  opacity: 0.4;
}

.empty-icon {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 1.25rem;
  color: var(--coral);
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
  background: var(--coral);
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
