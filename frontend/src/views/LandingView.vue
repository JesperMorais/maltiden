<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useLandingStore } from '@/stores/landing'
import { useThemeStore } from '@/stores/theme'
import HeroSection from '@/components/landing/HeroSection.vue'
import FeaturesSection from '@/components/landing/FeaturesSection.vue'
import CtaSection from '@/components/landing/CtaSection.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import WavesBackground from '@/components/vue-bits/WavesBackground.vue'
const landingStore = useLandingStore()
const themeStore = useThemeStore()

const waveLineColor = computed(() =>
  themeStore.isDarkMode ? 'rgba(255, 138, 125, 0.15)' : 'rgba(255, 107, 91, 0.12)',
)

onMounted(() => {
  landingStore.fetchLandingData()
})
</script>

<template>
  <main class="landing-page">
    <!-- Single wave background covering entire page (fixed = viewport-sized only, no jank) -->
    <WavesBackground
      :line-color="waveLineColor"
      background-color="transparent"
      :wave-speed-x="0.01"
      :wave-speed-y="0.004"
      :wave-amp-x="40"
      :wave-amp-y="20"
      :x-gap="12"
      :y-gap="36"
      :friction="0.92"
      :tension="0.006"
      :max-cursor-move="120"
      :style="{ position: 'fixed', zIndex: 0, pointerEvents: 'none' }"
    />

    <!-- Loading state -->
    <div v-if="landingStore.isLoading" class="loading-state" role="status" aria-label="Laddar sidan">
      <div class="loader">
        <span class="loader-icon" aria-hidden="true">🍳</span>
        <p class="loader-text">Laddar...</p>
      </div>
    </div>

    <!-- Error state -->
    <div v-else-if="landingStore.error" class="error-state">
      <div class="error-content">
        <span class="error-icon" aria-hidden="true">😅</span>
        <h2>Något gick fel</h2>
        <p>{{ landingStore.error }}</p>
        <BaseButton variant="primary" size="md" @click="landingStore.fetchLandingData(true)">
          Försök igen
        </BaseButton>
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
  position: relative;
  min-height: 100vh;
  background: linear-gradient(
    180deg,
    var(--bg-secondary) 0%,
    var(--bg-primary) 30%,
    var(--bg-secondary) 60%,
    var(--bg-secondary) 100%
  );
  overflow: clip;
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

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-20px); }
}
</style>
