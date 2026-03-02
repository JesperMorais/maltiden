<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRoute } from 'vue-router'
import { AnimatePresence, Motion } from 'motion-v'
import { MessageSquarePlus, SmilePlus, Meh, Frown, X, Check } from 'lucide-vue-next'
import { submitFeedback } from '@/api/feedback.api'
import { useToast } from '@/composables/useToast'

type Mood = 'good' | 'okay' | 'bad'
type Category = 'recipes' | 'menu' | 'shopping' | 'design' | 'other'
type Step = 'mood' | 'category' | 'comment' | 'success'

const route = useRoute()
const { error: showError } = useToast()

const COOLDOWN_KEY = 'feedback_cooldown_until'
const COOLDOWN_MS = 24 * 60 * 60 * 1000

// State
const isOpen = ref(false)
const isHidden = ref(false)
const step = ref<Step>('mood')
const selectedMood = ref<Mood | null>(null)
const selectedCategories = ref<Category[]>([])
const comment = ref('')
const isSubmitting = ref(false)
const modalEl = ref<HTMLElement | null>(null)
const triggerEl = ref<HTMLElement | null>(null)

const commentLength = computed(() => comment.value.length)

const categories: { key: Category; label: string }[] = [
  { key: 'recipes', label: 'Recept' },
  { key: 'menu', label: 'Menyn' },
  { key: 'shopping', label: 'Inköpslistan' },
  { key: 'design', label: 'Design' },
  { key: 'other', label: 'Övrigt' },
]

const moodOptions: { key: Mood; label: string; icon: typeof SmilePlus }[] = [
  { key: 'good', label: 'Bra', icon: SmilePlus },
  { key: 'okay', label: 'Okej', icon: Meh },
  { key: 'bad', label: 'Dåligt', icon: Frown },
]

// Check cooldown on mount
onMounted(() => {
  const cooldownUntil = localStorage.getItem(COOLDOWN_KEY)
  if (cooldownUntil && Date.now() < Number(cooldownUntil)) {
    isHidden.value = true
  }
})

// Focus trap
let previouslyFocusedElement: HTMLElement | null = null

function openModal() {
  previouslyFocusedElement = document.activeElement as HTMLElement | null
  isOpen.value = true
  nextTick(() => {
    modalEl.value?.focus()
  })
}

function closeModal() {
  isOpen.value = false
  step.value = 'mood'
  selectedMood.value = null
  selectedCategories.value = []
  comment.value = ''
  previouslyFocusedElement?.focus()
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    closeModal()
    return
  }

  // Focus trap
  if (e.key === 'Tab' && modalEl.value) {
    const focusable = modalEl.value.querySelectorAll<HTMLElement>(
      'button, textarea, [tabindex]:not([tabindex="-1"])',
    )
    if (focusable.length === 0) return
    const first = focusable[0]!
    const last = focusable[focusable.length - 1]!

    if (e.shiftKey) {
      if (document.activeElement === first) {
        e.preventDefault()
        last.focus()
      }
    } else {
      if (document.activeElement === last) {
        e.preventDefault()
        first.focus()
      }
    }
  }
}

// Close on click outside
function handleOverlayClick(e: MouseEvent) {
  if (e.target === e.currentTarget) {
    closeModal()
  }
}

// Step handlers
function selectMood(mood: Mood) {
  selectedMood.value = mood
  step.value = 'category'
}

function toggleCategory(cat: Category) {
  const idx = selectedCategories.value.indexOf(cat)
  if (idx >= 0) {
    selectedCategories.value.splice(idx, 1)
  } else {
    selectedCategories.value.push(cat)
  }
}

function skipToComment() {
  step.value = 'comment'
}

function goToComment() {
  step.value = 'comment'
}

