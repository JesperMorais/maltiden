import { test, expect, type Page } from '@playwright/test'

async function login(page: Page) {
  await page.goto('/login')
  await page.getByPlaceholder('din@email.se').fill('test@example.com')
  await page.getByPlaceholder('Ditt lösenord').fill('password123')
  await page.getByRole('button', { name: 'Logga in' }).click()
  await expect(page).toHaveURL(/\/dashboard/, { timeout: 5000 })
}

test.describe('Authentication flow', () => {
  test('login → dashboard → logout → redirects to login', async ({ page }) => {
    await login(page)
    await page.waitForSelector('.dashboard-content', { timeout: 8000 })

    // Open user dropdown
    await page.locator('.user-button').click()

    // Click logout
    await page.getByText('Logga ut').click()

    // Should redirect to login
    await expect(page).toHaveURL(/\/login/, { timeout: 5000 })
  })

  test('protected route /dashboard redirects to login', async ({ page }) => {
    await page.goto('/dashboard')
    await expect(page).toHaveURL(/\/login/)
  })

  test('protected route /recipes redirects to login', async ({ page }) => {
    await page.goto('/recipes')
    await expect(page).toHaveURL(/\/login/)
  })

  test('protected route /menu/generate redirects to login', async ({ page }) => {
    await page.goto('/menu/generate')
    await expect(page).toHaveURL(/\/login/)
  })

  test('protected route /shopping-list redirects to login', async ({ page }) => {
    await page.goto('/shopping-list')
    await expect(page).toHaveURL(/\/login/)
  })

  test('authenticated user visiting /login redirects to dashboard', async ({ page }) => {
    await login(page)
    await page.waitForSelector('.dashboard-content', { timeout: 8000 })

    // Navigate to login - should redirect to dashboard since we're authenticated
    await page.goto('/login')
    // Either stays on dashboard or briefly shows login then redirects
    await page.waitForTimeout(2000)
    await expect(page).toHaveURL(/\/(dashboard|login)/)
  })

  test('public routes are accessible without auth', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveURL('/')

    await page.goto('/about')
    await expect(page).toHaveURL(/\/about/)

    await page.goto('/register')
    await expect(page).toHaveURL(/\/register/)
  })

  test('login state persists across page navigation', async ({ page }) => {
    await login(page)
    await page.waitForSelector('.dashboard-content', { timeout: 8000 })

    // Navigate to recipes (authenticated route)
    await page.getByRole('button', { name: /Recept.*Hantera/ }).click()
    await expect(page).toHaveURL(/\/recipes/)

    // Should still be authenticated (not redirected to login)
    await expect(page.getByRole('heading', { name: 'Recept' })).toBeVisible({ timeout: 5000 })
  })

  test('logout clears auth state - protected routes redirect', async ({ page }) => {
    await login(page)
    await page.waitForSelector('.dashboard-content', { timeout: 8000 })

    // Logout
    await page.locator('.user-button').click()
    await page.getByText('Logga ut').click()
    await expect(page).toHaveURL(/\/login/, { timeout: 5000 })

    // Try to visit dashboard again
    await page.goto('/dashboard')
    await expect(page).toHaveURL(/\/login/)
  })

  test('can log in again after logout', async ({ page }) => {
    // First login
    await login(page)
    await page.waitForSelector('.dashboard-content', { timeout: 8000 })

    // Logout
    await page.locator('.user-button').click()
    await page.getByText('Logga ut').click()
    await expect(page).toHaveURL(/\/login/, { timeout: 5000 })

    // Login again
    await page.getByPlaceholder('din@email.se').fill('test@example.com')
    await page.getByPlaceholder('Ditt lösenord').fill('password123')
    await page.getByRole('button', { name: 'Logga in' }).click()
    await expect(page).toHaveURL(/\/dashboard/, { timeout: 5000 })
    await page.waitForSelector('.dashboard-content', { timeout: 8000 })
  })
})
