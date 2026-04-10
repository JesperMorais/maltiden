<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDashboardStore } from '@/stores/dashboard'
import { useUserStore } from '@/stores/user'
import { usePlanningPreferencesStore } from '@/stores/planningPreferences'
import { createInvite, removeMember } from '@/api/household.api'
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

// Invite modal state
const showInviteModal = ref(false)
const generatedInviteCode = ref('')
const isGeneratingInvite = ref(false)
const inviteExpiresAt = ref('')

onMounted(() => {
  dashboardStore.fetchDashboard()
  userStore.initFromToken()
  prefsStore.initPreferences()
})

function handleDayClick(day: MenuDay) {
  if (day.meal) {
    router.push({ name: 'recipes' })
  }
}

function handleMealClick() {
  const meal = dashboardStore.todaysMeal
  if (meal) {
    router.push({ name: 'recipes' })
  }
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

async function handleInviteMember() {
  await generateAndShowInvite()
}

async function handleShowInvite() {
  await generateAndShowInvite()
}

async function generateAndShowInvite() {
  isGeneratingInvite.value = true
  showInviteModal.value = true

  try {
    const response = await createInvite()
    generatedInviteCode.value = response.code
    inviteExpiresAt.value = new Date(response.expiresAt).toLocaleDateString('sv-SE')
  } catch {
    generatedInviteCode.value = ''
  } finally {
    isGeneratingInvite.value = false
  }
}

function closeInviteModal() {
  showInviteModal.value = false
  generatedInviteCode.value = ''
}

async function handleRemoveMember(memberId: string) {
  try {
    await removeMember(memberId)
    // Refresh dashboard to reflect the change
    await dashboardStore.fetchDashboard(true)
  } catch (e) {
    console.error('Failed to remove member:', e)
  }
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

      <!-- Invite Modal -->
      <Teleport to="body">
        <Transition name="modal">
          <div v-if="showInviteModal" class="invite-overlay" @click.self="closeInviteModal">
            <div class="invite-dialog">
              <button class="invite-close" @click="closeInviteModal" aria-label="Stäng">x</button>
              <h3 class="invite-title">Bjud in till hushållet</h3>
              <template v-if="isGeneratingInvite">
                <p class="invite-loading">Skapar inbjudningskod...</p>
              </template>
              <template v-else-if="generatedInviteCode">
                <p class="invite-description">Dela denna kod med den du vill bjuda in:</p>
                <div class="invite-code-display">{{ generatedInviteCode }}</div>
                <p class="invite-expires">Giltig till {{ inviteExpiresAt }}</p>
              </template>
              <template v-else>
                <p class="invite-error">Kunde inte skapa inbjudningskod. Försök igen.</p>
              </template>
            </div>
          </div>
        </Transition>
      </Teleport>
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

<!-- Non-scoped styles for teleported invite modal -->
<style>
.invite-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 1rem;
}

.invite-dialog {
  background: var(--bg-primary);
  border-radius: 20px;
  padding: 2rem;
  max-width: 400px;
  width: 100%;
  text-align: center;
  box-shadow: var(--shadow-lg);
  border: 1px solid var(--border-color);
  position: relative;
}

.invite-close {
  position: absolute;
  top: 1rem;
  right: 1rem;
  background: none;
  border: none;
  font-size: 1.25rem;
  color: var(--text-secondary);
  cursor: pointer;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}

.invite-close:hover {
  background: var(--bg-hover);
}

.invite-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.25rem;
  color: var(--text-primary);
  margin: 0 0 1rem;
}

.invite-description {
  font-family: 'Nunito', sans-serif;
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin: 0 0 1rem;
}

.invite-code-display {
  font-family: 'Courier New', monospace;
  font-weight: 800;
  font-size: 2rem;
  letter-spacing: 0.15em;
  color: var(--accent);
  background: var(--bg-hover);
  border-radius: 12px;
  padding: 1rem;
  margin: 0 0 0.75rem;
  user-select: all;
}

.invite-expires {
  font-family: 'Nunito', sans-serif;
  font-size: 0.8rem;
  color: var(--text-secondary);
  margin: 0;
}

.invite-loading {
  font-family: 'Nunito', sans-serif;
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin: 1rem 0;
}

.invite-error {
  font-family: 'Nunito', sans-serif;
  font-size: 0.9rem;
  color: var(--error);
  margin: 1rem 0;
}

.modal-enter-active,
.modal-leave-active {
  transition: all 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
