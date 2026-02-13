import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { createApp, defineComponent, type App } from 'vue'
import { useProgressBar } from '../useProgressBar'

type ProgressBarReturn = ReturnType<typeof useProgressBar>

/**
 * Run useProgressBar inside a real Vue component setup context
 * so that lifecycle hooks (onUnmounted) register without warnings.
 */
function setup(
  options?: Parameters<typeof useProgressBar>[0],
): { result: ProgressBarReturn; app: App } {
  let result!: ProgressBarReturn
  const app = createApp(
    defineComponent({
      setup() {
        result = useProgressBar(options)
        return () => null
      },
    }),
  )
  app.mount(document.createElement('div'))
  return { result, app }
}

describe('useProgressBar', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.spyOn(window, 'requestAnimationFrame').mockImplementation((cb) => {
      return setTimeout(() => cb(Date.now()), 16) as unknown as number
    })
    vi.spyOn(window, 'cancelAnimationFrame').mockImplementation((id) => {
      clearTimeout(id)
    })
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('initializes with progress 0 and inactive', () => {
    const { result } = setup()

    expect(result.progress.value).toBe(0)
    expect(result.isActive.value).toBe(false)
  })

  it('sets isActive to true on start', () => {
    const { result } = setup()

    result.start()

    expect(result.isActive.value).toBe(true)
  })

  it('progress increases over time', () => {
    const { result } = setup({ duration: 10000 })

    result.start()
    vi.advanceTimersByTime(1000)

    expect(result.progress.value).toBeGreaterThan(0)
  })

  it('progress never exceeds 90 before finish', () => {
    const { result } = setup({ duration: 1000 })

    result.start()
    vi.advanceTimersByTime(100000)

    expect(result.progress.value).toBeLessThanOrEqual(90)
  })

  it('finish sets progress to 100', () => {
    const { result } = setup()

    result.start()
    vi.advanceTimersByTime(500)
    result.finish()

    expect(result.progress.value).toBe(100)
  })

  it('finish resets to inactive after 400ms', () => {
    const { result } = setup()

    result.start()
    result.finish()

    expect(result.isActive.value).toBe(true)
    expect(result.progress.value).toBe(100)

    vi.advanceTimersByTime(400)

    expect(result.isActive.value).toBe(false)
    expect(result.progress.value).toBe(0)
  })

  it('start resets a previous finish timeout', () => {
    const { result } = setup()

    result.start()
    result.finish()
    expect(result.progress.value).toBe(100)

    // Restart before the 400ms reset fires
    result.start()
    expect(result.progress.value).toBe(0)
    expect(result.isActive.value).toBe(true)

    // The old finish timeout should not fire and reset state
    vi.advanceTimersByTime(400)
    expect(result.isActive.value).toBe(true)
  })

  it('progress is faster at start and slower later', () => {
    const { result } = setup({ duration: 10000 })

    result.start()

    vi.advanceTimersByTime(1000)
    const earlyProgress = result.progress.value

    vi.advanceTimersByTime(4000)
    const midProgress = result.progress.value

    expect(earlyProgress).toBeGreaterThan(20)
    const earlyRate = earlyProgress / 1000
    const lateRate = (midProgress - earlyProgress) / 4000
    expect(earlyRate).toBeGreaterThan(lateRate)
  })

  it('uses custom duration for easing', () => {
    const { result: fast } = setup({ duration: 1000 })
    const { result: slow } = setup({ duration: 100000 })

    fast.start()
    slow.start()
    vi.advanceTimersByTime(500)

    expect(fast.progress.value).toBeGreaterThan(slow.progress.value)
  })

  it('defaults to duration 10000', () => {
    const { result } = setup()

    result.start()
    vi.advanceTimersByTime(5000)

    // At 5s of a 10s duration, should be roughly 78% (90 * (1 - e^(-1.5)))
    expect(result.progress.value).toBeGreaterThan(60)
    expect(result.progress.value).toBeLessThan(90)
  })

  it('cleans up on unmount', () => {
    const { result, app } = setup()

    result.start()
    vi.advanceTimersByTime(100)

    const cancelSpy = vi.mocked(window.cancelAnimationFrame)
    app.unmount()

    expect(cancelSpy).toHaveBeenCalled()
  })
})
