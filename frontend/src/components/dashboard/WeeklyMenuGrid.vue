<script setup lang="ts">
import { ref, computed } from 'vue'
import type { MenuDay } from '@/api/types/dashboard.types'
import { usePlanningPreferencesStore, type DayIndex } from '@/stores/planningPreferences'
import { useClickOutside } from '@/composables/useClickOutside'
import DayPickerPopover from './DayPickerPopover.vue'

interface Props {
  weeklyMenu: MenuDay[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'day-click': [day: MenuDay]
}>()

const prefsStore = usePlanningPreferencesStore()

const isPickerOpen = ref(false)
const headerRef = ref<HTMLElement | null>(null)

useClickOutside(headerRef, () => {
  isPickerOpen.value = false
})

const filteredMenu = computed(() =>
  props.weeklyMenu.filter((_, index) => prefsStore.isDayActive(index as DayIndex))
)

const gridColumns = computed(() => filteredMenu.value.length)

const badgeLabel = computed(() => `${prefsStore.activeDayCount} dagar`)
</script>

<template>
  <section class="weekly-menu">
    <header class="menu-header">
      <h3 class="menu-title">Veckans meny</h3>
      <div ref="headerRef" class="days-badge-wrapper">
        <button
          class="days-badge"
          :class="{ open: isPickerOpen }"
          :aria-expanded="isPickerOpen"
          aria-haspopup="true"
          @click="isPickerOpen = !isPickerOpen"
        >
          {{ badgeLabel }}
          <svg class="badge-chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M19 9l-7 7-7-7" />
          </svg>
        </button>

        <Transition name="dropdown">
          <DayPickerPopover v-if="isPickerOpen" @close="isPickerOpen = false" />
        </Transition>
      </div>
    </header>

    <TransitionGroup
      name="day-list"
      tag="div"
      class="days-grid"
      :style="{ gridTemplateColumns: `repeat(${gridColumns}, 1fr)` }"
    >
      <button
        v-for="day in filteredMenu"
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
    </TransitionGroup>
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

.days-badge-wrapper {
  position: relative;
}

.days-badge {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.8rem;
  color: var(--text-secondary);
  background: var(--bg-hover);
  padding: 0.25rem 0.6rem 0.25rem 0.75rem;
  border-radius: 100px;
  border: 1.5px solid transparent;
  cursor: pointer;
  transition: all 0.2s ease;
}

.days-badge:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.days-badge.open {
  border-color: var(--accent);
  color: var(--accent);
}

.badge-chevron {
  width: 14px;
  height: 14px;
  transition: transform 0.2s ease;
}

.days-badge.open .badge-chevron {
  transform: rotate(180deg);
}

/* Dropdown animation (matches DashboardHeader) */
.dropdown-enter-active {
  transition: all 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.dropdown-leave-active {
  transition: all 0.15s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-8px) scale(0.95);
}

.days-grid {
  display: grid;
  gap: 0.5rem;
  overflow: hidden;
}

/* Day list TransitionGroup animations */
.day-list-enter-active {
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.day-list-leave-active {
  transition: all 0.2s ease;
  position: absolute;
  width: 0;
  overflow: hidden;
  padding: 0;
  opacity: 0;
}

.day-list-enter-from {
  opacity: 0;
  transform: scale(0.8);
}

.day-list-leave-to {
  opacity: 0;
  transform: scale(0.8);
}

.day-list-move {
  transition: transform 0.3s ease;
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
  background-image: repeating-linear-gradient(
    -45deg,
    transparent,
    transparent 4px,
    var(--border-color) 4px,
    var(--border-color) 5px
  );
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
  color: var(--accent-text);
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
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
    scroll-snap-type: x proximity;
    padding-bottom: 0.25rem;
  }

  .day-card {
    padding: 0.6rem 0.35rem;
    border-radius: 12px;
    min-width: 52px;
    min-height: 80px;
    scroll-snap-align: start;
  }

  .day-name {
    font-size: 0.7rem;
  }

  .day-meal {
    width: 36px;
    height: 36px;
  }

  .meal-emoji {
    font-size: 1.35rem;
  }
}
</style>
