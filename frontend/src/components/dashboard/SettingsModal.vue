<script setup lang="ts">
import { ref } from 'vue'
import type { User } from '@/api/types/dashboard.types'
import { useThemeStore } from '@/stores/theme'

interface Props {
  user: User | null
  isOpen: boolean
}

const props = defineProps<Props>()

const emit = defineEmits<{
  close: []
  logout: []
}>()

// Theme store for dark mode
const themeStore = useThemeStore()

// Settings state
const notificationsEnabled = ref(true)
const mealReminders = ref(true)
const shoppingReminders = ref(true)

function handleClose() {
  emit('close')
}

function handleLogout() {
  emit('logout')
}

function handleOverlayClick(e: MouseEvent) {
  if (e.target === e.currentTarget) {
    handleClose()
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="isOpen" class="settings-overlay" @click="handleOverlayClick">
        <div class="settings-modal">
          <!-- Header -->
          <div class="modal-header">
            <h2>Inställningar</h2>
            <button class="close-btn" @click="handleClose">
              <span>×</span>
            </button>
          </div>

          <!-- Content -->
          <div class="modal-content">
            <!-- Profile Section -->
            <section class="settings-section">
              <h3>
                <span class="section-icon">👤</span>
                Profil
              </h3>
              <div class="profile-card">
                <div class="profile-avatar">
                  {{ user?.name?.charAt(0).toUpperCase() || '?' }}
                </div>
                <div class="profile-info">
                  <p class="profile-name">{{ user?.name || 'Okänd' }}</p>
                  <p class="profile-email">{{ user?.email || 'Ingen e-post' }}</p>
                  <span class="profile-role" :class="user?.role">
                    {{ user?.role === 'owner' ? 'Ägare' : user?.role === 'member' ? 'Medlem' : 'Gäst' }}
                  </span>
                </div>
              </div>
            </section>

            <!-- Notifications Section -->
            <section class="settings-section">
              <h3>
                <span class="section-icon">🔔</span>
                Notifikationer
              </h3>
              <div class="settings-options">
                <label class="toggle-option">
                  <span class="option-label">
                    <span class="option-title">Aktivera notifikationer</span>
                    <span class="option-desc">Få påminnelser om måltider</span>
                  </span>
                  <input
                    type="checkbox"
                    v-model="notificationsEnabled"
                    class="toggle-input"
                  />
                  <span class="toggle-slider"></span>
                </label>

                <label class="toggle-option" :class="{ disabled: !notificationsEnabled }">
                  <span class="option-label">
                    <span class="option-title">Måltidspåminnelser</span>
                    <span class="option-desc">Påminn mig när det är dags att laga mat</span>
                  </span>
                  <input
                    type="checkbox"
                    v-model="mealReminders"
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
                    type="checkbox"
                    v-model="shoppingReminders"
                    :disabled="!notificationsEnabled"
                    class="toggle-input"
                  />
                  <span class="toggle-slider"></span>
                </label>
              </div>
            </section>

            <!-- Appearance Section -->
            <section class="settings-section">
              <h3>
                <span class="section-icon">🎨</span>
                Utseende
              </h3>
              <div class="settings-options">
                <label class="toggle-option">
                  <span class="option-label">
                    <span class="option-title">Mörkt läge</span>
                    <span class="option-desc">Använd mörkt färgschema</span>
                  </span>
                  <input
                    type="checkbox"
                    :checked="themeStore.isDarkMode"
                    @change="themeStore.toggleDarkMode()"
                    class="toggle-input"
                  />
                  <span class="toggle-slider"></span>
                </label>
              </div>
              <p class="coming-soon">Fler teman kommer snart!</p>
            </section>

            <!-- Account Section -->
            <section class="settings-section">
              <h3>
                <span class="section-icon">⚙️</span>
                Konto
              </h3>
              <div class="account-actions">
                <button class="action-btn logout-btn" @click="handleLogout">
                  <span class="btn-icon">👋</span>
                  Logga ut
                </button>
              </div>
            </section>
          </div>

          <!-- Footer -->
          <div class="modal-footer">
            <p class="version">Måltiden v1.0.0</p>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.settings-overlay {
  position: fixed;
  inset: 0;
  background: var(--overlay-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.settings-modal {
  background: var(--bg-primary);
  border-radius: 24px;
  width: 100%;
  max-width: 480px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: var(--shadow-lg);
  border: 1px solid var(--border-color);
}

/* Header */
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.5rem 2rem;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h2 {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.5rem;
  color: var(--text-primary);
  margin: 0;
}

.close-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: var(--bg-hover);
  color: var(--text-secondary);
  font-size: 1.5rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: var(--bg-hover);
  color: var(--accent);
}

/* Content */
.modal-content {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem 2rem;
}

/* Sections */
.settings-section {
  margin-bottom: 2rem;
}

.settings-section:last-child {
  margin-bottom: 0;
}

.settings-section h3 {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 0 0 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.section-icon {
  font-size: 1rem;
}

/* Profile Card */
.profile-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
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
  color: white;
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
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 0.25rem 0.75rem;
  border-radius: 100px;
}

.profile-role.owner {
  background: var(--role-owner-bg);
  color: white;
}

.profile-role.member {
  background: var(--role-member-bg);
  color: white;
}

.profile-role.guest {
  background: var(--role-guest-bg);
  color: var(--role-guest-text);
}

/* Toggle Options */
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

/* Toggle Switch */
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

/* Account Actions */
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

.logout-btn {
  background: rgba(229, 62, 62, 0.1);
  color: var(--error);
}

.logout-btn:hover {
  background: var(--error);
  color: white;
  transform: translateY(-2px);
}

.btn-icon {
  font-size: 1.1rem;
}

/* Footer */
.modal-footer {
  padding: 1rem 2rem;
  border-top: 1px solid var(--border-color);
  text-align: center;
}

.version {
  font-family: 'Nunito', sans-serif;
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin: 0;
}

/* Modal animation */
.modal-enter-active {
  transition: all 0.3s ease;
}

.modal-leave-active {
  transition: all 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .settings-modal,
.modal-leave-to .settings-modal {
  transform: scale(0.95) translateY(20px);
}

/* Responsive */
@media (max-width: 520px) {
  .settings-modal {
    max-height: 100vh;
    border-radius: 0;
  }

  .modal-header,
  .modal-content,
  .modal-footer {
    padding-left: 1.5rem;
    padding-right: 1.5rem;
  }
}
</style>
