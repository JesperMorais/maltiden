<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import BaseButton from '@/components/common/BaseButton.vue'
import { useUserStore } from '@/stores/user'
import { useThemeStore } from '@/stores/theme'
import {
  Sparkles,
  CalendarCheck,
  ShoppingCart,
  Check,
  BookOpen,
  Clock,
  Users as UsersIcon,
  UtensilsCrossed,
} from 'lucide-vue-next'

interface Props {
  title: string
  subtitle: string
  ctaButtonText: string
  ctaButtonLink: string
}

defineProps<Props>()

const router = useRouter()
const userStore = useUserStore()
const themeStore = useThemeStore()

function goToDashboard() {
  router.push('/dashboard')
}

const previewDays = [
  { day: 'Mån', meal: 'Pasta carbonara' },
  { day: 'Tis', meal: 'Kycklingwok' },
  { day: 'Ons', meal: 'Laxfilé med dill' },
  { day: 'Tor', meal: 'Tacos' },
]

const previewItems = [
  { name: 'Pasta 500g', checked: true },
  { name: 'Kycklingfilé', checked: false },
  { name: 'Lax 400g', checked: false },
  { name: 'Grädde 3dl', checked: false },
]

const previewRecipe = {
  title: 'Pasta carbonara',
  time: '25 min',
  servings: '4 port',
  ingredients: ['Spaghetti 400g', 'Bacon 150g', 'Ägg 3st', 'Parmesan 100g'],
  tags: ['Snabb', 'Klassiker'],
}

// Cycling card stack
const CARD_COUNT = 3
const activeCard = ref(0)
let timer: ReturnType<typeof setInterval> | undefined
let paused = false

function nextCard() {
  activeCard.value = (activeCard.value + 1) % CARD_COUNT
}

function startTimer() {
  stopTimer()
  timer = setInterval(() => {
    if (!paused) nextCard()
  }, 4000)
}

function stopTimer() {
  if (timer) {
    clearInterval(timer)
    timer = undefined
  }
}

function onPreviewClick() {
  nextCard()
  startTimer()
}

function onPreviewEnter() {
  paused = true
}

function onPreviewLeave() {
  paused = false
}

onMounted(startTimer)
onUnmounted(stopTimer)
</script>

