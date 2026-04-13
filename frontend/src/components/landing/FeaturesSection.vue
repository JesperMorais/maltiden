<script setup lang="ts">
import type { Feature } from '@/api/types/landing.types'
import FeatureCard from './FeatureCard.vue'

interface Props {
  sectionTitle: string
  features: Feature[]
}

defineProps<Props>()
</script>

<template>
  <section class="features-section">
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
  padding: 4rem 2rem 6rem;
  background: transparent;
  overflow: visible;
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
  color: var(--accent-text);
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
  grid-template-columns: 1fr;
  gap: 1.5rem;
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

/* Tablet: 2 equal columns */
@media (min-width: 768px) {
  .features-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 2rem;
  }
}

/* Desktop: bento layout — row 1: 3/5 + 2/5, row 2: 2/5 + 3/5 */
@media (min-width: 1280px) {
  .features-grid {
    grid-template-columns: repeat(5, 1fr);
  }

  .feature-item:nth-child(1) { grid-column: span 3; }
  .feature-item:nth-child(2) { grid-column: span 2; }
  .feature-item:nth-child(3) { grid-column: span 2; }
  .feature-item:nth-child(4) { grid-column: span 3; }
}

/* Responsive */
@media (max-width: 767px) {
  .features-section {
    padding: 3rem 1.5rem;
  }

  .section-header {
    margin-bottom: 3rem;
  }
}

@media (max-width: 480px) {
  .features-section {
    padding: 3rem 1rem;
    padding-bottom: 6rem;
  }

  .section-header {
    margin-bottom: 2rem;
  }
}
</style>
