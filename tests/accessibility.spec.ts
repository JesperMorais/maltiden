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
    // Log violations but don't fail — we collect them for the report
    if (violations.length > 0) {
      console.log(`[A11Y] Landing page: ${violations.length} violations`)
      for (const v of violations) {
        console.log(`  - [${v.impact}] ${v.id}: ${v.description}`)
      }
    }
  })

  test('Login page', async ({ page }) => {
    await page.goto('/login')
    await page.waitForTimeout(1000)
    const violations = await runAxeOnView(page, 'Login page')
    if (violations.length > 0) {
      console.log(`[A11Y] Login page: ${violations.length} violations`)
      for (const v of violations) {
        console.log(`  - [${v.impact}] ${v.id}: ${v.description}`)
      }
    }
  })

  test('Register page', async ({ page }) => {
    await page.goto('/register')
    await page.waitForTimeout(1000)
    const violations = await runAxeOnView(page, 'Register page')
    if (violations.length > 0) {
      console.log(`[A11Y] Register page: ${violations.length} violations`)
      for (const v of violations) {
        console.log(`  - [${v.impact}] ${v.id}: ${v.description}`)
      }
    }
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
    if (violations.length > 0) {
      console.log(`[A11Y] Dashboard: ${violations.length} violations`)
      for (const v of violations) {
        console.log(`  - [${v.impact}] ${v.id}: ${v.description}`)
      }
    }
  })

  test('Recipe list', async ({ page }) => {
    await navigateTo(page, '/recipes')
    await expect(page).toHaveURL(/\/recipes/)
    await page.waitForTimeout(2000)
    const violations = await runAxeOnView(page, 'Recipe list')
    if (violations.length > 0) {
      console.log(`[A11Y] Recipe list: ${violations.length} violations`)
      for (const v of violations) {
        console.log(`  - [${v.impact}] ${v.id}: ${v.description}`)
      }
    }
  })

  test('Add recipe tab', async ({ page }) => {
    await navigateTo(page, '/recipes')
    await expect(page).toHaveURL(/\/recipes/)
    await page.waitForTimeout(1000)
    await page.getByRole('tab', { name: 'Lägg till' }).click()
    await page.waitForTimeout(1000)
    const violations = await runAxeOnView(page, 'Add recipe tab')
    if (violations.length > 0) {
      console.log(`[A11Y] Add recipe tab: ${violations.length} violations`)
      for (const v of violations) {
        console.log(`  - [${v.impact}] ${v.id}: ${v.description}`)
      }
    }
  })

  test('Menu generator', async ({ page }) => {
    await page.getByRole('button', { name: /Generera meny/ }).click()
    await expect(page).toHaveURL(/\/menu\/generate/)
    await page.waitForTimeout(1500)
    const violations = await runAxeOnView(page, 'Menu generator')
    if (violations.length > 0) {
      console.log(`[A11Y] Menu generator: ${violations.length} violations`)
      for (const v of violations) {
        console.log(`  - [${v.impact}] ${v.id}: ${v.description}`)
      }
    }
  })

  test('Shopping list', async ({ page }) => {
    const widget = page.locator('[class*="shopping-widget"]').first()
    if (await widget.isVisible({ timeout: 5000 }).catch(() => false)) {
      await widget.click()
      await expect(page).toHaveURL(/\/shopping-list/)
      await page.waitForTimeout(1500)
      const violations = await runAxeOnView(page, 'Shopping list')
      if (violations.length > 0) {
        console.log(`[A11Y] Shopping list: ${violations.length} violations`)
        for (const v of violations) {
          console.log(`  - [${v.impact}] ${v.id}: ${v.description}`)
        }
      }
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
    if (violations.length > 0) {
      console.log(`[A11Y] Settings modal: ${violations.length} violations`)
      for (const v of violations) {
        console.log(`  - [${v.impact}] ${v.id}: ${v.description}`)
      }
    }
  })

  test('Invite modal', async ({ page }) => {
    await page.waitForSelector('.dashboard-content', { timeout: 8000 })
    await page.waitForTimeout(1500)
    const inviteBtn = page.getByText('Bjud in fler')
    if (await inviteBtn.isVisible({ timeout: 5000 }).catch(() => false)) {
      await inviteBtn.click()
      await page.waitForTimeout(800)
      const violations = await runAxeOnView(page, 'Invite modal')
      if (violations.length > 0) {
        console.log(`[A11Y] Invite modal: ${violations.length} violations`)
        for (const v of violations) {
          console.log(`  - [${v.impact}] ${v.id}: ${v.description}`)
        }
      }
    }
  })
})

// After all tests, write a JSON summary
test.afterAll(async () => {
  fs.mkdirSync(RESULTS_DIR, { recursive: true })
  const summaryPath = path.join(RESULTS_DIR, 'axe-results.json')
  fs.writeFileSync(summaryPath, JSON.stringify(allResults, null, 2))
})
