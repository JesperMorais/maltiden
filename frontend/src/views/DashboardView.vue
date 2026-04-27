<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
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
import InviteModal from '@/components/dashboard/InviteModal.vue'
import RecipeDetailModal from '@/components/recipes/RecipeDetailModal.vue'
import DashboardSkeleton from '@/components/skeleton/layouts/DashboardSkeleton.vue'
import ErrorState from '@/components/common/ErrorState.vue'
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

// Modal state
const showInvite = ref(false)
const selectedRecipeId = ref<string | null>(null)

const todayExtraPortions = computed(() => {
  const today = dashboardStore.todayFromMenu
  if (!today) return 0
  return dashboardStore.getDayLunchBoxCount(today.date)
})

onMounted(() => {
  dashboardStore.fetchDashboard()
  dashboardStore.initLunchBoxDays()
  dashboardStore.initDayMemberExclusions()
  prefsStore.initPreferences()
})

function handleViewRecipe(day: MenuDay) {
  selectedRecipeId.value = day.meal?.id ?? null
}

function handleMealClick() {
  selectedRecipeId.value = dashboardStore.todaysMeal?.id ?? null
}

function handleLogout() {
  userStore.logout()
}

function handleSettings() {
  router.push({ name: 'account' })
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
  showInvite.value = true
}

function handleShowInvite() {
  showInvite.value = true
}

function handleCloseInvite() {
  showInvite.value = false
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
        <ErrorState :description="dashboardStore.error" @retry="dashboardStore.fetchDashboard(true)" />
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
                :extra-portions="todayExtraPortions"
                @click="handleMealClick"
              />
            </FadeContent>

            <FadeContent :duration="800" :delay="150" :blur="true">
              <WeeklyMenuGrid
                :weekly-menu="dashboardStore.weeklyMenu"
                @view-recipe="handleViewRecipe"
              />
            </FadeContent>

            <FadeContent :duration="600" :delay="250">
              <HouseholdWidget
                :members="dashboardStore.householdMembers"
                :invite-code="dashboardStore.inviteCode"
                :selected-date="dashboardStore.selectedDate"
                horizontal
                @show-invite="handleShowInvite"
                @remove-member="handleRemoveMember"
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
              <ShoppingListWidget
                :shopping-list="dashboardStore.shoppingList"
                @view-list="handleViewShoppingList"
              />
            </FadeContent>
          </aside>
        </div>
      </main>

      <!-- Invite Modal -->
      <InviteModal
        :is-open="showInvite"
        @close="handleCloseInvite"
      />

      <!-- Recipe Detail Modal -->
      <RecipeDetailModal
        :recipe-id="selectedRecipeId"
        @close="selectedRecipeId = null"
        @updated="dashboardStore.fetchDashboard(true)"
        @deleted="selectedRecipeId = null; dashboardStore.fetchDashboard(true)"
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
  max-width: 400px;
}

/* Dashboard greeting */
.dashboard-greeting {
  max-width: 1400px;
  margin: 0 auto;
  padding: 2rem 2rem 1rem;
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: clamp(1.25rem, 3vw, 1.75rem);
  color: var(--text-primary);
  min-height: 3em;
  display: flex;
  align-items: center;
  overflow: hidden;
  contain: layout style;
}

:deep(.greeting-text) {
  overflow: hidden;
}

/* Dashboard content */
.dashboard-content {
  max-width: 1400px;
  margin: 0 auto;
  padding: 2rem 2rem 2rem;
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
  top: calc(var(--header-height) + 2rem);
  z-index: 1;
}

/* Responsive */
@media (max-width: 1024px) {
  .dashboard-grid {
    grid-template-columns: 1fr;
  }

  .sidebar {
    position: static;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }
}

@media (max-width: 768px) {
  .dashboard-greeting {
    padding: 0.75rem 0.75rem 0.25rem;
    font-size: 1.1rem;
    min-height: 2.5em;
  }

  .dashboard-content {
    padding: 0.75rem;
  }

  .dashboard-grid {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .main-area,
  .sidebar {
    display: contents;
  }

  .main-area :deep(> div),
  .sidebar :deep(> div) {
    opacity: 1 !important;
    filter: none !important;
    width: 100%;
  }

  /* Interleave: TodaysMeal(1) → QuickActions(2) → WeeklyMenu(3) → ShoppingList(4) */
  .main-area > :first-child { order: 1; }
  .sidebar > :nth-child(1) { order: 2; }
  .main-area > :nth-child(2) { order: 3; }
  .sidebar > :nth-child(2) { order: 4; }
}
</style>
