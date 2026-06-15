<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, onBeforeRouteLeave } from 'vue-router'
import { useMenuGeneratorStore } from '@/stores/menuGenerator'
import { usePlanningPreferencesStore } from '@/stores/planningPreferences'
import { useSlotMachine, type DisplayRecipe } from '@/composables/useSlotMachine'
import { useToast } from '@/composables/useToast'
import { useFocusTrap } from '@/composables/useFocusTrap'
import { ChefHat } from 'lucide-vue-next'
import MenuDayCard from '@/components/menu/MenuDayCard.vue'
import GenerateMenuEmptyState from '@/components/menu/GenerateMenuEmptyState.vue'
import MenuGeneratorActions from '@/components/menu/MenuGeneratorActions.vue'
import MenuEconomyBar from '@/components/menu/MenuEconomyBar.vue'
import ErrorState from '@/components/common/ErrorState.vue'
import GenerateMenuSkeleton from '@/components/skeleton/layouts/GenerateMenuSkeleton.vue'
import { useSkeleton } from '@/composables/useSkeleton'

const router = useRouter()
const store = useMenuGeneratorStore()
const prefsStore = usePlanningPreferencesStore()
const slotMachine = useSlotMachine()
const toast = useToast()

// Local state
const showUnsavedWarning = ref(false)
const unsavedModalRef = ref<HTMLElement | null>(null)

useFocusTrap(unsavedModalRef, {
  isActive: showUnsavedWarning,
  onEscape: cancelLeave,
})
const hasNavigatedFromSave = ref(false)

// Computed
const days = computed(() => store.orderedDays)
const hasMenu = computed(() => store.hasMenu)

// Prep-läge (batch cooking) toggle, two-way bound to the store + localStorage.
const prepMode = computed({
  get: () => store.prepMode,
  set: (value: boolean) => store.setPrepMode(value),
})
const isLoading = computed(() => store.isLoading)
const { showSkeleton } = useSkeleton(
  computed(() => store.isGenerating && !slotMachine.isAnimating.value),
  { minDuration: 400 }
)

// Show grid during slot animation even before recipes arrive
const showGrid = computed(() => {
  return hasMenu.value || slotMachine.animationPhase.value !== 'idle'
})

// ============================================
// HANDLERS
// ============================================

/**
 * Handle initial menu generation with slot machine animation
 */
async function handleInitialGenerate() {
  // Initialize week so we have dates to work with
  store.initializeWeek()
  store.startSlotAnimation()

  // Get dates for all days
  const dates = store.orderedDays.map((d) => d.date)

  // Start rolling animation on all 5 cards
  slotMachine.startRolling(dates, store.lockedDays)

  try {
    // API call runs concurrently with rolling animation
    await store.generateInitialMenu()

    // Build final recipes map from store data
    const finalRecipes = new Map<string, DisplayRecipe>()
    store.orderedDays.forEach((day) => {
      if (day.recipeId) {
        finalRecipes.set(day.date, {
          recipeName: day.recipeName || 'Recept',
          emoji: day.emoji || '🍽️',
        })
      }
    })

    // Land sequentially left-to-right
    await slotMachine.landSequentially(finalRecipes, store.lockedDays)
    store.onSlotAnimationComplete()
  } catch {
    toast.error('Kunde inte generera meny. Försök igen.')
    store.setError('Kunde inte generera meny. Försök igen.')
    slotMachine.reset()
    store.onSlotAnimationComplete()
  }
}

/**
 * Handle regenerating unlocked days with slot machine animation
 */
async function handleRegenerate() {
  store.startSlotAnimation()

  // Start rolling only unlocked days
  const unlockedDates = store.orderedDays
    .filter((d) => !store.isDayLocked(d.date))
    .map((d) => d.date)

  slotMachine.startRolling(unlockedDates, store.lockedDays)

  try {
    await store.regenerateUnlockedDays()

    // Build final recipes map for unlocked days
    const finalRecipes = new Map<string, DisplayRecipe>()
    store.orderedDays.forEach((day) => {
      if (!store.isDayLocked(day.date) && day.recipeId) {
        finalRecipes.set(day.date, {
          recipeName: day.recipeName || 'Recept',
          emoji: day.emoji || '🍽️',
        })
      }
    })

    await slotMachine.landSequentially(finalRecipes, store.lockedDays)
    store.onSlotAnimationComplete()
  } catch {
    toast.error('Kunde inte generera nya recept. Försök igen.')
    store.setError('Kunde inte generera nya recept. Försök igen.')
    slotMachine.reset()
    store.onSlotAnimationComplete()
  }
}

/**
 * Handle lock toggle for a day
 */
function handleLockToggle(date: string) {
  // Don't allow toggling during animation
  if (slotMachine.isAnimating.value) return
  store.toggleDayLock(date)
}

/**
 * Handle saving the menu
 */
