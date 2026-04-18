<script setup lang="ts">
import { ref, computed } from 'vue'
import { useThemeStore } from '@/stores/theme'
import { useDashboardStore } from '@/stores/dashboard'
import { useUserStore } from '@/stores/user'
import BackLink from '@/components/common/BackLink.vue'
import {
  User as UserIcon,
  Bell,
  Palette,
  Settings as SettingsIcon,
  LogOut,
  MessageCircle,
  Home,
  Check,
  X,
} from 'lucide-vue-next'

const themeStore = useThemeStore()
const dashboardStore = useDashboardStore()
const userStore = useUserStore()

const user = computed(() => userStore.currentUser)

// Ephemeral notification toggles — placeholder UI; not yet wired to backend.
const notificationsEnabled = ref(true)
const mealReminders = ref(true)
const shoppingReminders = ref(true)

// Household name editing
const isEditingHouseholdName = ref(false)
const draftHouseholdName = ref('')
const householdNameError = ref('')
const isSavingHouseholdName = ref(false)

const canEditHousehold = computed(
  () => userStore.isOwner || userStore.currentUser?.role === 'member',
)

function startEditHouseholdName() {
  draftHouseholdName.value = dashboardStore.householdName
  householdNameError.value = ''
  isEditingHouseholdName.value = true
}

function cancelEditHouseholdName() {
  isEditingHouseholdName.value = false
  householdNameError.value = ''
}

async function saveHouseholdName() {
  const name = draftHouseholdName.value.trim()
  if (!name) {
    householdNameError.value = 'Ange ett namn för hushållet'
    return
  }
  if (name.length > 100) {
    householdNameError.value = 'Namnet får vara högst 100 tecken'
    return
  }
  if (name === dashboardStore.householdName) {
    isEditingHouseholdName.value = false
    return
  }

  isSavingHouseholdName.value = true
  const ok = await dashboardStore.updateHouseholdName(name)
  isSavingHouseholdName.value = false

  if (ok) {
    isEditingHouseholdName.value = false
    householdNameError.value = ''
  } else {
    householdNameError.value = 'Kunde inte spara — försök igen'
  }
}

function handleLogout() {
  userStore.logout()
}
</script>

