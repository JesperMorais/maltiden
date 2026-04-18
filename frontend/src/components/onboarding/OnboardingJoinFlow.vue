<script setup lang="ts">
import { ref, computed } from 'vue'
import { Eye, EyeOff, Star, User, Lightbulb } from 'lucide-vue-next'
import { useRouter } from 'vue-router'
import BaseButton from '@/components/common/BaseButton.vue'
import PasswordStrength from '@/components/common/PasswordStrength.vue'
import { useUserStore } from '@/stores/user'
import { joinHousehold } from '@/api/household.api'
import { isAxiosError } from 'axios'

type JoinStep = 'code' | 'member-or-guest' | 'member-form' | 'guest-form'

const router = useRouter()
const userStore = useUserStore()

const emit = defineEmits<{
  success: [payload: { type: 'member' | 'guest'; family: string }]
}>()

const joinCode = ref('')
const joinStep = ref<JoinStep>('code')
const matchedFamily = ref('')
const isValidatingCode = ref(false)
const codeError = ref('')

const joinForm = ref({
  name: '',
  email: '',
  password: '',
  passwordConfirm: '',
})

const showJoinPassword = ref(false)
const showJoinPasswordConfirm = ref(false)
const isSubmitting = ref(false)
const submitError = ref('')

// Normalize to match backend: codes are uppercase alphanumerics. Trim + upper
// so pasted values with casing/whitespace variations still resolve.
const normalizedCode = computed(() => joinCode.value.trim().toUpperCase())

const canSubmitCode = computed(() => normalizedCode.value.length >= 4)

function passwordCharTypes(pw: string): number {
  let upper = false,
    lower = false,
    digit = false,
    special = false
  for (const ch of pw) {
    if (/[A-Z]/.test(ch)) upper = true
    else if (/[a-z]/.test(ch)) lower = true
    else if (/[0-9]/.test(ch)) digit = true
    else special = true
  }
  return [upper, lower, digit, special].filter(Boolean).length
}

const passwordStrongEnoughJoin = computed(
  () => joinForm.value.password.length >= 8 && passwordCharTypes(joinForm.value.password) >= 3,
)

const passwordsMatchJoin = computed(
  () => joinForm.value.password === joinForm.value.passwordConfirm,
)

const canSubmitJoinMember = computed(
  () =>
    joinForm.value.name.length >= 2 &&
    joinForm.value.email.includes('@') &&
    passwordStrongEnoughJoin.value &&
    passwordsMatchJoin.value,
)

const canSubmitJoinGuest = computed(() => joinForm.value.name.length >= 2)

// There's no pre-join "preview by code" endpoint on the backend yet, so we
// skip the welcome-preview step and only validate the code at submit time
// (when the user registers + joins in the same flow).
function validateCode() {
  if (!canSubmitCode.value) return
  codeError.value = ''
  joinStep.value = 'member-or-guest'
}

function selectJoinType(type: 'member' | 'guest') {
  joinStep.value = type === 'member' ? 'member-form' : 'guest-form'
  submitError.value = ''
}

function joinErrorMessage(e: unknown, fallback: string): string {
  if (isAxiosError(e)) {
    const code = e.response?.data?.error as string | undefined
    if (code === 'invalid_code') return 'Koden är ogiltig eller har gått ut.'
    if (code === 'already_member') return 'Du är redan medlem i hushållet.'
    if (code === 'code_required') return 'Ange en kod.'
    if (!e.response) return 'Kunde inte nå servern — kontrollera din internetanslutning.'
  }
  return fallback
}

async function handleJoinAsMember() {
  if (!canSubmitJoinMember.value) return
  isSubmitting.value = true
  submitError.value = ''

  // Two-step: create the user (which auto-provisions a solo household), then
  // join the invited household. Backend enforces single-household membership,
  // so joining swaps the user out of the auto-created one. If register
  // succeeds but join fails, the user is still registered — surface the
  // error so they know why the family isn't appearing.
  const registered = await userStore.register(
    joinForm.value.name,
    joinForm.value.email,
    joinForm.value.password,
  )

  if (!registered) {
    isSubmitting.value = false
    submitError.value = userStore.error || 'Registreringen misslyckades'
    return
  }

  try {
    await joinHousehold(normalizedCode.value)
    emit('success', { type: 'member', family: matchedFamily.value })
    setTimeout(() => {
      router.push('/dashboard')
    }, 1500)
  } catch (e) {
    submitError.value = joinErrorMessage(e, 'Kunde inte gå med i hushållet.')
  } finally {
    isSubmitting.value = false
  }
}

async function handleJoinAsGuest() {
  if (!canSubmitJoinGuest.value) return
  isSubmitting.value = true

  // Guest accounts aren't backed by a real endpoint yet — keep the UI
  // behaving the same way while we flesh out the backend surface.
  await new Promise((resolve) => setTimeout(resolve, 1500))

  isSubmitting.value = false
  emit('success', { type: 'guest', family: matchedFamily.value })
}
</script>

