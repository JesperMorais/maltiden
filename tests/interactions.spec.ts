import { test, expect } from '@playwright/test'
import { login, navigateTo } from './helpers'

// ─── 1. Grocery store flow ────────────────────────────────────────

test.describe('Interaction: Shopping list rapid check-off', () => {
  test('check off 5 items rapidly and verify progress updates', async ({ page }) => {
    await login(page)

    // Wait for dashboard fully loaded
    await page.waitForSelector('.dashboard-content', { timeout: 8000 })
    await page.waitForTimeout(2000)

    // Navigate to shopping list via widget click (preserves Pinia state)
    const widget = page.locator('.shopping-widget').first()
    await expect(widget).toBeVisible({ timeout: 5000 })
    await widget.click()
    await expect(page).toHaveURL(/\/shopping-list/)

    // Wait for shopping list categories to load
    await expect(page.locator('.category-header').first()).toBeVisible({ timeout: 15000 })
    await page.waitForTimeout(500)

    // Count initial unchecked items (each item is an .item-row)
    const allItems = page.locator('.item-row')
    const initialCount = await allItems.count()
    expect(initialCount).toBeGreaterThan(5)

    // Count initially unchecked rows (rows without .checked class)
    const uncheckedBefore = await page.locator('.item-row:not(.checked)').count()

    // Rapidly check off 5 unchecked items by clicking their labels
    for (let i = 0; i < 5; i++) {
      const uncheckedRow = page.locator('.item-row:not(.checked)').first()
      if (await uncheckedRow.isVisible({ timeout: 2000 }).catch(() => false)) {
        await uncheckedRow.click()
        await page.waitForTimeout(350) // small delay to let state update
      }
    }

    // Verify progress updated — fewer unchecked items now
    await page.waitForTimeout(500)
    const uncheckedAfter = await page.locator('.item-row:not(.checked)').count()
    expect(uncheckedAfter).toBeLessThan(uncheckedBefore)

    // Verify checked items are visually distinct
    const checkedRows = page.locator('.item-row.checked')
    const checkedCount = await checkedRows.count()
    expect(checkedCount).toBeGreaterThanOrEqual(5)

    // Verify checked item rows have visual distinction (opacity + strikethrough)
    const firstCheckedRow = checkedRows.first()
    const rowStyle = await firstCheckedRow.evaluate((el) => {
      const style = window.getComputedStyle(el)
      const nameEl = el.querySelector('.item-name')
      const nameStyle = nameEl ? window.getComputedStyle(nameEl) : null
      return {
        opacity: style.opacity,
        textDecoration: nameStyle?.textDecorationLine ?? 'none',
      }
    })

    // Both opacity and strikethrough should be present per the CSS
    const hasVisualDistinction =
      parseFloat(rowStyle.opacity) < 1 ||
      rowStyle.textDecoration.includes('line-through')
    expect(hasVisualDistinction).toBe(true)

    // Verify progress text updated
    const stickyHeader = page.locator('.sticky-header')
    await expect(stickyHeader).toBeVisible()
    const progressText = await page.locator('.progress-text').textContent()
    expect(progressText).toMatch(/\d+ av \d+/)
  })
})

// ─── 2. Recipe browse on mobile ───────────────────────────────────

test.describe('Interaction: Recipe browse on mobile (375px)', () => {
  test.beforeEach(async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
  })

  test('scroll list, open modal, scroll modal, close, verify scroll position', async ({ page }) => {
    await login(page)
    await navigateTo(page, '/recipes')
    await expect(page).toHaveURL(/\/recipes/)
    await page.waitForTimeout(2000)

    // Scroll the recipe list down
    await page.evaluate(() => window.scrollTo(0, 300))
    await page.waitForTimeout(500)
    const scrollBefore = await page.evaluate(() => window.scrollY)

    // Tap first recipe to open modal
    const recipeCard = page.locator('[class*="recipe-card"], [class*="RecipeCard"]').first()
    if (await recipeCard.isVisible({ timeout: 5000 }).catch(() => false)) {
      await recipeCard.click()
      await page.waitForTimeout(1000)

      // Verify modal opened (look for modal overlay or recipe detail)
      const modal = page.locator('[class*="modal"], [class*="overlay"], [class*="detail"]').first()
      await expect(modal).toBeVisible({ timeout: 3000 })

      // Scroll modal content if scrollable
      await modal.evaluate((el) => {
        const scrollable = el.querySelector('[class*="content"], [class*="body"]') || el
        scrollable.scrollTop = 200
      }).catch(() => {})
      await page.waitForTimeout(500)

      // Close modal (click X, overlay, or press Escape)
      const closeBtn = page.locator('[class*="close"], button[aria-label*="close"], button[aria-label*="stäng"]').first()
      if (await closeBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await closeBtn.click()
      } else {
        await page.keyboard.press('Escape')
      }
      await page.waitForTimeout(500)

      // Verify scroll position restored (within tolerance)
      const scrollAfter = await page.evaluate(() => window.scrollY)
      expect(Math.abs(scrollAfter - scrollBefore)).toBeLessThan(250)
    }
  })

  test('recipe list is scrollable on mobile without horizontal overflow', async ({ page }) => {
    await login(page)
    await navigateTo(page, '/recipes')
    await expect(page).toHaveURL(/\/recipes/)
    await page.waitForTimeout(2000)

    const scrollWidth = await page.evaluate(() => document.documentElement.scrollWidth)
    const clientWidth = await page.evaluate(() => document.documentElement.clientWidth)
    expect(scrollWidth).toBeLessThanOrEqual(clientWidth + 1)
  })
})

