import { test, expect, type Page } from '@playwright/test'
import { registerUser, loginUser } from '../e2e/helpers'

/**
 * Feedback Widget E2E Tests
 *
 * Tests run against the mock dev server (port 5174).
 * The feedback API is mocked (returns fb_mock-{timestamp} after 500ms).
 */

// Helper: clear feedback cooldown so button is visible
async function clearCooldown(page: Page) {
  await page.evaluate(() => localStorage.removeItem('feedback_cooldown_until'))
}

// Helper: register + land on dashboard with feedback button ready
// Each test gets a fresh browser context so there's no cooldown to worry about.
async function setupAuthenticatedUser(page: Page) {
  await registerUser(page)
  // Wait for async FeedbackWidget to load + 1s animation delay
  await page.waitForTimeout(2000)
}

// Helper: open the feedback modal and wait for it to be ready
async function openFeedbackModal(page: Page) {
  await page.locator('button[aria-label="Ge feedback"]').click()
  await expect(page.locator('[role="dialog"]')).toBeVisible({ timeout: 3_000 })
}

// Helper: select a mood and wait for step 2 (categories) to be ready
async function selectMoodAndWaitForCategories(page: Page, mood: string) {
  const modal = page.locator('[role="dialog"]')
  await modal.getByRole('button', { name: mood }).click()
  await expect(modal.getByText('Vad gäller det?')).toBeVisible({ timeout: 3_000 })
}

// Helper: advance to step 3 (comment) via "Nästa" and wait for it to be ready
async function advanceToCommentStep(page: Page) {
  const modal = page.locator('[role="dialog"]')
  await modal.getByRole('button', { name: 'Nästa' }).click()
  await expect(modal.getByText('Berätta mer')).toBeVisible({ timeout: 3_000 })
}

// ─── Visibility & Positioning ───────────────────────────────────────────────

test.describe('Feedback Widget — Visibility', () => {
  test('feedback button visible on dashboard (authenticated)', async ({ page }) => {
    await setupAuthenticatedUser(page)
    const btn = page.locator('button[aria-label="Ge feedback"]')
    await expect(btn).toBeVisible({ timeout: 5_000 })
  })

  test('feedback button NOT visible on landing page (unauthenticated)', async ({ page }) => {
    await page.goto('/')
    await page.evaluate(() => localStorage.removeItem('maltiden_token'))
    await page.goto('/')
    // Wait for page to settle
    await page.waitForTimeout(1500)
    const btn = page.locator('button[aria-label="Ge feedback"]')
    await expect(btn).not.toBeVisible()
  })

  test('feedback button NOT visible on login page', async ({ page }) => {
    await page.goto('/login')
    await page.waitForTimeout(1500)
    const btn = page.locator('button[aria-label="Ge feedback"]')
    await expect(btn).not.toBeVisible()
  })

  test('feedback button NOT visible on register page', async ({ page }) => {
    await page.goto('/register')
    await page.waitForTimeout(1500)
    const btn = page.locator('button[aria-label="Ge feedback"]')
    await expect(btn).not.toBeVisible()
  })

  test('feedback button has high z-index (visible after scroll)', async ({ page }) => {
    await setupAuthenticatedUser(page)
    // Scroll the page down
    await page.evaluate(() => window.scrollTo(0, 500))
    await page.waitForTimeout(300)
    const btn = page.locator('button[aria-label="Ge feedback"]')
    await expect(btn).toBeVisible()
  })
})

// ─── Visibility at different viewport sizes ─────────────────────────────────

test.describe('Feedback Widget — Mobile (390px)', () => {
  test.use({ viewport: { width: 390, height: 844 } })

  test('feedback button visible on mobile', async ({ page }) => {
    await setupAuthenticatedUser(page)
    const btn = page.locator('button[aria-label="Ge feedback"]')
    await expect(btn).toBeVisible({ timeout: 5_000 })
  })

  test('trigger label hidden on mobile (icon-only)', async ({ page }) => {
    await setupAuthenticatedUser(page)
    const label = page.locator('.trigger-label')
    await expect(label).not.toBeVisible()
  })

  test('feedback button does not overlap bottom nav on mobile', async ({ page }) => {
    await setupAuthenticatedUser(page)
    const btn = page.locator('button[aria-label="Ge feedback"]')
    await expect(btn).toBeVisible({ timeout: 5_000 })

    const btnBox = await btn.boundingBox()
    expect(btnBox).not.toBeNull()
    // Button should be within viewport
    expect(btnBox!.x + btnBox!.width).toBeLessThanOrEqual(390)
    expect(btnBox!.y + btnBox!.height).toBeLessThanOrEqual(844)
  })
})