<template>
  <div class="join-flow">
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
          :aria-describedby="codeError ? 'code-error' : undefined"
          @keyup.enter="validateCode"
        />
      </label>
      <p v-if="codeError" id="code-error" role="alert" class="form-error">{{ codeError }}</p>
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

    <!-- Step 2: Member or Guest choice -->
    <div v-else-if="joinStep === 'member-or-guest'" class="form-card member-choice">
      <h3 class="choice-title">Hur vill du gå med?</h3>

      <div class="join-options">
        <button class="join-option" @click="selectJoinType('member')">
          <div class="option-icon"><Star :size="24" :stroke-width="1.75" /></div>
          <div class="option-content">
            <h4>Bli medlem</h4>
            <p>Skapa ett konto med e-post</p>
          </div>
          <span class="option-badge recommended">Rekommenderas</span>
        </button>

        <button class="join-option guest" @click="selectJoinType('guest')">
          <div class="option-icon"><User :size="24" :stroke-width="1.75" /></div>
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
    <form
      v-else-if="joinStep === 'member-form'"
      class="form-card"
      @submit.prevent="handleJoinAsMember"
    >
      <button type="button" class="back-to-choice" @click="joinStep = 'member-or-guest'">
        ← Tillbaka
      </button>

      <h3 class="form-title">Skapa ditt medlemskonto</h3>
      <p class="form-subtitle">Du går med i {{ matchedFamily }}</p>

      <label class="form-label">
        <span>Ditt namn</span>
        <input
          v-model="joinForm.name"
          type="text"
          name="name"
          autocomplete="name"
          placeholder="Anna Andersson"
          class="form-input"
        />
      </label>

      <label class="form-label">
        <span>E-post</span>
        <input
          v-model="joinForm.email"
          type="email"
          name="email"
          autocomplete="email"
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
            name="password"
            autocomplete="new-password"
            placeholder="Minst 8 tecken"
            class="form-input"
          />
          <button
            type="button"
            class="password-toggle"
            :aria-label="showJoinPassword ? 'Dölj lösenord' : 'Visa lösenord'"
            @click="showJoinPassword = !showJoinPassword"
          >
            <component :is="showJoinPassword ? EyeOff : Eye" :size="18" :stroke-width="2" />
          </button>
        </div>
        <PasswordStrength
          v-if="joinForm.password.length > 0"
          :password="joinForm.password"
        />
      </label>

      <label class="form-label">
        <span>Bekräfta lösenord</span>
        <div class="password-input-wrapper">
          <input
            v-model="joinForm.passwordConfirm"
            :type="showJoinPasswordConfirm ? 'text' : 'password'"
            autocomplete="new-password"
            placeholder="Skriv lösenordet igen"
            class="form-input"
            :class="{ 'input-error': joinForm.passwordConfirm && !passwordsMatchJoin }"
          />
          <button
            type="button"
            class="password-toggle"
            :aria-label="showJoinPasswordConfirm ? 'Dölj lösenord' : 'Visa lösenord'"
            @click="showJoinPasswordConfirm = !showJoinPasswordConfirm"
          >
            <component :is="showJoinPasswordConfirm ? EyeOff : Eye" :size="18" :stroke-width="2" />
          </button>
        </div>
        <span v-if="joinForm.passwordConfirm && !passwordsMatchJoin" class="field-error">
          Lösenorden matchar inte
        </span>
      </label>

      <p v-if="submitError" role="alert" class="form-error">{{ submitError }}</p>

      <BaseButton
        type="submit"
        variant="primary"
        size="lg"
        :disabled="!canSubmitJoinMember"
        :loading="isSubmitting"
      >
        Skapa konto och gå med
      </BaseButton>
    </form>

    <!-- Step 4b: Guest form -->
    <form
      v-else-if="joinStep === 'guest-form'"
      class="form-card"
      @submit.prevent="handleJoinAsGuest"
    >
      <button type="button" class="back-to-choice" @click="joinStep = 'member-or-guest'">
        ← Tillbaka
      </button>

      <h3 class="form-title">Gå med som gäst</h3>
      <p class="form-subtitle">Du går med i {{ matchedFamily }}</p>

      <label class="form-label">
        <span>Ditt namn</span>
        <input
          v-model="joinForm.name"
          type="text"
          name="name"
          autocomplete="name"
          placeholder="Anna"
          class="form-input"
        />
      </label>

      <p class="guest-note">
        <Lightbulb :size="16" :stroke-width="2" class="guest-note-icon" />
        Du kan uppgradera till medlem när som helst för att få full tillgång.
      </p>

      <BaseButton
        type="submit"
        variant="primary"
        size="lg"
        :disabled="!canSubmitJoinGuest"
        :loading="isSubmitting"
      >
        Gå med som gäst
      </BaseButton>
    </form>
  </div>
</template>

<style scoped>
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
  box-shadow: 0 0 0 4px var(--accent-focus-ring);
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
  color: var(--text-secondary);
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
  color: var(--text-primary);
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

.form-error {
  font-family: 'Nunito', sans-serif;
  font-size: 0.9rem;
  color: var(--error);
  margin: -0.5rem 0 1.5rem;
  padding: 0.5rem 0.75rem;
  background: var(--error-bg);
  border-radius: 8px;
}

.input-error {
  border-color: var(--error) !important;
  background: var(--error-bg);
}

.input-error:focus {
  box-shadow: 0 0 0 4px var(--error-bg) !important;
}

.field-error {
  display: block;
  font-family: 'Nunito', sans-serif;
  font-size: 0.8rem;
  color: var(--error);
  margin-top: 0.4rem;
}

.field-hint {
  display: block;
  font-family: 'Nunito', sans-serif;
  font-size: 0.8rem;
  color: var(--text-secondary, #a0a0a0);
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
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--accent);
  margin-bottom: 1rem;
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
  color: var(--accent);
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
  color: var(--text-on-accent);
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
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  font-family: 'Nunito', sans-serif;
  font-size: 0.85rem;
  color: var(--text-secondary);
  background: var(--bg-hover);
  padding: 0.75rem 1rem;
  border-radius: 10px;
  margin: 0 0 1.5rem;
}

.guest-note-icon {
  flex-shrink: 0;
  color: var(--accent);
  margin-top: 0.1rem;
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

/* Responsive */
@media (max-width: 480px) {
  .benefits-comparison {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .form-card {
    padding: 1.5rem;
  }
}
</style>
