<script setup lang="ts">
import { RouterLink } from 'vue-router'
import BaseButton from '@/components/common/BaseButton.vue'
import { UtensilsCrossed, Lock } from 'lucide-vue-next'

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

    <!-- Background elements -->
    <div class="cta-bg">
      <div class="gradient-orb orb-1"></div>
      <div class="gradient-orb orb-2"></div>
    </div>

    <div class="cta-container">
      <div class="cta-content">
        <div class="cta-icon">
          <UtensilsCrossed :size="32" :stroke-width="1.75" />
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
          <Lock :size="16" :stroke-width="2" class="note-icon" />
          Gratis att börja. Inga kreditkort krävs.
        </p>
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
  background: transparent;
  overflow: visible;
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
  color: var(--accent);
  transform: rotate(-6deg);
  box-shadow: var(--shadow-accent);
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
  color: var(--text-secondary);
  flex-shrink: 0;
}

/* Animations */
@keyframes float-slow {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(20px, 15px); }
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

@media (max-width: 480px) {
  .cta-section {
    padding: 3rem 1rem;
  }

  .cta-content {
    padding: 2rem 1.25rem;
    border-radius: 20px;
  }

  .cta-actions {
    flex-direction: column;
    width: 100%;
  }

  .cta-link {
    display: block;
    width: 100%;
  }
}
</style>
