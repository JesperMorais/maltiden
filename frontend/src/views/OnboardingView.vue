<script setup lang="ts">
import { ref, computed } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseCard from '@/components/common/BaseCard.vue'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

type Choice = 'none' | 'join' | 'create'
type JoinStep = 'code' | 'welcome' | 'member-or-guest' | 'member-form' | 'guest-form'

const selectedChoice = ref<Choice>('none')
const joinCode = ref('')
const joinStep = ref<JoinStep>('code')
const matchedFamily = ref('')
const isValidatingCode = ref(false)
const codeError = ref('')

const joinForm = ref({
  name: '',
  email: '',
  password: '',
  passwordConfirm: ''
})

const createForm = ref({
  name: '',
  email: '',
  password: '',
  passwordConfirm: '',
  householdName: ''
})
const isSubmitting = ref(false)
const submitSuccess = ref(false)
const joinedAsMember = ref(false)

// Password visibility toggles
const showJoinPassword = ref(false)
const showJoinPasswordConfirm = ref(false)
const showCreatePassword = ref(false)
const showCreatePasswordConfirm = ref(false)

// Mock family database
const mockFamilies: Record<string, string> = {
  'ABC123': 'Familjen Andersson',
  'FAM456': 'Johanssons Hushåll',
  'TEST99': 'Testfamiljen',
  'DEMO01': 'Demo Hushåll'
}

const canSubmitCode = computed(() => joinCode.value.length >= 4)

const passwordsMatchCreate = computed(() =>
  createForm.value.password === createForm.value.passwordConfirm
)
const passwordsMatchJoin = computed(() =>
  joinForm.value.password === joinForm.value.passwordConfirm
)

const canSubmitCreate = computed(() =>
  createForm.value.name.length >= 2 &&
  createForm.value.email.includes('@') &&
  createForm.value.password.length >= 8 &&
  passwordsMatchCreate.value &&
  createForm.value.householdName.length >= 2
)
const canSubmitJoinMember = computed(() =>
  joinForm.value.name.length >= 2 &&
  joinForm.value.email.includes('@') &&
  joinForm.value.password.length >= 8 &&
  passwordsMatchJoin.value
)
const canSubmitJoinGuest = computed(() => joinForm.value.name.length >= 2)

function selectChoice(choice: Choice) {
  selectedChoice.value = choice
  submitSuccess.value = false
  joinStep.value = 'code'
  codeError.value = ''
  joinCode.value = ''
  matchedFamily.value = ''
}

async function validateCode() {
  if (!canSubmitCode.value) return
  isValidatingCode.value = true
  codeError.value = ''

  // Mock API call - check if code exists
  await new Promise(resolve => setTimeout(resolve, 1000))

  const upperCode = joinCode.value.toUpperCase()
  if (mockFamilies[upperCode]) {
    matchedFamily.value = mockFamilies[upperCode]
    joinStep.value = 'welcome'

    // Auto-advance to choice after showing welcome
    setTimeout(() => {
      joinStep.value = 'member-or-guest'
    }, 1500)
  } else {
    codeError.value = 'Koden hittades inte. Kontrollera och försök igen.'
  }

  isValidatingCode.value = false
}

function selectJoinType(type: 'member' | 'guest') {
  joinStep.value = type === 'member' ? 'member-form' : 'guest-form'
}

async function handleJoinAsMember() {
  if (!canSubmitJoinMember.value) return
  isSubmitting.value = true

  await new Promise(resolve => setTimeout(resolve, 1500))

  isSubmitting.value = false
  joinedAsMember.value = true
  submitSuccess.value = true
}

async function handleJoinAsGuest() {
  if (!canSubmitJoinGuest.value) return
  isSubmitting.value = true

  await new Promise(resolve => setTimeout(resolve, 1500))

  isSubmitting.value = false
  joinedAsMember.value = false
  submitSuccess.value = true
}

const createError = ref('')

