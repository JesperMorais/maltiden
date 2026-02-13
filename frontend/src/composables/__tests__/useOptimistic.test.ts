import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useOptimistic } from '../useOptimistic'

describe('useOptimistic', () => {
  it('updates target immediately on execute', () => {
    const { execute } = useOptimistic()
    const target = ref(false)

    execute(target, true, () => new Promise(() => {}))

    expect(target.value).toBe(true)
  })

  it('sets isPending to true during API call', () => {
    const { execute, isPending } = useOptimistic()
    const target = ref('a')

    expect(isPending.value).toBe(false)

    execute(target, 'b', () => new Promise(() => {}))

    expect(isPending.value).toBe(true)
  })

  it('sets isPending to false after successful API call', async () => {
    const { execute, isPending } = useOptimistic()
    const target = ref(0)

    const apiCall = vi.fn().mockResolvedValue({ ok: true })
    execute(target, 1, apiCall)

    // Wait for promise chain to settle
    await vi.waitFor(() => {
      expect(isPending.value).toBe(false)
    })

    expect(target.value).toBe(1)
  })

  it('rolls back target on API failure', async () => {
    const { execute } = useOptimistic()
    const target = ref('original')

    const apiCall = vi.fn().mockRejectedValue(new Error('network error'))
    execute(target, 'optimistic', apiCall)

    // Immediately the value is optimistic
    expect(target.value).toBe('optimistic')

    // After the promise rejects, it rolls back
    await vi.waitFor(() => {
      expect(target.value).toBe('original')
    })
  })

  it('calls onError callback on API failure', async () => {
    const onError = vi.fn()
    const { execute } = useOptimistic({ onError })
    const target = ref(true)

    const error = new Error('server error')
    execute(target, false, () => Promise.reject(error))

    await vi.waitFor(() => {
      expect(onError).toHaveBeenCalledWith(error)
    })
  })

  it('does not call onError on success', async () => {
    const onError = vi.fn()
    const { execute } = useOptimistic({ onError })
    const target = ref(1)

    execute(target, 2, () => Promise.resolve({ ok: true }))

    await vi.waitFor(() => {
      expect(onError).not.toHaveBeenCalled()
    })

    expect(target.value).toBe(2)
  })

  it('passes new value to API call', () => {
    const { execute } = useOptimistic()
    const target = ref(10)
    const apiCall = vi.fn().mockResolvedValue({})

    execute(target, 42, apiCall)

    expect(apiCall).toHaveBeenCalledWith(42)
  })

  it('handles multiple sequential executions', async () => {
    const { execute } = useOptimistic()
    const target = ref('a')

    // First call succeeds
    execute(target, 'b', () => Promise.resolve({}))
    expect(target.value).toBe('b')

    // Second call while first is in flight — overwrites
    execute(target, 'c', () => Promise.resolve({}))
    expect(target.value).toBe('c')

    await vi.waitFor(() => {
      expect(target.value).toBe('c')
    })
  })
})