<template>
  <div class="account-view">
    <header class="page-header">
      <div class="header-content">
        <BackLink :to="{ name: 'dashboard' }" label="Dashboard" />
        <h1 class="title">Mitt konto</h1>
        <p class="description">Profil, hushåll och inställningar.</p>
      </div>
    </header>

    <main class="content">
      <div class="content-container">
        <!-- Profile -->
        <section class="settings-section">
          <h2>
            <span class="section-icon"><UserIcon :size="16" /></span>
            Profil
          </h2>
          <div class="profile-card">
            <div class="profile-avatar">
              {{ user?.name?.charAt(0).toUpperCase() || '?' }}
            </div>
            <div class="profile-info">
              <p class="profile-name">{{ user?.name || 'Okänd' }}</p>
              <p class="profile-email">{{ user?.email || 'Ingen e-post' }}</p>
              <span class="profile-role" :class="user?.role">
                {{
                  user?.role === 'owner'
                    ? 'Ägare'
                    : user?.role === 'member'
                      ? 'Medlem'
                      : 'Gäst'
                }}
              </span>
            </div>
          </div>
        </section>

        <!-- Household -->
        <section class="settings-section">
          <h2>
            <span class="section-icon"><Home :size="16" /></span>
            Hushåll
          </h2>
          <div class="household-card">
            <div v-if="!isEditingHouseholdName" class="household-display">
              <div class="household-info">
                <span class="household-label">Namn</span>
                <span class="household-name">{{ dashboardStore.householdName }}</span>
              </div>
              <button
                v-if="canEditHousehold"
                type="button"
                class="edit-btn"
                @click="startEditHouseholdName"
              >
                Ändra
              </button>
            </div>
            <div v-else class="household-edit">
              <label class="household-label" for="household-name-input">Namn</label>
              <div class="household-edit-row">
                <input
                  id="household-name-input"
                  v-model="draftHouseholdName"
                  type="text"
                  class="household-input"
                  maxlength="100"
                  placeholder="Ange hushållsnamn"
                  :disabled="isSavingHouseholdName"
                  @keydown.enter.prevent="saveHouseholdName"
                  @keydown.esc.prevent="cancelEditHouseholdName"
                />
                <button
                  type="button"
                  class="icon-btn save-btn"
                  :disabled="isSavingHouseholdName"
                  aria-label="Spara"
                  @click="saveHouseholdName"
                >
                  <Check :size="18" />
                </button>
                <button
                  type="button"
                  class="icon-btn cancel-btn"
                  :disabled="isSavingHouseholdName"
                  aria-label="Avbryt"
                  @click="cancelEditHouseholdName"
                >
                  <X :size="18" />
                </button>
              </div>
              <p v-if="householdNameError" role="alert" class="household-error">
                {{ householdNameError }}
              </p>
            </div>
          </div>
        </section>

        <!-- Notifications -->
        <section class="settings-section">
          <h2>
            <span class="section-icon"><Bell :size="16" /></span>
            Notifikationer
          </h2>
          <div class="settings-options">
            <label class="toggle-option">
              <span class="option-label">
                <span class="option-title">Aktivera notifikationer</span>
                <span class="option-desc">Få påminnelser om måltider</span>
              </span>
              <input v-model="notificationsEnabled" type="checkbox" class="toggle-input" />
              <span class="toggle-slider"></span>
            </label>

            <label class="toggle-option" :class="{ disabled: !notificationsEnabled }">
              <span class="option-label">
                <span class="option-title">Måltidspåminnelser</span>
                <span class="option-desc">Påminn mig när det är dags att laga mat</span>
              </span>
              <input
                v-model="mealReminders"
                type="checkbox"
                :disabled="!notificationsEnabled"
                class="toggle-input"
              />
              <span class="toggle-slider"></span>
            </label>

            <label class="toggle-option" :class="{ disabled: !notificationsEnabled }">
              <span class="option-label">
                <span class="option-title">Inköpspåminnelser</span>
                <span class="option-desc">Påminn mig om inköp som behöver göras</span>
              </span>
              <input
                v-model="shoppingReminders"
                type="checkbox"
                :disabled="!notificationsEnabled"
                class="toggle-input"
              />
              <span class="toggle-slider"></span>
            </label>
          </div>
        </section>

        <!-- Appearance -->
        <section class="settings-section">
          <h2>
            <span class="section-icon"><Palette :size="16" /></span>
            Utseende
          </h2>
          <div class="settings-options">
            <label class="toggle-option">
              <span class="option-label">
                <span class="option-title">Mörkt läge</span>
                <span class="option-desc">Använd mörkt färgschema</span>
              </span>
              <input
                type="checkbox"
                :checked="themeStore.isDarkMode"
                class="toggle-input"
                @change="themeStore.toggleDarkMode()"
              />
              <span class="toggle-slider"></span>
            </label>
          </div>
          <p class="coming-soon">Fler teman kommer snart!</p>
        </section>

        <!-- Feedback -->
        <section class="settings-section">
          <h2>
            <span class="section-icon"><MessageCircle :size="16" /></span>
            Feedback
          </h2>
          <div class="account-actions">
            <a
              href="mailto:maltiden.app@gmail.com?subject=Feedback%20-%20Måltiden"
              class="action-btn feedback-btn"
            >
              <span class="btn-icon"><MessageCircle :size="18" /></span>
              Skicka feedback
            </a>
          </div>
          <p class="coming-soon">Hjälp oss bli bättre — vi läser all feedback!</p>
        </section>

        <!-- Account -->
        <section class="settings-section">
          <h2>
            <span class="section-icon"><SettingsIcon :size="16" /></span>
            Konto
          </h2>
          <div class="account-actions">
            <button type="button" class="action-btn logout-btn" @click="handleLogout">
              <span class="btn-icon"><LogOut :size="18" /></span>
              Logga ut
            </button>
          </div>
        </section>

        <footer class="page-footer">
          <p class="version">Måltiden v1.0.0</p>
        </footer>
      </div>
    </main>
  </div>
</template>

<style scoped>
.account-view {
  min-height: 100vh;
  background: var(--bg-primary);
  display: flex;
  flex-direction: column;
}

.page-header {
  padding: 2rem 2rem 1.5rem;
  background: linear-gradient(180deg, var(--bg-card) 0%, var(--bg-primary) 100%);
  border-bottom: 1px solid var(--border-color);
}

.header-content {
  max-width: 680px;
  margin: 0 auto;
}

.header-content :deep(.back-link) {
  margin-bottom: 0.5rem;
  margin-left: -1rem;
}

.title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 2.5rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
  line-height: 1.2;
}

.description {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1.05rem;
  color: var(--text-secondary);
  margin: 0;
}

.content {
  flex: 1;
  padding: 2rem;
}

.content-container {
  max-width: 680px;
  margin: 0 auto;
}

/* Sections */
.settings-section {
  margin-bottom: 2rem;
}

.settings-section:last-of-type {
  margin-bottom: 1.5rem;
}

.settings-section h2 {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 0.9rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 0 0 0.85rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.section-icon {
  color: var(--accent);
}

/* Profile */
.profile-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem;
  background: var(--bg-card);
  border-radius: 16px;
  border: 1px solid var(--border-color);
}

.profile-avatar {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--accent-gradient);
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 1.5rem;
  color: var(--text-on-accent);
  flex-shrink: 0;
}

.profile-info {
  flex: 1;
  min-width: 0;
}

