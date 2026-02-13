import { ref, onUnmounted } from 'vue'

interface UseProgressBarOptions {
  /** Estimated total duration in ms. Controls the speed curve. Default: 10000 */
  duration?: number
}

/**
 * Composable for simulated progress bar animation.
 *
 * Creates a fake progress value that advances quickly at first,
 * then slows down as it approaches ~90%, never reaching 100%
 * until `finish()` is called.
 *
 * @example
 * const { progress, isActive, start, finish } = useProgressBar({ duration: 15000 })
 *
 * async function handleGenerate() {
 *   start()
 *   try {
 *     await store.generateMenu()
 *   } finally {
 *     finish()
 *   }
 * }
 */
export function useProgressBar(options: UseProgressBarOptions = {}) {
  const { duration = 10000 } = options

  const progress = ref(0)
  const isActive = ref(false)

  let animationFrame: number | null = null
  let startTime = 0
  let finishTimeout: ReturnType<typeof setTimeout> | null = null

  function easeProgress(elapsed: number): number {
    // Asymptotic curve: fast start, approaches 90% but never reaches it
    const t = elapsed / duration
    return 90 * (1 - Math.exp(-3 * t))
  }

  function tick() {
    if (!isActive.value) return
    const elapsed = Date.now() - startTime
    progress.value = Math.min(easeProgress(elapsed), 90)
    animationFrame = requestAnimationFrame(tick)
  }

  function start() {
    cleanup()
    progress.value = 0
    isActive.value = true
    startTime = Date.now()
    animationFrame = requestAnimationFrame(tick)
  }

  function finish() {
    if (animationFrame) cancelAnimationFrame(animationFrame)
    animationFrame = null
    progress.value = 100
    finishTimeout = setTimeout(() => {
      isActive.value = false
      progress.value = 0
    }, 400)
  }

  function cleanup() {
    if (animationFrame) cancelAnimationFrame(animationFrame)
    if (finishTimeout) clearTimeout(finishTimeout)
    animationFrame = null
    finishTimeout = null
  }

  onUnmounted(cleanup)

  return { progress, isActive, start, finish }
}