async function handleCreate() {
  if (!canSubmitCreate.value) return
  isSubmitting.value = true
  createError.value = ''

  // Call real backend registration
  const success = await userStore.register(
    createForm.value.name,
    createForm.value.email,
    createForm.value.password
  )

  isSubmitting.value = false

  if (success) {
    submitSuccess.value = true
    // Redirect to dashboard after showing success
    setTimeout(() => {
      router.push('/dashboard')
    }, 2000)
  } else {
    createError.value = userStore.error || 'Registreringen misslyckades'
  }
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
      <!-- Back link -->
      <RouterLink to="/" class="back-link">
        <span class="back-arrow">←</span>
        <span>Tillbaka</span>
      </RouterLink>

      <!-- Header -->
      <header class="onboarding-header">
        <div class="logo-icon">🍽️</div>
        <h1>Välkommen till Måltiden</h1>
        <p>Hur vill du komma igång?</p>
      </header>

      <!-- Success state -->
      <div v-if="submitSuccess" class="success-state">
        <div class="success-icon">🎉</div>
        <h2>{{ selectedChoice === 'join'
          ? `Välkommen till ${matchedFamily}!`
          : 'Konto skapat!'
        }}</h2>
        <p v-if="selectedChoice === 'join' && joinedAsMember">
          Du har gått med som medlem. Du har full tillgång till alla funktioner!
        </p>
        <p v-else-if="selectedChoice === 'join' && !joinedAsMember">
          Du har gått med som gäst. Du kan se menyn och inköpslistan.
        </p>
        <p v-else>
          Ditt hushåll är redo. Bjud in familjen!
        </p>
        <BaseButton variant="primary" size="lg">
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
              <div class="choice-icon">👨‍👩‍👧‍👦</div>
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
                <!-- Step 1: Enter code -->
                <div v-if="joinStep === 'code'" class="form-card">
                  <label class="form-label">
                    <span>Hushållskod</span>
                    <input
                      v-model="joinCode"
                      type="text"
                      placeholder="T.ex. ABC123"
                      class="form-input code-input"
                      maxlength="8"
                      @keyup.enter="validateCode"
                    />
                  </label>
                  <p v-if="codeError" class="form-error">{{ codeError }}</p>
                  <p v-else class="form-hint">Fråga den som skapade kontot efter koden</p>

                  <BaseButton
                    variant="primary"
                    size="lg"
                    :disabled="!canSubmitCode"
                    :loading="isValidatingCode"
                    @click="validateCode"
                  >
                    Fortsätt
                  </BaseButton>
                </div>

                <!-- Step 2: Welcome message -->
                <div v-else-if="joinStep === 'welcome'" class="welcome-step">
                  <div class="welcome-icon">🏠</div>
                  <h3>Välkommen till</h3>
                  <h2 class="family-name">{{ matchedFamily }}</h2>
                  <div class="loading-dots">
                    <span></span>
                    <span></span>
                    <span></span>
                  </div>
                </div>

                <!-- Step 3: Member or Guest choice -->
                <div v-else-if="joinStep === 'member-or-guest'" class="form-card member-choice">
                  <h3 class="choice-title">Hur vill du gå med?</h3>

                  <div class="join-options">
                    <button class="join-option" @click="selectJoinType('member')">
                      <div class="option-icon">⭐</div>
                      <div class="option-content">
                        <h4>Bli medlem</h4>
                        <p>Skapa ett konto med e-post</p>
                      </div>
                      <span class="option-badge recommended">Rekommenderas</span>
                    </button>

                    <button class="join-option guest" @click="selectJoinType('guest')">
                      <div class="option-icon">👤</div>
                      <div class="option-content">
                        <h4>Gå med som gäst</h4>
                        <p>Bara ange ditt namn</p>
                      </div>
                    </button>
                  </div>

                  <div class="benefits-comparison">
                    <div class="benefits-column member">
                      <h5>Som medlem får du:</h5>
                      <ul>
                        <li>✓ Skapa och ändra menyer</li>
                        <li>✓ Lägga till recept</li>
                        <li>✓ Hantera inköpslistan</li>
                        <li>✓ Bjuda in fler familjemedlemmar</li>
                        <li>✓ Synkas mellan enheter</li>
                      </ul>
                    </div>
                    <div class="benefits-column guest">
                      <h5>Som gäst kan du:</h5>
                      <ul>
                        <li>✓ Se veckomenyn</li>
                        <li>✓ Se inköpslistan</li>
                        <li>✗ <span class="muted">Ändra menyer</span></li>
                        <li>✗ <span class="muted">Lägga till recept</span></li>
                        <li>✗ <span class="muted">Synkas mellan enheter</span></li>
                      </ul>
                    </div>
                  </div>
                </div>

                <!-- Step 4a: Member form -->
                <div v-else-if="joinStep === 'member-form'" class="form-card">
                  <button class="back-to-choice" @click="joinStep = 'member-or-guest'">
                    ← Tillbaka
                  </button>

                  <h3 class="form-title">Skapa ditt medlemskonto</h3>
                  <p class="form-subtitle">Du går med i {{ matchedFamily }}</p>

                  <label class="form-label">
                    <span>Ditt namn</span>
                    <input
                      v-model="joinForm.name"
                      type="text"
                      placeholder="Anna Andersson"
                      class="form-input"
                    />
                  </label>

                  <label class="form-label">
                    <span>E-post</span>
                    <input
                      v-model="joinForm.email"
                      type="email"
                      placeholder="anna@exempel.se"
                      class="form-input"
                    />
                  </label>

                  <label class="form-label">
                    <span>Lösenord</span>
                    <div class="password-input-wrapper">
                      <input
                        v-model="joinForm.password"
                        :type="showJoinPassword ? 'text' : 'password'"
                        placeholder="Minst 8 tecken"
                        class="form-input"
                      />
                      <button
                        type="button"
                        class="password-toggle"
                        @click="showJoinPassword = !showJoinPassword"
                      >
                        {{ showJoinPassword ? '🙈' : '👁️' }}
                      </button>
                    </div>
                  </label>

                  <label class="form-label">
                    <span>Bekräfta lösenord</span>
                    <div class="password-input-wrapper">
                      <input
                        v-model="joinForm.passwordConfirm"
                        :type="showJoinPasswordConfirm ? 'text' : 'password'"
                        placeholder="Skriv lösenordet igen"
                        class="form-input"
                        :class="{ 'input-error': joinForm.passwordConfirm && !passwordsMatchJoin }"
                      />
                      <button
                        type="button"
                        class="password-toggle"
                        @click="showJoinPasswordConfirm = !showJoinPasswordConfirm"
                      >
                        {{ showJoinPasswordConfirm ? '🙈' : '👁️' }}
                      </button>
                    </div>
                    <span v-if="joinForm.passwordConfirm && !passwordsMatchJoin" class="field-error">
                      Lösenorden matchar inte
                    </span>
                  </label>

                  <BaseButton
                    variant="primary"
                    size="lg"
                    :disabled="!canSubmitJoinMember"
                    :loading="isSubmitting"
                    @click="handleJoinAsMember"
                  >
                    Skapa konto och gå med
                  </BaseButton>
                </div>

                <!-- Step 4b: Guest form -->
                <div v-else-if="joinStep === 'guest-form'" class="form-card">
                  <button class="back-to-choice" @click="joinStep = 'member-or-guest'">
                    ← Tillbaka
                  </button>

                  <h3 class="form-title">Gå med som gäst</h3>
                  <p class="form-subtitle">Du går med i {{ matchedFamily }}</p>

                  <label class="form-label">
                    <span>Ditt namn</span>
                    <input
                      v-model="joinForm.name"
                      type="text"
                      placeholder="Anna"
                      class="form-input"
                    />
                  </label>

                  <p class="guest-note">
                    💡 Du kan uppgradera till medlem när som helst för att få full tillgång.
                  </p>

                  <BaseButton
                    variant="primary"
                    size="lg"
                    :disabled="!canSubmitJoinGuest"
                    :loading="isSubmitting"
                    @click="handleJoinAsGuest"
                  >
                    Gå med som gäst
                  </BaseButton>
                </div>
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
              <div class="choice-icon">✨</div>
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
                <div class="form-card">
                  <label class="form-label">
                    <span>Ditt namn</span>
                    <input
                      v-model="createForm.name"
                      type="text"
                      placeholder="Anna Andersson"
                      class="form-input"
                    />
                  </label>

                  <label class="form-label">
                    <span>E-post</span>
                    <input
                      v-model="createForm.email"
                      type="email"
                      placeholder="anna@exempel.se"
                      class="form-input"
                    />
                  </label>

                  <label class="form-label">
                    <span>Lösenord</span>
                    <div class="password-input-wrapper">
                      <input
                        v-model="createForm.password"
                        :type="showCreatePassword ? 'text' : 'password'"
                        placeholder="Minst 8 tecken"
                        class="form-input"
                      />
                      <button
                        type="button"
                        class="password-toggle"
                        @click="showCreatePassword = !showCreatePassword"
                      >
                        {{ showCreatePassword ? '🙈' : '👁️' }}
                      </button>
                    </div>
                  </label>

                  <label class="form-label">
                    <span>Bekräfta lösenord</span>
                    <div class="password-input-wrapper">
                      <input
                        v-model="createForm.passwordConfirm"
                        :type="showCreatePasswordConfirm ? 'text' : 'password'"
                        placeholder="Skriv lösenordet igen"
                        class="form-input"
                        :class="{ 'input-error': createForm.passwordConfirm && !passwordsMatchCreate }"
                      />
                      <button
                        type="button"
                        class="password-toggle"
                        @click="showCreatePasswordConfirm = !showCreatePasswordConfirm"
                      >
                        {{ showCreatePasswordConfirm ? '🙈' : '👁️' }}
                      </button>
                    </div>
                    <span v-if="createForm.passwordConfirm && !passwordsMatchCreate" class="field-error">
                      Lösenorden matchar inte
                    </span>
                  </label>

                  <label class="form-label">
                    <span>Namn på hushållet</span>
                    <input
                      v-model="createForm.householdName"
                      type="text"
                      placeholder="T.ex. Familjen Andersson"
                      class="form-input"
                    />
                  </label>

                  <p v-if="createError" class="form-error">{{ createError }}</p>

                  <BaseButton
                    variant="primary"
                    size="lg"
                    :disabled="!canSubmitCreate"
                    :loading="isSubmitting"
                    @click="handleCreate"
                  >
                    Skapa konto
                  </BaseButton>

                  <p class="form-terms">
                    Genom att skapa konto godkänner du våra
                    <a href="#">villkor</a> och <a href="#">integritetspolicy</a>.
                  </p>
                </div>
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

/* Back link */
.back-link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--text-secondary);
  text-decoration: none;
  padding: 0.5rem 1rem;
  border-radius: 100px;
  transition: all 0.3s ease;
  margin-bottom: 2rem;
}

