<script setup lang="ts">
import { ref, toRefs } from 'vue'
import { useFocusTrap } from '@/composables/useFocusTrap'

interface Props {
  isOpen: boolean
  inviteCode: string
}

const props = defineProps<Props>()
const { isOpen } = toRefs(props)

const emit = defineEmits<{
  close: []
}>()

const inviteModalRef = ref<HTMLElement | null>(null)

useFocusTrap(inviteModalRef, {
  isActive: isOpen,
  onEscape: () => emit('close'),
})

const copied = ref(false)
let copyTimer: ReturnType<typeof setTimeout> | undefined

async function copyCode(code: string) {
  try {
    await navigator.clipboard.writeText(code)
    copied.value = true
    clearTimeout(copyTimer)
    copyTimer = setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch {
    // Fallback for older browsers
    const textarea = document.createElement('textarea')
    textarea.value = code
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
    copied.value = true
    clearTimeout(copyTimer)
    copyTimer = setTimeout(() => {
      copied.value = false
    }, 2000)
  }
}

function handleOverlayClick(e: MouseEvent) {
  if (e.target === e.currentTarget) {
    emit('close')
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="isOpen" class="invite-overlay" @click="handleOverlayClick">
        <div ref="inviteModalRef" class="invite-modal" role="dialog" aria-modal="true" aria-labelledby="invite-modal-title">
          <!-- Header -->
          <div class="modal-header">
            <h2 id="invite-modal-title">Bjud in familjemedlem</h2>
            <button class="close-btn" @click="emit('close')">
              <span>&times;</span>
            </button>
          </div>

          <!-- Content -->
          <div class="modal-content">
            <p class="instructions">
              Dela denna kod med din familj så de kan gå med i hushållet.
            </p>

            <div class="code-display">
              <span class="code-label">Inbjudningskod</span>
              <span class="code-value">{{ inviteCode }}</span>
            </div>

            <button
              class="copy-btn"
              :class="{ copied }"
              @click="copyCode(inviteCode)"
            >
              {{ copied ? 'Kopierad!' : 'Kopiera kod' }}
            </button>

            <p class="hint">
              Den inbjudna personen väljer "Gå med i hushåll" vid registrering och anger koden.
            </p>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.invite-overlay {
  position: fixed;
  inset: 0;
  background: var(--overlay-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.invite-modal {
  background: var(--bg-primary);
  border-radius: 24px;
  width: 100%;
  max-width: 420px;
  overflow: hidden;
  box-shadow: var(--shadow-lg);
  border: 1px solid var(--border-color);
}

/* Header */
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.5rem 2rem;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h2 {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.5rem;
  color: var(--text-primary);
  margin: 0;
}

.close-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: var(--bg-hover);
  color: var(--text-secondary);
  font-size: 1.5rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.close-btn:hover {
  color: var(--accent);
}

/* Content */
.modal-content {
  padding: 2rem;
  text-align: center;
}

.instructions {
  font-family: 'Nunito', sans-serif;
  font-size: 0.95rem;
  color: var(--text-secondary);
  line-height: 1.5;
  margin: 0 0 1.5rem;
}

.code-display {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 1.25rem;
  background: var(--bg-card);
  border-radius: 16px;
  border: 2px dashed var(--border-color-hover);
  margin-bottom: 1.25rem;
}

.code-label {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
}

.code-value {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 2rem;
  color: var(--accent-text);
  letter-spacing: 0.15em;
  user-select: all;
}

.copy-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.85rem 2rem;
  border: none;
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  background: linear-gradient(135deg, var(--accent) 0%, var(--accent-dark) 100%);
  color: var(--text-on-accent);
  box-shadow: var(--shadow-accent);
  margin-bottom: 1.25rem;
}

.copy-btn:hover {
  transform: translateY(-2px) scale(1.02);
}

.copy-btn:active {
  transform: translateY(-1px) scale(0.98);
}

.copy-btn.copied {
  background: linear-gradient(135deg, var(--success) 0%, #38a169 100%);
  box-shadow: 0 6px 20px rgba(72, 187, 120, 0.35);
}

.hint {
  font-family: 'Nunito', sans-serif;
  font-size: 0.8rem;
  color: var(--text-muted);
  line-height: 1.5;
  margin: 0;
}

/* Modal animation */
.modal-enter-active {
  transition: all 0.3s ease;
}

.modal-leave-active {
  transition: all 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .invite-modal,
.modal-leave-to .invite-modal {
  transform: scale(0.95) translateY(20px);
}

/* Responsive */
@media (max-width: 480px) {
  .invite-modal {
    max-width: none;
    margin: 0 0.5rem;
  }

  .modal-header,
  .modal-content {
    padding-left: 1.5rem;
    padding-right: 1.5rem;
  }

  .code-value {
    font-size: 1.5rem;
  }
}
</style>
