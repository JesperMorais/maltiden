import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import MockAdapter from 'axios-mock-adapter'

// We need to test the interceptor logic, so we import the configured client
// and mock the HTTP layer underneath it.
let apiClient: typeof import('../client').default
let markAuthSuccess: typeof import('../client').markAuthSuccess
let resetAuthState: typeof import('../client').resetAuthState
let mock: MockAdapter

// Mock tokenUtils
const mockTokenUtils = {
  get: vi.fn(),
  set: vi.fn(),
  remove: vi.fn(),
  exists: vi.fn(),
}
vi.mock('@/utils/token', () => ({
  tokenUtils: mockTokenUtils,
}))

// Capture window.location.href assignments
const originalLocation = window.location

beforeEach(async () => {
  // Reset mocks
  vi.clearAllMocks()
  sessionStorage.clear()

  // Mock window.location (non-configurable in happy-dom, so use defineProperty)
  Object.defineProperty(window, 'location', {
    writable: true,
    value: { ...originalLocation, pathname: '/dashboard', href: '' },
  })

  // Import the shared axios instance (ES modules are cached, same singleton)
  const mod = await import('../client')
  apiClient = mod.default
  markAuthSuccess = mod.markAuthSuccess
  resetAuthState = mod.resetAuthState

  // Reset module-level auth timing state to prevent leaks between tests
  resetAuthState()

  // Attach mock adapter
  mock = new MockAdapter(apiClient)
})

afterEach(() => {
  mock.restore()
  Object.defineProperty(window, 'location', { writable: true, value: originalLocation })
})

describe('API client request interceptor', () => {
  it('attaches Authorization header when token exists', async () => {
    mockTokenUtils.get.mockReturnValue('test-jwt-token')
    mock.onGet('/health').reply(200, { status: 'ok' })

    await apiClient.get('/health')

    const request = mock.history.get[0]!
    expect(request.headers?.Authorization).toBe('Bearer test-jwt-token')
  })

  it('does not attach Authorization header when no token', async () => {
    mockTokenUtils.get.mockReturnValue(null)
    mock.onGet('/health').reply(200, { status: 'ok' })

    await apiClient.get('/health')

    const request = mock.history.get[0]!
    expect(request.headers?.Authorization).toBeUndefined()
  })
})

