import { test, expect } from '@playwright/test'

test.describe('About page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/about')
  })

  test('shows page title', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Om Måltiden', exact: true })).toBeVisible({ timeout: 5000 })
  })

  test('shows hero badge', async ({ page }) => {
    await expect(page.getByText('Vår historia')).toBeVisible({ timeout: 5000 })
  })

  test('shows hero subtitle', async ({ page }) => {
    await expect(
      page.getByText(/måltidsplanering till en fröjd/i),
    ).toBeVisible({ timeout: 5000 })
  })

  test('shows back link', async ({ page }) => {
    const backLink = page.getByRole('link', { name: /Tillbaka/ })
    await expect(backLink).toBeVisible()
  })

  test('back link navigates to landing', async ({ page }) => {
    await page.getByRole('link', { name: /Tillbaka/ }).click()
    await expect(page).toHaveURL('/')
  })

  test('shows mission section', async ({ page }) => {
    await expect(page.getByText('Vår Mission')).toBeVisible({ timeout: 5000 })
  })

  test('shows mission description text', async ({ page }) => {
    await expect(
      page.getByText(/smart måltidsplaneringsapp skapad för svenska familjer/i),
    ).toBeVisible({ timeout: 5000 })
  })

  test('shows three value cards', async ({ page }) => {
    await expect(page.getByText('Enklare vardag')).toBeVisible({ timeout: 5000 })
    await expect(page.getByText('Mindre matsvinn')).toBeVisible()
    await expect(page.getByText('Familjetid')).toBeVisible()
  })

  test('shows mission quote', async ({ page }) => {
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight / 3))
    await page.waitForTimeout(500)
    await expect(
      page.getByText(/varje familj ska kunna njuta av god mat/i),
    ).toBeVisible({ timeout: 5000 })
  })

  test('shows team section', async ({ page }) => {
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight / 2))
    await page.waitForTimeout(500)
    await expect(page.getByText('Teamet bakom Måltiden')).toBeVisible({ timeout: 5000 })
  })

  test('shows team members', async ({ page }) => {
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight / 2))
    await page.waitForTimeout(500)

    await expect(page.getByText('David')).toBeVisible({ timeout: 5000 })
    await expect(page.getByText('Jesper')).toBeVisible()
  })

  test('shows team member roles', async ({ page }) => {
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight / 2))
    await page.waitForTimeout(500)

    // Both team members have the role 'Utvecklare'
    const roles = page.locator('.member-role')
    await expect(roles.first()).toBeVisible({ timeout: 5000 })
    const count = await roles.count()
    expect(count).toBe(2)
    await expect(roles.first()).toHaveText('Utvecklare')
    await expect(roles.nth(1)).toHaveText('Utvecklare')
  })

  test('shows team member names with initials', async ({ page }) => {
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight / 2))
    await page.waitForTimeout(500)

    // Team cards show member initials (D and J) instead of descriptions
    const initials = page.locator('.member-initial')
    await expect(initials.first()).toBeVisible({ timeout: 5000 })
    const count = await initials.count()
    expect(count).toBe(2)
  })

  test('shows team member avatars', async ({ page }) => {
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight / 2))
    await page.waitForTimeout(500)

    const avatars = page.locator('.member-initial')
    await expect(avatars.first()).toBeVisible({ timeout: 5000 })
    const count = await avatars.count()
    expect(count).toBe(2)
  })

  test('shows CTA section at bottom', async ({ page }) => {
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight))
    await page.waitForTimeout(500)

    await expect(page.getByText('Redo att förenkla din matvardag?')).toBeVisible({ timeout: 5000 })
  })

  test('CTA button navigates to register', async ({ page }) => {
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight))
    await page.waitForTimeout(500)

    await page.getByText('Kom igång gratis').last().click()
    await expect(page).toHaveURL(/\/register/)
  })

  test('CTA section subtitle is visible', async ({ page }) => {
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight))
    await page.waitForTimeout(500)

    await expect(
      page.getByText(/Gör måltidsplaneringen enklare för hela familjen/),
    ).toBeVisible({ timeout: 5000 })
  })

  test('page is accessible without login', async ({ page }) => {
    // About page should load without auth
    await expect(page).toHaveURL(/\/about/)
    await expect(page.getByRole('heading', { name: 'Om Måltiden', exact: true })).toBeVisible({ timeout: 5000 })
  })
})