<template>
  <section class="hero">
    <!-- Navigation -->
    <nav class="hero-nav">
      <RouterLink to="/" class="nav-logo">
        <UtensilsCrossed :size="22" :stroke-width="2" class="logo-icon" />
        <span class="logo-text">Måltiden</span>
      </RouterLink>
      <div class="nav-actions">
        <RouterLink v-prefetch="'about'" to="/about" class="nav-link">Om oss</RouterLink>
        <button
          class="theme-toggle"
          @click="themeStore.toggleDarkMode()"
          :aria-label="themeStore.isDarkMode ? 'Byt till ljust läge' : 'Byt till mörkt läge'"
        >
          <svg v-if="themeStore.isDarkMode" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
          <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
        </button>
      </div>
    </nav>

    <!-- Decorative background blobs (reduced) -->
    <div class="hero-bg">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
    </div>

    <div class="hero-inner">
      <div class="hero-content">
        <div class="badge">
          <Sparkles :size="16" :stroke-width="2" class="badge-icon" />
          <span>Smartare matplanering</span>
        </div>

        <h1 class="hero-title">{{ title }}</h1>

        <p class="hero-subtitle">{{ subtitle }}</p>

        <div class="hero-actions">
          <!-- Logged in: Go to dashboard -->
          <template v-if="userStore.isAuthenticated">
            <BaseButton variant="primary" size="lg" @click="goToDashboard">
              Gå till Dashboard
              <span class="btn-arrow">→</span>
            </BaseButton>
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
        </div>
      </div>

      <!-- App preview — cycling card stack -->
      <div
        class="hero-preview"
        aria-hidden="true"
        @click="onPreviewClick"
        @mouseenter="onPreviewEnter"
        @mouseleave="onPreviewLeave"
      >
        <div class="preview-stack">
          <!-- Card 0: Weekly menu -->
          <div
            class="preview-card"
            :class="{
              'stack-front': activeCard === 0,
              'stack-mid': activeCard === 2,
              'stack-back': activeCard === 1,
            }"
          >
            <div class="preview-header">
              <CalendarCheck :size="16" :stroke-width="2" class="preview-header-icon" />
              <span class="preview-title">Veckans meny</span>
            </div>
            <ul class="preview-menu">
              <li v-for="item in previewDays" :key="item.day" class="preview-day">
                <span class="preview-day-label">{{ item.day }}</span>
                <span class="preview-meal">{{ item.meal }}</span>
              </li>
            </ul>
          </div>

          <!-- Card 1: Shopping list -->
          <div
            class="preview-card"
            :class="{
              'stack-front': activeCard === 1,
              'stack-mid': activeCard === 0,
              'stack-back': activeCard === 2,
            }"
          >
            <div class="preview-header">
              <ShoppingCart :size="16" :stroke-width="2" class="preview-header-icon" />
              <span class="preview-title">Inköpslista</span>
              <span class="preview-count">{{ previewItems.length }} varor</span>
            </div>
            <ul class="preview-list">
              <li
                v-for="item in previewItems"
                :key="item.name"
                class="preview-list-item"
                :class="{ checked: item.checked }"
              >
                <span class="preview-checkbox">
                  <Check v-if="item.checked" :size="12" :stroke-width="3" />
                </span>
                <span class="preview-item-name">{{ item.name }}</span>
              </li>
            </ul>
          </div>

          <!-- Card 2: Recipe -->
          <div
            class="preview-card"
            :class="{
              'stack-front': activeCard === 2,
              'stack-mid': activeCard === 1,
              'stack-back': activeCard === 0,
            }"
          >
            <div class="preview-header">
              <BookOpen :size="16" :stroke-width="2" class="preview-header-icon" />
              <span class="preview-title">{{ previewRecipe.title }}</span>
            </div>
            <div class="recipe-meta">
              <span class="recipe-badge"><Clock :size="12" :stroke-width="2" /> {{ previewRecipe.time }}</span>
              <span class="recipe-badge"><UsersIcon :size="12" :stroke-width="2" /> {{ previewRecipe.servings }}</span>
            </div>
            <ul class="recipe-ingredients">
              <li v-for="ing in previewRecipe.ingredients" :key="ing">{{ ing }}</li>
            </ul>
            <div class="recipe-tags">
              <span v-for="tag in previewRecipe.tags" :key="tag" class="recipe-tag">{{ tag }}</span>
            </div>
          </div>
        </div>

        <!-- Dots indicator -->
        <div class="preview-dots">
          <button
            v-for="i in CARD_COUNT"
            :key="i"
            class="preview-dot"
            :class="{ active: activeCard === i - 1 }"
            @click.stop="activeCard = i - 1; startTimer()"
          />
        </div>
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
  padding-bottom: 2rem;
  background: transparent;
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
  color: var(--accent);
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
  width: 44px;
  height: 44px;
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

/* Background decorations — reduced opacity, no center blob */
.hero-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.3;
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

/* Split layout wrapper */
.hero-inner {
  position: relative;
  z-index: 10;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3rem;
  max-width: 1200px;
  width: 100%;
}

/* Content */
.hero-content {
  text-align: center;
  max-width: 700px;
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--bg-hover);
  color: var(--accent-text);
  padding: 0.5rem 1rem;
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.875rem;
  margin-bottom: 1.5rem;
  animation: fade-in-up 0.8s ease-out;
}

.badge-icon {
  color: var(--accent);
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
  gap: 1rem;
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
  color: var(--accent-text);
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

/* App preview — cycling card stack */
.hero-preview {
  animation: fade-in-up 0.8s ease-out 0.4s backwards;
  width: 100%;
  max-width: 380px;
  cursor: pointer;
  user-select: none;
  -webkit-tap-highlight-color: transparent;
}

.preview-stack {
  position: relative;
  perspective: 900px;
  min-height: 260px;
}

.preview-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  padding: 1.25rem;
  box-shadow: var(--shadow-lg);
  position: absolute;
  inset: 0;
  transition:
    transform 0.5s cubic-bezier(0.34, 1.56, 0.64, 1),
    opacity 0.4s ease,
    box-shadow 0.4s ease;
}

/* Front card: fully visible */
.stack-front {
  z-index: 3;
  transform: rotate(0deg) translateY(0);
  opacity: 1;
}

/* Middle card: peeking behind, shifted right + down */
.stack-mid {
  z-index: 2;
  transform: rotate(3deg) translate(8%, 6%);
  opacity: 0.7;
}

