import { type Page, expect } from '@playwright/test'

/** Generate unique email using timestamp + random suffix */
export function generateUniqueEmail(): string {
  const ts = Date.now()
  const rand = Math.random().toString(36).slice(2, 6)
  return `test-${ts}-${rand}@maltiden.test`
}

/** Generate unique name for test data */
export function generateUniqueName(prefix: string): string {
  return `${prefix}-${Date.now()}`
}

/** Register a new user via the Onboarding "Create household" UI flow. */
export async function registerUser(
  page: Page,
  options?: { email?: string; password?: string; name?: string; householdName?: string },
): Promise<{ email: string; password: string; name: string }> {
  const email = options?.email ?? generateUniqueEmail()
  const password = options?.password ?? 'TestPassword123!'
  const name = options?.name ?? 'Test User'
  const householdName = options?.householdName ?? 'Test Hushåll'

  await page.goto('/register')
  await page.getByText('Skapa nytt hushåll').click()

  await page.getByPlaceholder('Anna Andersson').fill(name)
  await page.getByPlaceholder('anna@exempel.se').fill(email)
  await page.getByPlaceholder('Minst 8 tecken').fill(password)
  await page.getByPlaceholder('Skriv lösenordet igen').fill(password)
  await page.getByPlaceholder('Familjen Andersson').fill(householdName)

  await page.getByRole('button', { name: 'Skapa konto' }).click()
  await expect(page.getByText('Konto skapat!')).toBeVisible({ timeout: 10_000 })
  await page.waitForURL('**/dashboard', { timeout: 10_000 })

  return { email, password, name }
}

/** Login via the Login page UI. */
export async function loginUser(page: Page, email: string, password: string): Promise<void> {
  await page.goto('/login')
  await page.getByPlaceholder('din@email.se').fill(email)
  await page.getByPlaceholder('Ditt lösenord').fill(password)
  await page.getByRole('button', { name: 'Logga in' }).click()
  await page.waitForURL('**/dashboard', { timeout: 10_000 })
}

/** Get JWT token from localStorage. Key is 'maltiden_token'. */
export async function getToken(page: Page): Promise<string | null> {
  return page.evaluate(() => localStorage.getItem('maltiden_token'))
}

/** Make an authenticated API call from within the browser context. */
export async function apiCall(
  page: Page,
  method: string,
  path: string,
  body?: Record<string, unknown>,
): Promise<{ status: number; data: Record<string, unknown> }> {
  const token = await getToken(page)
  const baseURL = 'http://localhost:8080'

  return page.evaluate(
    async ({ baseURL, method, path, body, token }) => {
      const res = await fetch(`${baseURL}${path}`, {
        method,
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: body ? JSON.stringify(body) : undefined,
      })
      const data = await res.json().catch(() => ({}))
      return { status: res.status, data }
    },
    { baseURL, method, path, body, token },
  )
}

/** Register a user via API (not UI). Faster for setup/teardown. */
export async function registerUserViaAPI(options?: {
  email?: string
  password?: string
  name?: string
}): Promise<{
  token: string
  email: string
  password: string
  userId: string
  householdId: string
}> {
  const email = options?.email ?? generateUniqueEmail()
  const password = options?.password ?? 'TestPassword123!'
  const name = options?.name ?? 'Test User'

  const res = await fetch('http://localhost:8080/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password, name }),
  })

  if (!res.ok) {
    throw new Error(`Registration failed: ${res.status} ${await res.text()}`)
  }

  const data = await res.json()
  return {
    token: data.token,
    email,
    password,
    userId: data.user.id,
    householdId: data.user.householdId,
  }
}

/** Set JWT token in localStorage. Key is 'maltiden_token'. */
export async function setAuthToken(page: Page, token: string): Promise<void> {
  await page.evaluate((t) => localStorage.setItem('maltiden_token', t), token)
}

/**
 * Navigate within the SPA without full page reload.
 * Uses window.location.hash or direct URL bar navigation followed by
 * waiting for the Vue Router to handle the route.
 *
 * IMPORTANT: page.goto() causes a full page reload which loses Pinia state.
 * After login/register, the user store has auth state in memory. A full reload
 * loses it, and the router guard redirects to /login. This helper navigates
 * via the SPA's Vue Router to preserve state.
 */
export async function navigateTo(page: Page, path: string): Promise<void> {
  await page.evaluate((p) => {
    window.history.pushState({}, '', p)
    window.dispatchEvent(new PopStateEvent('popstate'))
  }, path)
  // Wait for Vue Router to handle the navigation
  await page.waitForURL(`**${path}`, { timeout: 10_000 })
}
