<script setup lang="ts">
import type { Meal } from '@/api/types/dashboard.types'
import { UtensilsCrossed, Coffee, HelpCircle } from 'lucide-vue-next'

interface Props {
  meal: Meal | null
  isDayOff?: boolean
  extraPortions?: number
}

withDefaults(defineProps<Props>(), {
  extraPortions: 0,
})

const emit = defineEmits<{
  click: []
}>()
</script>

<template>
  <article
    class="todays-meal"
    role="button"
    tabindex="0"
    @click="emit('click')"
    @keydown.enter="emit('click')"
    @keydown.space.prevent="emit('click')"
  >
    <!-- Background decorations -->
    <div class="meal-bg">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="shimmer"></div>
    </div>

    <div class="meal-content">
      <!-- Label -->
      <div class="meal-label">
        <span class="label-dot"></span>
        <span>Idag</span>
      </div>

      <!-- Day off state -->
      <template v-if="isDayOff">
        <div class="empty-state day-off-state">
          <div class="day-off-icon">
            <Coffee :size="32" :stroke-width="1.75" />
          </div>
          <h2 class="empty-title">Ledig dag</h2>
          <p class="empty-text">Ingen matlagning planerad idag</p>
        </div>
      </template>

      <!-- Has meal -->
      <template v-else-if="meal">
        <div class="meal-emoji">
          <UtensilsCrossed :size="32" :stroke-width="1.75" />
        </div>
        <h2 class="meal-name">{{ meal.name }}</h2>
        <p class="meal-portions">{{ meal.portions }} portioner</p>
        <div class="meal-action">
          <span>Se recept</span>
          <span class="arrow">→</span>
        </div>
      </template>

      <!-- Empty state -->
      <template v-else>
        <div class="empty-state">
          <div class="empty-icon">
            <UtensilsCrossed :size="28" :stroke-width="1.75" class="plate-icon" />
            <HelpCircle :size="16" :stroke-width="2" class="question-icon" />
          </div>
          <h2 class="empty-title">Ingen måltid planerad</h2>
          <p class="empty-text">Klicka för att lägga till något gott!</p>
        </div>
      </template>
    </div>
  </article>
</template>

<style scoped>
.todays-meal {
  position: relative;
  background: var(--bg-primary);
  border-radius: 32px;
  padding: 2.5rem;
  min-height: 320px;
  cursor: pointer;
  overflow: hidden;
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
  box-shadow: var(--shadow-md);
  border: 1px solid var(--border-color);
}

.todays-meal:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
  border-color: var(--border-color-hover);
}

/* Background decorations */
.meal-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
  border-radius: inherit;
}

.blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(60px);
  opacity: 0.5;
}

.blob-1 {
  width: 300px;
  height: 300px;
  background: var(--accent-gradient);
  top: -100px;
  right: -50px;
}

.blob-2 {
  width: 200px;
  height: 200px;
  background: var(--warning);
  bottom: -50px;
  left: -50px;
  opacity: 0.3;
}

.shimmer {
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(255, 255, 255, 0.12) 50%,
    transparent 100%
  );
  animation: shimmer 8s ease-in-out infinite;
}

@keyframes shimmer {
  0%, 70%, 100% { left: -100%; }
  30% { left: 100%; }
}

/* Content */
.meal-content {
  position: relative;
  z-index: 2;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 240px;
}

.meal-label {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 1rem;
  background: var(--bg-hover);
  border-radius: 100px;
  margin-bottom: 1.5rem;
}

.label-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent);
}

.meal-label span:last-child {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.85rem;
  color: var(--accent-text);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.meal-emoji {
  margin-bottom: 1rem;
  color: var(--text-primary);
}

.meal-name {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: clamp(1.75rem, 4vw, 2.5rem);
  color: var(--text-primary);
  margin: 0 0 0.5rem;
  line-height: 1.2;
}

.meal-portions {
  font-family: 'Nunito', sans-serif;
  font-size: 1rem;
  color: var(--text-secondary);
  margin: 0 0 1.5rem;
}

.meal-action {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  background: var(--accent-text);
  color: var(--text-on-accent);
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.95rem;
  transition: all 0.3s ease;
  box-shadow: var(--shadow-accent);
}

.todays-meal:hover .meal-action {
  background: var(--accent);
  box-shadow: var(--shadow-accent);
}

.arrow {
  transition: transform 0.3s ease;
}

.todays-meal:hover .arrow {
  transform: translateX(4px);
}

/* Empty state */
.empty-state {
  text-align: center;
}

.empty-icon {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.5rem;
}

.empty-icon .plate-icon {
  color: var(--text-secondary);
  opacity: 0.4;
}

.empty-icon .question-icon {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: var(--accent);
}

.empty-title {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.5rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
}

.empty-text {
  font-family: 'Nunito', sans-serif;
  font-size: 1rem;
  color: var(--text-secondary);
  margin: 0;
}

/* Day off state */
.day-off-icon {
  margin-bottom: 1rem;
  color: var(--text-secondary);
}

/* Responsive */
@media (max-width: 768px) {
  .todays-meal {
    padding: 1rem;
    min-height: 0;
    border-radius: 20px;
  }

  .meal-bg {
    display: none;
  }

  .meal-content {
    position: static;            /* Let .meal-label anchor to .todays-meal */
    display: grid;
    grid-template-columns: auto 1fr;
    align-items: center;
    text-align: left;
    min-height: 0;
    gap: 0.15rem 1rem;
  }

  .meal-label {
    position: absolute;
    top: 0.75rem;
    right: 0.75rem;
    margin-bottom: 0;
    padding: 0.25rem 0.6rem;
    font-size: 0.7rem;
  }

  .meal-emoji {
    grid-column: 1;
    grid-row: 1 / -1;
    align-self: center;
    margin-bottom: 0;
  }

  .meal-name {
    grid-column: 2;
    font-size: 1.25rem;
    margin: 0;
  }

  .meal-portions {
    grid-column: 2;
    font-size: 0.85rem;
    margin: 0 0 0.25rem;
  }

  .meal-action {
    grid-column: 2;
    justify-self: start;
    padding: 0.5rem 1rem;
    font-size: 0.85rem;
  }

  /* Empty/day-off states use full width with horizontal layout */
  .empty-state {
    grid-column: 1 / -1;
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: 1rem;
  }

  .empty-icon {
    margin-bottom: 0;
  }

  .day-off-icon {
    margin-bottom: 0;
  }
}
</style>
