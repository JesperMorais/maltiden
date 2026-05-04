<script setup lang="ts">
import { ref, computed } from 'vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BackLink from '@/components/common/BackLink.vue'
import BaseThemeToggle from '@/components/common/BaseThemeToggle.vue'
import { Mail, CheckCircle2 } from 'lucide-vue-next'
import { forgotPassword } from '@/api/auth.api'

const email = ref('')
const isLoading = ref(false)
const submitted = ref(false)

const canSubmit = computed(() => email.value.includes('@') && !isLoading.value)

async function handleSubmit() {
  if (!canSubmit.value) return

  isLoading.value = true
  try {
    // Backend always returns 200 to prevent account-existence leaks.
    // We swallow any unexpected error here too — the user always sees the
    // generic confirmation message regardless of outcome.
    await forgotPassword(email.value.trim())
  } catch {
    // Intentionally swallow — see comment above.
  } finally {
    isLoading.value = false
    submitted.value = true
  }
}
</script>

<template>
  <main class="forgot-page">
    <div class="bg-decorations">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
    </div>

    <div class="forgot-container">
      <div class="forgot-top-bar">
        <BackLink to="/login" />
        <BaseThemeToggle />
      </div>

      <div class="forgot-card">
        <div v-if="!submitted" class="forgot-content">
          <div class="logo-icon">
            <Mail :size="40" :stroke-width="1.75" />
          </div>
          <h1>Glömt lösenord?</h1>
          <p class="lead">
            Ange din e-postadress nedan så skickar vi en återställningslänk.
          </p>

          <form @submit.prevent="handleSubmit" class="forgot-form">
            <label class="form-label">
              <span>E-post</span>
              <input
                v-model="email"
                type="email"
                name="email"
                autocomplete="email"
                placeholder="din@email.se"
                class="form-input"
                required
              />
            </label>

            <BaseButton
              type="submit"
              variant="primary"
              size="lg"
              :disabled="!canSubmit"
              :loading="isLoading"
            >
              Skicka återställningslänk
            </BaseButton>
          </form>
        </div>

        <div v-else class="forgot-success">
          <div class="success-icon">
            <CheckCircle2 :size="48" :stroke-width="1.75" />
          </div>
          <h1>Kolla din inkorg</h1>
          <p class="lead">
            Om kontot finns har vi skickat ett mejl med en återställningslänk.
            Länken är giltig i 1 timme.
          </p>
          <p class="hint">
            Inget mejl? Kolla skräpposten eller försök igen om ett par minuter.
          </p>
          <RouterLink to="/login" class="back-to-login">Tillbaka till inloggning</RouterLink>
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
.forgot-page {
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

.forgot-container {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 420px;
}

.forgot-top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.forgot-card {
  background: var(--bg-card);
  border-radius: 24px;
  padding: 2.5rem;
  box-shadow: var(--shadow-lg);
  text-align: center;
}

.logo-icon,
.success-icon {
  color: var(--accent);
  margin-bottom: 1rem;
}

.forgot-card h1 {
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

.hint {
  font-family: 'Nunito', sans-serif;
  font-size: 0.85rem;
  color: var(--text-muted);
  margin: -1rem 0 1.5rem;
}

.forgot-form {
  text-align: left;
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
  background: var(--bg-primary);
  transition: all 0.3s ease;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 4px var(--accent-focus-ring);
}

.forgot-form .base-button {
  width: 100%;
}

.back-to-login {
  display: inline-block;
  margin-top: 1rem;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  color: var(--accent-text);
  text-decoration: none;
}

.back-to-login:hover {
  text-decoration: underline;
}

@media (max-width: 480px) {
  .forgot-card {
    padding: 2rem 1.5rem;
  }
}
</style>
