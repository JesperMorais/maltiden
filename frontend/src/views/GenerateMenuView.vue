<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, onBeforeRouteLeave } from 'vue-router'
import { useMenuGeneratorStore } from '@/stores/menuGenerator'
import MenuDayCard from '@/components/menu/MenuDayCard.vue'
import GenerateMenuEmptyState from '@/components/menu/GenerateMenuEmptyState.vue'
import MenuGeneratorActions from '@/components/menu/MenuGeneratorActions.vue'
import ProgressBar from '@/components/common/ProgressBar.vue'
import { useProgressBar } from '@/composables/useProgressBar'

const router = useRouter()
const store = useMenuGeneratorStore()

// Progress bar for generation
const { progress: genProgress, isActive: genActive, start: genStart, finish: genFinish } =
  useProgressBar({ duration: 8000 })

// Local state
const showUnsavedWarning = ref(false)
const hasNavigatedFromSave = ref(false)

// Computed
const days = computed(() => store.orderedDays)
const hasMenu = computed(() => store.hasMenu)
const isLoading = computed(() => store.isLoading)

// ============================================
// HANDLERS
// ============================================

/**
 * Handle initial menu generation
 */
async function handleInitialGenerate() {
  genStart()
  try {
    await store.generateInitialMenu()
  } catch (error) {
    console.error('Generation failed:', error)
  } finally {
    genFinish()
  }
}

/**
 * Handle regenerating unlocked days
 */
async function handleRegenerate() {
  genStart()
  try {
    await store.regenerateUnlockedDays()
  } catch (error) {
    console.error('Regeneration failed:', error)
  } finally {
    genFinish()
  }
}

/**
 * Handle lock toggle for a day
 */
function handleLockToggle(date: string) {
  store.toggleDayLock(date)
}

/**
 * Handle saving the menu
 */
async function handleSave() {
  hasNavigatedFromSave.value = true
  try {
    const success = await store.saveMenu(router)
    if (success) {
      // Navigation handled by store
    }
  } catch (error) {
    console.error('Save failed:', error)
    hasNavigatedFromSave.value = false
  }
}

/**
 * Handle cancel/back navigation
 */
function handleCancel() {
  if (hasMenu.value) {
    showUnsavedWarning.value = true
  } else {
    router.push({ name: 'dashboard' })
  }
}

/**
 * Confirm leave without saving
 */
function confirmLeave() {
  showUnsavedWarning.value = false
  store.clearDraft()
  router.push({ name: 'dashboard' })
}

/**
 * Stay on page
 */
function cancelLeave() {
  showUnsavedWarning.value = false
}

// ============================================
// LIFECYCLE
// ============================================

onMounted(() => {
  // Initialize the week
  store.initializeWeek()
})

onBeforeUnmount(() => {
  // Optionally clear draft when leaving
  // Commented out to preserve state if user navigates back
  // store.clearDraft()
})

// Route guard for unsaved changes
onBeforeRouteLeave((to, from, next) => {
  if (hasMenu.value && !hasNavigatedFromSave.value) {
    showUnsavedWarning.value = true
    next(false)
  } else {
    next()
  }
})
</script>

<template>
  <div class="generate-menu-view">
    <!-- Header -->
    <header class="header">
      <div class="header-content">
        <h1 class="title">Generera veckomeny</h1>
        <p class="subtitle">Måndag - Fredag</p>
        <p class="description">
          Skapa en meny för 5 dagar med slumpmässiga recept. Lås dagar du vill behålla och generera nya för resten.
        </p>
      </div>
    </header>

    <!-- Main content -->
    <main class="content">
      <div class="content-container">
        <!-- Empty state -->
        <GenerateMenuEmptyState v-if="!hasMenu" @generate="handleInitialGenerate" />

        <!-- Menu grid -->
        <div v-else class="menu-grid">
          <MenuDayCard
            v-for="day in days"
            :key="day.date"
            :day="day"
            :is-locked="store.isDayLocked(day.date)"
            :is-loading="store.isRegenerating && !store.isDayLocked(day.date)"
            @toggle-lock="handleLockToggle(day.date)"
          />
        </div>

        <!-- Loading overlay (initial generation) -->
        <div v-if="store.isGenerating" class="loading-overlay">
          <div class="loading-spinner">
            <div class="spinner-emoji">✨</div>
            <p class="loading-text">Genererar meny...</p>
            <ProgressBar :progress="genProgress" :active="genActive" />
          </div>
        </div>

        <!-- Error state -->
        <div v-if="store.error" class="error-state">
          <div class="error-icon">⚠️</div>
          <p class="error-message">{{ store.error }}</p>
          <button class="retry-button" @click="handleInitialGenerate">
            Försök igen
          </button>
        </div>
      </div>
    </main>

    <!-- Actions bar -->
    <MenuGeneratorActions
      v-if="hasMenu || store.isGenerating"
      :has-menu="hasMenu"
      :is-loading="isLoading"
      :locked-count="store.lockedDaysCount"
      @regenerate="handleRegenerate"
      @save="handleSave"
      @cancel="handleCancel"
    />

    <!-- Unsaved changes warning modal -->
    <Teleport to="body">
      <div v-if="showUnsavedWarning" class="modal-overlay" @click="cancelLeave">
        <div class="modal-card" @click.stop>
          <div class="modal-header">
            <h3 class="modal-title">Osparade ändringar</h3>
          </div>
          <div class="modal-body">
            <p class="modal-text">
              Osparade ändringar kommer gå förlorade. Är du säker på att du vill lämna?
            </p>
          </div>
          <div class="modal-actions">
            <button class="modal-btn modal-btn-cancel" @click="cancelLeave">
              Stanna kvar
            </button>
            <button class="modal-btn modal-btn-leave" @click="confirmLeave">
              Lämna utan att spara
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.generate-menu-view {
  min-height: 100vh;
  background: var(--bg-primary);
  display: flex;
  flex-direction: column;
}

