<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BaseButton from '@/components/common/BaseButton.vue'
import BackLink from '@/components/common/BackLink.vue'
import BaseThemeToggle from '@/components/common/BaseThemeToggle.vue'
import PasswordStrength from '@/components/common/PasswordStrength.vue'
import { KeyRound } from 'lucide-vue-next'
import { resetPassword } from '@/api/auth.api'
import { useToast } from '@/composables/useToast'
import { parseAxiosError } from '@/api/types/error.types'

const route = useRoute()
const router = useRouter()
const toast = useToast()

const token = ref('')
const password = ref('')
const confirm = ref('')
const isLoading = ref(false)
const error = ref('')

onMounted(() => {
  const t = route.query.token
  token.value = typeof t === 'string' ? t : ''
  if (!token.value) {
    error.value = 'Återställningslänken är ogiltig eller saknas.'
  }
})

const lengthOk = computed(() => password.value.length >= 8)
const hasUpper = computed(() => /[A-Z]/.test(password.value))
const hasLower = computed(() => /[a-z]/.test(password.value))
const hasDigit = computed(() => /[0-9]/.test(password.value))
const hasSpecial = computed(() => /[^A-Za-z0-9]/.test(password.value))
const typesMet = computed(
  () => [hasUpper.value, hasLower.value, hasDigit.value, hasSpecial.value].filter(Boolean).length,
)
const strongEnough = computed(() => lengthOk.value && typesMet.value >= 3)
const passwordsMatch = computed(() => password.value.length > 0 && password.value === confirm.value)

const canSubmit = computed(
  () => !!token.value && strongEnough.value && passwordsMatch.value && !isLoading.value,
)

async function handleSubmit() {
  if (!canSubmit.value) return

  isLoading.value = true
  error.value = ''

  try {
    await resetPassword(token.value, password.value)
    toast.success('Lösenord uppdaterat')
    router.push('/login')
  } catch (err: unknown) {
    const apiError = parseAxiosError(err)
    switch (apiError.code) {
      case 'invalid_reset_token':
        error.value = 'Länken är ogiltig. Begär en ny återställning.'
        break
      case 'expired_reset_token':
        error.value = 'Länken har gått ut. Begär en ny återställning.'
        break
      case 'used_reset_token':
        error.value = 'Länken har redan använts. Begär en ny återställning.'
        break
      case 'weak_password':
        error.value = 'Lösenordet är för svagt.'
        break
      default:
        error.value = 'Något gick fel. Försök igen.'
    }
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <main class="reset-page">
    <div class="bg-decorations">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
    </div>

    <div class="reset-container">
      <div class="reset-top-bar">
        <BackLink to="/login" />
        <BaseThemeToggle />
      </div>

      <div class="reset-card">
        <div class="logo-icon">
          <KeyRound :size="40" :stroke-width="1.75" />
        </div>
        <h1>Välj nytt lösenord</h1>
        <p class="lead">Ditt nya lösenord måste vara minst 8 tecken.</p>

        <form @submit.prevent="handleSubmit" class="reset-form">
          <label class="form-label">
            <span>Nytt lösenord</span>
            <input
              v-model="password"
              type="password"
              name="new-password"
              autocomplete="new-password"
              placeholder="Minst 8 tecken"
              class="form-input"
            />
          </label>

          <PasswordStrength :password="password" />

          <label class="form-label" style="margin-top: 1rem">
            <span>Bekräfta lösenord</span>
            <input
              v-model="confirm"
              type="password"
              name="confirm-password"
              autocomplete="new-password"
              placeholder="Skriv samma lösenord igen"
              class="form-input"
              :class="{ mismatch: confirm.length > 0 && !passwordsMatch }"
            />
          </label>

          <p v-if="confirm.length > 0 && !passwordsMatch" class="form-hint">
            Lösenorden matchar inte.
          </p>

          <p v-show="error" role="alert" class="form-error">{{ error }}</p>

          <BaseButton
            type="submit"
            variant="primary"
            size="lg"
            :disabled="!canSubmit"
            :loading="isLoading"
          >
            Uppdatera lösenord
          </BaseButton>
        </form>
      </div>
    </div>
  </main>
</template>

<style scoped>
.reset-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
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

.bg-decorations {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.4;
}

.blob-1 {
  width: 480px;
  height: 480px;
  background: linear-gradient(135deg, var(--peach) 0%, var(--accent-light) 100%);
  top: -180px;
  right: -120px;
}

.blob-2 {
  width: 360px;
  height: 360px;
  background: linear-gradient(135deg, var(--yellow-soft) 0%, var(--peach) 100%);
  bottom: -120px;
  left: -80px;
}

.reset-container {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 460px;
}

.reset-top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.reset-card {
  background: var(--bg-card);
  border-radius: 24px;
  padding: 2.5rem;
  box-shadow: var(--shadow-lg);
  text-align: center;
}

.logo-icon {
  color: var(--accent);
  margin-bottom: 1rem;
}

.reset-card h1 {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.75rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
}

.lead {
  font-family: 'Nunito', sans-serif;
  font-size: 1rem;
  color: var(--text-secondary);
  margin: 0 0 2rem;
}

.reset-form {
  text-align: left;
}

.form-label {
  display: block;
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
  background: var(--bg-primary);
  transition: all 0.3s ease;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 4px var(--accent-focus-ring);
}

.form-input.mismatch {
  border-color: var(--error);
}

.form-hint {
  font-family: 'Nunito', sans-serif;
  font-size: 0.85rem;
  color: var(--error);
  margin: 0.5rem 0 0;
}

.form-error {
  font-family: 'Nunito', sans-serif;
  font-size: 0.9rem;
  color: var(--error);
  margin: 1rem 0;
  padding: 0.5rem 0.75rem;
  background: var(--error-bg);
  border-radius: 8px;
  text-align: center;
}

.reset-form .base-button {
  width: 100%;
  margin-top: 1.25rem;
}

@media (max-width: 480px) {
  .reset-card {
    padding: 2rem 1.5rem;
  }
}
</style>
