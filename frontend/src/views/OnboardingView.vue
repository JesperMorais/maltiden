<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { UtensilsCrossed, UsersRound, Sparkles, PartyPopper } from 'lucide-vue-next'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseCard from '@/components/common/BaseCard.vue'
import BackLink from '@/components/common/BackLink.vue'
import BaseThemeToggle from '@/components/common/BaseThemeToggle.vue'
import OnboardingJoinFlow from '@/components/onboarding/OnboardingJoinFlow.vue'
import OnboardingCreateForm from '@/components/onboarding/OnboardingCreateForm.vue'

type Choice = 'none' | 'join' | 'create'

const router = useRouter()
const selectedChoice = ref<Choice>('none')
const submitSuccess = ref(false)
const joinedAsMember = ref(false)

function selectChoice(choice: Choice) {
  selectedChoice.value = choice
  submitSuccess.value = false
}

function handleJoinSuccess(payload: { type: 'member' | 'guest'; family: string }) {
  void payload.family
  joinedAsMember.value = payload.type === 'member'
  submitSuccess.value = true
}

function handleCreateSuccess() {
  submitSuccess.value = true
}

function goToDashboard() {
  router.push('/dashboard')
}
</script>

<template>
  <main class="onboarding-page">
    <!-- Background decorations -->
    <div class="bg-decorations">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="grain"></div>
    </div>

    <div class="onboarding-container">
      <!-- Top bar with back link and theme toggle -->
      <div class="onboarding-top-bar">
        <BackLink to="/" />
        <BaseThemeToggle />
      </div>

      <!-- Header -->
      <header class="onboarding-header">
        <div class="logo-icon"><UtensilsCrossed :size="32" :stroke-width="1.75" /></div>
        <h1>Välkommen till Måltiden</h1>
        <p>Hur vill du komma igång?</p>
      </header>

      <!-- Success state -->
      <div v-if="submitSuccess" class="success-state">
        <div class="success-icon"><PartyPopper :size="36" :stroke-width="1.75" /></div>
        <h2>{{ selectedChoice === 'join' ? 'Välkommen!' : 'Konto skapat!' }}</h2>
        <p v-if="selectedChoice === 'join' && joinedAsMember">
          Du har gått med som medlem. Du har full tillgång till alla funktioner!
        </p>
        <p v-else-if="selectedChoice === 'join' && !joinedAsMember">
          Du har gått med som gäst. Du kan se menyn och inköpslistan.
        </p>
        <p v-else>
          Ditt hushåll är redo. Bjud in familjen!
        </p>
        <BaseButton variant="primary" size="lg" @click="goToDashboard">
          Gå till dashboard →
        </BaseButton>
      </div>

      <!-- Choice cards -->
      <div v-else class="choices-wrapper">
        <div class="choices-grid" :class="{ 'choice-made': selectedChoice !== 'none' }">

          <!-- Join existing -->
          <div
            class="choice-card-wrapper"
            :class="{
              active: selectedChoice === 'join',
              inactive: selectedChoice === 'create'
            }"
          >
            <BaseCard
              class="choice-card"
              padding="lg"
              @click="selectChoice('join')"
            >
              <div class="choice-icon"><UsersRound :size="28" :stroke-width="1.75" /></div>
              <h2>Gå med i hushåll</h2>
              <p>Någon i din familj har redan skapat ett konto? Ange koden för att gå med.</p>

              <div class="choice-indicator">
                <span v-if="selectedChoice !== 'join'">Välj</span>
                <span v-else class="check">✓</span>
              </div>
            </BaseCard>

            <!-- Join flow -->
            <Transition name="form-slide">
              <div v-if="selectedChoice === 'join'" class="inline-form">
                <OnboardingJoinFlow @success="handleJoinSuccess" />
              </div>
            </Transition>
          </div>

          <!-- Create new -->
          <div
            class="choice-card-wrapper"
            :class="{
              active: selectedChoice === 'create',
              inactive: selectedChoice === 'join'
            }"
          >
            <BaseCard
              class="choice-card"
              padding="lg"
              @click="selectChoice('create')"
            >
              <div class="choice-icon"><Sparkles :size="28" :stroke-width="1.75" /></div>
              <h2>Skapa nytt hushåll</h2>
              <p>Starta ett nytt konto och bjud in din familj att planera måltider tillsammans.</p>

              <div class="choice-indicator">
                <span v-if="selectedChoice !== 'create'">Välj</span>
                <span v-else class="check">✓</span>
              </div>
            </BaseCard>

            <!-- Create form -->
            <Transition name="form-slide">
              <div v-if="selectedChoice === 'create'" class="inline-form">
                <OnboardingCreateForm @success="handleCreateSuccess" />
              </div>
            </Transition>
          </div>
        </div>

        <!-- Reset choice -->
        <Transition name="fade">
          <button
            v-if="selectedChoice !== 'none'"
            class="reset-choice"
            @click="selectChoice('none')"
          >
            ← Ändra val
          </button>
        </Transition>
      </div>
    </div>
  </main>
