<script setup lang="ts">
import { AnimatePresence, Motion } from 'motion-v'
import { Lock, LockOpen, RefreshCw } from 'lucide-vue-next'
import type { DraftMenuDay } from '@/stores/menuGenerator'
import type { DisplayRecipe } from '@/composables/useSlotMachine'

interface Props {
  day: DraftMenuDay
  isLocked: boolean
  isLoading?: boolean
  isRolling?: boolean
  hasLanded?: boolean
  displayRecipe?: DisplayRecipe
  showSwap?: boolean
}

withDefaults(defineProps<Props>(), {
  showSwap: false,
})

const emit = defineEmits<{
  'toggle-lock': []
  'swap': []
}>()
</script>

<template>
  <article
    class="menu-day-card"
    :class="{
      locked: isLocked,
      loading: isLoading,
      empty: !day.recipeId && !isRolling,
      rolling: isRolling,
      'just-landed': hasLanded,
    }"
  >
    <!-- Loading shimmer overlay -->
    <div v-if="isLoading" class="shimmer-overlay"></div>

    <!-- Day name header -->
    <div class="day-header">
      <h3 class="day-name">{{ day.dayName }}</h3>
    </div>

    <!-- Recipe content -->
    <div class="recipe-content">
      <!-- SLOT ROLLING STATE -->
      <template v-if="isRolling && displayRecipe">
        <div class="slot-reel" aria-hidden="true">
          <AnimatePresence mode="wait">
            <Motion
              :key="displayRecipe.emoji"
              tag="div"
              class="slot-reel-emoji"
              :initial="{ y: 24, opacity: 0 }"
              :animate="{ y: 0, opacity: 1 }"
              :exit="{ y: -24, opacity: 0 }"
              :transition="{ duration: 0.06 }"
            >
              {{ displayRecipe.emoji }}
            </Motion>
          </AnimatePresence>
          <AnimatePresence mode="wait">
            <Motion
              :key="displayRecipe.recipeName"
              tag="div"
              class="slot-reel-name"
              :initial="{ y: 14, opacity: 0 }"
              :animate="{ y: 0, opacity: 0.5 }"
              :exit="{ y: -14, opacity: 0 }"
              :transition="{ duration: 0.06 }"
            >
              {{ displayRecipe.recipeName }}
            </Motion>
          </AnimatePresence>
        </div>
        <span class="sr-only">Genererar recept...</span>
      </template>

      <!-- LANDED / FILLED STATE -->
      <template v-else-if="day.recipeId">
        <!-- Recipe emoji -->
        <div class="recipe-emoji" aria-hidden="true" :class="{ 'landing-bounce': hasLanded }">
          {{ day.emoji || '🍽️' }}
        </div>

        <!-- Recipe name -->
        <h4 class="recipe-name" :class="{ 'landing-bounce': hasLanded }">
          {{ day.recipeName }}
        </h4>

        <!-- Servings -->
        <p class="recipe-servings">{{ day.servings }} portioner</p>
      </template>

      <!-- EMPTY STATE -->
      <template v-else>
        <div class="empty-recipe">
          <div class="plus-icon">+</div>
          <p class="empty-text">Ingen måltid</p>
        </div>
      </template>
    </div>

    <!-- Lock button -->
    <button
      v-if="day.recipeId && !isRolling"
      class="lock-button"
      :class="{ locked: isLocked }"
      :aria-label="isLocked ? `Lås upp ${day.dayName}` : `Lås ${day.dayName}`"
      :disabled="isLoading"
      @click.stop="emit('toggle-lock')"
    >
      <component :is="isLocked ? Lock : LockOpen" :size="14" class="lock-icon" />
    </button>

    <!-- Swap button -->
    <button
      v-if="showSwap && day.recipeId && !isRolling"
      class="swap-button"
      :aria-label="`Byt recept för ${day.dayName}`"
      :disabled="isLoading"
      @click.stop="emit('swap')"
    >
      <RefreshCw :size="14" />
    </button>
  </article>
</template>

<style scoped>
.menu-day-card {
  position: relative;
  background: var(--bg-card);
  border: 2px solid var(--border-color);
  border-radius: 20px;
  padding: 1.5rem;
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
  cursor: default;
  overflow: hidden;
  min-height: 280px;
  display: flex;
  flex-direction: column;
}

.menu-day-card:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
}

/* Locked state */
.menu-day-card.locked {
  border-color: var(--accent);
  background: linear-gradient(
    165deg,
    var(--bg-card) 0%,
    var(--bg-hover) 100%
  );
}

/* Empty state */
.menu-day-card.empty {
  border-style: dashed;
  border-color: var(--border-color-light);
}

