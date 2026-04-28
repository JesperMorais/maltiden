import { test, expect } from '@playwright/test'
import AxeBuilder from '@axe-core/playwright'
import { login, navigateTo } from './helpers'
import fs from 'fs'
import path from 'path'

const RESULTS_DIR = path.join(__dirname, 'screenshots')

interface AxeViolation {
  id: string
  impact: string | undefined
  description: string
  helpUrl: string
  nodes: { html: string; target: string[] }[]
}

interface ViewResult {
  view: string
  violations: AxeViolation[]
  passes: number
  incomplete: number
}

const allResults: ViewResult[] = []

function saveResult(view: string, violations: AxeViolation[], passes: number, incomplete: number) {
  allResults.push({ view, violations, passes, incomplete })
}

// Logs all violations and fails the test if any are 'critical'.
// 'serious'/'moderate'/'minor' are warnings only — they get logged but don't fail.
function reportViolations(view: string, violations: AxeViolation[]) {
  if (violations.length === 0) return
  console.log(`[A11Y] ${view}: ${violations.length} violations`)
  for (const v of violations) {
    console.log(`  - [${v.impact}] ${v.id}: ${v.description}`)
  }
  const critical = violations.filter((v) => v.impact === 'critical')
  if (critical.length > 0) {
    const summary = critical.map((v) => `${v.id} (${v.nodes.length} node(s))`).join(', ')
    throw new Error(`[A11Y] ${view}: ${critical.length} critical violation(s): ${summary}`)
  }
}

async function runAxeOnView(
  page: import('@playwright/test').Page,
  viewName: string,
): Promise<AxeViolation[]> {
  const results = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
    .analyze()

  const violations = results.violations as AxeViolation[]
  saveResult(viewName, violations, results.passes.length, results.incomplete.length)

  return violations
}

// ─── Logged-out views ─────────────────────────────────────────────

test.describe('Accessibility: Logged-out views', () => {
  test('Landing page', async ({ page }) => {
    await page.goto('/')
    await page.waitForTimeout(1500)
    const violations = await runAxeOnView(page, 'Landing page')
    reportViolations('Landing page', violations)
  })

  test('Login page', async ({ page }) => {
    await page.goto('/login')
    await page.waitForTimeout(1000)
    const violations = await runAxeOnView(page, 'Login page')
    reportViolations('Login page', violations)
  })

  test('Register page', async ({ page }) => {
    await page.goto('/register')
    await page.waitForTimeout(1000)
    const violations = await runAxeOnView(page, 'Register page')
    reportViolations('Register page', violations)
  })
})

// ─── Logged-in views ──────────────────────────────────────────────

test.describe('Accessibility: Logged-in views', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('Dashboard', async ({ page }) => {
    await page.waitForSelector('.dashboard-content', { timeout: 8000 })
    await page.waitForTimeout(2000)
    const violations = await runAxeOnView(page, 'Dashboard')
    reportViolations('Dashboard', violations)
  })

  test('Recipe list', async ({ page }) => {
    await navigateTo(page, '/recipes')
    await expect(page).toHaveURL(/\/recipes/)
    await page.waitForTimeout(2000)
    const violations = await runAxeOnView(page, 'Recipe list')
    reportViolations('Recipe list', violations)
  })

  test('Add recipe tab', async ({ page }) => {
    await navigateTo(page, '/recipes')
    await expect(page).toHaveURL(/\/recipes/)
    await page.waitForTimeout(1000)
    await page.getByRole('tab', { name: 'Lägg till' }).click()
    await page.waitForTimeout(1000)
    const violations = await runAxeOnView(page, 'Add recipe tab')
    reportViolations('Add recipe tab', violations)
  })

  test('Menu generator', async ({ page }) => {
    await page.getByRole('button', { name: /Generera meny/ }).click()
    await expect(page).toHaveURL(/\/menu\/generate/)
    await page.waitForTimeout(1500)
    const violations = await runAxeOnView(page, 'Menu generator')
    reportViolations('Menu generator', violations)
  })

  test('Shopping list', async ({ page }) => {
    const widget = page.locator('[class*="shopping-widget"]').first()
    if (await widget.isVisible({ timeout: 5000 }).catch(() => false)) {
      await widget.click()
      await expect(page).toHaveURL(/\/shopping-list/)
      await page.waitForTimeout(1500)
      const violations = await runAxeOnView(page, 'Shopping list')
      reportViolations('Shopping list', violations)
    }
  })

  test('Settings modal', async ({ page }) => {
    await page.waitForSelector('.dashboard-content', { timeout: 8000 })
    await page.waitForTimeout(1500)
    await page.locator('.user-button').click()
    await page.waitForTimeout(300)
    await page.getByText('Inställningar').click()
    await page.waitForTimeout(800)
    const violations = await runAxeOnView(page, 'Settings modal')
    reportViolations('Settings modal', violations)
  })

  test('Invite modal', async ({ page }) => {
    await page.waitForSelector('.dashboard-content', { timeout: 8000 })
    await page.waitForTimeout(1500)
    const inviteBtn = page.getByText('Bjud in fler')
    if (await inviteBtn.isVisible({ timeout: 5000 }).catch(() => false)) {
      await inviteBtn.click()
      await page.waitForTimeout(800)
      const violations = await runAxeOnView(page, 'Invite modal')
      reportViolations('Invite modal', violations)
    }
  })
})

// After all tests, write a JSON summary
test.afterAll(async () => {
  fs.mkdirSync(RESULTS_DIR, { recursive: true })
  const summaryPath = path.join(RESULTS_DIR, 'axe-results.json')
  fs.writeFileSync(summaryPath, JSON.stringify(allResults, null, 2))
})