async function submit() {
  if (!selectedMood.value || isSubmitting.value) return

  isSubmitting.value = true
  try {
    await submitFeedback({
      mood: selectedMood.value,
      categories: selectedCategories.value.length > 0 ? selectedCategories.value : undefined,
      comment: comment.value.trim() || undefined,
      page: route.path,
      viewportWidth: window.innerWidth,
    })

    step.value = 'success'

    // Set cooldown
    localStorage.setItem(COOLDOWN_KEY, String(Date.now() + COOLDOWN_MS))

    // Close after showing success
    setTimeout(() => {
      closeModal()
      isHidden.value = true
    }, 2000)
  } catch {
    showError('Kunde inte skicka feedback. Försök igen.')
  } finally {
    isSubmitting.value = false
  }
}

// Body scroll lock when modal is open
watch(isOpen, (open) => {
  if (open) {
    document.addEventListener('keydown', handleKeydown)
  } else {
    document.removeEventListener('keydown', handleKeydown)
  }
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <Teleport to="body">
    <AnimatePresence>
      <!-- Floating button -->
      <Motion
        v-if="!isHidden"
        :initial="{ opacity: 0, scale: 0.8, y: 10 }"
        :animate="{ opacity: 1, scale: 1, y: 0 }"
        :transition="{ duration: 0.35, ease: 'easeOut', delay: 1 }"
      >
        <button
          ref="triggerEl"
          class="feedback-trigger"
          aria-label="Ge feedback"
          @click="openModal"
        >
          <MessageSquarePlus :size="18" :stroke-width="2.2" />
          <span class="trigger-label">Feedback</span>
        </button>
      </Motion>
    </AnimatePresence>

    <!-- Modal overlay -->
    <Transition name="overlay">
      <div
        v-if="isOpen"
        class="feedback-overlay"
        @click="handleOverlayClick"
      >
        <div
          ref="modalEl"
          class="feedback-modal"
          role="dialog"
          aria-modal="true"
          aria-label="Ge feedback"
          tabindex="-1"
        >
          <!-- Close button -->
          <button class="modal-close" aria-label="Stäng" @click="closeModal">
            <X :size="18" />
          </button>

          <!-- Step 1: Mood -->
          <Transition name="step" mode="out-in">
            <div v-if="step === 'mood'" key="mood" class="step-content">
              <h3 class="step-heading">Hur upplever du Måltiden?</h3>
              <div class="mood-options">
                <button
                  v-for="opt in moodOptions"
                  :key="opt.key"
                  class="mood-btn"
                  :class="{ selected: selectedMood === opt.key }"
                  :aria-label="opt.label"
                  @click="selectMood(opt.key)"
                >
                  <span class="mood-icon" :class="`mood-${opt.key}`">
                    <component :is="opt.icon" :size="28" :stroke-width="1.8" />
                  </span>
                  <span class="mood-label">{{ opt.label }}</span>
                </button>
              </div>
            </div>

            <!-- Step 2: Category -->
            <div v-else-if="step === 'category'" key="category" class="step-content">
              <h3 class="step-heading">Vad gäller det?</h3>
              <div class="category-chips">
                <button
                  v-for="cat in categories"
                  :key="cat.key"
                  class="chip"
                  :class="{ active: selectedCategories.includes(cat.key) }"
                  @click="toggleCategory(cat.key)"
                >
                  {{ cat.label }}
                </button>
              </div>
              <div class="step-actions">
                <button class="text-link" @click="skipToComment">Hoppa över</button>
                <button class="btn-next" @click="goToComment">Nästa</button>
              </div>
            </div>

            <!-- Step 3: Comment -->
            <div v-else-if="step === 'comment'" key="comment" class="step-content">
              <h3 class="step-heading">Berätta mer (valfritt)</h3>
              <div class="textarea-wrap">
                <textarea
                  v-model="comment"
                  class="feedback-textarea"
                  placeholder="Skriv din feedback här..."
                  maxlength="500"
                  rows="4"
                ></textarea>
                <span class="char-count" :class="{ near: commentLength > 450 }">
                  {{ commentLength }}/500
                </span>
              </div>
              <button
                class="btn-submit"
                :disabled="isSubmitting"
                :aria-busy="isSubmitting"
                @click="submit"
              >
                <span v-if="isSubmitting" class="submit-loading">Skickar...</span>
                <span v-else>Skicka</span>
              </button>
            </div>

            <!-- Step 4: Success -->
            <div v-else-if="step === 'success'" key="success" class="step-content success-content">
              <Motion
                :initial="{ scale: 0, opacity: 0 }"
                :animate="{ scale: 1, opacity: 1 }"
                :transition="{ duration: 0.4, ease: [0.34, 1.56, 0.64, 1] }"
              >
                <span class="success-icon">
                  <Check :size="32" :stroke-width="2.5" />
                </span>
              </Motion>
              <p class="success-text">Tack för din feedback!</p>
            </div>
          </Transition>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
/* Floating trigger button */
.feedback-trigger {
  position: fixed;
  bottom: 24px;
  right: 24px;
  z-index: 9000;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  background: var(--bg-card);
  color: var(--text-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-full);
  box-shadow: var(--shadow-md);
  cursor: pointer;
  font-family: 'Nunito', sans-serif;
  font-size: 0.8125rem;
  font-weight: 600;
  transition:
    transform 0.25s var(--ease-default),
    box-shadow 0.25s ease,
    color 0.2s ease;
}

.feedback-trigger:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
  color: var(--accent-text);
}

.feedback-trigger:active {
  transform: translateY(0);
}

.trigger-label {
  display: inline;
}

/* Overlay */
.feedback-overlay {
  position: fixed;
  inset: 0;
  z-index: 9500;
  display: flex;
  align-items: flex-end;
  justify-content: flex-end;
  padding: 24px;
  background: rgba(0, 0, 0, 0.15);
}

.overlay-enter-active,
.overlay-leave-active {
  transition: opacity 0.2s ease;
}

.overlay-enter-active .feedback-modal {
  transition:
    opacity 0.25s ease,
    transform 0.25s var(--ease-default);
}

.overlay-leave-active .feedback-modal {
  transition:
    opacity 0.15s ease,
    transform 0.15s ease;
}

.overlay-enter-from {
  opacity: 0;
}

.overlay-enter-from .feedback-modal {
  opacity: 0;
  transform: translateY(12px) scale(0.95);
}

.overlay-leave-to {
  opacity: 0;
}

.overlay-leave-to .feedback-modal {
  opacity: 0;
  transform: translateY(8px) scale(0.97);
}

/* Modal card */
.feedback-modal {
  position: relative;
  width: 340px;
  background: var(--bg-card);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  border: 1px solid var(--border-color);
  padding: 24px;
  outline: none;
}

.modal-close {
  position: absolute;
  top: 12px;
  right: 12px;
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.2s ease;
}

.modal-close:hover {
  color: var(--text-primary);
}

/* Step transitions */
.step-enter-active,
.step-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}