async function handleSave() {
  hasNavigatedFromSave.value = true
  const success = await store.saveDraftMenu(router)
  if (!success) {
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
  slotMachine.reset()
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
  slotMachine.destroy()
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
        <p class="subtitle">Måndag – Söndag</p>
        <p class="description">
          Skapa en meny för hela veckan med slumpmässiga recept. Lås dagar du vill behålla och generera nya för resten.
        </p>

        <!-- Prep-läge (batch cooking) toggle -->
        <label class="prep-toggle" :class="{ active: prepMode }">
          <span class="prep-toggle-icon">
            <ChefHat :size="20" :stroke-width="2.25" />
          </span>
          <span class="prep-toggle-text">
            <span class="prep-toggle-label">Prep-läge</span>
            <span class="prep-toggle-desc">Laga en gång, ät i två dagar</span>
          </span>
          <input
            v-model="prepMode"
            type="checkbox"
            class="prep-toggle-input"
            aria-label="Prep-läge: laga en gång, ät i två dagar"
          />
          <span class="prep-toggle-slider"></span>
        </label>
      </div>
    </header>

    <!-- Main content -->
    <main class="content">
      <div class="content-container">
        <!-- Skeleton loading state -->
        <GenerateMenuSkeleton v-if="showSkeleton" :day-count="prefsStore.activeDayCount" />

        <!-- Empty state -->
        <GenerateMenuEmptyState v-else-if="!showGrid" @generate="handleInitialGenerate" />

        <!-- Menu grid -->
        <div v-else class="menu-grid">
          <MenuDayCard
            v-for="day in days"
            :key="day.date"
            :day="day"
            :is-locked="store.isDayLocked(day.date)"
            :is-loading="store.isRegenerating && !store.isDayLocked(day.date) && !slotMachine.isAnimating.value"
            :is-rolling="slotMachine.isSlotRolling(day.date)"
            :has-landed="slotMachine.hasSlotLanded(day.date)"
            :display-recipe="slotMachine.getDisplayRecipe(day.date)"
            @toggle-lock="handleLockToggle(day.date)"
          />
        </div>

        <!-- Ingredient economy: shared ingredients across the week -->
        <MenuEconomyBar v-if="hasMenu" :economy="store.economy" />

        <!-- Error state -->
        <ErrorState v-if="store.error" :description="store.error" @retry="handleInitialGenerate" />
      </div>
    </main>

    <!-- Actions bar -->
    <MenuGeneratorActions
      v-if="hasMenu || store.isGenerating || slotMachine.isAnimating.value"
      :has-menu="hasMenu"
      :is-loading="isLoading || slotMachine.isAnimating.value"
      :locked-count="store.lockedDaysCount"
      @regenerate="handleRegenerate"
      @save="handleSave"
      @cancel="handleCancel"
    />

    <!-- Unsaved changes warning modal -->
    <Teleport to="body">
      <div v-if="showUnsavedWarning" class="modal-overlay" @click="cancelLeave">
        <div ref="unsavedModalRef" class="modal-card" role="dialog" aria-modal="true" aria-labelledby="unsaved-modal-title" @click.stop>
          <div class="modal-header">
            <h3 id="unsaved-modal-title" class="modal-title">Osparade ändringar</h3>
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
  color: var(--accent-text);
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

/* Prep-läge toggle */
.prep-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.875rem;
  margin: 1.5rem auto 0;
  padding: 0.75rem 1.25rem;
  background: var(--bg-card);
  border: 2px solid var(--border-color);
  border-radius: var(--radius-full);
  cursor: pointer;
  text-align: left;
  transition: all var(--duration-normal) var(--ease-default);
}

.prep-toggle:hover {
  border-color: var(--border-color-hover);
}

.prep-toggle.active {
  border-color: var(--accent);
  background: var(--bg-hover);
}

.prep-toggle-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  transition: color var(--duration-normal) var(--ease-default);
}

.prep-toggle.active .prep-toggle-icon {
  color: var(--accent);
}

.prep-toggle-text {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
}

.prep-toggle-label {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.95rem;
  color: var(--text-primary);
}

.prep-toggle-desc {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.prep-toggle-input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.prep-toggle-slider {
  position: relative;
  width: 48px;
  height: 28px;
  background: var(--border-color);
  border-radius: var(--radius-full);
  transition: background var(--duration-normal) ease;
  flex-shrink: 0;
}

.prep-toggle-slider::after {
  content: '';
  position: absolute;
  top: 3px;
  left: 3px;
  width: 22px;
  height: 22px;
  background: var(--bg-card);
  border-radius: 50%;
  box-shadow: var(--shadow-sm);
  transition: transform var(--duration-normal) var(--ease-default);
}

.prep-toggle-input:checked + .prep-toggle-slider {
  background: var(--accent);
}

.prep-toggle-input:checked + .prep-toggle-slider::after {
  transform: translateX(20px);
}

.prep-toggle-input:focus-visible + .prep-toggle-slider {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
  box-shadow: 0 0 0 4px var(--accent-focus-ring);
}

@media (prefers-reduced-motion: reduce) {
  .prep-toggle,
  .prep-toggle-icon,
  .prep-toggle-slider,
  .prep-toggle-slider::after {
    transition: none;
  }
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

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: var(--overlay-bg);
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
  color: var(--text-on-accent);
}

.modal-btn-leave:hover {
  box-shadow: var(--shadow-accent-hover);
}

/* Responsive */
@media (max-width: 1024px) {
  .menu-grid {
    grid-template-columns: repeat(3, 1fr);
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
    grid-template-columns: repeat(2, 1fr);
    gap: 0.75rem;
  }
}

@media (max-width: 480px) {
  .menu-grid {
    grid-template-columns: 1fr;
    gap: 0.75rem;
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
