<script setup lang="ts">
import { ref, computed } from 'vue'
import type { MenuDay } from '@/api/types/dashboard.types'
import { usePlanningPreferencesStore, type DayIndex } from '@/stores/planningPreferences'
import { useDashboardStore } from '@/stores/dashboard'
import { useClickOutside } from '@/composables/useClickOutside'
import { UtensilsCrossed, Coffee, Plus, Users } from 'lucide-vue-next'
import DayPickerPopover from './DayPickerPopover.vue'
import DayActionPopover from './DayActionPopover.vue'

interface Props {
  weeklyMenu: MenuDay[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'view-recipe': [day: MenuDay]
}>()

const prefsStore = usePlanningPreferencesStore()
const dashboardStore = useDashboardStore()

const isPickerOpen = ref(false)
const headerRef = ref<HTMLElement | null>(null)
const sectionRef = ref<HTMLElement | null>(null)

useClickOutside(headerRef, () => {
  isPickerOpen.value = false
})

useClickOutside(sectionRef, () => {
  dashboardStore.setSelectedDate(null)
})

// Derive the day-of-week from the date itself (Mon=0..Sun=6) rather than
// the array position — the API may return fewer or out-of-order days, and
// the user's active-day preferences are keyed to the weekday, not the slot.
const filteredMenu = computed(() =>
  props.weeklyMenu.filter((day) => {
    const jsDay = new Date(day.date + 'T12:00:00').getDay()
    const mondayFirst = ((jsDay + 6) % 7) as DayIndex
    return prefsStore.isDayActive(mondayFirst)
  }),
)

const gridColumns = computed(() => filteredMenu.value.length)

const badgeLabel = computed(() => `${prefsStore.activeDayCount} dagar`)

const selectedDay = computed(() =>
  filteredMenu.value.find((d) => d.date === dashboardStore.selectedDate) ?? null
)

function dateNumber(dateStr: string): string {
  const d = new Date(dateStr + 'T00:00:00')
  return String(d.getDate())
}

function selectDay(day: MenuDay) {
  if (!day.meal) return
  dashboardStore.setSelectedDate(dashboardStore.selectedDate === day.date ? null : day.date)
}

function handleViewRecipe() {
  if (selectedDay.value) {
    emit('view-recipe', selectedDay.value)
  }
}

function handleLunchboxUpdate(count: number) {
  if (!selectedDay.value?.meal) return
  dashboardStore.setDayLunchBoxCount(selectedDay.value.date, count)
  dashboardStore.updateDayServings(
    selectedDay.value.date,
    dashboardStore.getMembersEatingDay(selectedDay.value.date).length,
    count,
  )
}
</script>

<template>
  <section ref="sectionRef" class="weekly-menu">
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
      <div
        v-for="day in filteredMenu"
        :key="day.date"
        role="button"
        tabindex="0"
        :aria-label="`${day.dayShort}: ${day.meal?.name ?? (day.isSkipped ? 'Ledig dag' : 'Inte planerad')}`"
        class="day-card"
        :class="{
          today: day.isToday,
          skipped: day.isSkipped,
          'no-meal': !day.meal && !day.isSkipped,
          selected: dashboardStore.selectedDate === day.date
        }"
        @click="selectDay(day)"
        @keydown.enter="selectDay(day)"
        @keydown.space.prevent="selectDay(day)"
      >
        <!-- Day header -->
        <div class="day-header">
          <span class="day-name">{{ day.dayShort }}</span>
          <span class="day-date">{{ dateNumber(day.date) }}</span>
        </div>

        <!-- Meal visual -->
        <div class="day-visual">
          <span v-if="day.meal && day.meal.emoji" class="meal-emoji" aria-hidden="true">{{ day.meal.emoji }}</span>
          <UtensilsCrossed v-else-if="day.meal" :size="24" :stroke-width="1.75" class="meal-icon" />
          <Coffee v-else-if="day.isSkipped" :size="22" :stroke-width="1.75" class="skipped-icon" />
          <Plus v-else :size="22" :stroke-width="2" class="empty-icon" />
        </div>

        <!-- Meal info -->
        <div class="day-info">
          <span v-if="day.meal" class="meal-name">{{ day.meal.name }}</span>
          <span v-else-if="day.isSkipped" class="meal-status">Ledig dag</span>
          <span v-else class="meal-status add-hint">Planera</span>
        </div>

        <!-- Portions (base + matlådor) -->
        <div v-if="day.meal" class="day-meta">
          <Users :size="12" :stroke-width="2" />
          <span>{{ dashboardStore.getMembersEatingDay(day.date).length + dashboardStore.getDayLunchBoxCount(day.date) }}</span>
        </div>

        <!-- Today glow ring -->
        <div v-if="day.isToday" class="today-indicator"></div>

        <!-- Action popover -->
        <Transition name="popover">
          <DayActionPopover
            v-if="dashboardStore.selectedDate === day.date && day.meal"
            :day="day"
            :lunch-box-count="dashboardStore.getDayLunchBoxCount(day.date)"
            :max-lunch-boxes="dashboardStore.getMembersEatingDay(day.date).length"
            :members="dashboardStore.householdMembers"
            :members-eating-day="dashboardStore.getMembersEatingDay(day.date)"
            :is-real-data="dashboardStore.isUsingRealData"
            @view-recipe="handleViewRecipe"
            @update-lunchbox="handleLunchboxUpdate"
            @toggle-member="(memberId: string) => dashboardStore.toggleMemberDay(day.date, memberId)"
            @close="dashboardStore.setSelectedDate(null)"
          />
        </Transition>
      </div>
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

