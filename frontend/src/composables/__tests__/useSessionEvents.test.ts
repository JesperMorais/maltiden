import { describe, it, expect, beforeEach, vi } from 'vitest'

describe('useSessionEvents', () => {
  beforeEach(() => {
    vi.resetModules()
  })

  it('does not record events when disabled (default)', async () => {
    vi.stubEnv('VITE_EVENTS', 'false')
    const { useSessionEvents } = await import('../useSessionEvents')
    const { events, recordRouteChange, recordAction } = useSessionEvents()

    recordRouteChange('dashboard')
    recordAction('some_action')

    expect(events.value).toHaveLength(0)
    vi.unstubAllEnvs()
  })

  it('records route change events when enabled', async () => {
    vi.stubEnv('VITE_EVENTS', 'true')
    const { useSessionEvents } = await import('../useSessionEvents')
    const { events, recordRouteChange } = useSessionEvents()

    recordRouteChange('dashboard')

    expect(events.value).toHaveLength(1)
    expect(events.value[0]).toMatchObject({ type: 'route_change', name: 'dashboard' })
    vi.unstubAllEnvs()
  })

  it('records named key action events when enabled', async () => {
    vi.stubEnv('VITE_EVENTS', 'true')
    const { useSessionEvents } = await import('../useSessionEvents')
    const { events, recordAction } = useSessionEvents()

    recordAction('generate_menu')

    expect(events.value).toHaveLength(1)
    expect(events.value[0]).toMatchObject({ type: 'action', name: 'generate_menu' })
    vi.unstubAllEnvs()
  })

  it('correlates events with the last request id', async () => {
    vi.stubEnv('VITE_EVENTS', 'true')
    const { useSessionEvents, setLastRequestId } = await import('../useSessionEvents')
    const { events, recordAction } = useSessionEvents()

    setLastRequestId('req-123')
    recordAction('generate_menu')

    expect(events.value[0]?.requestId).toBe('req-123')
    vi.unstubAllEnvs()
  })

  it('caps buffer size and drops oldest event', async () => {
    vi.stubEnv('VITE_EVENTS', 'true')
    const { useSessionEvents } = await import('../useSessionEvents')
    const { events, recordAction } = useSessionEvents()

    for (let i = 0; i < 201; i++) {
      recordAction(`action_${i}`)
    }

    expect(events.value).toHaveLength(200)
    expect(events.value[0]?.name).toBe('action_1')
    vi.unstubAllEnvs()
  })
})
