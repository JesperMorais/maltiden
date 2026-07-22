import { test, expect } from '@playwright/test'
import { login, navigateTo } from './helpers'
import path from 'path'

const VIEWPORTS = [
  { name: '375-iphone-se', width: 375, height: 667 },
  { name: '390-iphone-14', width: 390, height: 844 },
  { name: '768-ipad-mini', width: 768, height: 1024 },
  { name: '1440-desktop', width: 1440, height: 900 },
] as const

const SCREENSHOT_DIR = path.join(__dirname, 'screenshots')

function screenshotPath(view: string, viewport: string): string {
  return path.join(SCREENSHOT_DIR, `${view}--${viewport}.png`)
}

// ─── Logged-out views ─────────────────────────────────────────────

test.describe('Screenshots: Logged-out views', () => {
  for (const vp of VIEWPORTS) {
    test.describe(`@ ${vp.name}`, () => {
      test.beforeEach(async ({ page }) => {
        await page.setViewportSize({ width: vp.width, height: vp.height })
      })

      test(`Landing page`, async ({ page }) => {
        await page.goto('/')
        await page.waitForTimeout(1500)
        await page.screenshot({
          path: screenshotPath('landing', vp.name),
          fullPage: true,
        })
      })

      test(`Login page`, async ({ page }) => {
        await page.goto('/login')
        await page.waitForTimeout(1000)
        await page.screenshot({
          path: screenshotPath('login', vp.name),
          fullPage: true,
        })
      })

      test(`Register page`, async ({ page }) => {
        await page.goto('/register')
        await page.waitForTimeout(1000)
        await page.screenshot({
          path: screenshotPath('register', vp.name),
          fullPage: true,
        })
      })
    })
  }
})

// ─── Logged-in views ──────────────────────────────────────────────

test.describe('Screenshots: Logged-in views', () => {
  for (const vp of VIEWPORTS) {
    test.describe(`@ ${vp.name}`, () => {
      test.beforeEach(async ({ page }) => {
        await page.setViewportSize({ width: vp.width, height: vp.height })
        await login(page)
      })

      test(`Dashboard (with menu)`, async ({ page }) => {
        await page.waitForSelector('.dashboard-content', { timeout: 8000 })
        await page.waitForTimeout(2000) // let animations settle
        await page.screenshot({
          path: screenshotPath('dashboard', vp.name),
          fullPage: true,
        })
      })

      test(`Recipe list`, async ({ page }) => {
        await navigateTo(page, '/recipes')
        await expect(page).toHaveURL(/\/recipes/)
        await page.waitForTimeout(2000)
        await page.screenshot({
          path: screenshotPath('recipes-list', vp.name),
          fullPage: true,
        })
      })

      test(`Recipe detail modal`, async ({ page }) => {
        await navigateTo(page, '/recipes')
        await expect(page).toHaveURL(/\/recipes/)
        await page.waitForTimeout(2000)

        // Click first recipe card to open detail
        const recipeCard = page.locator('[class*="recipe-card"], [class*="RecipeCard"]').first()
        if (await recipeCard.isVisible({ timeout: 5000 }).catch(() => false)) {
          await recipeCard.click()
          await page.waitForTimeout(1000)
          await page.screenshot({
            path: screenshotPath('recipe-detail-modal', vp.name),
            fullPage: true,
          })
        }
      })

      test(`Add recipe / AI parse flow`, async ({ page }) => {
        await navigateTo(page, '/recipes')
        await expect(page).toHaveURL(/\/recipes/)
        await page.waitForTimeout(1000)

        await page.getByRole('tab', { name: 'Lägg till' }).click()
        await page.waitForTimeout(1000)
        await page.screenshot({
          path: screenshotPath('add-recipe', vp.name),
          fullPage: true,
        })
      })

      test(`Menu generator (empty state)`, async ({ page }) => {
        await page.getByRole('button', { name: /Generera meny/ }).click()
        await expect(page).toHaveURL(/\/menu\/generate/)
        await page.waitForTimeout(1500)
        await page.screenshot({
          path: screenshotPath('menu-generator-empty', vp.name),
          fullPage: true,
        })
      })

      test(`Menu generator (generated)`, async ({ page }) => {
        await page.getByRole('button', { name: /Generera meny/ }).click()
        await expect(page).toHaveURL(/\/menu\/generate/)
        await page.waitForTimeout(1000)

        const genBtn = page.locator('.generate-button')
        if (await genBtn.isVisible({ timeout: 5000 }).catch(() => false)) {
          await genBtn.click()
          await page.waitForTimeout(4000) // wait for animation
          await page.screenshot({
            path: screenshotPath('menu-generator-filled', vp.name),
            fullPage: true,
          })
        }
      })

      test(`Shopping list (with items)`, async ({ page }) => {
        // Navigate to shopping list
        const widget = page.locator('[class*="shopping-widget"]').first()
        if (await widget.isVisible({ timeout: 5000 }).catch(() => false)) {
          await widget.click()
          await expect(page).toHaveURL(/\/shopping-list/)
          await page.waitForTimeout(1500)
          await page.screenshot({
            path: screenshotPath('shopping-list', vp.name),
            fullPage: true,
          })
        }
      })

      test(`Settings modal`, async ({ page }) => {
        await page.waitForSelector('.dashboard-content', { timeout: 8000 })
        await page.waitForTimeout(1500)

        // Open user dropdown and click settings
        await page.locator('.user-button').click()
        await page.waitForTimeout(300)
        await page.getByText('Inställningar').click()
        await page.waitForTimeout(800)

        await page.screenshot({
          path: screenshotPath('settings-modal', vp.name),
          fullPage: true,
        })
      })

      test(`Household invite modal`, async ({ page }) => {
        await page.waitForSelector('.dashboard-content', { timeout: 8000 })
        await page.waitForTimeout(1500)

        // Click "Bjud in fler" in the household widget
        const inviteBtn = page.getByText('Bjud in fler')
        if (await inviteBtn.isVisible({ timeout: 5000 }).catch(() => false)) {
          await inviteBtn.click()
          await page.waitForTimeout(800)
          await page.screenshot({
            path: screenshotPath('invite-modal', vp.name),
            fullPage: true,
          })
        }
      })
    })
  }
})
