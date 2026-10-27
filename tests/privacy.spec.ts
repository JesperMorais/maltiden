import { test, expect } from '@playwright/test'

test.describe('Privacy policy', () => {
  test('is public and describes data handling', async ({ page }) => {
    await page.goto('/integritet')
    await expect(page.getByRole('heading', { level: 1, name: 'Integritetspolicy' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Dina rättigheter' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Mitt konto' })).toHaveAttribute('href', '/konto')
  })
})
