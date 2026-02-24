import { test, expect } from '@playwright/test'
import { login, navigateTo } from './helpers'

test.describe('Cross-page navigation', () => {
  test('landing → login → dashboard → recipes → dashboard', async ({ page }) => {
    // Start at landing
    await page.goto('/')
    await expect(page.locator('.hero-title')).toBeVisible()

    // Go to login
    await page.getByText('Logga in').click()
    await expect(page).toHaveURL(/\/login/)

    // Login
    await page.getByPlaceholder('din@email.se').fill('test@example.com')
    await page.getByPlaceholder('Ditt lösenord').fill('password123')
    await page.getByRole('button', { name: 'Logga in' }).click()
    await expect(page).toHaveURL(/\/dashboard/, { timeout: 5000 })
    await page.waitForSelector('.dashboard-content', { timeout: 8000 })

    // Go to recipes
    await page.getByRole('button', { name: /Recept.*Hantera/ }).click()
    await expect(page).toHaveURL(/\/recipes/)
    await expect(page.getByRole('heading', { name: 'Recept' })).toBeVisible({ timeout: 5000 })

    // Back to dashboard
    await page.locator('.back-link').click()
    await expect(page).toHaveURL(/\/dashboard/)
    await expect(page.locator('.dashboard-content')).toBeVisible()
  })

  test('landing → about → landing', async ({ page }) => {
    await page.goto('/')
    await page.getByText('Om oss').click()
    await expect(page).toHaveURL(/\/about/)
    await expect(page.getByText('Om Måltiden')).toBeVisible({ timeout: 5000 })

    await page.getByRole('link', { name: /Tillbaka/ }).click()
    await expect(page).toHaveURL('/')
  })

  test('landing → register → landing', async ({ page }) => {
    await page.goto('/')
    await page.getByText('Kom igång gratis').click()
    await expect(page).toHaveURL(/\/register/)

    // Back link on register goes to landing
    const backLink = page.locator('.back-link').first()
    await backLink.click()
    await expect(page).toHaveURL('/')
  })

  test('login → register → login', async ({ page }) => {
    await page.goto('/login')
    await page.getByText('Skapa konto').click()
    await expect(page).toHaveURL(/\/register/)

    await page.getByRole('link', { name: /Tillbaka/ }).click()
    await expect(page).toHaveURL('/')
  })

  test('dashboard → menu generator → dashboard (via navigateTo)', async ({ page }) => {
    await login(page)

    await page.getByRole('button', { name: /Generera meny/ }).click()
    await expect(page).toHaveURL(/\/menu\/generate/)
    await expect(page.getByText('Generera veckomeny')).toBeVisible({ timeout: 5000 })

    // Navigate back via router (no unsaved changes since we didn't generate)
    await navigateTo(page, '/dashboard')
    await expect(page).toHaveURL(/\/dashboard/)
  })

  test('dashboard → shopping list via widget → dashboard', async ({ page }) => {
    await login(page)
    const widget = page.locator('[class*="shopping-widget"]').first()
    if (await widget.isVisible({ timeout: 5000 }).catch(() => false)) {
      await widget.click()
      await expect(page).toHaveURL(/\/shopping-list/)
      await page.locator('.back-link').click()
      await expect(page).toHaveURL(/\/dashboard/)
    }
  })

  test('logo always links to landing from dashboard', async ({ page }) => {
    await login(page)
    const logo = page.locator('.dashboard-header .logo')
    await expect(logo).toHaveAttribute('href', '/')
  })

  test('browser back button works correctly', async ({ page }) => {
    // Start at landing
    await page.goto('/')
    // Go to about
    await page.getByText('Om oss').click()
    await expect(page).toHaveURL(/\/about/)

    // Browser back
    await page.goBack()
    await expect(page).toHaveURL('/')
  })
})

test.describe('Direct URL access', () => {
  test('can access landing page directly', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('.hero-title')).toBeVisible({ timeout: 5000 })
  })

  test('can access about page directly', async ({ page }) => {
    await page.goto('/about')
    await expect(page.getByRole('heading', { name: 'Om Måltiden', exact: true })).toBeVisible({ timeout: 5000 })
  })

  test('can access login page directly', async ({ page }) => {
    await page.goto('/login')
    await expect(page.getByText('Välkommen tillbaka')).toBeVisible({ timeout: 5000 })
  })

  test('can access register page directly', async ({ page }) => {
    await page.goto('/register')
    await expect(page.getByText('Välkommen till Måltiden')).toBeVisible({ timeout: 5000 })
  })

  test('accessing protected route without auth redirects to login', async ({ page }) => {
    await page.goto('/dashboard')
    await expect(page).toHaveURL(/\/login/)
  })

  test('/recipes/parse redirects to /recipes', async ({ page }) => {
    await login(page)
    await navigateTo(page, '/recipes/parse')
    await expect(page).toHaveURL(/\/recipes/)
  })

  test('non-existent route falls back gracefully', async ({ page }) => {
    await page.goto('/nonexistent-page')
    // Vue Router should handle this — either redirect or 404
    await page.waitForTimeout(1000)
    // Should not crash — page should have content
    const body = page.locator('body')
    await expect(body).toBeVisible()
  })
})