test.describe('Feedback Widget — Desktop (1440px)', () => {
  test.use({ viewport: { width: 1440, height: 900 } })

  test('feedback button visible with label on desktop', async ({ page }) => {
    await setupAuthenticatedUser(page)
    const btn = page.locator('button[aria-label="Ge feedback"]')
    await expect(btn).toBeVisible({ timeout: 5_000 })
    // Label should be visible at desktop widths
    await expect(btn.locator('.trigger-label')).toBeVisible()
    await expect(btn.locator('.trigger-label')).toHaveText('Feedback')
  })
})

// ─── Modal Interaction ──────────────────────────────────────────────────────

test.describe('Feedback Widget — Modal Steps', () => {
  test('click button opens modal with step 1 (mood selection)', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)

    const modal = page.locator('[role="dialog"][aria-label="Ge feedback"]')
    await expect(modal).toBeVisible({ timeout: 3_000 })
    await expect(modal.getByText('Hur upplever du Måltiden?')).toBeVisible()
  })

  test('three mood options visible with Swedish labels', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)

    const modal = page.locator('[role="dialog"]')
    await expect(modal.getByRole('button', { name: 'Bra' })).toBeVisible()
    await expect(modal.getByRole('button', { name: 'Okej' })).toBeVisible()
    await expect(modal.getByRole('button', { name: 'Dåligt' })).toBeVisible()
  })

  test('selecting mood auto-advances to step 2 (categories)', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)

    const modal = page.locator('[role="dialog"]')
    await modal.getByRole('button', { name: 'Bra' }).click()

    // Step 2 heading
    await expect(modal.getByText('Vad gäller det?')).toBeVisible({ timeout: 3_000 })
  })

  test('category chips visible with correct labels', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)
    await selectMoodAndWaitForCategories(page, 'Okej')

    const modal = page.locator('[role="dialog"]')
    await expect(modal.getByText('Recept')).toBeVisible()
    await expect(modal.getByText('Menyn')).toBeVisible()
    await expect(modal.getByText('Inköpslistan')).toBeVisible()
    await expect(modal.getByText('Design')).toBeVisible()
    await expect(modal.getByText('Övrigt')).toBeVisible()
  })

  test('can select multiple categories (toggle on/off)', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)
    await selectMoodAndWaitForCategories(page, 'Bra')

    const modal = page.locator('[role="dialog"]')
    const recept = modal.locator('.chip', { hasText: 'Recept' })
    const design = modal.locator('.chip', { hasText: 'Design' })

    // Wait for chips to be ready for interaction
    await expect(recept).toBeVisible({ timeout: 3_000 })

    // Select two
    await recept.click()
    await expect(recept).toHaveClass(/active/)
    await design.click()
    await expect(design).toHaveClass(/active/)

    // Deselect one
    await recept.click()
    await expect(recept).not.toHaveClass(/active/)
    await expect(design).toHaveClass(/active/)
  })

  test('"Hoppa \u00f6ver" skips categories to step 3', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)
    await selectMoodAndWaitForCategories(page, 'Okej')

    await page.locator('[role="dialog"]').getByText('Hoppa \u00f6ver').click()

    await expect(page.locator('[role="dialog"]').getByText('Berätta mer')).toBeVisible({
      timeout: 3_000,
    })
  })

  test('"Nästa" advances to step 3', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)
    await selectMoodAndWaitForCategories(page, 'Bra')
    await advanceToCommentStep(page)

    await expect(page.locator('[role="dialog"]').getByText('Berätta mer')).toBeVisible({
      timeout: 3_000,
    })
  })

  test('step 3 shows textarea with character counter', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)
    await selectMoodAndWaitForCategories(page, 'Bra')
    await advanceToCommentStep(page)

    const modal = page.locator('[role="dialog"]')
    await expect(modal.locator('textarea')).toBeVisible({ timeout: 3_000 })
    await expect(modal.locator('.char-count')).toHaveText('0/500')
  })

  test('textarea enforces 500 character limit', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)
    await selectMoodAndWaitForCategories(page, 'Bra')
    await advanceToCommentStep(page)

    const textarea = page.locator('[role="dialog"] textarea')
    await expect(textarea).toBeVisible({ timeout: 3_000 })
    const longText = 'a'.repeat(510)
    await textarea.fill(longText)

    // HTML maxlength=500 should truncate
    const value = await textarea.inputValue()
    expect(value.length).toBeLessThanOrEqual(500)
  })

  test('"Skicka" submits feedback and shows success', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)
    await selectMoodAndWaitForCategories(page, 'Bra')
    await advanceToCommentStep(page)

    const modal = page.locator('[role="dialog"]')
    await modal.getByRole('button', { name: 'Skicka' }).click()

    // Success state
    await expect(modal.getByText('Tack för din feedback!')).toBeVisible({
      timeout: 5_000,
    })
  })

  test('modal auto-closes after success', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)
    await selectMoodAndWaitForCategories(page, 'Bra')
    await advanceToCommentStep(page)

    const modal = page.locator('[role="dialog"]')
    await modal.getByRole('button', { name: 'Skicka' }).click()

    await expect(modal.getByText('Tack för din feedback!')).toBeVisible({
      timeout: 5_000,
    })

    // Wait for auto-close (~2 seconds)
    await expect(modal).not.toBeVisible({ timeout: 5_000 })
  })

  test('after submission, feedback button is hidden (24h cooldown)', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)
    await selectMoodAndWaitForCategories(page, 'Bra')
    await advanceToCommentStep(page)

    const modal = page.locator('[role="dialog"]')
    await modal.getByRole('button', { name: 'Skicka' }).click()

    // Wait for auto-close
    await expect(modal).not.toBeVisible({ timeout: 5_000 })

    // Button should be hidden
    await expect(page.locator('button[aria-label="Ge feedback"]')).not.toBeVisible()
  })

  test('clearing localStorage cooldown re-shows button after re-login', { timeout: 60_000 }, async ({ page }) => {
    const { email, password } = await registerUser(page)
    await page.waitForTimeout(2000)

    // Submit feedback to trigger cooldown
    await openFeedbackModal(page)
    await selectMoodAndWaitForCategories(page, 'Bra')
    await advanceToCommentStep(page)

    const modal = page.locator('[role="dialog"]')
    await modal.getByRole('button', { name: 'Skicka' }).click()
    await expect(modal.getByText('Tack för din feedback!')).toBeVisible({ timeout: 5_000 })
    await expect(modal).not.toBeVisible({ timeout: 5_000 })
    await expect(page.locator('button[aria-label="Ge feedback"]')).not.toBeVisible()

    // Clear cooldown and token, then re-login to get fresh widget mount
    await clearCooldown(page)
    await page.evaluate(() => localStorage.removeItem('maltiden_token'))
    await loginUser(page, email, password)
    await page.waitForTimeout(2000)

    await expect(page.locator('button[aria-label="Ge feedback"]')).toBeVisible({ timeout: 5_000 })
  })
})