.back-link:hover {
  color: var(--accent);
  background: var(--accent-bg);
}

.back-arrow {
  transition: transform 0.3s ease;
}

.back-link:hover .back-arrow {
  transform: translateX(-4px);
}

/* Header */
.onboarding-header {
  text-align: center;
  margin-bottom: 3rem;
}

.logo-icon {
  font-size: 3.5rem;
  margin-bottom: 1rem;
  animation: float 3s ease-in-out infinite;
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
  font-size: 4rem;
  margin-bottom: 1.5rem;
  animation: bounce 1s ease-in-out;
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
  font-size: 3rem;
  margin-bottom: 1rem;
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
  background: var(--accent-bg);
  color: var(--accent);
  transition: all 0.3s ease;
}

.choice-card:hover .choice-indicator {
  background: var(--accent);
  color: white;
}

.choice-indicator .check {
  font-size: 1.1rem;
}

.choice-card-wrapper.active .choice-indicator {
  background: var(--accent);
  color: white;
}

/* Inline form */
.inline-form {
  margin-top: 1.5rem;
}

.form-card {
  background: var(--bg-primary);
  border-radius: 20px;
  padding: 2rem;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
}

.form-label {
  display: block;
  margin-bottom: 1.25rem;
}

.form-label span {
  display: block;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
}