.step-enter-from {
  opacity: 0;
  transform: translateX(12px);
}

.step-leave-to {
  opacity: 0;
  transform: translateX(-12px);
}

/* Step content */
.step-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.step-heading {
  font-family: 'Nunito', sans-serif;
  font-size: 1rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
  padding-right: 24px;
}

/* Mood options */
.mood-options {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.mood-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 14px 16px;
  background: transparent;
  border: 2px solid var(--border-color);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    background 0.2s ease,
    transform 0.2s var(--ease-default);
  flex: 1;
  min-width: 0;
}

.mood-btn:hover {
  border-color: var(--border-color-hover);
  background: var(--bg-hover);
  transform: translateY(-2px);
}

.mood-btn:active {
  transform: translateY(0);
}

.mood-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.mood-good {
  color: var(--success);
  background: var(--success-bg);
}

.mood-okay {
  color: var(--warning-dark);
  background: var(--warning-bg);
}

.mood-bad {
  color: var(--error);
  background: var(--error-bg);
}

.mood-label {
  font-family: 'Nunito', sans-serif;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--text-secondary);
}

/* Category chips */
.category-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.chip {
  padding: 6px 14px;
  border-radius: var(--radius-full);
  border: 1.5px solid var(--border-color);
  background: transparent;
  color: var(--text-secondary);
  font-family: 'Nunito', sans-serif;
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    background 0.2s ease,
    color 0.2s ease;
}

