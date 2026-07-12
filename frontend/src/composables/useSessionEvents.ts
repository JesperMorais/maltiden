import { ref, readonly } from 'vue'

export type SessionEventType = 'route_change' | 'action'

export interface SessionEvent {
  id: number
  type: SessionEventType
  name: string
  requestId: string | null
  timestamp: number
}

// Vite env vars are always strings, so we check against the string 'true'
export const EVENTS_ENABLED = import.meta.env.VITE_EVENTS === 'true'

// Consent posture + retention are config consts (one-line flip, no migration)
export const CONSENT_MODE: 'opt-in' | 'opt-out' = 'opt-in'
export const RETENTION_DAYS = 1

const MAX_EVENTS = 200
let nextId = 0

const events = ref<SessionEvent[]>([])
let lastRequestId: string | null = null

/** Called by the axios response interceptor to correlate events with X-Request-ID. */
export function setLastRequestId(requestId: string | null) {
  lastRequestId = requestId
}

function record(type: SessionEventType, name: string) {
  if (!EVENTS_ENABLED) return

  const event: SessionEvent = {
    id: nextId++,
    type,
    name,
    requestId: lastRequestId,
    timestamp: Date.now(),
  }

  if (events.value.length >= MAX_EVENTS) {
    events.value = events.value.slice(1)
  }

  events.value = [...events.value, event]

  console.debug('[Events]', type, name, event.requestId ?? '')
}

function recordRouteChange(name: string) {
  record('route_change', name)
}

function recordAction(name: string) {
  record('action', name)
}

export function useSessionEvents() {
  return {
    events: readonly(events),
    recordRouteChange,
    recordAction,
  }
}
