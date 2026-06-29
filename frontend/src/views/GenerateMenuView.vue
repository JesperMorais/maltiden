<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, onBeforeRouteLeave } from 'vue-router'
import { useMenuGeneratorStore } from '@/stores/menuGenerator'
import { usePlanningPreferencesStore } from '@/stores/planningPreferences'
import { useSlotMachine, type DisplayRecipe } from '@/composables/useSlotMachine'
import { useToast } from '@/composables/useToast'
import { useFocusTrap } from '@/composables/useFocusTrap'
import { Sparkles, Info } from 'lucide-vue-next'
import MenuDayCard from '@/components/menu/MenuDayCard.vue'
import GenerateMenuEmptyState from '@/components/menu/GenerateMenuEmptyState.vue'
import MenuGeneratorActions from '@/components/menu/MenuGeneratorActions.vue'
import MenuRationale from '@/components/menu/MenuRationale.vue'
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
const isLoading = computed(() => store.isLoading)

// Free-text weekly wishes (Swedish), parsed server-side into constraints.
const wishes = computed({
  get: () => store.wishes,
  set: (value: string) => {
    store.wishes = value
  },
})

// AI-arrangemang opt-in toggle (persisted across reloads).
const arrange = computed({
  get: () => store.arrange,
  set: (value: boolean) => store.setArrange(value),
})
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
      </div>
    </header>

    <!-- Main content -->
    <main class="content">
      <div class="content-container">
        <!-- Prep-läge (batch cooking) toggle (#248 Phase 2). Affects the next
             generation/regeneration — batchable recipes are cooked once at
             double servings and reused as leftovers. -->
        <div v-if="!showSkeleton" class="prep-toggle">
          <label
            class="prep-toggle-control"
            :class="{ disabled: isLoading || slotMachine.isAnimating.value }"
          >
            <input
              v-model="store.prepMode"
              type="checkbox"
              class="prep-toggle-input"
              :disabled="isLoading || slotMachine.isAnimating.value"
            />
            <span class="prep-toggle-track"><span class="prep-toggle-thumb"></span></span>
            <span class="prep-toggle-text">
              <span class="prep-toggle-title">Matlagningsläge</span>
              <span class="prep-toggle-hint">
                Laga en gång, ät två gånger — fördubblar en rätt och återanvänder den som rester.
              </span>
            </span>
          </label>
        </div>

        <!-- Skeleton loading state -->
        <GenerateMenuSkeleton v-if="showSkeleton" :day-count="prefsStore.activeDayCount" />

        <!-- Pre-generation: AI experience controls (#248 Phase 4) + empty
             state. Set weekly wishes and opt into AI arrangement before
             generating; shown only before a menu exists. -->
        <div v-else-if="!showGrid" class="pre-generate">
          <div class="ai-controls">
            <div class="wishes-pref">
              <label for="wishes-input" class="wishes-label">Veckans önskemål</label>
              <textarea
                id="wishes-input"
                v-model="wishes"
                class="wishes-input"
                rows="2"
                maxlength="500"
                placeholder="T.ex. två vegetariska dagar och snabb vardagsmat"
              ></textarea>
            </div>

            <label class="ai-toggle" :class="{ active: arrange }">
              <span class="ai-toggle-icon">
                <Sparkles :size="20" :stroke-width="2.25" />
              </span>
              <span class="ai-toggle-text">
                <span class="ai-toggle-label">AI-arrangemang</span>
                <span class="ai-toggle-desc">Låt AI placera rätterna på veckodagar och förklara veckan</span>
              </span>
              <input
                v-model="arrange"
                type="checkbox"
                class="ai-toggle-input"
                aria-label="AI-arrangemang: låt AI placera rätterna och förklara veckan"
              />
              <span class="ai-toggle-slider"></span>
            </label>
          </div>

          <GenerateMenuEmptyState @generate="handleInitialGenerate" />
        </div>

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

        <!-- Shared-ingredient economy: highlights ingredients reused across the
             week's recipes so the overlap-aware generation is visible. -->
        <section
          v-if="hasMenu && store.sharedIngredientsCount > 0"
          class="shared-ingredients"
          aria-labelledby="shared-ingredients-title"
        >
          <div class="shared-header">
            <span class="shared-icon" aria-hidden="true">🛒</span>
            <h2 id="shared-ingredients-title" class="shared-title">
              {{ store.sharedIngredientsCount }}
              {{ store.sharedIngredientsCount === 1 ? 'delad ingrediens' : 'delade ingredienser' }}
              denna vecka
            </h2>
          </div>
          <p class="shared-subtitle">
            Recepten återanvänder ingredienser — färre varor att handla och mindre svinn.
          </p>
          <ul class="shared-chips">
            <li
              v-for="ing in store.sharedIngredients"
              :key="ing.name"
              class="shared-chip"
            >
              <span class="chip-name">{{ ing.name }}</span>
              <span class="chip-count" :title="`Används i ${ing.recipeCount} recept`">
                ×{{ ing.recipeCount }}
              </span>
            </li>
          </ul>
        </section>

        <!-- AI rationale: short Swedish explanation of the arranged week -->
        <MenuRationale v-if="hasMenu" :rationale="store.rationale" />

        <!-- Wishes-ignored note: wishes were sent but AI was unavailable -->
        <p v-if="store.wishesIgnored" class="wishes-ignored">
          <Info :size="16" :stroke-width="2" />
          AI-önskemål kräver konfiguration och hoppades över.
        </p>

        <!-- Weekly nutrition total (#248 Phase 3): aggregated macros across the
             cooked days, shown only when recipes carry nutrition data. -->
        <section
          v-if="hasMenu && store.weeklyNutrition"
          class="week-nutrition"
          aria-labelledby="week-nutrition-title"
        >
          <div class="nutrition-header">
            <span class="nutrition-icon" aria-hidden="true">🍎</span>
            <h2 id="week-nutrition-title" class="nutrition-title">Näring för veckan</h2>
          </div>
          <p v-if="store.weeklyNutrition.partial" class="nutrition-note">
            Delvis beräknat — vissa recept saknar näringsvärden.
          </p>
          <ul class="nutrition-stats">
            <li class="nutrition-stat">
              <span class="stat-value">{{ Math.round(store.weeklyNutrition.calories) }}</span>
              <span class="stat-label">kcal</span>
            </li>
            <li class="nutrition-stat">
              <span class="stat-value">{{ Math.round(store.weeklyNutrition.proteinG) }} g</span>
              <span class="stat-label">protein</span>
            </li>
            <li class="nutrition-stat">
              <span class="stat-value">{{ Math.round(store.weeklyNutrition.carbsG) }} g</span>
              <span class="stat-label">kolhydrater</span>
            </li>
            <li class="nutrition-stat">
              <span class="stat-value">{{ Math.round(store.weeklyNutrition.fatG) }} g</span>
              <span class="stat-label">fett</span>
            </li>
          </ul>
        </section>

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
      <div v-if="showUnsavedWarning" class="modal-overlay" role="presentation" @click="cancelLeave">
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

/* Prep-läge (batch cooking) toggle */
.prep-toggle {
  margin-bottom: 1.5rem;
}

.prep-toggle-control {
  display: flex;
  align-items: center;
  gap: 0.9rem;
  padding: 1rem 1.25rem;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  cursor: pointer;
  transition: border-color var(--duration-fast, 0.2s) var(--ease-default, ease);
}

.prep-toggle-control:hover {
  border-color: var(--accent);
}

.prep-toggle-control.disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.prep-toggle-input {
  position: absolute;
  opacity: 0;
  width: 1px;
  height: 1px;
}

.prep-toggle-track {
  flex-shrink: 0;
  position: relative;
  width: 46px;
  height: 26px;
  border-radius: 100px;
  background: var(--bg-secondary);
  border: 2px solid var(--border-color);
  transition: background 0.25s ease, border-color 0.25s ease;
}

.prep-toggle-thumb {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--text-secondary);
  transition: transform 0.25s cubic-bezier(0.34, 1.56, 0.64, 1), background 0.25s ease;
}

.prep-toggle-input:checked + .prep-toggle-track {
  background: var(--accent);
  border-color: var(--accent);
}

.prep-toggle-input:checked + .prep-toggle-track .prep-toggle-thumb {
  transform: translateX(20px);
  background: var(--text-on-accent);
}

.prep-toggle-input:focus-visible + .prep-toggle-track {
  box-shadow: 0 0 0 3px var(--accent-focus-ring);
}

.prep-toggle-text {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.prep-toggle-title {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 1rem;
  color: var(--text-primary);
}

.prep-toggle-hint {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-secondary);
  line-height: 1.4;
}

/* Weekly nutrition panel */
.week-nutrition {
  margin-top: 1.5rem;
  padding: 1.5rem 1.75rem;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  box-shadow: var(--shadow-sm);
}

.nutrition-header {
  display: flex;
  align-items: center;
  gap: 0.625rem;
}

.nutrition-icon {
  font-size: 1.5rem;
  line-height: 1;
}

.nutrition-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.35rem;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.3;
}

.nutrition-note {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin: 0.5rem 0 0;
}

.nutrition-stats {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  list-style: none;
  margin: 1rem 0 0;
  padding: 0;
}

.nutrition-stat {
  flex: 1 1 0;
  min-width: 100px;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  padding: 0.85rem 1rem;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 14px;
  text-align: center;
}

.stat-value {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.3rem;
  color: var(--accent-text);
}

.stat-label {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-secondary);
}

