import { ref, readonly } from 'vue'

export type RetentionDays = 30 | 120

export interface SessionEventsConfig {
  /** Master switch. Default false — no capture happens unless explicitly enabled. */
  enabled: boolean
  /** How long events may be retained before callers should purge them. Informational only; this composable does not persist anything. */
  retentionDays: RetentionDays
  /** When true, capture cannot be disabled by the user at runtime (forced collection). When false, disabling is allowed. */
  forced: boolean
  /** Max number of events kept in the in-memory buffer (oldest dropped first). */
  maxBufferSize: number
}

export type SessionEvent =
  | { kind: 'route'; name: string; timestamp: number }
  | { kind: 'action'; name: string; timestamp: number }

const DEFAULT_CONFIG: SessionEventsConfig = {
  enabled: false,
  retentionDays: 30,
  forced: false,
  maxBufferSize: 200,
}

let config: SessionEventsConfig = { ...DEFAULT_CONFIG }
const events = ref<SessionEvent[]>([])

function configure(overrides: Partial<SessionEventsConfig>) {
  config = { ...config, ...overrides }
}

function getConfig(): Readonly<SessionEventsConfig> {
  return readonly(config) as SessionEventsConfig
}

function isEnabled(): boolean {
  return config.enabled
}

function record(event: SessionEvent) {
  if (!config.enabled) return

  const next = [...events.value, event]
  if (next.length > config.maxBufferSize) {
    next.splice(0, next.length - config.maxBufferSize)
  }
  events.value = next
}

/** Log a route change by route name only — no full path, query, or params (avoids PII in dynamic segments). */
function logRoute(name: string) {
  record({ kind: 'route', name, timestamp: Date.now() })
}

/** Log a named key action (e.g. "menu:generate", "recipe:save"). No payload/DOM data captured. */
function logAction(name: string) {
  record({ kind: 'action', name, timestamp: Date.now() })
}

function clear() {
  events.value = []
}

/** Debug sink — reads the current in-memory buffer. Nothing is sent anywhere; there is no backend integration. */
function debugSink(): readonly SessionEvent[] {
  return events.value
}

export function useSessionEvents() {
  return {
    events: readonly(events),
    configure,
    getConfig,
    isEnabled,
    logRoute,
    logAction,
    clear,
    debugSink,
  }
}
