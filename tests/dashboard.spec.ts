import { test, expect } from '@playwright/test'
import { login, navigateTo } from './helpers'

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('shows dashboard content after login', async ({ page }) => {
    await expect(page.locator('.dashboard-content')).toBeVisible()
  })

  test('shows dashboard header with logo', async ({ page }) => {
    await expect(page.locator('.dashboard-header')).toBeVisible()
    await expect(page.getByText('Måltiden').first()).toBeVisible()
  })

  test('shows household name in header', async ({ page }) => {
    await expect(page.getByText('Familjen Andersson')).toBeVisible({ timeout: 5000 })
  })

  test('shows user button with name', async ({ page }) => {
    const userButton = page.locator('.user-button')
    await expect(userButton).toBeVisible()
  })

  test('shows role badge in header', async ({ page }) => {
    const roleBadge = page.locator('.role-badge')
    await expect(roleBadge).toBeVisible({ timeout: 5000 })
  })

  test('user dropdown opens and shows settings and logout', async ({ page }) => {
    await page.locator('.user-button').click()
    await expect(page.getByText('Inställningar')).toBeVisible()
    await expect(page.getByText('Logga ut')).toBeVisible()
  })

  test('shows weekly menu grid', async ({ page }) => {
    await expect(
      page.locator('[class*="weekly-menu"], [class*="menu-grid"]').first(),
    ).toBeVisible({ timeout: 5000 })
  })

  test('shows today\'s meal section', async ({ page }) => {
    const todaysMeal = page.locator('[class*="todays-meal"], [class*="TodaysMeal"]').first()
    await expect(todaysMeal).toBeVisible({ timeout: 5000 })
  })

  test('shows quick actions section', async ({ page }) => {
    await expect(page.getByText('Snabbåtgärder')).toBeVisible({ timeout: 5000 })
  })

  test('shows quick action buttons', async ({ page }) => {
    await expect(page.getByRole('button', { name: /Generera meny/ })).toBeVisible({
      timeout: 5000,
    })
    await expect(page.getByRole('button', { name: /Recept.*Hantera/ })).toBeVisible()
  })

  test('shows invite quick action', async ({ page }) => {
    await expect(page.getByText(/Bjud in/).first()).toBeVisible({ timeout: 5000 })
  })

  test('shows parse recipe quick action', async ({ page }) => {
    await expect(page.getByRole('button', { name: /Tolka recept/ })).toBeVisible({
      timeout: 5000,
    })
  })

  test('navigate to generate menu via quick action', async ({ page }) => {
    await page.getByRole('button', { name: /Generera meny/ }).click()
    await expect(page).toHaveURL(/\/menu\/generate/)
  })

  test('navigate to recipes via quick action', async ({ page }) => {
    await page.getByRole('button', { name: /Recept.*Hantera/ }).click()
    await expect(page).toHaveURL(/\/recipes/)
  })

  test('shows household widget', async ({ page }) => {
    await expect(page.getByText('Hushållet')).toBeVisible({ timeout: 5000 })
  })

  test('shows household member count', async ({ page }) => {
    await expect(page.getByText(/\d+ personer/)).toBeVisible({ timeout: 5000 })
  })

  test('shows household members list', async ({ page }) => {
    // Mock data has Anna, Erik, Lisa, Oscar
    await expect(page.getByText('Anna').first()).toBeVisible({ timeout: 5000 })
    await expect(page.getByText('Erik')).toBeVisible()
  })

  test('shows member eating status', async ({ page }) => {
    await expect(page.getByText('Äter idag').first()).toBeVisible({ timeout: 5000 })
  })

  test('shows owner badge for owner member', async ({ page }) => {
    await expect(page.getByText('Ägare').first()).toBeVisible({ timeout: 5000 })
  })

  test('shows guest badge for guest members', async ({ page }) => {
    await expect(page.getByText('Gäst').first()).toBeVisible({ timeout: 5000 })
  })

  test('shows invite button in household widget', async ({ page }) => {
    // In horizontal layout, the invite button shows as an avatar circle with label 'Bjud in'
    await expect(page.getByLabel('Bjud in fler')).toBeVisible({ timeout: 5000 })
  })

  test('shows shopping list widget', async ({ page }) => {
    await expect(page.getByText('Inköpslista')).toBeVisible({ timeout: 5000 })
  })

  test('shopping list widget shows items count', async ({ page }) => {
    // Shopping list widget shows stats: remaining "kvar", checked "klart", total "totalt"
    await expect(page.getByText('kvar')).toBeVisible({ timeout: 5000 })
    await expect(page.getByText('totalt')).toBeVisible()
  })

  test('shopping list widget has progress bar', async ({ page }) => {
    const progressBar = page.locator('.progress-bar, .progress-fill, [class*="progress"]').first()
    await expect(progressBar).toBeVisible({ timeout: 5000 })
  })

  test('logo in header links to landing page', async ({ page }) => {
    const logo = page.locator('.logo').first()
    await expect(logo).toHaveAttribute('href', '/')
  })

  test('dashboard greeting is visible', async ({ page }) => {
    const greeting = page.locator('.dashboard-greeting')
    await expect(greeting).toBeVisible({ timeout: 5000 })
  })

  test('weekly menu shows day names', async ({ page }) => {
    // Check for at least some Swedish day names
    const dayNames = ['Mån', 'Tis', 'Ons', 'Tor', 'Fre']
    let found = 0
    for (const day of dayNames) {
      const locator = page.getByText(day, { exact: false })
      if (await locator.first().isVisible({ timeout: 2000 }).catch(() => false)) {
        found++
      }
    }
    expect(found).toBeGreaterThanOrEqual(3) // At least 3 days visible
  })
})
