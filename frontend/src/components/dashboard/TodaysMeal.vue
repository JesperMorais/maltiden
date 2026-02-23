<script setup lang="ts">
import type { Meal } from '@/api/types/dashboard.types'

interface Props {
  meal: Meal | null
  isDayOff?: boolean
}

defineProps<Props>()

const emit = defineEmits<{
  click: []
}>()
</script>

<template>
  <article class="todays-meal" @click="emit('click')">
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
          <div class="day-off-icon">😌</div>
          <h2 class="empty-title">Ledig dag</h2>
          <p class="empty-text">Ingen matlagning planerad idag</p>
        </div>
      </template>

      <!-- Has meal -->
      <template v-else-if="meal">
        <div class="meal-emoji">{{ meal.emoji || '🍽️' }}</div>
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
            <span class="plate">🍽️</span>
            <span class="question">?</span>
          </div>
          <h2 class="empty-title">Ingen måltid planerad</h2>
          <p class="empty-text">Klicka för att lägga till något gott!</p>
        </div>
      </template>
    </div>

    <!-- Decorative food items -->
    <div class="floating-foods" v-if="meal">
      <span class="food food-1">🥬</span>
      <span class="food food-2">🧄</span>
      <span class="food food-3">🍅</span>
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
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(0.8); }
}

.meal-label span:last-child {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.85rem;
  color: var(--accent);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.meal-emoji {
  font-size: 5rem;
  line-height: 1;
  margin-bottom: 1rem;
  animation: float 3s ease-in-out infinite;
  filter: drop-shadow(0 8px 16px rgba(61, 44, 41, 0.1));
}

@keyframes float {
  0%, 100% { transform: translateY(0) rotate(-2deg); }
  50% { transform: translateY(-8px) rotate(2deg); }
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
  background: var(--accent);
  color: var(--text-on-accent);
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.95rem;
  transition: all 0.3s ease;
  box-shadow: var(--shadow-accent);
}

.todays-meal:hover .meal-action {
  background: var(--accent-light);
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
  display: inline-block;
  margin-bottom: 1.5rem;
}

.empty-icon .plate {
  font-size: 4rem;
  opacity: 0.4;
}

.empty-icon .question {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 2rem;
  color: var(--accent);
  animation: bounce 2s ease-in-out infinite;
}

@keyframes bounce {
  0%, 100% { transform: translate(-50%, -50%); }
  50% { transform: translate(-50%, -60%); }
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
  font-size: 4rem;
  line-height: 1;
  margin-bottom: 1rem;
  animation: float 3s ease-in-out infinite;
}

/* Floating foods */
.floating-foods {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 1;
}

.food {
  position: absolute;
  font-size: 1.5rem;
  opacity: 0.6;
  animation: float-food 4s ease-in-out infinite;
}

.food-1 {
  top: 15%;
  left: 10%;
  animation-delay: 0s;
}

.food-2 {
  bottom: 20%;
  right: 15%;
  animation-delay: 1s;
  font-size: 1.25rem;
}

.food-3 {
  top: 30%;
  right: 10%;
  animation-delay: 0.5s;
}

@keyframes float-food {
  0%, 100% { transform: translateY(0) rotate(0deg); }
  50% { transform: translateY(-10px) rotate(10deg); }
}

/* Responsive */
@media (max-width: 768px) {
  .todays-meal {
    padding: 2rem 1.5rem;
    min-height: 280px;
    border-radius: 24px;
  }

  .meal-emoji {
    font-size: 4rem;
  }

  .floating-foods {
    display: none;
  }
}
</style>
