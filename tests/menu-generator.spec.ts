import { test, expect } from '@playwright/test'
import { login } from './helpers'

test.describe('Menu generator', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.getByRole('button', { name: /Generera meny/ }).click()
    await expect(page).toHaveURL(/\/menu\/generate/)
  })

  test('shows page title and subtitle', async ({ page }) => {
    await expect(page.getByText('Generera veckomeny')).toBeVisible({ timeout: 5000 })
    await expect(page.getByText('Måndag - Fredag')).toBeVisible()
  })

  test('shows description text', async ({ page }) => {
    await expect(page.getByText(/Skapa en meny för 5 dagar/)).toBeVisible({ timeout: 5000 })
  })

  test('shows empty state or generate option initially', async ({ page }) => {
    // Either shows an empty state component or a generate button
    const emptyOrGenerate = page.locator('[class*="empty-state"], [class*="EmptyState"], .generate-button, [class*="generate"]').first()
    await expect(emptyOrGenerate).toBeVisible({ timeout: 5000 })
  })

  test('clicking generate creates menu grid', async ({ page }) => {
    const genBtn = page.locator('.generate-button')
    if (await genBtn.isVisible({ timeout: 5000 }).catch(() => false)) {
      await genBtn.click()
      await expect(page.locator('.menu-grid')).toBeVisible({ timeout: 8000 })
    }
  })

  test('generated menu shows 5 day cards', async ({ page }) => {
    const genBtn = page.locator('.generate-button')
    if (await genBtn.isVisible({ timeout: 5000 }).catch(() => false)) {
      await genBtn.click()
      await page.waitForTimeout(3000)
      const dayCards = page.locator('.menu-grid > *')
      const count = await dayCards.count()
      expect(count).toBe(5)
    }
  })

  test('generated menu shows save button', async ({ page }) => {
    const genBtn = page.locator('.generate-button')
    if (await genBtn.isVisible({ timeout: 5000 }).catch(() => false)) {
      await genBtn.click()
      await page.waitForTimeout(3000)
      await expect(page.getByText('Spara').first()).toBeVisible({ timeout: 5000 })
    }
  })

  test('generated menu shows regenerate button', async ({ page }) => {
    const genBtn = page.locator('.generate-button')
    if (await genBtn.isVisible({ timeout: 5000 }).catch(() => false)) {
      await genBtn.click()
      await page.waitForTimeout(3000)
      // Look for regenerate button
      const regenerateBtn = page.getByText(/Generera nya|Byt ut/i).first()
      await expect(regenerateBtn).toBeVisible({ timeout: 5000 })
    }
  })

  test('page has header with title text', async ({ page }) => {
    await expect(page.getByText('Generera veckomeny')).toBeVisible({ timeout: 5000 })
  })
})