// ─── Modal Dismissal ────────────────────────────────────────────────────────

test.describe('Feedback Widget — Dismissal', () => {
  test('Escape key closes modal', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)

    await page.keyboard.press('Escape')
    await expect(page.locator('[role="dialog"]')).not.toBeVisible({ timeout: 3_000 })
  })

  test('clicking overlay (outside modal) closes it', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)

    // Click the overlay (top-left, away from the bottom-right modal)
    await page.locator('.feedback-overlay').click({ position: { x: 10, y: 10 } })
    await expect(page.locator('[role="dialog"]')).not.toBeVisible({ timeout: 3_000 })
  })

  test('X button closes modal', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)

    await page.locator('button[aria-label="Stäng"]').click()
    await expect(page.locator('[role="dialog"]')).not.toBeVisible({ timeout: 3_000 })
  })

  test('closing at step 2 does NOT submit', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)

    // Advance to step 2
    await selectMoodAndWaitForCategories(page, 'Bra')

    // Close
    await page.keyboard.press('Escape')
    await expect(page.locator('[role="dialog"]')).not.toBeVisible({ timeout: 3_000 })

    // No cooldown set → button still visible
    await expect(page.locator('button[aria-label="Ge feedback"]')).toBeVisible()
  })

  test('closing at step 3 does NOT submit', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)
    await selectMoodAndWaitForCategories(page, 'Okej')
    await advanceToCommentStep(page)

    await page.locator('button[aria-label="Stäng"]').click()
    await expect(page.locator('[role="dialog"]')).not.toBeVisible({ timeout: 3_000 })

    // No cooldown → button still visible
    await expect(page.locator('button[aria-label="Ge feedback"]')).toBeVisible()
  })
})

// ─── Accessibility ──────────────────────────────────────────────────────────

