<script setup lang="ts">
import { RouterLink } from 'vue-router'
import BaseButton from '@/components/common/BaseButton.vue'
import { useUserStore } from '@/stores/user'
import { useThemeStore } from '@/stores/theme'

interface Props {
  title: string
  subtitle: string
  ctaButtonText: string
  ctaButtonLink: string
}

defineProps<Props>()

const userStore = useUserStore()
const themeStore = useThemeStore()
</script>

<template>
  <section class="hero">
    <!-- Navigation -->
    <nav class="hero-nav">
      <RouterLink to="/" class="nav-logo">
        <span class="logo-icon">🍽️</span>
        <span class="logo-text">Måltiden</span>
      </RouterLink>
      <div class="nav-actions">
        <RouterLink v-prefetch="'about'" to="/about" class="nav-link">Om oss</RouterLink>
        <button
          class="theme-toggle"
          @click="themeStore.toggleDarkMode()"
          :aria-label="themeStore.isDarkMode ? 'Byt till ljust läge' : 'Byt till mörkt läge'"
        >
          <span v-if="themeStore.isDarkMode">☀️</span>
          <span v-else>🌙</span>
        </button>
      </div>
    </nav>

    <!-- Decorative background elements -->
    <div class="hero-bg">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="blob blob-3"></div>
      <div class="grain"></div>
    </div>

    <!-- Floating food illustrations -->
    <div class="floating-elements">
      <span class="float-item float-1">🥕</span>
      <span class="float-item float-2">🍅</span>
      <span class="float-item float-3">🥦</span>
      <span class="float-item float-4">🧅</span>
      <span class="float-item float-5">🍋</span>
    </div>

    <div class="hero-content">
      <div class="badge">
        <span class="badge-icon">✨</span>
        <span>Smartare matplanering</span>
      </div>

      <h1 class="hero-title">{{ title }}</h1>

      <p class="hero-subtitle">{{ subtitle }}</p>

      <div class="hero-actions">
        <!-- Logged in: Go to dashboard -->
        <template v-if="userStore.isAuthenticated">
          <RouterLink v-prefetch="'dashboard'" to="/dashboard" class="cta-link">
            <BaseButton variant="primary" size="lg">
              Gå till Dashboard
              <span class="btn-arrow">→</span>
            </BaseButton>
          </RouterLink>
          <p class="logged-in-text">
            Inloggad som <strong>{{ userStore.userName }}</strong>
          </p>
        </template>

        <!-- Not logged in: Register + Login -->
        <template v-else>
          <div class="auth-buttons">
            <RouterLink v-prefetch="'register'" :to="ctaButtonLink" class="cta-link">
              <BaseButton variant="primary" size="lg">
                {{ ctaButtonText }}
                <span class="btn-arrow">→</span>
              </BaseButton>
            </RouterLink>

            <RouterLink v-prefetch="'login'" to="/login" class="login-link">
              Redan medlem? <span>Logga in</span>
            </RouterLink>
          </div>
        </template>

        <div class="trust-badges">
          <div class="trust-item">
            <span class="trust-icon">🏠</span>
            <span>1000+ hushåll</span>
          </div>
          <div class="trust-item">
            <span class="trust-icon">⭐</span>
            <span>4.9 betyg</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Decorative plate illustration -->
    <div class="hero-illustration">
      <div class="plate">
        <div class="plate-inner">
          <span class="plate-emoji">🍽️</span>
        </div>
        <div class="plate-shadow"></div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.hero {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: visible;
  padding: 2rem;
  padding-bottom: 10rem;
  background: linear-gradient(
    165deg,
    var(--bg-secondary) 0%,
    var(--bg-primary) 50%,
    var(--bg-secondary) 100%
  );
}

/* Navigation */
.hero-nav {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  z-index: 20;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem 2rem;
}

.nav-logo {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  text-decoration: none;
  transition: transform 0.3s ease;
}

.nav-logo:hover {
  transform: scale(1.05);
}

.nav-logo .logo-icon {
  font-size: 1.75rem;
}

.nav-logo .logo-text {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.35rem;
  color: var(--text-primary);
}

.nav-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.nav-link {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.95rem;
  color: var(--text-secondary);
  text-decoration: none;
  padding: 0.5rem 1.25rem;
  border-radius: 100px;
  background: var(--bg-hover);
  transition: all 0.3s ease;
}

.nav-link:hover {
  color: var(--accent);
  background: var(--border-color-hover);
}

.theme-toggle {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-hover);
  border: 1px solid var(--border-color);
  border-radius: 50%;
  cursor: pointer;
  font-size: 1.25rem;
  transition: all 0.3s ease;
}

.theme-toggle:hover {
  background: var(--border-color-hover);
  transform: scale(1.1);
}

/* Curved bottom transition */
.hero::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 150px;
  background: var(--bg-secondary);
  clip-path: ellipse(75% 100% at 50% 100%);
}

/* Background decorations */
.hero-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.6;
}

.blob-1 {
  width: 600px;
  height: 600px;
  background: linear-gradient(135deg, var(--peach) 0%, var(--accent-light) 100%);
  top: -200px;
  right: -100px;
  animation: float-slow 20s ease-in-out infinite;
}