.form-input {
  width: 100%;
  padding: 0.9rem 1.25rem;
  border: 2px solid var(--border-color);
  border-radius: 14px;
  font-family: 'Nunito', sans-serif;
  font-size: 1rem;
  color: var(--text-primary);
  background: var(--bg-card);
  transition: all 0.3s ease;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 4px var(--accent-bg);
}

.form-input::placeholder {
  color: var(--text-secondary);
  opacity: 0.6;
}

/* Password toggle */
.password-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.password-input-wrapper .form-input {
  padding-right: 3rem;
}

.password-toggle {
  position: absolute;
  right: 0.75rem;
  background: none;
  border: none;
  font-size: 1.25rem;
  cursor: pointer;
  padding: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0.6;
  transition: opacity 0.2s ease;
}

.password-toggle:hover {
  opacity: 1;
}

.code-input {
  font-size: 1.5rem;
  font-weight: 700;
  text-align: center;
  letter-spacing: 0.2em;
  text-transform: uppercase;
}

.form-hint {
  font-family: 'Nunito', sans-serif;
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin: -0.5rem 0 1.5rem;
}

.form-card .base-button {
  width: 100%;
}

.form-terms {
  font-family: 'Nunito', sans-serif;
  font-size: 0.8rem;
  color: var(--text-secondary);
  text-align: center;
  margin: 1rem 0 0;
}

.form-terms a {
  color: var(--accent);
  text-decoration: none;
}

