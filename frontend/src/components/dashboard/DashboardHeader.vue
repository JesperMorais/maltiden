<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import type { UserRole } from '@/api/types/dashboard.types'
import { Home, Settings, LogOut, UtensilsCrossed } from 'lucide-vue-next'

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
      <RouterLink v-prefetch="'landing'" to="/" class="logo">
        <UtensilsCrossed :size="20" :stroke-width="2" class="logo-icon" />
        <span class="logo-text">Måltiden</span>
      </RouterLink>

      <!-- Household name -->
      <div class="household-badge">
        <span class="household-icon"><Home :size="16" /></span>
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
              <span class="dropdown-icon"><Settings :size="18" /></span>
              <span>Inställningar</span>
            </button>
            <div class="dropdown-divider"></div>
            <button class="dropdown-item logout" @click="handleLogout">
              <span class="dropdown-icon"><LogOut :size="18" /></span>
              <span>Logga ut</span>
            </button>
          </div>
        </Transition>
      </div>
    </div>
  </header>
</template>

<style scoped>
.dashboard-header {
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border-color);
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
  background: var(--bg-hover);
}

.logo-icon {
  color: var(--accent);
}

.logo-text {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.35rem;
  color: var(--text-primary);
}

/* Household badge */
.household-badge {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: var(--bg-hover);
  border-radius: 100px;
  border: 1px solid var(--border-color);
}

.household-icon {
  font-size: 1rem;
}

.household-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--text-primary);
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
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 100px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.user-button:hover {
  border-color: var(--accent);
  box-shadow: var(--shadow-md);
}

.avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--accent-gradient);
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 1rem;
  color: var(--text-on-accent);
}

.user-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--text-primary);
}

.role-badge {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 0.2rem 0.5rem;
  border-radius: 100px;
}

.role-badge.owner {
  background: var(--role-owner-bg);
  color: var(--text-on-accent);
}

.role-badge.member {
  background: var(--role-member-bg);
  color: var(--text-on-accent);
}

.role-badge.guest {
  background: var(--role-guest-bg);
  color: var(--role-guest-text);
}

.chevron {
  width: 16px;
  height: 16px;
  color: var(--text-secondary);
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
  background: var(--bg-card);
  border-radius: 16px;
  box-shadow: var(--shadow-lg);
  border: 1px solid var(--border-color);
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
  color: var(--text-primary);
  transition: all 0.2s ease;
}

.dropdown-item:hover {
  background: var(--bg-hover);
  color: var(--accent);
}

.dropdown-item.logout:hover {
  background: var(--error-bg);
  color: var(--error);
}

.dropdown-icon {
  font-size: 1rem;
}

.dropdown-divider {
  height: 1px;
  background: var(--border-color);
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
    min-height: 44px;
    min-width: 44px;
    justify-content: center;
  }
}
</style>
