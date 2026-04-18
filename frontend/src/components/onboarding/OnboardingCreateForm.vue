<script setup lang="ts">
import { ref, computed } from 'vue'
import { Eye, EyeOff } from 'lucide-vue-next'
import { useRouter } from 'vue-router'
import BaseButton from '@/components/common/BaseButton.vue'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const emit = defineEmits<{
  success: []
}>()

const createForm = ref({
  name: '',
  lastName: '',
  email: '',
  password: '',
  passwordConfirm: '',
})

const showCreatePassword = ref(false)
const showCreatePasswordConfirm = ref(false)
const createError = ref('')
const isSubmitting = ref(false)

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

const passwordStrongEnoughCreate = computed(
  () =>
    createForm.value.password.length >= 8 && passwordCharTypes(createForm.value.password) >= 3,
)

const passwordsMatchCreate = computed(
  () => createForm.value.password === createForm.value.passwordConfirm,
)

const canSubmitCreate = computed(
  () =>
    createForm.value.name.length >= 2 &&
    createForm.value.email.includes('@') &&
    passwordStrongEnoughCreate.value &&
    passwordsMatchCreate.value,
)

async function handleCreate() {
  if (!canSubmitCreate.value) return
  isSubmitting.value = true
  createError.value = ''

  // Call real backend registration
  const success = await userStore.register(
    createForm.value.name,
    createForm.value.email,
    createForm.value.password,
    createForm.value.lastName || undefined,
  )

  isSubmitting.value = false

  if (success) {
    emit('success')
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
  <form class="form-card" @submit.prevent="handleCreate">
    <div class="form-row">
      <label class="form-label">
        <span>Förnamn</span>
        <input
          v-model="createForm.name"
          type="text"
          name="given-name"
          autocomplete="given-name"
          placeholder="Anna"
          class="form-input"
        />
      </label>

      <label class="form-label">
        <span>Efternamn <span class="field-hint-inline">(valfritt)</span></span>
        <input
          v-model="createForm.lastName"
          type="text"
          name="family-name"
          autocomplete="family-name"
          placeholder="Andersson"
          class="form-input"
        />
      </label>
    </div>

    <label class="form-label">
      <span>E-post</span>
      <input
        v-model="createForm.email"
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
          v-model="createForm.password"
          :type="showCreatePassword ? 'text' : 'password'"
          name="password"
          autocomplete="new-password"
          placeholder="Minst 8 tecken"
          class="form-input"
        />
        <button
          type="button"
          class="password-toggle"
          :aria-label="showCreatePassword ? 'Dölj lösenord' : 'Visa lösenord'"
          @click="showCreatePassword = !showCreatePassword"
        >
          <component :is="showCreatePassword ? EyeOff : Eye" :size="18" :stroke-width="2" />
        </button>
      </div>
      <span
        v-if="createForm.password.length > 0 && !passwordStrongEnoughCreate"
        class="field-hint"
      >
        Minst 8 tecken med minst 3 av: versaler, gemener, siffror, specialtecken
      </span>
    </label>

    <label class="form-label">
      <span>Bekräfta lösenord</span>
      <div class="password-input-wrapper">
        <input
          v-model="createForm.passwordConfirm"
          :type="showCreatePasswordConfirm ? 'text' : 'password'"
          autocomplete="new-password"
          placeholder="Skriv lösenordet igen"
          class="form-input"
          :class="{ 'input-error': createForm.passwordConfirm && !passwordsMatchCreate }"
        />
        <button
          type="button"
          class="password-toggle"
          :aria-label="showCreatePasswordConfirm ? 'Dölj lösenord' : 'Visa lösenord'"
          @click="showCreatePasswordConfirm = !showCreatePasswordConfirm"
        >
          <component :is="showCreatePasswordConfirm ? EyeOff : Eye" :size="18" :stroke-width="2" />
        </button>
      </div>
      <span v-if="createForm.passwordConfirm && !passwordsMatchCreate" class="field-error">
        Lösenorden matchar inte
      </span>
    </label>

    <p v-if="createError" id="create-error" role="alert" class="form-error">{{ createError }}</p>

    <BaseButton
      type="submit"
      variant="primary"
      size="lg"
      :disabled="!canSubmitCreate"
      :loading="isSubmitting"
    >
      Skapa konto
    </BaseButton>

    <p class="form-terms">
      Genom att skapa konto godkänner du våra
      <a href="#">villkor</a> och <a href="#">integritetspolicy</a>.
    </p>
  </form>
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

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-row .form-label {
  margin-bottom: 1.25rem;
}

.field-hint-inline {
  font-weight: 400;
  color: var(--text-secondary);
  opacity: 0.8;
  font-size: 0.8rem;
  margin-left: 0.25rem;
}

@media (max-width: 520px) {
  .form-row {
    grid-template-columns: 1fr;
    gap: 0;
  }
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

/* Responsive */
@media (max-width: 768px) {
  .form-card {
    padding: 1.5rem;
  }
}
</style>