describe('API client 401 response interceptor', () => {
  it('removes token and redirects on 401 from protected endpoint', async () => {
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { pathname: '/dashboard', href: '' },
    })
    mock.onGet('/households/me').reply(401, { error: 'unauthorized' })

    await expect(apiClient.get('/households/me')).rejects.toThrow()

    expect(mockTokenUtils.remove).toHaveBeenCalled()
    expect(sessionStorage.getItem('session_expired')).toBe('true')
    expect(window.location.href).toBe('/login')
  })

  it('does NOT redirect on 401 shortly after login (prevents redirect loop)', async () => {
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { pathname: '/dashboard', href: '' },
    })
    mock.onGet('/households/me').reply(401, { error: 'unauthorized' })

    // Simulate a fresh login
    markAuthSuccess()

    await expect(apiClient.get('/households/me')).rejects.toThrow()

    // Token is kept (not removed) — the 401 might be a timing issue right after login
    expect(mockTokenUtils.remove).not.toHaveBeenCalled()
    // No redirect — prevents loop
    expect(sessionStorage.getItem('session_expired')).toBeNull()
    expect(window.location.href).toBe('')
  })

  it('does NOT redirect on 401 from /auth/login (wrong credentials)', async () => {
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { pathname: '/login', href: '' },
    })
    mock.onPost('/auth/login').reply(401, { error: 'invalid_credentials' })

    await expect(apiClient.post('/auth/login', {})).rejects.toThrow()

    expect(mockTokenUtils.remove).not.toHaveBeenCalled()
    expect(sessionStorage.getItem('session_expired')).toBeNull()
    expect(window.location.href).toBe('')
  })

  it('does NOT redirect on 401 from /auth/register', async () => {
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { pathname: '/register', href: '' },
    })
    mock.onPost('/auth/register').reply(401, { error: 'unauthorized' })

    await expect(apiClient.post('/auth/register', {})).rejects.toThrow()

    expect(mockTokenUtils.remove).not.toHaveBeenCalled()
    expect(sessionStorage.getItem('session_expired')).toBeNull()
    expect(window.location.href).toBe('')
  })

  it('does NOT redirect when already on /login page', async () => {
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { pathname: '/login', href: '' },
    })
    mock.onGet('/households/me').reply(401, { error: 'unauthorized' })

    await expect(apiClient.get('/households/me')).rejects.toThrow()

    expect(mockTokenUtils.remove).toHaveBeenCalled()
    expect(sessionStorage.getItem('session_expired')).toBeNull()
    expect(window.location.href).toBe('')
  })

  it('does NOT redirect when already on /register page', async () => {
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { pathname: '/register', href: '' },
    })
    mock.onGet('/households/me').reply(401, { error: 'unauthorized' })

    await expect(apiClient.get('/households/me')).rejects.toThrow()

    expect(mockTokenUtils.remove).toHaveBeenCalled()
    expect(sessionStorage.getItem('session_expired')).toBeNull()
    expect(window.location.href).toBe('')
  })

  it('does NOT touch token or redirect on non-401 errors', async () => {
    mock.onGet('/households/me').reply(500, { error: 'internal_error' })

    await expect(apiClient.get('/households/me')).rejects.toThrow()

    expect(mockTokenUtils.remove).not.toHaveBeenCalled()
    expect(sessionStorage.getItem('session_expired')).toBeNull()
  })

  it('does NOT touch token or redirect on 403 errors', async () => {
    mock.onGet('/households/me').reply(403, { error: 'forbidden' })

    await expect(apiClient.get('/households/me')).rejects.toThrow()

    expect(mockTokenUtils.remove).not.toHaveBeenCalled()
    expect(sessionStorage.getItem('session_expired')).toBeNull()
  })

  it('handles network errors without crashing', async () => {
    const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})
    mock.onGet('/health').networkError()

    await expect(apiClient.get('/health')).rejects.toThrow()

    expect(consoleSpy).toHaveBeenCalledWith('[Auth] Network error — API may be unavailable')
    expect(mockTokenUtils.remove).not.toHaveBeenCalled()
    consoleSpy.mockRestore()
  })
})

describe('API client request-id correlation', () => {
  beforeEach(() => {
    vi.resetModules()
    vi.stubEnv('VITE_EVENTS', 'true')
  })

  afterEach(() => {
    vi.unstubAllEnvs()
  })

  it('forwards X-Request-ID from a successful response to setLastRequestId', async () => {
    const freshClient = (await import('../client')).default
    const { useSessionEvents } = await import('@/composables/useSessionEvents')
    const freshMock = new MockAdapter(freshClient)
    freshMock.onGet('/health').reply(200, { ok: true }, { 'x-request-id': 'req-success-1' })

    await freshClient.get('/health')

    const { events, recordAction } = useSessionEvents()
    recordAction('probe')
    expect(events.value[events.value.length - 1]?.requestId).toBe('req-success-1')
    freshMock.restore()
  })

  it('forwards X-Request-ID from an error response to setLastRequestId', async () => {
    const freshClient = (await import('../client')).default
    const { useSessionEvents } = await import('@/composables/useSessionEvents')
    const freshMock = new MockAdapter(freshClient)
    freshMock
      .onGet('/households/me')
      .reply(500, { error: 'internal_error' }, { 'x-request-id': 'req-error-1' })

    await expect(freshClient.get('/households/me')).rejects.toThrow()

    const { events, recordAction } = useSessionEvents()
    recordAction('probe')
    expect(events.value[events.value.length - 1]?.requestId).toBe('req-error-1')
    freshMock.restore()
  })
})