/* Dropdown animation */
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

/* Grid */
.days-grid {
  display: grid;
  gap: 0.5rem;
  overflow: visible;
}

/* TransitionGroup animations */
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

/* ═══ Day Card ═══ */
.day-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.35rem;
  padding: 0.85rem 0.5rem 0.75rem;
  background: var(--bg-card);
  border: 2px solid transparent;
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  position: relative;
  min-height: 170px;
  outline: none;
}

.day-card:hover {
  transform: translateY(-4px);
  border-color: var(--border-color-hover);
  box-shadow: var(--shadow-sm);
}

.day-card:focus-visible {
  box-shadow: 0 0 0 3px var(--accent-focus-ring);
}

/* Today state — warm ambient glow (passive) */
.day-card.today {
  background: var(--bg-hover);
  border-color: var(--accent);
  box-shadow: 0 0 0 3px var(--accent-focus-ring);
}

.day-card.today:hover {
  box-shadow: 0 0 0 3px var(--accent-focus-ring), var(--shadow-md);
}

/* Selected state — strong editing indicator (active) */
.day-card.selected {
  border-color: var(--text-primary);
  border-width: 2.5px;
  background: var(--bg-hover);
  box-shadow: inset 0 3px 0 0 var(--accent);
  transform: translateY(-2px);
}

.day-card.selected:hover {
  box-shadow: inset 0 3px 0 0 var(--accent), var(--shadow-md);
}

/* Both today AND selected */
.day-card.today.selected {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px var(--accent-focus-ring), inset 0 3px 0 0 var(--text-primary);
}

.day-card.today.selected:hover {
  box-shadow: 0 0 0 3px var(--accent-focus-ring), inset 0 3px 0 0 var(--text-primary), var(--shadow-md);
}

/* Skipped state */
.day-card.skipped {
  opacity: 0.55;
}

.day-card.skipped:hover {
  opacity: 0.75;
}

/* Empty / no meal state */
.day-card.no-meal {
  border-style: dashed;
  border-color: var(--border-color);
}

/* ═══ Day Header ═══ */
.day-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.1rem;
}

.day-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.7rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.day-card.today .day-name {
  color: var(--accent-text);
}

.day-date {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.15rem;
  color: var(--text-primary);
  line-height: 1.2;
}

.day-card.today .day-date {
  color: var(--accent-text);
}

.day-card.skipped .day-date {
  opacity: 0.6;
}

