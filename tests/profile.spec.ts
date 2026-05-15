import { test, expect } from '@playwright/test'
import { login, navigateTo } from './helpers'

test.describe('Profile', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('shows profile heading after navigation', async ({ page }) => {
    await navigateTo(page, '/profile')
    await expect(page.getByText('Min profil')).toBeVisible({ timeout: 5000 })
  })

  test('shows avatar initial', async ({ page }) => {
    await navigateTo(page, '/profile')
    await expect(page.locator('.avatar-circle')).toBeVisible({ timeout: 5000 })
  })

  test('shows logout button', async ({ page }) => {
    await navigateTo(page, '/profile')
    await expect(page.getByRole('button', { name: /Logga ut/ })).toBeVisible({ timeout: 5000 })
  })

  test('shows household section', async ({ page }) => {
    await navigateTo(page, '/profile')
    await expect(page.getByText('Hushåll')).toBeVisible({ timeout: 5000 })
  })

  test('shows members section', async ({ page }) => {
    await navigateTo(page, '/profile')
    await expect(page.getByText('Medlemmar')).toBeVisible({ timeout: 5000 })
  })

  test('redirects unauthenticated to login', async ({ page }) => {
    await page.goto('/profile')
    await expect(page).toHaveURL(/\/login/, { timeout: 5000 })
  })
})
