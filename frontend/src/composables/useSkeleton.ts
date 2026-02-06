import { ref, computed, watch, type Ref, type ComputedRef } from 'vue'

interface UseSkeletonOptions {
  /** Minimum time (ms) to show skeleton to prevent flash. Defaults to 0. */
  minDuration?: number
}

/**
 * Composable for managing skeleton loading state.
 *
 * Wraps a boolean loading ref and optionally enforces a minimum display
 * duration to prevent the skeleton from flashing on fast connections.
 *
 * @example
 * const { showSkeleton } = useSkeleton(
 *   computed(() => store.isLoading),
 *   { minDuration: 300 }
 * )
 */
export function useSkeleton(
  loadingRef: Ref<boolean> | ComputedRef<boolean>,
  options: UseSkeletonOptions = {}
) {
  const { minDuration = 0 } = options

  const startTime = ref<number | null>(null)
  const forceShow = ref(false)

  const showSkeleton = computed(() => {
    return loadingRef.value || forceShow.value
  })

  watch(loadingRef, (isLoading) => {
    if (isLoading) {
      startTime.value = Date.now()
    } else if (minDuration > 0 && startTime.value) {
      const elapsed = Date.now() - startTime.value
      if (elapsed < minDuration) {
        forceShow.value = true
        setTimeout(() => {
          forceShow.value = false
        }, minDuration - elapsed)
      }
      startTime.value = null
    }
  })

  return {
    showSkeleton,
    isLoading: loadingRef
  }
}
