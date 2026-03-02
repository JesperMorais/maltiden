import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'

// We need to test the interceptor logic, so we import the configured client
// and mock the HTTP layer underneath it.
let apiClient: typeof import('../client').default
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
const locationSpy = { href: '' }
const originalLocation = window.location

beforeEach(async () => {
  // Reset mocks
  vi.clearAllMocks()
  sessionStorage.clear()
  locationSpy.href = ''

  // Mock window.location (non-configurable in happy-dom, so use defineProperty)
  Object.defineProperty(window, 'location', {
    writable: true,
    value: { ...originalLocation, pathname: '/dashboard', href: '' },
  })

  // Fresh import to get a clean axios instance with interceptors
  const mod = await import('../client')
  apiClient = mod.default

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
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    mock.onGet('/health').networkError()

    await expect(apiClient.get('/health')).rejects.toThrow()

    expect(consoleSpy).toHaveBeenCalledWith('Network error - API may be unavailable')
    expect(mockTokenUtils.remove).not.toHaveBeenCalled()
    consoleSpy.mockRestore()
  })
})
