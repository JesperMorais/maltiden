<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import BaseButton from '@/components/common/BaseButton.vue'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const email = ref('')
const password = ref('')
const isLoading = ref(false)
const error = ref('')

const canSubmit = computed(() =>
  email.value.includes('@') && password.value.length >= 8
)

async function handleLogin() {
  if (!canSubmit.value) return

  isLoading.value = true
  error.value = ''

  // Mock login - simulate API call
  await new Promise(resolve => setTimeout(resolve, 1000))

  // Mock: Accept any valid-looking credentials
  if (email.value && password.value) {
    // Set mock user data
    userStore.setUser({
      id: 'user-1',
      name: email.value.split('@')[0],
      email: email.value,
      role: 'member'
    })

    router.push('/dashboard')
  } else {
    error.value = 'Fel e-post eller lösenord'
  }

  isLoading.value = false
}
</script>

<template>
  <main class="login-page">
    <!-- Background decorations -->
    <div class="bg-decorations">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="grain"></div>
    </div>

    <div class="login-container">
      <!-- Back link -->
      <RouterLink to="/" class="back-link">
        <span class="back-arrow">←</span>
        <span>Tillbaka</span>
      </RouterLink>

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
              placeholder="din@email.se"
              class="form-input"
            />
          </label>

          <label class="form-label">
            <span>Lösenord</span>
            <input
              v-model="password"
              type="password"
              placeholder="Ditt lösenord"
              class="form-input"
            />
          </label>

          <p v-if="error" class="form-error">{{ error }}</p>

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
            <RouterLink to="/register">Skapa konto</RouterLink>
          </p>
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Nunito:wght@400;600;700;800&family=Fraunces:wght@700;800&display=swap');

.login-page {
  --coral: #ff6b5b;
  --coral-dark: #e85a4a;
  --coral-light: #ff8a7d;
  --peach: #ffb599;
  --cream: #fff8f0;
  --warm-white: #fffcf7;
  --text-dark: #3d2c29;
  --text-muted: #6b5a56;
  --yellow-soft: #ffd93d;

  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  background: linear-gradient(
    165deg,
    var(--cream) 0%,
    var(--warm-white) 50%,
    #fff5eb 100%
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
  background: linear-gradient(135deg, var(--peach) 0%, var(--coral-light) 100%);
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

/* Back link */
.back-link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--text-muted);
  text-decoration: none;
  padding: 0.5rem 1rem;
  border-radius: 100px;
  transition: all 0.3s ease;
  margin-bottom: 2rem;
}

.back-link:hover {
  color: var(--coral);
  background: rgba(255, 107, 91, 0.1);
}

.back-arrow {
  transition: transform 0.3s ease;
}

.back-link:hover .back-arrow {
  transform: translateX(-4px);
}

/* Login card */
.login-card {
  background: var(--warm-white);
  border-radius: 24px;
  padding: 2.5rem;
  box-shadow:
    0 20px 60px rgba(61, 44, 41, 0.08),
    0 0 0 1px rgba(255, 107, 91, 0.08);
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
  color: var(--text-dark);
  margin: 0 0 0.5rem;
}

.login-card > p {
  font-family: 'Nunito', sans-serif;
  font-size: 1rem;
  color: var(--text-muted);
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
  color: var(--text-dark);
  margin-bottom: 0.5rem;
}

.form-input {
  width: 100%;
  padding: 0.9rem 1.25rem;
  border: 2px solid rgba(61, 44, 41, 0.1);
  border-radius: 14px;
  font-family: 'Nunito', sans-serif;
  font-size: 1rem;
  color: var(--text-dark);
  background: white;
  transition: all 0.3s ease;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: var(--coral);
  box-shadow: 0 0 0 4px rgba(255, 107, 91, 0.1);
}

.form-input::placeholder {
  color: #bbb;
}

.form-error {
  font-family: 'Nunito', sans-serif;
  font-size: 0.9rem;
  color: #e53e3e;
  margin: -0.5rem 0 1.5rem;
  padding: 0.5rem 0.75rem;
  background: rgba(229, 62, 62, 0.1);
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
  border-top: 1px solid rgba(61, 44, 41, 0.08);
}

.login-footer p {
  font-family: 'Nunito', sans-serif;
  font-size: 0.9rem;
  color: var(--text-muted);
  margin: 0;
}

.login-footer a {
  color: var(--coral);
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
