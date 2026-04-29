<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { BookOpen, ArrowRight, Sandwich, Minus, Plus, UtensilsCrossed, Users } from 'lucide-vue-next'
import type { MenuDay, HouseholdMember } from '@/api/types/dashboard.types'

interface Props {
  day: MenuDay
  lunchBoxCount: number
  maxLunchBoxes: number
  members: HouseholdMember[]
  membersEatingDay: HouseholdMember[]
  isRealData: boolean
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'view-recipe': []
  'update-lunchbox': [count: number]
  'toggle-member': [memberId: string]
  close: []
}>()

const eatingIds = computed(() => new Set(props.membersEatingDay.map((m) => m.id)))

function memberInitial(name: string): string {
  return name.charAt(0).toUpperCase()
}

const basePortions = computed(() => props.membersEatingDay.length)
const totalPortions = computed(() => basePortions.value + props.lunchBoxCount)

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    emit('close')
  }
}

onMounted(() => {
  document.addEventListener('keydown', onKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div ref="popoverRef" class="popover" @click.stop>
    <div class="popover-arrow"></div>

    <!-- Meal identity -->
    <div class="popover-header">
      <span v-if="day.meal?.emoji" class="meal-emoji" aria-hidden="true">{{ day.meal.emoji }}</span>
      <UtensilsCrossed v-else :size="20" class="meal-emoji-icon" />
      <span class="meal-name">{{ day.meal?.name }}</span>
    </div>

    <!-- View recipe link -->
    <button
      class="popover-action view-recipe-btn"
      :disabled="!isRealData"
      @click="emit('view-recipe')"
    >
      <BookOpen :size="16" />
      Se recept
      <ArrowRight :size="14" class="arrow-icon" />
    </button>

    <!-- Members eating this day -->
    <div v-if="members.length > 0" class="members-section">
      <div class="members-header">
        <Users :size="16" />
        <span>Äter denna dag</span>
      </div>
      <div class="members-avatars">
        <button
          v-for="member in members"
          :key="member.id"
          class="member-avatar"
          :class="[
            member.role,
            { eating: eatingIds.has(member.id), 'not-eating': !eatingIds.has(member.id) },
          ]"
          :title="member.name"
          :aria-label="`${member.name} – ${eatingIds.has(member.id) ? 'äter' : 'äter inte'}`"
          @click="emit('toggle-member', member.id)"
        >
          {{ memberInitial(member.name) }}
        </button>
      </div>
      <span class="members-count">
        {{ membersEatingDay.length }} av {{ members.length }} äter
      </span>
    </div>

    <!-- Matlada counter -->
    <div class="lunchbox-section">
      <div class="lunchbox-header">
        <Sandwich :size="16" />
        <span>Matlådor</span>
      </div>
      <div class="lunchbox-counter">
        <button
          class="counter-btn"
          aria-label="Minska matlådor"
          :disabled="lunchBoxCount <= 0"
          @click="emit('update-lunchbox', lunchBoxCount - 1)"
        >
          <Minus :size="16" />
        </button>
        <span class="counter-value">{{ lunchBoxCount }}</span>
        <button
          class="counter-btn"
          aria-label="Öka matlådor"
          :disabled="lunchBoxCount >= maxLunchBoxes"
          @click="emit('update-lunchbox', lunchBoxCount + 1)"
        >
          <Plus :size="16" />
        </button>
      </div>
      <span v-if="lunchBoxCount > 0" class="lunchbox-hint">
        {{ totalPortions }} portioner ({{ basePortions }} + {{ lunchBoxCount }})
      </span>
    </div>
  </div>
</template>

<style scoped>
.popover {
  position: absolute;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg, 12px);
  box-shadow: var(--shadow-lg);
  padding: 1rem;
  min-width: 220px;
  z-index: 50;
}

.popover-arrow {
  position: absolute;
  bottom: -6px;
  left: 50%;
  transform: translateX(-50%);
  width: 12px;
  height: 6px;
  overflow: hidden;
}

.popover-arrow::before {
  content: '';
  position: absolute;
  bottom: 3px;
  left: 50%;
  transform: translateX(-50%) rotate(45deg);
  width: 10px;
  height: 10px;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
}

.popover-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.meal-emoji {
  font-size: 1.5rem;
  line-height: 1;
}

.meal-emoji-icon {
  color: var(--text-muted);
}

.meal-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--text-primary);
}

.view-recipe-btn {
  display: flex;
  align-items: center;
  width: 100%;
  gap: 0.5rem;
  min-height: 40px;
  padding: 0 0.75rem;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-primary);
  background: transparent;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md, 8px);
  cursor: pointer;
  transition: all 0.2s ease;
}

.view-recipe-btn:hover:not(:disabled) {
  border-color: var(--accent);
  color: var(--accent);
}

.view-recipe-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.arrow-icon {
  margin-left: auto;
}

/* ═══ Members Section ═══ */
.members-section {
  border-top: 1px solid var(--border-color);
  padding-top: 0.75rem;
  margin-top: 0.75rem;
}

.members-header {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.members-avatars {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.member-avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  border: 2px solid transparent;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 0.7rem;
  color: var(--text-on-accent);
  cursor: pointer;
  transition: all 0.2s ease;
  padding: 0;
}

.member-avatar.owner {
  background: var(--role-owner-bg);
}

.member-avatar.member {
  background: var(--role-member-bg);
}

.member-avatar.guest {
  background: var(--role-guest-avatar);
}

.member-avatar.eating {
  opacity: 1;
  border-color: var(--success);
}

.member-avatar.not-eating {
  opacity: 0.4;
  border-color: transparent;
}

.member-avatar:hover {
  transform: scale(1.1);
}

.members-count {
  display: block;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.7rem;
  color: var(--text-muted);
  text-align: center;
  margin-top: 0.35rem;
}

/* ═══ Lunchbox Section ═══ */
.lunchbox-section {
  border-top: 1px solid var(--border-color);
  padding-top: 0.75rem;
  margin-top: 0.75rem;
}

.lunchbox-header {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.lunchbox-counter {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  margin-top: 0.35rem;
}

.counter-btn {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  border: 1.5px solid var(--border-color);
  background: var(--bg-card);
  color: var(--text-primary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  transition: all 0.2s ease;
}

.counter-btn:hover:not(:disabled) {
  border-color: var(--accent);
  color: var(--accent);
}

.counter-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.counter-value {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.25rem;
  min-width: 2rem;
  text-align: center;
  color: var(--text-primary);
}

.lunchbox-hint {
  display: block;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.7rem;
  color: var(--warning-dark);
  text-align: center;
  margin-top: 0.25rem;
}

@media (max-width: 640px) {
  .popover {
    min-width: 200px;
    padding: 0.85rem;
  }
}
</style>
