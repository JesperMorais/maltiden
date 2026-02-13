import { ref, type Ref } from 'vue'

interface UseOptimisticOptions {
  /** Called when the API call fails after optimistic update */
  onError?: (error: unknown) => void
}

/**
 * Composable for optimistic UI updates.
 *
 * Applies a local state change immediately, then fires an async API call
 * in the background. Rolls back the change if the API call fails.
 *
 * @example
 * const { execute, isPending } = useOptimistic<boolean>({
 *   onError: () => console.error('Failed to update')
 * })
 *
 * // In a click handler:
 * execute(
 *   isEatingRef,        // ref to mutate
 *   !isEatingRef.value, // new value
 *   (newVal) => api.updateStatus(id, { isEatingToday: newVal })
 * )
 */
export function useOptimistic(options: UseOptimisticOptions = {}) {
  const { onError } = options
  const isPending = ref(false)

  function execute<T>(
    target: Ref<T>,
    newValue: T,
    apiCall: (value: T) => Promise<unknown>,
  ): void {
    const previousValue = target.value
    target.value = newValue
    isPending.value = true

    apiCall(newValue)
      .catch((error: unknown) => {
        target.value = previousValue
        onError?.(error)
      })
      .finally(() => {
        isPending.value = false
      })
  }

  return { execute, isPending }
}