.menu-day-card.empty:hover {
  border-color: var(--border-color);
}

/* Loading state */
.menu-day-card.loading {
  pointer-events: none;
}

.shimmer-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(
    90deg,
    transparent 0%,
    var(--accent-bg) 50%,
    transparent 100%
  );
  animation: shimmer 1.5s ease-in-out infinite;
  z-index: 1;
}

@keyframes shimmer {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(100%);
  }
}

/* ========================================
   SLOT MACHINE ROLLING STATE
   ======================================== */

.menu-day-card.rolling {
  border-color: var(--accent);
  border-style: solid;
  box-shadow: 0 0 0 1px var(--accent-focus-ring),
    var(--shadow-sm);
}

.menu-day-card.rolling:hover {
  transform: none;
}

.slot-reel {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  height: 100%;
  min-height: 140px;
}

.slot-reel-emoji {
  font-size: 3.5rem;
  line-height: 1;
}

.slot-reel-name {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.05rem;
  color: var(--text-primary);
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}

/* ========================================
   LANDING ANIMATIONS
   ======================================== */

.menu-day-card.just-landed {
  animation: card-land 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.landing-bounce {
  animation: land-bounce 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes land-bounce {
  0% {
    transform: translateY(-10px);
    opacity: 0.7;
  }
  60% {
    transform: translateY(3px);
    opacity: 1;
  }
  100% {
    transform: translateY(0);
    opacity: 1;
  }
}

@keyframes card-land {
  0% {
    transform: translateY(-4px) scale(1.02);
  }
  60% {
    transform: translateY(1px) scale(0.99);
  }
  100% {
    transform: translateY(0) scale(1);
  }
}

/* Day header */
.day-header {
  margin-bottom: 1rem;
}

.day-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
  margin: 0;
}

/* Recipe content */
.recipe-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  gap: 0.75rem;
}

.recipe-emoji {
  font-size: 3.5rem;
  display: block;
  margin: 0.5rem auto;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

.recipe-name {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.1rem;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.3;
}

.recipe-servings {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin: 0;
}

/* Empty recipe state */
.empty-recipe {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  opacity: 0.6;
}

.plus-icon {
  font-size: 3rem;
  font-weight: 300;
  color: var(--text-secondary);
  line-height: 1;
}

.empty-text {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin: 0;
}

/* Lock button */
.lock-button {
  position: absolute;
  bottom: 1rem;
  right: 1rem;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--bg-secondary);
  border: 2px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  z-index: 2;
}

.lock-button:hover:not(:disabled) {
  transform: scale(1.15);
  border-color: var(--accent);
}

.lock-button:active:not(:disabled) {
  transform: scale(0.95);
}

.lock-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.lock-button.locked {
  background: var(--accent);
  border-color: var(--accent);
}

.lock-icon {
  font-size: 1.2rem;
  line-height: 1;
}

.lock-button.locked .lock-icon {
  filter: brightness(0) invert(1);
}

/* Swap button */
.swap-button {
  position: absolute;
  bottom: 1rem;
  right: 3.5rem;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--bg-secondary);
  border: 2px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  z-index: 2;
}

.swap-button:hover:not(:disabled) {
  transform: scale(1.15);
  border-color: var(--accent);
  color: var(--accent);
}

.swap-button:active:not(:disabled) {
  transform: scale(0.95);
}

.swap-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ========================================
   REDUCED MOTION
   ======================================== */

@media (prefers-reduced-motion: reduce) {
  .menu-day-card.rolling {
    animation: none;
  }

  .menu-day-card.just-landed {
    animation: none;
  }

  .landing-bounce {
    animation: none;
  }

  .recipe-emoji {
    animation: none;
  }

  .slot-reel-emoji,
  .slot-reel-name {
    animation: none;
  }
}

/* Responsive */
@media (max-width: 1024px) {
  .menu-day-card {
    min-height: 240px;
    padding: 1.25rem;
  }

  .recipe-emoji {
    font-size: 3rem;
  }

  .slot-reel-emoji {
    font-size: 3rem;
  }

  .recipe-name {
    font-size: 1rem;
  }
}

@media (max-width: 768px) {
  .menu-day-card {
    min-height: 220px;
    padding: 1rem;
  }

  .recipe-emoji {
    font-size: 2.5rem;
  }

  .slot-reel-emoji {
    font-size: 2.5rem;
  }

  .recipe-name {
    font-size: 0.95rem;
  }

  .lock-button {
    width: 36px;
    height: 36px;
  }

  .lock-icon {
    font-size: 1rem;
  }
}
</style>