.profile-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1.1rem;
  color: var(--text-primary);
  margin: 0 0 0.25rem;
}

.profile-email {
  font-family: 'Nunito', sans-serif;
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin: 0 0 0.5rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.profile-role {
  display: inline-block;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 0.25rem 0.75rem;
  border-radius: 100px;
}

.profile-role.owner {
  background: var(--role-owner-bg);
  color: var(--text-on-accent);
}

.profile-role.member {
  background: var(--role-member-bg);
  color: var(--text-on-accent);
}

.profile-role.guest {
  background: var(--role-guest-bg);
  color: var(--role-guest-text);
}

/* Household */
.household-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  padding: 1rem;
}

.household-display {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.household-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.household-label {
  font-family: 'Nunito', sans-serif;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
  margin-bottom: 0.25rem;
}

.household-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1.05rem;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.edit-btn {
  border: 1px solid var(--border-color);
  background: var(--bg-primary);
  color: var(--text-primary);
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.85rem;
  padding: 0.5rem 0.9rem;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.edit-btn:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.household-edit {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.household-edit-row {
  display: flex;
  gap: 0.5rem;
  align-items: stretch;
}

.household-input {
  flex: 1;
  padding: 0.65rem 0.9rem;
  border: 2px solid var(--border-color);
  border-radius: 10px;
  font-family: 'Nunito', sans-serif;
  font-size: 0.95rem;
  color: var(--text-primary);
  background: var(--bg-primary);
  transition: border-color 0.2s ease;
  min-width: 0;
}

.household-input:focus {
  outline: none;
  border-color: var(--accent);
}

.household-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  border: 1px solid var(--border-color);
  background: var(--bg-primary);
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.icon-btn:hover:not(:disabled) {
  transform: translateY(-1px);
}

.icon-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.save-btn:hover:not(:disabled) {
  background: var(--accent);
  color: var(--text-on-accent);
  border-color: var(--accent);
}

.cancel-btn:hover:not(:disabled) {
  background: var(--error-bg);
  color: var(--error);
  border-color: var(--error);
}

.household-error {
  margin: 0;
  font-family: 'Nunito', sans-serif;
  font-size: 0.85rem;
  color: var(--error);
}

/* Toggles */
.settings-options {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.toggle-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem;
  background: var(--bg-card);
  border-radius: 14px;
  border: 1px solid var(--border-color);
  cursor: pointer;
  transition: all 0.2s ease;
}

.toggle-option:hover {
  border-color: var(--border-color-hover);
}

.toggle-option.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.option-label {
  flex: 1;
}

.option-title {
  display: block;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.95rem;
  color: var(--text-primary);
}

.option-desc {
  display: block;
  font-family: 'Nunito', sans-serif;
  font-size: 0.8rem;
  color: var(--text-secondary);
  margin-top: 0.25rem;
}

.toggle-input {
  display: none;
}

.toggle-slider {
  position: relative;
  width: 48px;
  height: 28px;
  background: var(--border-color);
  border-radius: 100px;
  transition: all 0.3s ease;
  flex-shrink: 0;
}

.toggle-slider::after {
  content: '';
  position: absolute;
  top: 3px;
  left: 3px;
  width: 22px;
  height: 22px;
  background: var(--bg-primary);
  border-radius: 50%;
  box-shadow: var(--shadow-sm);
  transition: all 0.3s ease;
}

.toggle-input:checked + .toggle-slider {
  background: var(--accent-gradient);
}

.toggle-input:checked + .toggle-slider::after {
  transform: translateX(20px);
}

.coming-soon {
  font-family: 'Nunito', sans-serif;
  font-size: 0.8rem;
  color: var(--text-secondary);
  font-style: italic;
  margin: 0.75rem 0 0;
  text-align: center;
}

/* Action buttons */
.account-actions {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  width: 100%;
  padding: 1rem;
  border: none;
  border-radius: 14px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.3s ease;
}

.feedback-btn {
  background: var(--accent-bg, #e8f4fd);
  color: var(--accent-text);
  text-decoration: none;
}

.feedback-btn:hover {
  background: var(--accent);
  color: var(--text-on-accent);
  transform: translateY(-2px);
}

.logout-btn {
  background: var(--error-bg);
  color: var(--error);
}

.logout-btn:hover {
  background: var(--error);
  color: var(--text-on-accent);
  transform: translateY(-2px);
}

.btn-icon {
  font-size: 1.1rem;
}

.page-footer {
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border-color);
  text-align: center;
}

.version {
  font-family: 'Nunito', sans-serif;
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin: 0;
}

/* Responsive */
@media (max-width: 640px) {
  .page-header {
    padding: 1.5rem 1rem 1rem;
  }

  .content {
    padding: 1.5rem 1rem;
  }

  .title {
    font-size: 1.75rem;
  }

  .description {
    font-size: 0.95rem;
  }
}
</style>
