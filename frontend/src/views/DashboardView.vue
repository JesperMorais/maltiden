<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDashboardStore } from '@/stores/dashboard'
import { useUserStore } from '@/stores/user'
import { usePlanningPreferencesStore } from '@/stores/planningPreferences'
import DashboardHeader from '@/components/dashboard/DashboardHeader.vue'
import TodaysMeal from '@/components/dashboard/TodaysMeal.vue'
import WeeklyMenuGrid from '@/components/dashboard/WeeklyMenuGrid.vue'
import QuickActions from '@/components/dashboard/QuickActions.vue'
import HouseholdWidget from '@/components/dashboard/HouseholdWidget.vue'
import ShoppingListWidget from '@/components/dashboard/ShoppingListWidget.vue'
import SettingsModal from '@/components/dashboard/SettingsModal.vue'
import DashboardSkeleton from '@/components/skeleton/layouts/DashboardSkeleton.vue'
import FadeContent from '@/components/vue-bits/FadeContent.vue'
import RotatingText from '@/components/vue-bits/RotatingText.vue'
import type { MenuDay } from '@/api/types/dashboard.types'

const greetingTexts = [
  'Vad blir det till middag?',
  'Planera veckans mat',
  'Dags att laga gott!',
  'Inspireras av nya recept',
]

const router = useRouter()
const dashboardStore = useDashboardStore()
const userStore = useUserStore()
const prefsStore = usePlanningPreferencesStore()

// Settings modal state
const showSettings = ref(false)

onMounted(() => {
  dashboardStore.fetchDashboard()
  userStore.initFromToken()
  prefsStore.initPreferences()
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

function handleViewRecipes() {
  router.push({ name: 'recipes' })
}

function handleParseRecipe() {
  router.push('/recipes/parse')
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
  router.push({ name: 'shopping-list' })
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
        <!-- Rotating greeting -->
        <div class="dashboard-greeting">
          <RotatingText
            :texts="greetingTexts"
            :rotation-interval="4500"
            split-by="words"
            :stagger-duration="0.03"
            main-class-name="greeting-text"
          />
        </div>

        <div class="dashboard-grid">
          <!-- Main content area -->
          <div class="main-area">
            <FadeContent :duration="800" :blur="true">
              <TodaysMeal
                :meal="dashboardStore.todaysMeal"
                :is-day-off="!prefsStore.isTodayActive"
                @click="handleMealClick"
              />
            </FadeContent>

            <FadeContent :duration="800" :delay="150" :blur="true">
              <WeeklyMenuGrid
                :weekly-menu="dashboardStore.weeklyMenu"
                @day-click="handleDayClick"
              />
            </FadeContent>
          </div>

          <!-- Sidebar -->
          <aside class="sidebar">
            <FadeContent :duration="600" :delay="200">
              <QuickActions
                @generate-menu="handleGenerateMenu"
                @view-recipes="handleViewRecipes"
                @invite-member="handleInviteMember"
                @parse-recipe="handleParseRecipe"
              />
            </FadeContent>

            <FadeContent :duration="600" :delay="300">
              <HouseholdWidget
                :members="dashboardStore.householdMembers"
                :invite-code="dashboardStore.inviteCode"
                @show-invite="handleShowInvite"
                @remove-member="handleRemoveMember"
              />
            </FadeContent>

            <FadeContent :duration="600" :delay="400">
              <ShoppingListWidget
                :shopping-list="dashboardStore.shoppingList"
                @view-list="handleViewShoppingList"
              />
            </FadeContent>
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

/* Dashboard greeting */
.dashboard-greeting {
  max-width: 1400px;
  margin: 0 auto;
  padding: 2rem 2rem 0;
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: clamp(1.25rem, 3vw, 1.75rem);
  color: var(--text-primary);
  min-height: 2.5em;
  display: flex;
  align-items: center;
}

:deep(.greeting-text) {
  overflow: hidden;
}

/* Dashboard content */
.dashboard-content {
  max-width: 1400px;
  margin: 0 auto;
  padding: 1rem 2rem 2rem;
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