.form-terms a:hover {
  text-decoration: underline;
}

.form-error {
  font-family: 'Nunito', sans-serif;
  font-size: 0.9rem;
  color: #e53e3e;
  margin: -0.5rem 0 1.5rem;
  padding: 0.5rem 0.75rem;
  background: rgba(229, 62, 62, 0.1);
  border-radius: 8px;
}

.input-error {
  border-color: #e53e3e !important;
  background: rgba(229, 62, 62, 0.03);
}

.input-error:focus {
  box-shadow: 0 0 0 4px rgba(229, 62, 62, 0.15) !important;
}

.field-error {
  display: block;
  font-family: 'Nunito', sans-serif;
  font-size: 0.8rem;
  color: #e53e3e;
  margin-top: 0.4rem;
}

/* Welcome step */
.welcome-step {
  text-align: center;
  padding: 3rem 2rem;
  background: var(--bg-primary);
  border-radius: 20px;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
  animation: fade-in-up 0.5s ease-out;
}

.welcome-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
  animation: bounce 0.6s ease-out;
}

.welcome-step h3 {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1rem;
  color: var(--text-secondary);
  margin: 0;
}

.family-name {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.75rem;
  color: var(--accent);
  margin: 0.25rem 0 1.5rem;
}

.loading-dots {
  display: flex;
  justify-content: center;
  gap: 0.5rem;
}

.loading-dots span {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent);
  animation: dot-bounce 1.4s ease-in-out infinite;
}

.loading-dots span:nth-child(1) { animation-delay: 0s; }
.loading-dots span:nth-child(2) { animation-delay: 0.2s; }
.loading-dots span:nth-child(3) { animation-delay: 0.4s; }

@keyframes dot-bounce {
  0%, 80%, 100% { transform: scale(0.6); opacity: 0.4; }
  40% { transform: scale(1); opacity: 1; }
}

/* Member choice step */
.member-choice {
  padding: 2rem;
}

.choice-title {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.25rem;
  color: var(--text-primary);
  text-align: center;
  margin: 0 0 1.5rem;
}

.join-options {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-bottom: 2rem;
}

.join-option {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem;
  background: var(--bg-card);
  border: 2px solid var(--border-color);
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.3s ease;
  text-align: left;
  position: relative;
}

.join-option:hover {
  border-color: var(--accent);
  transform: translateY(-2px);
  box-shadow: var(--shadow-sm);
}

.option-icon {
  font-size: 2rem;
  flex-shrink: 0;
}

.option-content h4 {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1rem;
  color: var(--text-primary);
  margin: 0 0 0.25rem;
}

.option-content p {
  font-family: 'Nunito', sans-serif;
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin: 0;
}

.option-badge {
  position: absolute;
  top: -8px;
  right: 12px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 0.25rem 0.75rem;
  border-radius: 100px;
}

.option-badge.recommended {
  background: linear-gradient(135deg, var(--accent) 0%, var(--peach) 100%);
  color: white;
}

/* Benefits comparison */
.benefits-comparison {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border-color);
}

.benefits-column h5 {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.85rem;
  color: var(--text-primary);
  margin: 0 0 0.75rem;
}

.benefits-column ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.benefits-column li {
  font-family: 'Nunito', sans-serif;
  font-size: 0.8rem;
  color: var(--text-secondary);
  padding: 0.35rem 0;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.benefits-column.member li {
  color: var(--text-primary);
}

.benefits-column li .muted {
  opacity: 0.5;
  text-decoration: line-through;
}

/* Back to choice button */
.back-to-choice {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  background: none;
  border: none;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 0.5rem 0;
  margin-bottom: 1rem;
  transition: color 0.2s ease;
}

.back-to-choice:hover {
  color: var(--accent);
}

.form-title {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.25rem;
  color: var(--text-primary);
  margin: 0 0 0.25rem;
}

.form-subtitle {
  font-family: 'Nunito', sans-serif;
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin: 0 0 1.5rem;
}

.guest-note {
  font-family: 'Nunito', sans-serif;
  font-size: 0.85rem;
  color: var(--text-secondary);
  background: var(--accent-bg);
  padding: 0.75rem 1rem;
  border-radius: 10px;
  margin: 0 0 1.5rem;
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
  background: var(--accent-bg);
}

/* Animations */
@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

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

@keyframes bounce {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.2); }
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

  .form-card {
    padding: 1.5rem;
  }
}
</style>
