import { test, expect } from '@playwright/test'

test.describe('Landing page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('shows hero with title and CTA', async ({ page }) => {
    await expect(page.locator('.hero-title')).toBeVisible()
    await expect(page.getByText('Kom igång gratis')).toBeVisible()
    await expect(page.getByText('Logga in')).toBeVisible()
  })

  test('shows navigation with logo and links', async ({ page }) => {
    await expect(page.getByText('Måltiden').first()).toBeVisible()
    await expect(page.getByText('Om oss')).toBeVisible()
  })

  test('shows features section with cards', async ({ page }) => {
    const features = page.locator('.features-section, [class*="features"]').first()
    await expect(features).toBeVisible()
    // Should have multiple feature cards
    const cards = page.locator('[class*="feature-card"], [class*="FeatureCard"]')
    await expect(cards.first()).toBeVisible({ timeout: 5000 })
  })

  test('hero subtitle text is visible', async ({ page }) => {
    await expect(page.locator('.hero-title')).toBeVisible()
    // Hero should have a descriptive subtitle
    const subtitle = page.locator('[class*="hero-subtitle"], [class*="hero"] p').first()
    await expect(subtitle).toBeVisible()
  })

  test('CTA button navigates to register', async ({ page }) => {
    await page.getByText('Kom igång gratis').click()
    await expect(page).toHaveURL(/\/register/)
  })

  test('login link navigates to login page', async ({ page }) => {
    await page.getByText('Logga in').click()
    await expect(page).toHaveURL(/\/login/)
  })

  test('about link navigates to about page', async ({ page }) => {
    await page.getByText('Om oss').click()
    await expect(page).toHaveURL(/\/about/)
  })

  test('page has correct title', async ({ page }) => {
    const title = await page.title()
    expect(title.length).toBeGreaterThan(0)
  })

  test('page renders without console errors', async ({ page }) => {
    const errors: string[] = []
    page.on('console', (msg) => {
      if (msg.type() === 'error') errors.push(msg.text())
    })
    await page.goto('/')
    await page.waitForTimeout(1000)
    // Filter out known non-critical errors: favicons, external resource failures
    // (e.g. fonts, analytics), and network/proxy errors that aren't app bugs.
    const criticalErrors = errors.filter(
      (e) =>
        !e.includes('favicon') &&
        !e.includes('net::ERR') &&
        !e.includes('Failed to load resource'),
    )
    expect(criticalErrors).toHaveLength(0)
  })

  test('CTA section at bottom is visible', async ({ page }) => {
    // Scroll to bottom to find the CTA section
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight))
    await page.waitForTimeout(500)
    const cta = page.locator('[class*="cta-section"], [class*="Cta"]').first()
    await expect(cta).toBeVisible({ timeout: 5000 })
  })

  test('landing page is responsive - no horizontal scroll', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/')
    await page.waitForTimeout(500)
    const scrollWidth = await page.evaluate(() => document.documentElement.scrollWidth)
    const clientWidth = await page.evaluate(() => document.documentElement.clientWidth)
    expect(scrollWidth).toBeLessThanOrEqual(clientWidth + 1) // 1px tolerance
  })
})
