import { ref, readonly } from 'vue'

export type ToastType = 'success' | 'error' | 'info' | 'warning'

export interface Toast {
  id: number
  type: ToastType
  message: string
  duration: number
}

const MAX_TOASTS = 3
let nextId = 0

const toasts = ref<Toast[]>([])

function addToast(type: ToastType, message: string, duration?: number) {
  const defaultDuration = type === 'error' ? 5000 : 3000
  const toast: Toast = {
    id: nextId++,
    type,
    message,
    duration: duration ?? defaultDuration,
  }

  // If at max, remove the oldest
  if (toasts.value.length >= MAX_TOASTS) {
    toasts.value = toasts.value.slice(1)
  }

  toasts.value = [...toasts.value, toast]
}

function removeToast(id: number) {
  toasts.value = toasts.value.filter((t) => t.id !== id)
}

export function useToast() {
  return {
    toasts: readonly(toasts),
    success: (message: string, duration?: number) => addToast('success', message, duration),
    error: (message: string, duration?: number) => addToast('error', message, duration),
    info: (message: string, duration?: number) => addToast('info', message, duration),
    warning: (message: string, duration?: number) => addToast('warning', message, duration),
    remove: removeToast,
  }
}