.chip:hover {
  border-color: var(--border-color-hover);
}

.chip.active {
  background: var(--accent-bg);
  border-color: var(--accent);
  color: var(--accent-text);
}

/* Step actions */
.step-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.text-link {
  background: none;
  border: none;
  color: var(--text-muted);
  font-family: 'Nunito', sans-serif;
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
  padding: 4px 0;
  transition: color 0.2s ease;
}

.text-link:hover {
  color: var(--text-primary);
}

.btn-next {
  padding: 8px 20px;
  border-radius: var(--radius-full);
  border: none;
  background: var(--accent);
  color: var(--text-on-accent);
  font-family: 'Nunito', sans-serif;
  font-size: 0.875rem;
  font-weight: 700;
  cursor: pointer;
  transition:
    background 0.2s ease,
    transform 0.2s var(--ease-default);
}

.btn-next:hover {
  background: var(--accent-dark);
  transform: translateY(-1px);
}

/* Comment textarea */
.textarea-wrap {
  position: relative;
}

.feedback-textarea {
  width: 100%;
  padding: 12px;
  border: 1.5px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-primary);
  color: var(--text-primary);
  font-family: 'Nunito', sans-serif;
  font-size: 0.875rem;
  line-height: 1.5;
  resize: vertical;
  min-height: 80px;
  transition: border-color 0.2s ease;
}

.feedback-textarea::placeholder {
  color: var(--text-muted);
}

.feedback-textarea:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 3px var(--accent-focus-ring);
}

.char-count {
  position: absolute;
  bottom: 8px;
  right: 10px;
  font-size: 0.7rem;
  color: var(--text-muted);
  pointer-events: none;
}

.char-count.near {
  color: var(--warning-dark);
}

/* Submit button */
.btn-submit {
  align-self: flex-end;
  padding: 10px 24px;
  border-radius: var(--radius-full);
  border: none;
  background: linear-gradient(135deg, var(--accent) 0%, var(--accent-dark) 100%);
  color: var(--text-on-accent);
  font-family: 'Nunito', sans-serif;
  font-size: 0.875rem;
  font-weight: 700;
  cursor: pointer;
  box-shadow: var(--shadow-accent);
  transition:
    transform 0.25s var(--ease-default),
    box-shadow 0.25s ease,
    opacity 0.2s ease;
}

.btn-submit:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: var(--shadow-accent-hover);
}

.btn-submit:disabled {
  opacity: 0.6;
  cursor: wait;
}

.submit-loading {
  display: inline-block;
}

/* Success state */
.success-content {
  align-items: center;
  justify-content: center;
  min-height: 120px;
  text-align: center;
}

.success-icon {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--success-bg);
  color: var(--success);
  display: flex;
  align-items: center;
  justify-content: center;
}

.success-text {
  font-family: 'Nunito', sans-serif;
  font-size: 1rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

/* Mobile responsive */
@media (max-width: 768px) {
  .feedback-trigger {
    bottom: calc(var(--bottom-nav-height) + 16px);
    right: 16px;
    padding: 10px 12px;
  }

  .trigger-label {
    display: none;
  }

  .feedback-overlay {
    padding: 0;
    padding-bottom: var(--bottom-nav-height);
    align-items: flex-end;
    justify-content: stretch;
  }

  .feedback-modal {
    width: 100%;
    border-radius: var(--radius-lg) var(--radius-lg) 0 0;
    padding: 24px 20px;
    padding-bottom: max(24px, calc(env(safe-area-inset-bottom, 0px) + 24px));
  }
}
</style>
