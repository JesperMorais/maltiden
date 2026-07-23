import { describe, it, expect, beforeEach } from 'vitest'
import { useSessionEvents } from '../useSessionEvents'

describe('useSessionEvents', () => {
  beforeEach(() => {
    const { configure, clear } = useSessionEvents()
    configure({ enabled: false, retentionDays: 30, forced: false, maxBufferSize: 200 })
    clear()
  })

  it('defaults to disabled — no capture without opt-in', () => {
    const { isEnabled, logAction, debugSink } = useSessionEvents()

    expect(isEnabled()).toBe(false)

    logAction('menu:generate')

    expect(debugSink()).toHaveLength(0)
  })

  it('captures nothing until enabled via configure', () => {
    const { configure, logRoute, debugSink, isEnabled } = useSessionEvents()

    logRoute('dashboard')
    expect(debugSink()).toHaveLength(0)

    configure({ enabled: true })
    expect(isEnabled()).toBe(true)

    logRoute('dashboard')
    expect(debugSink()).toHaveLength(1)
  })

  it('logRoute captures only route name, no path/query/PII fields', () => {
    const { configure, logRoute, debugSink } = useSessionEvents()
    configure({ enabled: true })

    logRoute('recipe-detail')

    const [event] = debugSink()
    expect(event).toMatchObject({ kind: 'route', name: 'recipe-detail' })
    expect(Object.keys(event!)).toEqual(['kind', 'name', 'timestamp'])
  })

  it('logAction captures only the action name', () => {
    const { configure, logAction, debugSink } = useSessionEvents()
    configure({ enabled: true })

    logAction('recipe:save')

    const [event] = debugSink()
    expect(event).toMatchObject({ kind: 'action', name: 'recipe:save' })
  })

  it('trims the buffer to maxBufferSize, dropping oldest first', () => {
    const { configure, logAction, debugSink } = useSessionEvents()
    configure({ enabled: true, maxBufferSize: 3 })

    logAction('a')
    logAction('b')
    logAction('c')
    logAction('d')

    const names = debugSink().map((e) => e.name)
    expect(names).toEqual(['b', 'c', 'd'])
  })

  it('retention/forced flags are a one-line config flip', () => {
    const { configure, getConfig } = useSessionEvents()

    configure({ retentionDays: 120, forced: true })

    expect(getConfig().retentionDays).toBe(120)
    expect(getConfig().forced).toBe(true)
  })

  it('clear empties the buffer', () => {
    const { configure, logAction, clear, debugSink } = useSessionEvents()
    configure({ enabled: true })

    logAction('a')
    expect(debugSink()).toHaveLength(1)

    clear()
    expect(debugSink()).toHaveLength(0)
  })
})
