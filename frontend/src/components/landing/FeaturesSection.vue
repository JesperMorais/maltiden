<script setup lang="ts">
import type { Feature } from '@/api/types/landing.types'
import FeatureCard from './FeatureCard.vue'
import WavesBackground from '@/components/vue-bits/WavesBackground.vue'

interface Props {
  sectionTitle: string
  features: Feature[]
}

defineProps<Props>()
</script>

<template>
  <section class="features-section">
    <WavesBackground
      line-color="rgba(255, 140, 100, 0.06)"
      background-color="transparent"
      :wave-speed-x="0.005"
      :wave-speed-y="0.002"
      :wave-amp-x="20"
      :wave-amp-y="10"
      :x-gap="18"
      :y-gap="48"
      :friction="0.94"
      :tension="0.003"
    />
    <div class="features-bg">
      <div class="dot-pattern"></div>
    </div>

    <div class="features-container">
      <div class="section-header">
        <span class="section-badge">Funktioner</span>
        <h2 class="section-title">{{ sectionTitle }}</h2>
      </div>

      <div class="features-grid">
        <FeatureCard
          v-for="(feature, index) in features"
          :key="feature.id"
          :icon="feature.icon"
          :title="feature.title"
          :description="feature.description"
          class="feature-item"
          :style="{ '--delay': `${index * 0.1}s` }"
        />
      </div>
    </div>
  </section>
</template>

<style scoped>
.features-section {
  position: relative;
  padding: 6rem 2rem;
  padding-bottom: 12rem;
  background: var(--bg-secondary);
  overflow: visible;
}

/* Curved bottom transition to CTA */
.features-section::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 150px;
  background: var(--bg-secondary);
  clip-path: ellipse(70% 100% at 50% 100%);
}

.features-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.dot-pattern {
  position: absolute;
  inset: 0;
  background-image: radial-gradient(var(--accent) 1px, transparent 1px);
  background-size: 40px 40px;
  opacity: 0.03;
}

.features-container {
  position: relative;
  z-index: 1;
  max-width: 1200px;
  margin: 0 auto;
}

.section-header {
  text-align: center;
  margin-bottom: 4rem;
}

.section-badge {
  display: inline-block;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  color: var(--accent);
  background: var(--bg-hover);
  padding: 0.5rem 1.25rem;
  border-radius: 100px;
  margin-bottom: 1.25rem;
}

.section-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: clamp(2rem, 4vw, 2.75rem);
  color: var(--text-primary);
  line-height: 1.2;
  margin: 0;
  max-width: 600px;
  margin-inline: auto;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 2rem;
}

.feature-item {
  animation: fade-in-up 0.6s ease-out backwards;
  animation-delay: var(--delay);
}

@keyframes fade-in-up {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Responsive */
@media (max-width: 768px) {
  .features-section {
    padding: 6rem 1.5rem;
  }

  .section-header {
    margin-bottom: 3rem;
  }

  .features-grid {
    gap: 1.5rem;
  }
}

@media (min-width: 1024px) {
  .features-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 1280px) {
  .features-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}
</style>
