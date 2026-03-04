import { test, expect } from '@playwright/test'
import { login } from './helpers'

test.describe('Responsive design - mobile viewport', () => {
  test.use({ viewport: { width: 375, height: 667 } })

  test('landing page renders on mobile', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('.hero-title')).toBeVisible({ timeout: 5000 })
    await expect(page.getByText('Kom igång gratis')).toBeVisible()
  })

  test('login page renders on mobile', async ({ page }) => {
    await page.goto('/login')
    await expect(page.getByText('Välkommen tillbaka')).toBeVisible({ timeout: 5000 })
    await expect(page.getByPlaceholder('din@email.se')).toBeVisible()
    await expect(page.getByPlaceholder('Ditt lösenord')).toBeVisible()
  })

  test('login works on mobile', async ({ page }) => {
    await page.goto('/login')
    await page.getByPlaceholder('din@email.se').fill('test@example.com')
    await page.getByPlaceholder('Ditt lösenord').fill('password123')
    await page.getByRole('button', { name: 'Logga in' }).click()
    await expect(page).toHaveURL(/\/dashboard/, { timeout: 5000 })
  })

  test('dashboard renders on mobile', async ({ page }) => {
    await login(page)
    await expect(page.locator('.dashboard-content')).toBeVisible()
  })

  test('dashboard sidebar content is accessible on mobile', async ({ page }) => {
    await login(page)
    // On mobile (375px), .sidebar uses display:contents so it dissolves into the layout.
    // Verify the sidebar children (quick actions, shopping list) are visible instead.
    await expect(page.getByText('Snabbåtgärder')).toBeVisible({ timeout: 5000 })
    await expect(page.getByText('Inköpslista')).toBeVisible()
  })

  test('about page renders on mobile', async ({ page }) => {
    await page.goto('/about')
    await expect(page.getByRole('heading', { name: 'Om Måltiden', exact: true })).toBeVisible({ timeout: 5000 })
  })

  test('register page renders on mobile', async ({ page }) => {
    await page.goto('/register')
    await expect(page.getByText('Välkommen till Måltiden')).toBeVisible({ timeout: 5000 })
  })

  test('recipes page renders on mobile', async ({ page }) => {
    await login(page)
    // Use the mobile bottom nav button to navigate (preserves Vue Router state)
    await page.getByRole('button', { name: 'Gå till Recept' }).click()
    await expect(page).toHaveURL(/\/recipes/)
    await expect(page.getByRole('heading', { name: 'Recept' })).toBeVisible({ timeout: 5000 })
  })
})

test.describe('Responsive design - tablet viewport', () => {
  test.use({ viewport: { width: 768, height: 1024 } })

  test('landing page renders on tablet', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('.hero-title')).toBeVisible({ timeout: 5000 })
  })

  test('dashboard renders on tablet', async ({ page }) => {
    await login(page)
    await expect(page.locator('.dashboard-content')).toBeVisible()
  })

  test('about page renders on tablet', async ({ page }) => {
    await page.goto('/about')
    await expect(page.getByRole('heading', { name: 'Om Måltiden', exact: true })).toBeVisible({ timeout: 5000 })
  })
})

test.describe('Responsive design - wide viewport', () => {
  test.use({ viewport: { width: 1920, height: 1080 } })

  test('dashboard uses full width on wide screens', async ({ page }) => {
    await login(page)
    await expect(page.locator('.dashboard-content')).toBeVisible()
    // Dashboard grid should have sidebar visible
    const sidebar = page.locator('.sidebar')
    await expect(sidebar).toBeVisible({ timeout: 5000 })
  })

  test('landing page renders on wide screen', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('.hero-title')).toBeVisible({ timeout: 5000 })
  })
})