</template>

<style scoped>
.onboarding-page {
  min-height: 100vh;
  padding: 2rem;
  background: linear-gradient(
    165deg,
    var(--bg-secondary) 0%,
    var(--bg-primary) 50%,
    var(--bg-secondary) 100%
  );
  position: relative;
  overflow: hidden;
}

/* Background */
.bg-decorations {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.5;
}

.blob-1 {
  width: 500px;
  height: 500px;
  background: linear-gradient(135deg, var(--peach) 0%, var(--accent-light) 100%);
  top: -200px;
  right: -150px;
}

.blob-2 {
  width: 400px;
  height: 400px;
  background: linear-gradient(135deg, var(--yellow-soft) 0%, var(--peach) 100%);
  bottom: -150px;
  left: -100px;
}

.grain {
  position: absolute;
  inset: 0;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 400 400' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)'/%3E%3C/svg%3E");
  opacity: 0.03;
}

/* Container */
.onboarding-container {
  position: relative;
  z-index: 1;
  max-width: 900px;
  margin: 0 auto;
}

/* Top bar */
.onboarding-top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

/* Header */
.onboarding-header {
  text-align: center;
  margin-bottom: 3rem;
}

.logo-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border-radius: 18px;
  background: var(--accent-bg);
  color: var(--accent);
  margin: 0 auto 1rem;
}

.onboarding-header h1 {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: clamp(1.75rem, 4vw, 2.5rem);
  color: var(--text-primary);
  margin: 0 0 0.5rem;
}

.onboarding-header p {
  font-family: 'Nunito', sans-serif;
  font-size: 1.15rem;
  color: var(--text-secondary);
  margin: 0;
}

/* Success state */
.success-state {
  text-align: center;
  padding: 3rem;
  animation: fade-in-up 0.5s ease-out;
}

.success-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--accent);
  margin-bottom: 1.5rem;
}

.success-state h2 {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 2rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
}

.success-state p {
  font-family: 'Nunito', sans-serif;
  font-size: 1.1rem;
  color: var(--text-secondary);
  margin: 0 0 2rem;
}

/* Choices */
.choices-wrapper {
  position: relative;
}

.choices-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 2rem;
  transition: all 0.5s cubic-bezier(0.4, 0, 0.2, 1);
}

.choices-grid.choice-made {
  grid-template-columns: 1fr;
  max-width: 500px;
  margin: 0 auto;
}

.choice-card-wrapper {
  transition: all 0.5s cubic-bezier(0.4, 0, 0.2, 1);
}

.choice-card-wrapper.inactive {
  opacity: 0;
  transform: scale(0.8);
  position: absolute;
  pointer-events: none;
}

.choice-card {
  cursor: pointer;
  text-align: center;
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
  border: 3px solid transparent;
}

.choice-card:hover {
  transform: translateY(-8px);
  border-color: var(--accent);
}

.choice-card-wrapper.active .choice-card {
  border-color: var(--accent);
  transform: none;
}

.choice-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: 16px;
  background: var(--accent-bg);
  color: var(--accent);
  margin: 0 auto 1rem;
}

.choice-card h2 {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.35rem;
  color: var(--text-primary);
  margin: 0 0 0.75rem;
}

.choice-card p {
  font-family: 'Nunito', sans-serif;
  font-size: 1rem;
  color: var(--text-secondary);
  line-height: 1.6;
  margin: 0 0 1.5rem;
}

.choice-indicator {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 80px;
  padding: 0.5rem 1.25rem;
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  background: var(--bg-hover);
  color: var(--accent);
  transition: all 0.3s ease;
}

.choice-card:hover .choice-indicator {
  background: var(--accent);
  color: var(--text-on-accent);
}

.choice-indicator .check {
  font-size: 1.1rem;
}

.choice-card-wrapper.active .choice-indicator {
  background: var(--accent);
  color: var(--text-on-accent);
}

/* Inline form */
.inline-form {
  margin-top: 1.5rem;
}

/* Reset choice button */
.reset-choice {
  display: block;
  margin: 2rem auto 0;
  padding: 0.75rem 1.5rem;
  background: none;
  border: none;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 100px;
  transition: all 0.3s ease;
}

.reset-choice:hover {
  color: var(--accent);
  background: var(--bg-hover);
}

/* Animations */
@keyframes fade-in-up {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Transitions */
.form-slide-enter-active {
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.form-slide-leave-active {
  transition: all 0.3s ease;
}

.form-slide-enter-from,
.form-slide-leave-to {
  opacity: 0;
  transform: translateY(-20px);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* Responsive */
@media (max-width: 768px) {
  .onboarding-page {
    padding: 1.5rem;
  }

  .choices-grid {
    grid-template-columns: 1fr;
  }

  .choice-card-wrapper.inactive {
    display: none;
  }
}
</style>
