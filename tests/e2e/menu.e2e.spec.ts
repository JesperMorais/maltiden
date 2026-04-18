import { test, expect } from '@playwright/test'
import { registerUser, loginUser, navigateTo } from './helpers'

test.describe('Menu E2E', () => {
  test.describe.configure({ mode: 'serial' })

  // Shared state across serial tests
  let userEmail: string
  let userPassword: string

  test('register user for menu tests', async ({ page }) => {
    const { email, password } = await registerUser(page)
    userEmail = email
    userPassword = password

    // Should be on dashboard after registration
    await expect(page).toHaveURL(/\/dashboard/)
  })

  test('generate menu → 7 days populated with recipes', async ({ page }) => {
    await loginUser(page, userEmail, userPassword)

    // Navigate to menu generation page
    await navigateTo(page, '/menu/generate')
    await expect(page.getByRole('heading', { name: 'Generera veckomeny' })).toBeVisible({
      timeout: 10_000,
    })
    await expect(page.getByText(/Måndag\s*[-–]\s*Söndag/)).toBeVisible()

    // Click the "Generera nya" button to generate the initial menu
    await page.getByRole('button', { name: 'Generera nya' }).click()

    // Wait for the slot machine animation to complete:
    // The "Spara meny" button only becomes enabled after animation finishes
    // (isLoading || slotMachine.isAnimating === false and hasMenu === true)
    const saveButton = page.getByRole('button', { name: 'Spara meny' })
    await expect(saveButton).toBeVisible({ timeout: 30_000 })
    await expect(saveButton).toBeEnabled({ timeout: 15_000 })

    // Verify 7 day cards are visible in the menu grid
    const dayCards = page.locator('.menu-day-card')
    await expect(dayCards).toHaveCount(7)

    // Each day card should be in the filled state (not showing empty "Ingen måltid")
    for (let i = 0; i < 7; i++) {
      const card = dayCards.nth(i)
      // The servings text only shows when a recipe is assigned
      await expect(card.locator('.recipe-servings')).toBeVisible({ timeout: 10_000 })
    }

    // Verify day names are present (Mon–Sun in Swedish) — target the day-name heading
    const expectedDays = ['Måndag', 'Tisdag', 'Onsdag', 'Torsdag', 'Fredag', 'Lördag', 'Söndag']
    for (const dayName of expectedDays) {
      await expect(page.locator('.day-name', { hasText: dayName })).toBeVisible()
    }

    // Verify the "Generera nya" (regenerate) button is also visible
    await expect(page.getByRole('button', { name: 'Generera nya' })).toBeVisible()
  })

  test('lock one day, regenerate → locked day unchanged', async ({ page }) => {
    await loginUser(page, userEmail, userPassword)
    await navigateTo(page, '/menu/generate')

    // Generate a menu first
    await page.getByRole('button', { name: 'Generera nya' }).click()
    const saveButton = page.getByRole('button', { name: 'Spara meny' })
    await expect(saveButton).toBeVisible({ timeout: 30_000 })
    await expect(saveButton).toBeEnabled({ timeout: 15_000 })

    const dayCards = page.locator('.menu-day-card')
    await expect(dayCards).toHaveCount(7)

    // Capture the first day's recipe emoji before locking (recipe names may be empty
    // if the API omits recipeName, but emoji is always rendered from recipeId state)
    const firstCard = dayCards.nth(0)
    await expect(firstCard.locator('.recipe-servings')).toBeVisible({ timeout: 10_000 })
    const firstEmoji = await firstCard.locator('.recipe-emoji').textContent()
    expect(firstEmoji?.trim()).toBeTruthy()

    // Lock the first day card by clicking its lock button
    const lockButton = firstCard.locator('.lock-button')
    await expect(lockButton).toBeVisible()
    await lockButton.click()

    // Verify the card is now in locked state
    await expect(firstCard).toHaveClass(/locked/)

    // Verify the lock status shows "1 av 7 dagar låsta"
    await expect(page.getByText('1 av 7 dagar låsta')).toBeVisible()

    // Click "Generera nya" to regenerate unlocked days
    await page.getByRole('button', { name: 'Generera nya' }).click()

    // Wait for regeneration animation to complete
    await expect(saveButton).toBeEnabled({ timeout: 15_000 })

    // Verify the locked first day still has the same emoji (recipe unchanged)
    const firstEmojiAfter = await firstCard.locator('.recipe-emoji').textContent()
    expect(firstEmojiAfter?.trim()).toBe(firstEmoji?.trim())

    // Verify the card is still locked
    await expect(firstCard).toHaveClass(/locked/)
  })

  test('save menu → redirect to dashboard → menu persists', async ({ page }) => {
    await loginUser(page, userEmail, userPassword)
    await navigateTo(page, '/menu/generate')

    // Generate a menu
    await page.getByRole('button', { name: 'Generera nya' }).click()
    const saveButton = page.getByRole('button', { name: 'Spara meny' })
    await expect(saveButton).toBeVisible({ timeout: 30_000 })
    await expect(saveButton).toBeEnabled({ timeout: 15_000 })

    // Click "Spara meny" to save
    await saveButton.click()

    // Should redirect to dashboard
    await page.waitForURL('**/dashboard', { timeout: 10_000 })
    await expect(page).toHaveURL(/\/dashboard/)

    // The dashboard should show the weekly menu section ("Veckans meny")
    await expect(page.getByText('Veckans meny')).toBeVisible({ timeout: 10_000 })

    // The weekly menu grid should have day cards with meal emojis
    // Dashboard uses .meal-emoji for days that have a meal assigned
    const mealEmojis = page.locator('.meal-emoji')
    const emojiCount = await mealEmojis.count()
    expect(emojiCount).toBeGreaterThan(0)
  })

  test('navigate to shopping list from saved menu → items exist', async ({ page }) => {
    await loginUser(page, userEmail, userPassword)

    // Go to shopping list page
    await navigateTo(page, '/shopping-list')

    // Wait for the page to load
    await expect(page.getByRole('heading', { name: 'Inköpslista' })).toBeVisible({
      timeout: 10_000,
    })

    // Should show shopping list categories and items (not the empty state)
    // The progress card ("X av Y varor") indicates items are loaded
    await expect(page.locator('.progress-card')).toBeVisible({ timeout: 15_000 })
    await expect(page.getByText(/\d+ av \d+ varor/)).toBeVisible()

    // Should have at least one category section with items
    const categorySections = page.locator('.category-section')
    const categoryCount = await categorySections.count()
    expect(categoryCount).toBeGreaterThan(0)

    // Each category should have a heading and item rows
    const firstCategory = categorySections.first()
    await expect(firstCategory.locator('.category-heading')).toBeVisible()

    const itemRows = page.locator('.item-row')
    const itemCount = await itemRows.count()
    expect(itemCount).toBeGreaterThan(0)
  })
})