/* Shared-ingredient economy panel */
.shared-ingredients {
  margin-top: 2rem;
  padding: 1.5rem 1.75rem;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  box-shadow: var(--shadow-sm);
}

.shared-header {
  display: flex;
  align-items: center;
  gap: 0.625rem;
}

.shared-icon {
  font-size: 1.5rem;
  line-height: 1;
}

.shared-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.35rem;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.3;
}

.shared-subtitle {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1rem;
  color: var(--text-secondary);
  margin: 0.5rem 0 1rem;
  line-height: 1.5;
}

.shared-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.625rem;
  list-style: none;
  margin: 0;
  padding: 0;
}

.shared-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4rem 0.85rem;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
}

.chip-name {
  font-weight: 700;
  font-size: 0.95rem;
  color: var(--text-primary);
}

.chip-count {
  font-weight: 800;
  font-size: 0.85rem;
  color: var(--accent-text);
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

/* ============================================
   AI experience controls (#248 Phase 4)
   ============================================ */

.pre-generate {
  display: flex;
  flex-direction: column;
  gap: var(--space-lg);
}

.ai-controls {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.wishes-pref {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.wishes-label {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--text-primary);
}

.wishes-input {
  font-family: 'Nunito', sans-serif;
  font-size: 0.9rem;
  color: var(--text-primary);
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--space-sm);
  resize: vertical;
  transition: border-color var(--duration-fast) var(--ease-default);
}

.wishes-input::placeholder {
  color: var(--text-muted);
}

.wishes-input:focus {
  outline: none;
  border-color: var(--accent);
}

.ai-toggle {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  position: relative;
  padding: var(--space-md);
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition:
    border-color var(--duration-fast) var(--ease-default),
    background var(--duration-fast) var(--ease-default);
}

.ai-toggle.active {
  border-color: var(--accent);
  background: var(--bg-card-hover, var(--bg-card));
}

.ai-toggle-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--accent);
  flex-shrink: 0;
}

.ai-toggle-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
}

.ai-toggle-label {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.95rem;
  color: var(--text-primary);
}

.ai-toggle-desc {
  font-family: 'Nunito', sans-serif;
  font-size: 0.8rem;
  color: var(--text-secondary);
  line-height: 1.3;
}

/* Visually-hidden native checkbox; the slider is the visible control. */
.ai-toggle-input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.ai-toggle-slider {
  position: relative;
  flex-shrink: 0;
  width: 44px;
  height: 24px;
  border-radius: var(--radius-full);
  background: var(--border-color);
  transition: background var(--duration-fast) var(--ease-default);
}

.ai-toggle-slider::after {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 20px;
  height: 20px;
  border-radius: var(--radius-full);
  background: #fff;
  transition: transform var(--duration-fast) var(--ease-default);
}

.ai-toggle.active .ai-toggle-slider {
  background: var(--accent);
}

.ai-toggle.active .ai-toggle-slider::after {
  transform: translateX(20px);
}

.ai-toggle-input:focus-visible + .ai-toggle-slider {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.wishes-ignored {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  margin-top: var(--space-md);
  font-family: 'Nunito', sans-serif;
  font-size: 0.85rem;
  color: var(--text-secondary);
}
</style>