test.describe('Feedback Widget — Accessibility', () => {
  test('modal has role="dialog" and aria-label', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)

    const modal = page.locator('[role="dialog"]')
    await expect(modal).toBeVisible({ timeout: 3_000 })
    await expect(modal).toHaveAttribute('aria-modal', 'true')
    await expect(modal).toHaveAttribute('aria-label', 'Ge feedback')
  })

  test('mood buttons have accessible labels', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)

    const modal = page.locator('[role="dialog"]')
    await expect(modal).toBeVisible({ timeout: 3_000 })
    await expect(modal.locator('button[aria-label="Bra"]')).toBeVisible({ timeout: 3_000 })
    await expect(modal.locator('button[aria-label="Okej"]')).toBeVisible()
    await expect(modal.locator('button[aria-label="Dåligt"]')).toBeVisible()
  })

  test('focus trap: Tab does not leave modal', async ({ page }) => {
    await setupAuthenticatedUser(page)
    await openFeedbackModal(page)

    // Tab through all focusable elements
    for (let i = 0; i < 10; i++) {
      await page.keyboard.press('Tab')
    }

    // Active element should still be inside the modal
    const isInsideModal = await page.evaluate(() => {
      const modal = document.querySelector('[role="dialog"]')
      return modal?.contains(document.activeElement) ?? false
    })
    expect(isInsideModal).toBe(true)
  })
})

// ─── Full End-to-End Flows ──────────────────────────────────────────────────

test.describe('Feedback Widget — Full Flows', () => {
  test('happy path: mood → category → comment → submit → success → hidden', async ({ page }) => {
    await setupAuthenticatedUser(page)

    // Open modal
    await openFeedbackModal(page)
    const modal = page.locator('[role="dialog"]')

    // Step 1: Select mood → Step 2
    await selectMoodAndWaitForCategories(page, 'Bra')

    // Step 2: Select categories
    const recept = modal.locator('.chip', { hasText: 'Recept' })
    await expect(recept).toBeVisible({ timeout: 3_000 })
    await recept.click()
    await modal.locator('.chip', { hasText: 'Design' }).click()

    // Step 2 → Step 3
    await advanceToCommentStep(page)

    // Step 3: Write comment
    await modal.locator('textarea').fill('Jättebra app!')
    await modal.getByRole('button', { name: 'Skicka' }).click()

    // Step 4: Success
    await expect(modal.getByText('Tack för din feedback!')).toBeVisible({ timeout: 5_000 })

    // Auto-close and button hidden
    await expect(modal).not.toBeVisible({ timeout: 5_000 })
    await expect(page.locator('button[aria-label="Ge feedback"]')).not.toBeVisible()
  })

  test('negative feedback: bad mood → skip categories → comment → submit', async ({ page }) => {
    await setupAuthenticatedUser(page)

    await openFeedbackModal(page)
    const modal = page.locator('[role="dialog"]')

    // Step 1: Bad mood
    await selectMoodAndWaitForCategories(page, 'Dåligt')

    // Step 2: Skip categories
    await modal.getByText('Hoppa \u00f6ver').click()
    await expect(modal.getByText('Berätta mer')).toBeVisible({ timeout: 3_000 })

    // Step 3: Write comment and submit
    await modal.locator('textarea').fill('Listan buggar')
    await modal.getByRole('button', { name: 'Skicka' }).click()

    // Success
    await expect(modal.getByText('Tack för din feedback!')).toBeVisible({ timeout: 5_000 })
  })

  test('minimal feedback: mood only → skip → empty comment → submit', async ({ page }) => {
    await setupAuthenticatedUser(page)

    await openFeedbackModal(page)
    const modal = page.locator('[role="dialog"]')

    // Step 1: Select mood
    await selectMoodAndWaitForCategories(page, 'Okej')

    // Step 2: Skip
    await modal.getByText('Hoppa \u00f6ver').click()
    await expect(modal.getByText('Berätta mer')).toBeVisible({ timeout: 3_000 })

    // Step 3: Submit without comment
    await modal.getByRole('button', { name: 'Skicka' }).click()

    await expect(modal.getByText('Tack för din feedback!')).toBeVisible({ timeout: 5_000 })
  })

  test('re-opening modal resets to step 1', async ({ page }) => {
    await setupAuthenticatedUser(page)

    // Open and advance to step 2
    await openFeedbackModal(page)
    await selectMoodAndWaitForCategories(page, 'Bra')

    // Close
    await page.keyboard.press('Escape')
    await expect(page.locator('[role="dialog"]')).not.toBeVisible({ timeout: 3_000 })

    // Re-open → should be back on step 1
    await openFeedbackModal(page)
    await expect(
      page.locator('[role="dialog"]').getByText('Hur upplever du Måltiden?'),
    ).toBeVisible({ timeout: 3_000 })
  })
})
