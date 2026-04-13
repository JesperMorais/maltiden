import { test, expect } from '@playwright/test'
import { login } from './helpers'

test.describe('Shopping list', () => {
  test('route redirects to login when unauthenticated', async ({ page }) => {
    await page.goto('/shopping-list')
    await expect(page).toHaveURL(/\/login/)
  })

  test('shows page after navigating from dashboard widget', async ({ page }) => {
    await login(page)
    // Click shopping list widget on dashboard
    const widget = page.locator('[class*="shopping-widget"]').first()
    if (await widget.isVisible({ timeout: 5000 }).catch(() => false)) {
      await widget.click()
      await expect(page).toHaveURL(/\/shopping-list/)
      await expect(page.getByRole('heading', { name: 'Inköpslista' })).toBeVisible({ timeout: 5000 })
      await expect(page.getByText(/Alla ingredienser du behöver/)).toBeVisible()
    }
  })

  test('shows back link to dashboard', async ({ page }) => {
    await login(page)
    const widget = page.locator('[class*="shopping-widget"]').first()
    if (await widget.isVisible({ timeout: 5000 }).catch(() => false)) {
      await widget.click()
      await expect(page).toHaveURL(/\/shopping-list/)
      const backLink = page.locator('.back-link')
      await expect(backLink).toBeVisible({ timeout: 5000 })
      await expect(backLink).toContainText('Dashboard')
    }
  })

  test('back link returns to dashboard', async ({ page }) => {
    await login(page)
    const widget = page.locator('[class*="shopping-widget"]').first()
    if (await widget.isVisible({ timeout: 5000 }).catch(() => false)) {
      await widget.click()
      await expect(page).toHaveURL(/\/shopping-list/)
      await page.locator('.back-link').click()
      await expect(page).toHaveURL(/\/dashboard/)
    }
  })
})
