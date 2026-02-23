import { watch, onUnmounted, type Ref, type ComputedRef } from 'vue'

const FOCUSABLE_SELECTOR = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(', ')

export function useFocusTrap(
  containerRef: Ref<HTMLElement | null>,
  options: {
    isActive: Ref<boolean> | ComputedRef<boolean>
    onEscape?: () => void
  },
) {
  let previouslyFocused: HTMLElement | null = null

  function getFocusableElements(): HTMLElement[] {
    const el = containerRef.value
    if (!el) return []
    return Array.from(el.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR))
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape' && options.onEscape) {
      e.preventDefault()
      options.onEscape()
      return
    }

    if (e.key !== 'Tab') return

    const focusable = getFocusableElements()
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

  function activate() {
    previouslyFocused = document.activeElement as HTMLElement | null
    document.addEventListener('keydown', handleKeydown)

    requestAnimationFrame(() => {
      const focusable = getFocusableElements()
      if (focusable.length > 0) {
        focusable[0]!.focus()
      }
    })
  }

  function deactivate() {
    document.removeEventListener('keydown', handleKeydown)
    if (previouslyFocused && typeof previouslyFocused.focus === 'function') {
      previouslyFocused.focus()
    }
    previouslyFocused = null
  }

  watch(
    () => options.isActive.value,
    (active) => {
      if (active) {
        activate()
      } else {
        deactivate()
      }
    },
    { flush: 'post' },
  )

  onUnmounted(() => {
    deactivate()
  })
}