.blob-2 {
  width: 400px;
  height: 400px;
  background: linear-gradient(135deg, var(--yellow-soft) 0%, var(--orange-soft) 100%);
  bottom: -100px;
  left: -100px;
  animation: float-slow 25s ease-in-out infinite reverse;
}

.blob-3 {
  width: 300px;
  height: 300px;
  background: var(--accent-light);
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  opacity: 0.3;
  animation: pulse 8s ease-in-out infinite;
}

.grain {
  position: absolute;
  inset: 0;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 400 400' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)'/%3E%3C/svg%3E");
  opacity: 0.03;
}

/* Floating food elements */
.floating-elements {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.float-item {
  position: absolute;
  font-size: 2.5rem;
  animation: float 6s ease-in-out infinite;
  filter: drop-shadow(0 4px 8px var(--shadow-sm));
}

.float-1 { top: 15%; left: 10%; animation-delay: 0s; }
.float-2 { top: 25%; right: 15%; animation-delay: 1s; font-size: 2rem; }
.float-3 { bottom: 30%; left: 8%; animation-delay: 2s; }
.float-4 { bottom: 20%; right: 10%; animation-delay: 1.5s; font-size: 2rem; }
.float-5 { top: 40%; left: 20%; animation-delay: 0.5s; font-size: 1.8rem; }

/* Content */
.hero-content {
  position: relative;
  z-index: 10;
  text-align: center;
  max-width: 700px;
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--bg-hover);
  color: var(--accent);
  padding: 0.5rem 1rem;
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.875rem;
  margin-bottom: 1.5rem;
  animation: fade-in-up 0.8s ease-out;
}

.badge-icon {
  animation: sparkle 2s ease-in-out infinite;
}

.hero-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: clamp(2.5rem, 6vw, 4rem);
  color: var(--text-primary);
  line-height: 1.1;
  margin: 0 0 1.5rem;
  animation: fade-in-up 0.8s ease-out 0.1s backwards;
}

.hero-subtitle {
  font-family: 'Nunito', sans-serif;
  font-size: clamp(1.1rem, 2.5vw, 1.35rem);
  color: var(--text-secondary);
  line-height: 1.7;
  margin: 0 0 2.5rem;
  max-width: 550px;
  margin-inline: auto;
  animation: fade-in-up 0.8s ease-out 0.2s backwards;
}

.hero-actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2rem;
  animation: fade-in-up 0.8s ease-out 0.3s backwards;
}

.auth-buttons {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.cta-link {
  text-decoration: none;
}

.btn-arrow {
  display: inline-block;
  transition: transform 0.3s ease;
}

.cta-link:hover .btn-arrow {
  transform: translateX(4px);
}

.login-link {
  font-family: 'Nunito', sans-serif;
  font-size: 0.95rem;
  color: var(--text-secondary);
  text-decoration: none;
  transition: color 0.2s ease;
}

.login-link span {
  color: var(--accent);
  font-weight: 700;
}

.login-link:hover {
  color: var(--text-primary);
}

.login-link:hover span {
  text-decoration: underline;
}

.logged-in-text {
  font-family: 'Nunito', sans-serif;
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin: 0;
}

.logged-in-text strong {
  color: var(--accent);
}

.trust-badges {
  display: flex;
  gap: 2rem;
  flex-wrap: wrap;
  justify-content: center;
}

.trust-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.trust-icon {
  font-size: 1.2rem;
}

/* Hero illustration */
.hero-illustration {
  position: absolute;
  bottom: 5%;
  right: 5%;
  z-index: 5;
  display: none;
}

@media (min-width: 1024px) {
  .hero-illustration {
    display: block;
  }
}

.plate {
  position: relative;
  animation: float 4s ease-in-out infinite;
}

.plate-inner {
  width: 180px;
  height: 180px;
  background: linear-gradient(165deg, var(--bg-card) 0%, var(--bg-secondary) 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--shadow-lg);
}

.plate-emoji {
  font-size: 4rem;
  animation: wiggle 3s ease-in-out infinite;
}

.plate-shadow {
  position: absolute;
  bottom: -20px;
  left: 50%;
  transform: translateX(-50%);
  width: 140px;
  height: 20px;
  background: radial-gradient(ellipse, rgba(61, 44, 41, 0.15) 0%, transparent 70%);
}

/* Animations */
@keyframes float {
  0%, 100% { transform: translateY(0) rotate(0deg); }
  50% { transform: translateY(-15px) rotate(3deg); }
}

@keyframes float-slow {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(30px, 20px); }
}

@keyframes pulse {
  0%, 100% { transform: translate(-50%, -50%) scale(1); opacity: 0.3; }
  50% { transform: translate(-50%, -50%) scale(1.1); opacity: 0.4; }
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

@keyframes sparkle {
  0%, 100% { transform: scale(1) rotate(0deg); }
  50% { transform: scale(1.2) rotate(10deg); }
}

@keyframes wiggle {
  0%, 100% { transform: rotate(-5deg); }
  50% { transform: rotate(5deg); }
}

/* Responsive */
@media (max-width: 768px) {
  .hero {
    min-height: auto;
    padding: 4rem 1.5rem;
  }

  .float-item {
    font-size: 1.5rem;
  }

  .floating-elements .float-3,
  .floating-elements .float-4,
  .floating-elements .float-5 {
    display: none;
  }
}
</style>
