<script setup lang="ts">
import { computed } from 'vue'
import {
  usePlanningPreferencesStore,
  DAY_LABELS,
  ALL_DAYS,
  WEEKDAYS,
  type DayIndex
} from '@/stores/planningPreferences'

const store = usePlanningPreferencesStore()

defineEmits<{
  close: []
}>()

const isAllDays = computed(() => store.activeDayCount === 7)
const isWeekdays = computed(() => {
  if (store.activeDayCount !== 5) return false
  return WEEKDAYS.every((d) => store.isDayActive(d))
})
</script>

<template>
  <div class="day-picker-popover">
    <div class="day-chips">
      <button
        v-for="(label, index) in DAY_LABELS"
        :key="index"
        class="day-chip"
        :class="{ active: store.isDayActive(index as DayIndex) }"
        @click="store.toggleDay(index as DayIndex)"
      >
        {{ label.short }}
      </button>
    </div>

    <div class="presets">
      <button
        class="preset-btn"
        :class="{ active: isAllDays }"
        @click="store.setPreset(ALL_DAYS)"
      >
        Hela veckan
      </button>
      <button
        class="preset-btn"
        :class="{ active: isWeekdays }"
        @click="store.setPreset(WEEKDAYS)"
      >
        Vardagar
      </button>
    </div>
  </div>
</template>

<style scoped>
.day-picker-popover {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  background: var(--bg-card);
  border-radius: 16px;
  box-shadow: var(--shadow-lg);
  border: 1px solid var(--border-color);
  padding: 1rem;
  z-index: 10;
  min-width: 300px;
}

.day-chips {
  display: flex;
  gap: 0.4rem;
  justify-content: center;
  margin-bottom: 0.75rem;
}

.day-chip {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 2px solid var(--border-color);
  background: transparent;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.75rem;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
}

.day-chip:hover {
  border-color: var(--accent);
}

.day-chip.active {
  background: var(--btn-primary-bg);
  border-color: var(--btn-primary-bg);
  color: var(--btn-primary-text);
}

.presets {
  display: flex;
  gap: 0.5rem;
  justify-content: center;
}

.preset-btn {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.75rem;
  padding: 0.35rem 0.85rem;
  border-radius: 100px;
  border: 1.5px solid var(--border-color);
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.preset-btn:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.preset-btn.active {
  background: var(--btn-primary-bg);
  border-color: var(--btn-primary-bg);
  color: var(--btn-primary-text);
}
</style>
