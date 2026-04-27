import { test, expect } from '@playwright/test'
import { registerUser, loginUser, generateUniqueEmail, getToken } from './helpers'

test.describe('Authentication E2E', () => {
  test.describe.configure({ mode: 'serial' })

  let registeredEmail: string
  let registeredPassword: string

  test('register a new user → redirect to dashboard with JWT', async ({ page }) => {
    const { email, password } = await registerUser(page)
    registeredEmail = email
    registeredPassword = password

    await expect(page).toHaveURL(/\/dashboard/)
    const token = await getToken(page)
    expect(token).toBeTruthy()
    // Backend generates household name as "<name>'s household"
    await expect(page.getByText("Test User's household")).toBeVisible({ timeout: 10_000 })
  })

  test('register with existing email → shows error', async ({ page }) => {
    await page.goto('/register')
    await page.getByText('Skapa nytt hushåll').click()

    await page.getByPlaceholder('Anna', { exact: true }).fill('Duplicate')
    await page.getByPlaceholder('Andersson').fill('User')
    await page.getByPlaceholder('anna@exempel.se').fill(registeredEmail)
    await page.getByPlaceholder('Minst 8 tecken').fill('TestPassword123!')
    await page.getByPlaceholder('Skriv lösenordet igen').fill('TestPassword123!')

    await page.getByRole('button', { name: 'Skapa konto' }).click()

    // Should show Swedish error about duplicate email
    await expect(
      page.getByText(/redan registrerad|misslyckades/)
    ).toBeVisible({ timeout: 10_000 })
  })

  test('login with valid credentials → dashboard loads', async ({ page }) => {
    await loginUser(page, registeredEmail, registeredPassword)
    await expect(page).toHaveURL(/\/dashboard/)
    await expect(page.getByText("Test User's household")).toBeVisible({ timeout: 10_000 })
  })

  test('login with wrong password → shows error', async ({ page }) => {
    await page.goto('/login')
    await page.getByPlaceholder('din@email.se').fill(registeredEmail)
    await page.getByPlaceholder('Ditt lösenord').fill('WrongPassword999')
    await page.getByRole('button', { name: 'Logga in' }).click()

    await expect(
      page.getByText(/Fel e-post|lösenord|misslyckades/)
    ).toBeVisible({ timeout: 10_000 })
  })

  test('access protected route without JWT → redirect to login', async ({ page }) => {
    await page.goto('/')
    await page.evaluate(() => localStorage.removeItem('maltiden_token'))
    await page.goto('/dashboard')
    await expect(page).toHaveURL(/\/login/)
  })

  test('expired/invalid JWT → redirect to login', async ({ page }) => {
    await page.goto('/')
    await page.evaluate(() => localStorage.setItem('maltiden_token', 'expired.invalid.token'))
    await page.goto('/dashboard')
    // Axios interceptor catches 401, removes token, redirects to /login
    await expect(page).toHaveURL(/\/login/, { timeout: 10_000 })
  })
})