/* ═══ Meal Visual ═══ */
.day-visual {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0.15rem 0;
}

.meal-emoji {
  font-size: 2rem;
  line-height: 1;
  transition: transform 0.3s ease;
}

.day-card:hover .meal-emoji {
  transform: scale(1.15) rotate(5deg);
}

.meal-icon {
  color: var(--text-secondary);
  transition: transform 0.3s ease;
}

.day-card:hover .meal-icon {
  transform: scale(1.15) rotate(5deg);
}

.skipped-icon {
  color: var(--text-muted);
  opacity: 0.5;
}

.empty-icon {
  color: var(--accent-text);
  opacity: 0.35;
  transition: all 0.3s ease;
}

.day-card:hover .empty-icon {
  opacity: 1;
  transform: scale(1.2);
}

/* ═══ Meal Info ═══ */
.day-info {
  text-align: center;
  min-height: 2.4em;
  display: flex;
  align-items: center;
  padding: 0 0.25rem;
}

.meal-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.75rem;
  color: var(--text-primary);
  line-height: 1.3;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.meal-status {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.75rem;
  color: var(--text-muted);
}

.meal-status.add-hint {
  color: var(--accent-text);
  opacity: 0.55;
}

.day-card:hover .meal-status.add-hint {
  opacity: 1;
}

/* ═══ Meta / Portions ═══ */
.day-meta {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.65rem;
  color: var(--text-muted);
  background: var(--bg-hover);
  padding: 0.15rem 0.45rem;
  border-radius: var(--radius-full);
}

/* ═══ Today Indicator ═══ */
.today-indicator {
  position: absolute;
  bottom: 6px;
  left: 50%;
  transform: translateX(-50%);
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--accent);
}

/* ═══ Popover transition ═══ */
.popover-enter-active {
  transition: all 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.popover-leave-active {
  transition: all 0.15s ease;
}

.popover-enter-from,
.popover-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(4px) scale(0.95);
}

/* Position the popover above the card */
.day-card :deep(.popover) {
  bottom: calc(100% + 8px);
  left: 50%;
  transform: translateX(-50%);
}

/* ═══ Responsive — vertical list ═══ */
@media (max-width: 640px) {
  .weekly-menu {
    padding: 1rem;
  }

  .menu-header {
    margin-bottom: 0.75rem;
  }

  .days-grid {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    overflow-x: visible;
  }

  .day-card {
    flex-direction: row;
    align-items: center;
    gap: 0.65rem;
    padding: 0.5rem 0.75rem;
    border-radius: 12px;
    min-height: 0;
  }

  .day-card:hover {
    transform: none;
  }

  .day-card.selected {
    transform: none;
  }

  .day-card.today {
    box-shadow: none;
  }

  .day-card.today.selected {
    box-shadow: inset 0 3px 0 0 var(--text-primary);
  }

  .day-header {
    flex-direction: row;
    gap: 0.35rem;
    min-width: 52px;
  }

  .day-name {
    font-size: 0.65rem;
  }

  .day-date {
    font-size: 0.85rem;
  }

  .day-visual {
    width: 32px;
    height: 32px;
    flex-shrink: 0;
    margin: 0;
  }

  .meal-emoji {
    font-size: 1.35rem;
  }

  .day-info {
    flex: 1;
    min-width: 0;
    min-height: 0;
    text-align: left;
    padding: 0;
  }

  .meal-name {
    font-size: 0.85rem;
    -webkit-line-clamp: 1;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    display: block;
  }

  .meal-status {
    font-size: 0.8rem;
  }

  .day-meta {
    flex-shrink: 0;
  }

  .today-indicator {
    position: static;
    transform: none;
    flex-shrink: 0;
  }

  /* Popover: full-width below card on mobile */
  .day-card :deep(.popover) {
    left: 0;
    right: 0;
    transform: none;
    min-width: unset;
  }

  .popover-enter-from,
  .popover-leave-to {
    transform: translateY(4px) scale(0.95);
  }
}
</style>
