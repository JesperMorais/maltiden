import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

export type DayIndex = 0 | 1 | 2 | 3 | 4 | 5 | 6

const STORAGE_KEY = 'maltiden_active_days'

export const ALL_DAYS: DayIndex[] = [0, 1, 2, 3, 4, 5, 6]
export const WEEKDAYS: DayIndex[] = [0, 1, 2, 3, 4]

export const DAY_LABELS: { short: string; full: string }[] = [
  { short: 'Mån', full: 'Måndag' },
  { short: 'Tis', full: 'Tisdag' },
  { short: 'Ons', full: 'Onsdag' },
  { short: 'Tor', full: 'Torsdag' },
  { short: 'Fre', full: 'Fredag' },
  { short: 'Lör', full: 'Lördag' },
  { short: 'Sön', full: 'Söndag' }
]

export const usePlanningPreferencesStore = defineStore('planningPreferences', () => {
  const activeDays = ref<DayIndex[]>([...ALL_DAYS])

  const activeDayCount = computed(() => activeDays.value.length)

  const todayDayIndex = computed<DayIndex>(() => {
    // JS getDay(): 0=Sun,1=Mon..6=Sat -> our DayIndex: 0=Mon..6=Sun
    return ((new Date().getDay() + 6) % 7) as DayIndex
  })

  const isTodayActive = computed(() => activeDays.value.includes(todayDayIndex.value))

  function isDayActive(day: DayIndex): boolean {
    return activeDays.value.includes(day)
  }

  function toggleDay(day: DayIndex) {
    if (activeDays.value.includes(day)) {
      if (activeDays.value.length <= 1) return
      activeDays.value = activeDays.value.filter((d) => d !== day)
    } else {
      activeDays.value = [...activeDays.value, day].sort()
    }
    persist()
  }

  function setPreset(days: DayIndex[]) {
    activeDays.value = [...days]
    persist()
  }

  function initPreferences() {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved) {
      try {
        const parsed: number[] = JSON.parse(saved)
        const valid = parsed.filter((d): d is DayIndex => d >= 0 && d <= 6)
        if (valid.length > 0) {
          activeDays.value = valid
        }
      } catch {
        // ignore invalid data
      }
    }
  }

  function persist() {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(activeDays.value))
  }

  return {
    activeDays,
    activeDayCount,
    todayDayIndex,
    isTodayActive,
    isDayActive,
    toggleDay,
    setPreset,
    initPreferences
  }
})
