import { onMounted, onUnmounted, type Ref } from 'vue'

export function useClickOutside(
  elementRef: Ref<HTMLElement | null>,
  callback: () => void
) {
  function handler(event: MouseEvent) {
    const el = elementRef.value
    if (el && !el.contains(event.target as Node)) {
      callback()
    }
  }

  onMounted(() => {
    document.addEventListener('mousedown', handler)
  })

  onUnmounted(() => {
    document.removeEventListener('mousedown', handler)
  })
}
