<script setup lang="ts">
import { onMounted } from 'vue'
import { useLandingStore } from '@/stores/landing'
import HeroSection from '@/components/landing/HeroSection.vue'
import FeaturesSection from '@/components/landing/FeaturesSection.vue'
import CtaSection from '@/components/landing/CtaSection.vue'
const landingStore = useLandingStore()

onMounted(() => {
  landingStore.fetchLandingData()
})
</script>

<template>
  <main class="landing-page">
    <!-- Loading state -->
    <div v-if="landingStore.isLoading" class="loading-state">
      <div class="loader">
        <span class="loader-icon">🍳</span>
        <p class="loader-text">Laddar...</p>
      </div>
    </div>

    <!-- Error state -->
    <div v-else-if="landingStore.error" class="error-state">
      <div class="error-content">
        <span class="error-icon">😅</span>
        <h2>Något gick fel</h2>
        <p>{{ landingStore.error }}</p>
        <button class="retry-btn" @click="landingStore.fetchLandingData(true)">
          Försök igen
        </button>
      </div>
    </div>

    <!-- Content -->
    <template v-else-if="landingStore.landingData">
      <HeroSection
        v-if="landingStore.hero"
        :title="landingStore.hero.title"
        :subtitle="landingStore.hero.subtitle"
        :cta-button-text="landingStore.hero.ctaButtonText"
        :cta-button-link="landingStore.hero.ctaButtonLink"
      />

      <FeaturesSection
        v-if="landingStore.features"
        :section-title="landingStore.features.sectionTitle"
        :features="landingStore.sortedFeatures"
      />

      <CtaSection
        v-if="landingStore.cta"
        :title="landingStore.cta.title"
        :description="landingStore.cta.description"
        :primary-button="landingStore.cta.primaryButton"
        :secondary-button="landingStore.cta.secondaryButton"
      />
    </template>
  </main>
</template>

<style scoped>
.landing-page {
  min-height: 100vh;
  background: var(--bg-secondary);
  overflow-x: hidden;
}

/* Loading state */
.loading-state {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(165deg, var(--bg-secondary) 0%, var(--bg-primary) 100%);
}

.loader {
  text-align: center;
}

.loader-icon {
  font-size: 4rem;
  display: block;
  animation: bounce 1s ease-in-out infinite;
}

.loader-text {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1.25rem;
  color: var(--text-secondary);
  margin-top: 1rem;
}

/* Error state */
.error-state {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(165deg, var(--bg-secondary) 0%, var(--bg-primary) 100%);
  padding: 2rem;
}

.error-content {
  text-align: center;
  max-width: 400px;
}

.error-icon {
  font-size: 4rem;
  display: block;
  margin-bottom: 1rem;
}

.error-content h2 {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.75rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
}

.error-content p {
  font-family: 'Nunito', sans-serif;
  color: var(--text-secondary);
  margin: 0 0 1.5rem;
}

.retry-btn {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1rem;
  color: white;
  background: var(--accent);
  border: none;
  border-radius: 100px;
  padding: 0.85em 2em;
  cursor: pointer;
  transition: all 0.3s ease;
}

.retry-btn:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-accent);
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-20px); }
}
</style>
