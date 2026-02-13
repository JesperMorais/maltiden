<script setup lang="ts">
import { RouterLink } from 'vue-router'
import BaseButton from '@/components/common/BaseButton.vue'
import WavesBackground from '@/components/vue-bits/WavesBackground.vue'

interface Props {
  title: string
  description: string
  primaryButton: {
    text: string
    link: string
  }
  secondaryButton?: {
    text: string
    link: string
  }
}

defineProps<Props>()
</script>

<template>
  <section class="cta-section">
    <!-- Interactive wave background -->
    <WavesBackground
      line-color="rgba(255, 180, 130, 0.1)"
      background-color="transparent"
      :wave-speed-x="0.006"
      :wave-speed-y="0.003"
      :wave-amp-x="25"
      :wave-amp-y="12"
      :x-gap="16"
      :y-gap="44"
      :friction="0.93"
      :tension="0.004"
    />

    <!-- Background elements -->
    <div class="cta-bg">
      <div class="gradient-orb orb-1"></div>
      <div class="gradient-orb orb-2"></div>
      <div class="sparkles">
        <span class="sparkle" v-for="n in 6" :key="n" :style="{ '--i': n }">✦</span>
      </div>
    </div>

    <div class="cta-container">
      <div class="cta-content">
        <div class="cta-icon">
          <span>🍳</span>
        </div>

        <h2 class="cta-title">{{ title }}</h2>
        <p class="cta-description">{{ description }}</p>

        <div class="cta-actions">
          <RouterLink v-prefetch:path="primaryButton.link" :to="primaryButton.link" class="cta-link">
            <BaseButton variant="primary" size="lg">
              {{ primaryButton.text }}
              <span class="btn-icon">→</span>
            </BaseButton>
          </RouterLink>

          <RouterLink
            v-if="secondaryButton"
            v-prefetch:path="secondaryButton.link"
            :to="secondaryButton.link"
            class="cta-link"
          >
            <BaseButton variant="outline" size="lg">
              {{ secondaryButton.text }}
            </BaseButton>
          </RouterLink>
        </div>

        <p class="cta-note">
          <span class="note-icon">🔒</span>
          Gratis att börja. Inga kreditkort krävs.
        </p>
      </div>

      <!-- Decorative food items -->
      <div class="food-decoration">
        <span class="food-item food-1">🥗</span>
        <span class="food-item food-2">🍝</span>
        <span class="food-item food-3">🥘</span>
      </div>
    </div>
  </section>
</template>

<style scoped>
.cta-section {
  position: relative;
  padding: 6rem 2rem;
  padding-top: 4rem;
  padding-bottom: 6rem;
  background: var(--bg-secondary);
  overflow: hidden;
}

.cta-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.gradient-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(60px);
}

.orb-1 {
  width: 500px;
  height: 500px;
  background: linear-gradient(135deg, var(--peach) 0%, var(--accent-light) 100%);
  top: -150px;
  left: -150px;
  opacity: 0.4;
  animation: float-slow 15s ease-in-out infinite;
}

.orb-2 {
  width: 400px;
  height: 400px;
  background: linear-gradient(135deg, var(--yellow-soft) 0%, var(--peach) 100%);
  bottom: -100px;
  right: -100px;
  opacity: 0.3;
  animation: float-slow 20s ease-in-out infinite reverse;
}

.sparkles {
  position: absolute;
  inset: 0;
}

.sparkle {
  position: absolute;
  font-size: 1rem;
  color: var(--accent);
  opacity: 0.4;
  animation: twinkle 3s ease-in-out infinite;
  animation-delay: calc(var(--i) * 0.5s);
}

.sparkle:nth-child(1) { top: 10%; left: 15%; }
.sparkle:nth-child(2) { top: 20%; right: 20%; font-size: 0.8rem; }
.sparkle:nth-child(3) { bottom: 30%; left: 10%; }
.sparkle:nth-child(4) { top: 40%; right: 10%; font-size: 1.2rem; }
.sparkle:nth-child(5) { bottom: 15%; right: 25%; font-size: 0.7rem; }
.sparkle:nth-child(6) { top: 60%; left: 20%; }

.cta-container {
  position: relative;
  z-index: 1;
  max-width: 800px;
  margin: 0 auto;
}

.cta-content {
  text-align: center;
  background: var(--bg-primary);
  border-radius: 32px;
  padding: 4rem 3rem;
  box-shadow: var(--shadow-lg);
  border: 1px solid var(--border-color);
}

.cta-icon {
  width: 80px;
  height: 80px;
  margin: 0 auto 1.5rem;
  background: linear-gradient(135deg, var(--peach) 0%, var(--accent-light) 100%);
  border-radius: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2.5rem;
  transform: rotate(-6deg);
  box-shadow: var(--shadow-accent);
  animation: wiggle 4s ease-in-out infinite;
}

.cta-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: clamp(1.75rem, 4vw, 2.5rem);
  color: var(--text-primary);
  margin: 0 0 1rem;
  line-height: 1.2;
}

.cta-description {
  font-family: 'Nunito', sans-serif;
  font-size: 1.15rem;
  color: var(--text-secondary);
  line-height: 1.7;
  margin: 0 0 2rem;
  max-width: 500px;
  margin-inline: auto;
}

.cta-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.cta-link {
  text-decoration: none;
}

.btn-icon {
  display: inline-block;
  transition: transform 0.3s ease;
}

.cta-link:hover .btn-icon {
  transform: translateX(4px);
}

.cta-note {
  font-family: 'Nunito', sans-serif;
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.note-icon {
  font-size: 1rem;
}

/* Food decoration */
.food-decoration {
  position: absolute;
  inset: 0;
  pointer-events: none;
  display: none;
}

@media (min-width: 1024px) {
  .food-decoration {
    display: block;
  }
}

.food-item {
  position: absolute;
  font-size: 3rem;
  filter: drop-shadow(var(--shadow-sm));
  animation: float 5s ease-in-out infinite;
}

.food-1 {
  top: 10%;
  left: -60px;
  animation-delay: 0s;
}

.food-2 {
  bottom: 20%;
  left: -40px;
  font-size: 2.5rem;
  animation-delay: 1s;
}

.food-3 {
  top: 30%;
  right: -50px;
  animation-delay: 0.5s;
}

/* Animations */
@keyframes float-slow {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(20px, 15px); }
}

@keyframes twinkle {
  0%, 100% { opacity: 0.3; transform: scale(1); }
  50% { opacity: 0.7; transform: scale(1.3); }
}

@keyframes wiggle {
  0%, 100% { transform: rotate(-6deg); }
  50% { transform: rotate(6deg); }
}

@keyframes float {
  0%, 100% { transform: translateY(0) rotate(0deg); }
  50% { transform: translateY(-15px) rotate(5deg); }
}

/* Responsive */
@media (max-width: 768px) {
  .cta-section {
    padding: 6rem 1.5rem;
  }

  .cta-content {
    padding: 3rem 2rem;
    border-radius: 24px;
  }

  .cta-actions {
    flex-direction: column;
  }
}
</style>