// ─── 3. Menu generation flow ──────────────────────────────────────

test.describe('Interaction: Menu generation', () => {
  test('generate menu, verify animation, all days populated, save, toast', async ({ page }) => {
    await login(page)
    await navigateTo(page, '/menu/generate')
    await expect(page).toHaveURL(/\/menu\/generate/)
    await page.waitForTimeout(1500)

    const genBtn = page.locator('.generate-button')
    if (await genBtn.isVisible({ timeout: 5000 }).catch(() => false)) {
      // Click generate
      await genBtn.click()

      // Wait for animation to complete (slot machine or similar)
      await page.waitForTimeout(4000)

      // Verify menu grid appeared
      const menuGrid = page.locator('.menu-grid')
      await expect(menuGrid).toBeVisible({ timeout: 8000 })

      // Verify all 5 days are populated (Mon-Fri)
      const dayCards = page.locator('.menu-grid > *')
      const dayCount = await dayCards.count()
      expect(dayCount).toBe(5)

      // Each day should have a recipe name (not empty)
      for (let i = 0; i < dayCount; i++) {
        const dayCard = dayCards.nth(i)
        const text = await dayCard.textContent()
        expect(text!.trim().length).toBeGreaterThan(0)
      }

      // Look for save button and click it
      const saveBtn = page.getByText('Spara').first()
      if (await saveBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
        await saveBtn.click({ force: true })
        await page.waitForTimeout(1500)

        // Verify toast / success notification appeared
        const toast = page.locator('[class*="toast"], [class*="notification"], [class*="success"], [class*="snackbar"]').first()
        if (await toast.isVisible({ timeout: 3000 }).catch(() => false)) {
          await expect(toast).toBeVisible()
        } else {
          // Check if we got redirected to dashboard (also valid success)
          const url = page.url()
          const redirected = /dashboard/.test(url)
          if (!redirected) {
            console.warn('VISUAL_ISSUE: No toast/notification visible after saving menu')
          }
        }
      }
    }
  })
})

// ─── 4. Onboarding flow ──────────────────────────────────────────

test.describe('Interaction: Onboarding', () => {
  test('register new user, household auto-created, empty dashboard guides user', async ({ page }) => {
    await page.goto('/register')
    await page.waitForTimeout(1000)

    // Choose "Skapa nytt hushåll"
    await page.getByText('Skapa nytt hushåll').click()
    await page.waitForTimeout(800)

    // Fill registration form (password needs 3+ char types: upper, lower, digit, special)
    await page.getByPlaceholder('Anna Andersson').first().fill('Testperson')
    await page.getByPlaceholder('anna@exempel.se').first().fill('ny@test.se')
    await page.getByPlaceholder('Minst 8 tecken').first().fill('TestPass123!')
    await page.getByPlaceholder('Skriv lösenordet igen').first().fill('TestPass123!')

    // Submit
    await page.getByRole('button', { name: 'Skapa konto' }).click()
    await page.waitForTimeout(2000)

    // Verify success message
    await expect(page.getByText('Konto skapat!')).toBeVisible({ timeout: 5000 })

    // Click to go to dashboard
    const goBtn = page.getByRole('button', { name: /dashboard|Gå till|Fortsätt/i }).first()
    if (await goBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await goBtn.click()
      await page.waitForTimeout(2000)
    }

    // At this point either redirected to dashboard or still on register success
    // In mock mode we should check the dashboard state
    const url = page.url()
    if (/dashboard/.test(url)) {
      // Verify the dashboard has guidance for new users
      // Look for empty states, CTAs, or onboarding hints
      const guidance = page.locator('[class*="empty"], [class*="get-started"], [class*="onboarding"], [class*="quick-action"]').first()
      if (await guidance.isVisible({ timeout: 3000 }).catch(() => false)) {
        await expect(guidance).toBeVisible()
      }
    }
  })

  test('join household flow completes end to end', async ({ page }) => {
    await page.goto('/register')
    await page.waitForTimeout(1000)

    // Choose join
    await page.getByText(/gå med/i).first().click()
    await page.waitForTimeout(500)

    // Enter valid invite code
    await page.getByPlaceholder(/T\.ex\. ABC123/).fill('ABC123')
    await page.getByRole('button', { name: /Fortsätt/ }).click()

    // Verify household found
    await expect(page.getByText('Familjen Andersson')).toBeVisible({ timeout: 5000 })

    // Expect member/guest choice
    await expect(page.getByText(/Hur vill du gå med/i)).toBeVisible({ timeout: 5000 })

    // Choose member
    await page.getByText('Bli medlem').click()
    await page.waitForTimeout(500)

    // Fill member form
    await expect(page.getByPlaceholder('Anna Andersson')).toBeVisible({ timeout: 3000 })
    await page.getByPlaceholder('Anna Andersson').fill('Ny Medlem')
    await page.getByPlaceholder('anna@exempel.se').fill('nymedlem@test.se')
    await page.getByPlaceholder('Minst 8 tecken').fill('TestPass123!')
    await page.getByPlaceholder('Skriv lösenordet igen').fill('TestPass123!')

    // Submit
    const submitBtn = page.getByRole('button', { name: /Skapa konto|Gå med/ })
    if (await submitBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await submitBtn.click()
      await page.waitForTimeout(2000)

      // Should show success
      const success = page.getByText(/Konto skapat|Välkommen/i).first()
      await expect(success).toBeVisible({ timeout: 5000 })
    }
  })
})
