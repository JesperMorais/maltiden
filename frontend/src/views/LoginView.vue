<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import BaseButton from '@/components/common/BaseButton.vue'
import BackLink from '@/components/common/BackLink.vue'
import { UtensilsCrossed } from 'lucide-vue-next'
import { useUserStore } from '@/stores/user'
import { useThemeStore } from '@/stores/theme'
import { useToast } from '@/composables/useToast'
import WavesBackground from '@/components/vue-bits/WavesBackground.vue'
import BaseThemeToggle from '@/components/common/BaseThemeToggle.vue'

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
        <BackLink to="/" />
        <BaseThemeToggle />
      </div>

      <!-- Login card -->
      <div class="login-card">
        <div class="logo-icon">
          <UtensilsCrossed :size="40" :stroke-width="1.75" />
        </div>
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
              aria-describedby="login-error"
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
              aria-describedby="login-error"
            />
          </label>

          <p v-show="error" id="login-error" role="alert" class="form-error">{{ error }}</p>

          <div class="forgot-password">
            <a
              href="mailto:maltiden.app@gmail.com?subject=Glömt%20lösenord%20-%20Måltiden&body=Hej!%20Jag%20har%20glömt%20mitt%20lösenord.%20Min%20e-post%3A%20"
              class="forgot-link"
            >
              Glömt lösenord? Mejla support
            </a>
          </div>

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

/* Login card */
.login-card {
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

.forgot-password {
  text-align: right;
  margin: -0.5rem 0 0.5rem;
}

.forgot-link {
  font-family: 'Nunito', sans-serif;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--accent-text);
  text-decoration: none;
}

.forgot-link:hover {
  text-decoration: underline;
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
  color: var(--accent-text);
  font-weight: 700;
  font-size: 1rem;
  text-decoration: none;
}

.login-footer a:hover {
  text-decoration: underline;
}

/* Responsive */
@media (max-width: 480px) {
  .login-card {
    padding: 2rem 1.5rem;
  }
}
</style>