/* Back card: further behind */
.stack-back {
  z-index: 1;
  transform: rotate(6deg) translate(16%, 12%);
  opacity: 0.45;
}

/* Hover lifts the front card */
.hero-preview:hover .stack-front {
  transform: translateY(-4px);
  box-shadow: 0 20px 48px rgba(61, 44, 41, 0.22);
}

/* Dots — pill indicator */
.preview-dots {
  display: flex;
  justify-content: center;
  gap: 0.375rem;
  margin-top: 1.25rem;
}

.preview-dot {
  height: 6px;
  width: 6px;
  border-radius: 100px;
  border: none;
  background: var(--border-color-hover);
  cursor: pointer;
  padding: 0;
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.preview-dot.active {
  width: 24px;
  background: var(--accent);
}

.preview-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.875rem;
  padding-bottom: 0.625rem;
  border-bottom: 1px solid var(--border-color);
}

.preview-header-icon {
  color: var(--accent);
  flex-shrink: 0;
}

.preview-title {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 0.95rem;
  color: var(--text-primary);
}

.preview-count {
  margin-left: auto;
  font-family: 'Nunito', sans-serif;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-muted);
}

/* Menu card rows */
.preview-menu {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.preview-day {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0.75rem;
  border-radius: 10px;
  background: var(--bg-hover);
}

.preview-day-label {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.75rem;
  color: var(--accent-text);
  min-width: 2rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.preview-meal {
  font-family: 'Nunito', sans-serif;
  font-size: 0.85rem;
  color: var(--text-primary);
}

/* Shopping list rows */
.preview-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.preview-list-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.375rem 0.625rem;
  border-radius: 8px;
}

.preview-checkbox {
  width: 18px;
  height: 18px;
  border-radius: 5px;
  border: 2px solid var(--border-color-hover);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--text-on-accent);
}

.preview-list-item.checked .preview-checkbox {
  background: var(--accent);
  border-color: var(--accent);
}

.preview-item-name {
  font-family: 'Nunito', sans-serif;
  font-size: 0.8rem;
  color: var(--text-primary);
}

.preview-list-item.checked .preview-item-name {
  text-decoration: line-through;
  color: var(--text-muted);
}

/* Recipe card */
.recipe-meta {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.recipe-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-family: 'Nunito', sans-serif;
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--text-secondary);
  background: var(--bg-hover);
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
}

.recipe-ingredients {
  list-style: none;
  padding: 0;
  margin: 0 0 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.recipe-ingredients li {
  font-family: 'Nunito', sans-serif;
  font-size: 0.78rem;
  color: var(--text-primary);
  padding: 0.25rem 0;
  border-bottom: 1px solid var(--border-color);
}

.recipe-ingredients li:last-child {
  border-bottom: none;
}

.recipe-tags {
  display: flex;
  gap: 0.375rem;
}

.recipe-tag {
  font-family: 'Nunito', sans-serif;
  font-size: 0.68rem;
  font-weight: 700;
  color: var(--accent-text);
  background: var(--bg-hover);
  padding: 0.2rem 0.5rem;
  border-radius: 100px;
}

/* Animations */
@keyframes float-slow {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(30px, 20px); }
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

/* Desktop: split layout */
@media (min-width: 1024px) {
  .hero-inner {
    flex-direction: row;
    align-items: center;
    gap: 4rem;
  }

  .hero-content {
    flex: 1 1 60%;
    text-align: left;
  }

  .hero-subtitle {
    margin-inline: 0;
  }

  .hero-actions {
    align-items: flex-start;
  }

  .hero-preview {
    flex: 0 0 auto;
    max-width: 400px;
  }
}

/* Responsive */
@media (max-width: 768px) {
  .hero {
    min-height: 100svh;
    padding: 4rem 1.5rem;
    background: transparent;
  }

  .hero-preview {
    max-width: 320px;
  }
}

@media (max-width: 480px) {
  .hero {
    padding: 3rem 1rem 2rem;
  }

  .hero-content {
    padding-top: 1.5rem;
  }

  .badge {
    margin-bottom: 1rem;
  }

  .hero-nav {
    padding: 1rem;
  }

  .nav-link {
    display: none;
  }

  .hero-preview {
    max-width: 300px;
  }

  .preview-stack {
    min-height: 230px;
  }

  .preview-card {
    padding: 1rem;
  }
}
</style>