/* Header */
.header {
  padding: 3rem 2rem 2rem;
  background: linear-gradient(
    180deg,
    var(--bg-card) 0%,
    var(--bg-primary) 100%
  );
  border-bottom: 1px solid var(--border-color);
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  text-align: center;
}

.title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 2.5rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem 0;
  line-height: 1.2;
}

.subtitle {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1.25rem;
  color: var(--accent);
  margin: 0 0 1rem 0;
}

.description {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1.1rem;
  color: var(--text-secondary);
  margin: 0;
  line-height: 1.6;
  max-width: 700px;
  margin: 0 auto;
}

/* Content */
.content {
  flex: 1;
  padding: 3rem 2rem;
  position: relative;
}

.content-container {
  max-width: 1200px;
  margin: 0 auto;
  position: relative;
}

/* Menu grid */
.menu-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 1.5rem;
}

/* Loading overlay */
.loading-overlay {
  position: absolute;
  inset: 0;
  background: rgba(255, 252, 247, 0.9);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.loading-spinner {
  text-align: center;
  width: 280px;
}

.spinner-emoji {
  font-size: 4rem;
  animation: spin 1.5s ease-in-out infinite;
}

@keyframes spin {
  0% {
    transform: rotate(0deg) scale(1);
  }
  50% {
    transform: rotate(180deg) scale(1.2);
  }
  100% {
    transform: rotate(360deg) scale(1);
  }
}

.loading-text {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1.25rem;
  color: var(--text-primary);
  margin: 1rem 0 0 0;
}

/* Error state */
.error-state {
  text-align: center;
  padding: 3rem 2rem;
}

.error-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.error-message {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1.1rem;
  color: var(--text-primary);
  margin: 0 0 1.5rem 0;
}

.retry-button {
  padding: 0.875rem 2rem;
  background: var(--accent);
  color: white;
  border: none;
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.retry-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(255, 107, 91, 0.4);
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 2rem;
}

.modal-card {
  background: var(--bg-card);
  border-radius: 24px;
  box-shadow: var(--shadow-lg);
  max-width: 500px;
  width: 100%;
  overflow: hidden;
}

.modal-header {
  padding: 2rem 2rem 1rem;
}

.modal-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.5rem;
  color: var(--text-primary);
  margin: 0;
}

.modal-body {
  padding: 0 2rem 2rem;
}

.modal-text {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1.05rem;
  color: var(--text-secondary);
  margin: 0;
  line-height: 1.6;
}

.modal-actions {
  display: flex;
  gap: 1rem;
  padding: 1.5rem 2rem 2rem;
  border-top: 1px solid var(--border-color);
}

.modal-btn {
  flex: 1;
  padding: 0.875rem 1.5rem;
  border: 2px solid transparent;
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.modal-btn:hover {
  transform: translateY(-2px);
}

.modal-btn-cancel {
  background: var(--bg-secondary);
  color: var(--text-primary);
  border-color: var(--border-color);
}

.modal-btn-cancel:hover {
  border-color: var(--text-primary);
}

.modal-btn-leave {
  background: var(--accent);
  color: white;
}

.modal-btn-leave:hover {
  box-shadow: 0 4px 12px rgba(255, 107, 91, 0.4);
}

/* Responsive */
@media (max-width: 1024px) {
  .menu-grid {
    grid-template-columns: repeat(5, 1fr);
    gap: 1rem;
  }

  .title {
    font-size: 2rem;
  }

  .description {
    font-size: 1rem;
  }
}

@media (max-width: 768px) {
  .header {
    padding: 2rem 1rem 1.5rem;
  }

  .content {
    padding: 2rem 1rem;
  }

  .menu-grid {
    grid-template-columns: 1fr;
    gap: 1rem;
  }

  .title {
    font-size: 1.75rem;
  }

  .subtitle {
    font-size: 1.1rem;
  }

  .description {
    font-size: 0.95rem;
  }
}
</style>
