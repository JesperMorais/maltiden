<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDashboardStore } from '@/stores/dashboard'
import { useUserStore } from '@/stores/user'
import DashboardHeader from '@/components/dashboard/DashboardHeader.vue'
import TodaysMeal from '@/components/dashboard/TodaysMeal.vue'
import WeeklyMenuGrid from '@/components/dashboard/WeeklyMenuGrid.vue'
import QuickActions from '@/components/dashboard/QuickActions.vue'
import HouseholdWidget from '@/components/dashboard/HouseholdWidget.vue'
import ShoppingListWidget from '@/components/dashboard/ShoppingListWidget.vue'
import SettingsModal from '@/components/dashboard/SettingsModal.vue'
import DashboardSkeleton from '@/components/skeleton/layouts/DashboardSkeleton.vue'
import type { MenuDay } from '@/api/types/dashboard.types'

const router = useRouter()
const dashboardStore = useDashboardStore()
const userStore = useUserStore()

// Settings modal state
const showSettings = ref(false)

onMounted(() => {
  dashboardStore.fetchDashboard()
  userStore.initFromToken()
})

function handleDayClick(day: MenuDay) {
  console.log('Day clicked:', day)
  // TODO: Open day detail modal
}

function handleMealClick() {
  console.log('Today meal clicked')
  // TODO: Open recipe detail
}

function handleLogout() {
  userStore.logout()
}

function handleSettings() {
  showSettings.value = true
}

function handleCloseSettings() {
  showSettings.value = false
}

function handleGenerateMenu() {
  router.push({ name: 'generate-menu' })
}

function handleAddRecipe() {
  console.log('Add recipe')
  // TODO: Navigate to add recipe
}

function handleParseRecipe() {
  router.push({ name: 'parse-recipe' })
}

function handleInviteMember() {
  console.log('Invite member')
  // TODO: Show invite modal
}

function handleShowInvite() {
  console.log('Show invite code')
  // TODO: Show invite modal
}

function handleRemoveMember(memberId: string) {
  console.log('Remove member:', memberId)
  // TODO: Call API to remove member from household
}

function handleViewShoppingList() {
  console.log('View shopping list')
  // TODO: Navigate to shopping list
}
</script>

<template>
  <div class="dashboard-page">
    <!-- Skeleton loading state -->
    <DashboardSkeleton v-if="dashboardStore.isLoading" />

    <!-- Error state -->
    <div v-else-if="dashboardStore.error" class="error-state">
      <div class="error-content">
        <span class="error-icon">😅</span>
        <h2>Något gick fel</h2>
        <p>{{ dashboardStore.error }}</p>
        <button class="retry-btn" @click="dashboardStore.fetchDashboard(true)">
          Försök igen
        </button>
      </div>
    </div>

    <!-- Dashboard content -->
    <template v-else-if="dashboardStore.dashboardData">
      <DashboardHeader
        :household-name="dashboardStore.householdName"
        :user-name="userStore.userName"
        :user-role="userStore.userRole || 'guest'"
        @logout="handleLogout"
        @settings="handleSettings"
      />

      <main class="dashboard-content">
        <div class="dashboard-grid">
          <!-- Main content area -->
          <div class="main-area">
            <TodaysMeal
              :meal="dashboardStore.todaysMeal"
              @click="handleMealClick"
            />

            <WeeklyMenuGrid
              :weekly-menu="dashboardStore.weeklyMenu"
              @day-click="handleDayClick"
            />
          </div>

          <!-- Sidebar -->
          <aside class="sidebar">
            <QuickActions
              @generate-menu="handleGenerateMenu"
              @add-recipe="handleAddRecipe"
              @invite-member="handleInviteMember"
              @parse-recipe="handleParseRecipe"
            />

            <HouseholdWidget
              :members="dashboardStore.householdMembers"
              :invite-code="dashboardStore.inviteCode"
              @show-invite="handleShowInvite"
              @remove-member="handleRemoveMember"
            />

            <ShoppingListWidget
              :shopping-list="dashboardStore.shoppingList"
              @view-list="handleViewShoppingList"
            />
          </aside>
        </div>
      </main>

      <!-- Settings Modal -->
      <SettingsModal
        :user="userStore.currentUser"
        :is-open="showSettings"
        @close="handleCloseSettings"
        @logout="handleLogout"
      />
    </template>
  </div>
</template>

<style scoped>
.dashboard-page {
  min-height: 100vh;
  background: var(--bg-secondary);
}

/* Error state */
.error-state {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-primary);
  padding: 2rem;
}

.error-content {
  text-align: center;
  max-width: 400px;
}

.error-icon {
  font-size: 4rem;
  display: block;
  margin-bottom: 1rem;
}

.error-content h2 {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.75rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
}

.error-content p {
  font-family: 'Nunito', sans-serif;
  color: var(--text-secondary);
  margin: 0 0 1.5rem;
}

.retry-btn {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1rem;
  color: white;
  background: var(--accent);
  border: none;
  border-radius: 100px;
  padding: 0.85em 2em;
  cursor: pointer;
  transition: all 0.3s ease;
}

.retry-btn:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-accent);
}

/* Dashboard content */
.dashboard-content {
  max-width: 1400px;
  margin: 0 auto;
  padding: 2rem;
}

.dashboard-grid {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 2rem;
  align-items: start;
}

.main-area {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.sidebar {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  position: sticky;
  top: calc(70px + 2rem); /* Header height + padding */
}

/* Responsive */
@media (max-width: 1024px) {
  .dashboard-grid {
    grid-template-columns: 1fr;
  }

  .sidebar {
    position: static;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 1rem;
  }
}

@media (max-width: 768px) {
  .dashboard-content {
    padding: 1rem;
  }

  .sidebar {
    grid-template-columns: 1fr;
  }
}
</style>
