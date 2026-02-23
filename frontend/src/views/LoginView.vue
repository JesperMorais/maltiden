<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import BaseButton from '@/components/common/BaseButton.vue'
import { useUserStore } from '@/stores/user'
import { useThemeStore } from '@/stores/theme'
import { useToast } from '@/composables/useToast'
import WavesBackground from '@/components/vue-bits/WavesBackground.vue'

const router = useRouter()
const userStore = useUserStore()
const themeStore = useThemeStore()
const toast = useToast()

onMounted(() => {
  if (sessionStorage.getItem('session_expired')) {
    sessionStorage.removeItem('session_expired')
    toast.info('Din session har löpt ut. Logga in igen.')
  }
})

const email = ref('')
const password = ref('')
const isLoading = ref(false)
const error = ref('')

const canSubmit = computed(() =>
  email.value.includes('@') && password.value.length >= 8
)

const waveLineColor = computed(() =>
  themeStore.isDarkMode ? 'rgba(255, 138, 125, 0.15)' : 'rgba(255, 107, 91, 0.12)',
)

async function handleLogin() {
  if (!canSubmit.value) return

  isLoading.value = true
  error.value = ''

  // Use real backend login
  const success = await userStore.login(email.value, password.value)

  if (success) {
    router.push('/dashboard')
  } else {
    error.value = userStore.error || 'Fel e-post eller lösenord'
  }

  isLoading.value = false
}
</script>

<template>
  <main class="login-page">
    <!-- Animated wave background -->
    <WavesBackground
      :line-color="waveLineColor"
      background-color="transparent"
      :wave-speed-x="0.01"
      :wave-speed-y="0.004"
      :wave-amp-x="40"
      :wave-amp-y="20"
      :x-gap="12"
      :y-gap="36"
    />

    <!-- Background decorations (kept for blobs + grain texture) -->
    <div class="bg-decorations">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="grain"></div>
    </div>

    <div class="login-container">
      <!-- Top bar with back link and theme toggle -->
      <div class="login-top-bar">
        <RouterLink v-prefetch="'landing'" to="/" class="back-link">
          <span class="back-arrow">←</span>
          <span>Tillbaka</span>
        </RouterLink>
        <button
          class="theme-toggle"
          @click="themeStore.toggleDarkMode()"
          :aria-label="themeStore.isDarkMode ? 'Byt till ljust läge' : 'Byt till mörkt läge'"
        >
          <svg v-if="themeStore.isDarkMode" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
          <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
        </button>
      </div>

      <!-- Login card -->
      <div class="login-card">
        <div class="logo-icon">🍽️</div>
        <h1>Välkommen tillbaka</h1>
        <p>Logga in på ditt Måltiden-konto</p>

        <form @submit.prevent="handleLogin" class="login-form">
          <label class="form-label">
            <span>E-post</span>
            <input
              v-model="email"
              type="email"
              name="email"
              autocomplete="email"
              placeholder="din@email.se"
              class="form-input"
              :aria-describedby="error ? 'login-error' : undefined"
            />
          </label>

          <label class="form-label">
            <span>Lösenord</span>
            <input
              v-model="password"
              type="password"
              name="password"
              autocomplete="current-password"
              placeholder="Ditt lösenord"
              class="form-input"
              :aria-describedby="error ? 'login-error' : undefined"
            />
          </label>

          <p v-if="error" id="login-error" role="alert" class="form-error">{{ error }}</p>

          <BaseButton
            type="submit"
            variant="primary"
            size="lg"
            :disabled="!canSubmit"
            :loading="isLoading"
          >
            Logga in
          </BaseButton>
        </form>

        <div class="login-footer">
          <p>
            Har du inget konto?
            <RouterLink v-prefetch="'register'" to="/register">Skapa konto</RouterLink>
          </p>
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
.login-page {
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
.login-container {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 420px;
}

/* Top bar */
.login-top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
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
}

.back-link:hover {
  color: var(--accent);
  background: var(--bg-hover);
}

.back-arrow {
  transition: transform 0.3s ease;
}

.back-link:hover .back-arrow {
  transform: translateX(-4px);
}

/* Theme toggle */
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
  color: var(--text-secondary);
  transition: all 0.3s ease;
}

.theme-toggle:hover {
  background: var(--border-color-hover);
  color: var(--text-primary);
  transform: scale(1.1);
}

/* Login card */
.login-card {
  background: var(--bg-card);
  border-radius: 24px;
  padding: 2.5rem;
  box-shadow: var(--shadow-lg);
  text-align: center;
}

.logo-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
  animation: float 3s ease-in-out infinite;
}

.login-card h1 {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.75rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
}

.login-card > p {
  font-family: 'Nunito', sans-serif;
  font-size: 1rem;
  color: var(--text-secondary);
  margin: 0 0 2rem;
}

/* Form */
.login-form {
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

.form-input::placeholder {
  color: var(--text-muted);
}

.form-error {
  font-family: 'Nunito', sans-serif;
  font-size: 0.9rem;
  color: var(--error);
  margin: -0.5rem 0 1.5rem;
  padding: 0.5rem 0.75rem;
  background: var(--error-bg);
  border-radius: 8px;
  text-align: center;
}

.login-form .base-button {
  width: 100%;
  margin-top: 0.5rem;
}

/* Footer */
.login-footer {
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border-color);
}

.login-footer p {
  font-family: 'Nunito', sans-serif;
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin: 0;
}

.login-footer a {
  color: var(--accent);
  font-weight: 700;
  text-decoration: none;
}

.login-footer a:hover {
  text-decoration: underline;
}

/* Animation */
@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

/* Responsive */
@media (max-width: 480px) {
  .login-card {
    padding: 2rem 1.5rem;
  }
}
</style>
