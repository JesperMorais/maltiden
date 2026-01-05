<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import type { UserRole } from '@/api/types/dashboard.types'

interface Props {
  householdName: string
  userName: string
  userRole: UserRole
}

defineProps<Props>()

const emit = defineEmits<{
  logout: []
  settings: []
}>()

const isDropdownOpen = ref(false)

function toggleDropdown() {
  isDropdownOpen.value = !isDropdownOpen.value
}

function closeDropdown() {
  isDropdownOpen.value = false
}

function handleLogout() {
  closeDropdown()
  emit('logout')
}

function handleSettings() {
  closeDropdown()
  emit('settings')
}
</script>

<template>
  <header class="dashboard-header">
    <div class="header-content">
      <!-- Logo (link to landing) -->
      <RouterLink to="/" class="logo">
        <span class="logo-icon">🍽️</span>
        <span class="logo-text">Måltiden</span>
      </RouterLink>

      <!-- Household name -->
      <div class="household-badge">
        <span class="household-icon">🏠</span>
        <span class="household-name">{{ householdName }}</span>
      </div>

      <!-- User menu -->
      <div class="user-menu" v-click-outside="closeDropdown">
        <button class="user-button" @click="toggleDropdown">
          <div class="avatar">
            <span>{{ userName.charAt(0).toUpperCase() }}</span>
          </div>
          <span class="user-name">{{ userName }}</span>
          <span class="role-badge" :class="userRole">
            {{ userRole === 'owner' ? 'Ägare' : userRole === 'member' ? 'Medlem' : 'Gäst' }}
          </span>
          <svg class="chevron" :class="{ open: isDropdownOpen }" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>

        <Transition name="dropdown">
          <div v-if="isDropdownOpen" class="dropdown">
            <button class="dropdown-item" @click="handleSettings">
              <span class="dropdown-icon">⚙️</span>
              <span>Inställningar</span>
            </button>
            <div class="dropdown-divider"></div>
            <button class="dropdown-item logout" @click="handleLogout">
              <span class="dropdown-icon">👋</span>
              <span>Logga ut</span>
            </button>
          </div>
        </Transition>
      </div>
    </div>
  </header>
</template>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Nunito:wght@600;700;800&family=Fraunces:wght@700&display=swap');

.dashboard-header {
  --coral: #ff6b5b;
  --coral-light: #ff8a7d;
  --peach: #ffb599;
  --cream: #fff8f0;
  --warm-white: #fffcf7;
  --text-dark: #3d2c29;
  --text-muted: #6b5a56;

  background: var(--warm-white);
  border-bottom: 1px solid rgba(61, 44, 41, 0.08);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-content {
  max-width: 1400px;
  margin: 0 auto;
  padding: 1rem 2rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

/* Logo */
.logo {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  text-decoration: none;
  padding: 0.5rem 0.75rem;
  margin: -0.5rem;
  border-radius: 12px;
  transition: all 0.2s ease;
}

.logo:hover {
  background: rgba(255, 107, 91, 0.08);
}

.logo-icon {
  font-size: 1.75rem;
}

.logo-text {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.35rem;
  color: var(--text-dark);
}

/* Household badge */
.household-badge {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: linear-gradient(135deg, rgba(255, 107, 91, 0.08) 0%, rgba(255, 181, 153, 0.08) 100%);
  border-radius: 100px;
  border: 1px solid rgba(255, 107, 91, 0.1);
}

.household-icon {
  font-size: 1rem;
}

.household-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--text-dark);
}

/* User menu */
.user-menu {
  position: relative;
}

.user-button {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem;
  padding-right: 1rem;
  background: white;
  border: 1px solid rgba(61, 44, 41, 0.1);
  border-radius: 100px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.user-button:hover {
  border-color: var(--coral);
  box-shadow: 0 4px 12px rgba(255, 107, 91, 0.15);
}

.avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--coral) 0%, var(--peach) 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 1rem;
  color: white;
}

.user-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--text-dark);
}

.role-badge {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.65rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 0.2rem 0.5rem;
  border-radius: 100px;
}

.role-badge.owner {
  background: linear-gradient(135deg, #f6ad55 0%, #ed8936 100%);
  color: white;
}

.role-badge.member {
  background: linear-gradient(135deg, var(--coral) 0%, var(--peach) 100%);
  color: white;
}

.role-badge.guest {
  background: rgba(61, 44, 41, 0.1);
  color: var(--text-muted);
}

.chevron {
  width: 16px;
  height: 16px;
  color: var(--text-muted);
  transition: transform 0.3s ease;
}

.chevron.open {
  transform: rotate(180deg);
}

/* Dropdown */
.dropdown {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  min-width: 180px;
  background: white;
  border-radius: 16px;
  box-shadow:
    0 10px 40px rgba(61, 44, 41, 0.15),
    0 4px 12px rgba(61, 44, 41, 0.1);
  border: 1px solid rgba(61, 44, 41, 0.08);
  padding: 0.5rem;
  overflow: hidden;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
  padding: 0.75rem 1rem;
  background: none;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--text-dark);
  transition: all 0.2s ease;
}

.dropdown-item:hover {
  background: rgba(255, 107, 91, 0.08);
  color: var(--coral);
}

.dropdown-item.logout:hover {
  background: rgba(229, 62, 62, 0.08);
  color: #e53e3e;
}

.dropdown-icon {
  font-size: 1rem;
}

.dropdown-divider {
  height: 1px;
  background: rgba(61, 44, 41, 0.08);
  margin: 0.25rem 0.5rem;
}

/* Dropdown animation */
.dropdown-enter-active {
  transition: all 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.dropdown-leave-active {
  transition: all 0.15s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-8px) scale(0.95);
}

/* Responsive */
@media (max-width: 768px) {
  .header-content {
    padding: 0.75rem 1rem;
  }

  .logo-text {
    display: none;
  }

  .household-badge {
    flex: 1;
    justify-content: center;
  }

  .user-name,
  .role-badge {
    display: none;
  }

  .user-button {
    padding: 0.35rem;
  }
}
</style>
