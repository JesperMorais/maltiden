import { test, expect } from '@playwright/test'
import { registerUser, apiCall, loginUser, navigateTo } from './helpers'

test.describe('Recipes E2E', () => {
  test.describe.configure({ mode: 'serial' })

  // Shared state across serial tests
  let userEmail: string
  let userPassword: string
  let createdRecipeName: string

  test('register and navigate to recipes', async ({ page }) => {
    const { email, password } = await registerUser(page)
    userEmail = email
    userPassword = password

    // Should be on dashboard after registration
    await expect(page).toHaveURL(/\/dashboard/)

    // Navigate to recipes page (use SPA navigation to preserve auth state)
    await navigateTo(page, '/recipes')
    await expect(page).toHaveURL(/\/recipes/)

    // Should see the page title and the "Mina recept" tab active
    await expect(page.getByRole('heading', { name: 'Recept' })).toBeVisible({ timeout: 10_000 })
    await expect(page.getByText('Mina recept')).toBeVisible()
  })

  test('browse seed recipes — at least 15 visible', async ({ page }) => {
    // Login and navigate to recipes
    await page.goto('/login')
    await page.getByPlaceholder('din@email.se').fill(userEmail)
    await page.getByPlaceholder('Ditt lösenord').fill(userPassword)
    await page.getByRole('button', { name: 'Logga in' }).click()
    await page.waitForURL('**/dashboard', { timeout: 10_000 })
    await navigateTo(page, '/recipes')

    // Wait for recipe grid to load (skeleton disappears, cards appear)
    await expect(page.locator('.recipe-card').first()).toBeVisible({ timeout: 15_000 })

    // Count recipe cards — 20 seed recipes exist, at least 15 should be visible
    const recipeCards = page.locator('.recipe-card')
    const count = await recipeCards.count()
    expect(count).toBeGreaterThanOrEqual(15)

    // Verify some known seed recipe names exist in the DOM
    // (some may be scrolled below viewport, so use toBeAttached instead of toBeVisible)
    const knownRecipes = [
      'Pasta Carbonara',
      'Kycklingwok',
      'Tacos',
      'Pannkakor',
      'Köttfärssås',
    ]

    for (const name of knownRecipes) {
      await expect(page.getByText(name, { exact: true }).first()).toBeAttached({ timeout: 10_000 })
    }

    // Verify search input is present
    await expect(page.getByPlaceholder('Sök recept...')).toBeVisible()
  })

  test('view recipe detail — ingredients and instructions render', async ({ page }) => {
    // Login and navigate to recipes
    await page.goto('/login')
    await page.getByPlaceholder('din@email.se').fill(userEmail)
    await page.getByPlaceholder('Ditt lösenord').fill(userPassword)
    await page.getByRole('button', { name: 'Logga in' }).click()
    await page.waitForURL('**/dashboard', { timeout: 10_000 })
    await navigateTo(page, '/recipes')

    // Wait for recipes to load
    await expect(page.locator('.recipe-card').first()).toBeVisible({ timeout: 15_000 })

    // Click the first recipe card to open the detail modal
    await page.locator('.recipe-card').first().click()

    // The modal should appear with a dialog role
    const modal = page.locator('[role="dialog"]')
    await expect(modal).toBeVisible({ timeout: 10_000 })

    // Modal should show the "Ingredienser" section with list items
    await expect(modal.getByText('Ingredienser')).toBeVisible()
    const ingredientItems = modal.locator('.ingredient-item')
    const ingredientCount = await ingredientItems.count()
    expect(ingredientCount).toBeGreaterThan(0)

    // Modal should show the "Instruktioner" section with list items
    await expect(modal.getByText('Instruktioner')).toBeVisible()
    const instructionItems = modal.locator('.instruction-item')
    const instructionCount = await instructionItems.count()
    expect(instructionCount).toBeGreaterThan(0)

    // Modal should show servings info (e.g., "4 portioner")
    await expect(modal.getByText(/\d+ portioner/)).toBeVisible()

    // Close the modal
    await modal.getByRole('button', { name: 'Stäng' }).click()
    await expect(modal).not.toBeVisible()
  })

  test('create recipe manually — appears in list', async ({ page }) => {
    // Login and navigate to recipes
    await page.goto('/login')
    await page.getByPlaceholder('din@email.se').fill(userEmail)
    await page.getByPlaceholder('Ditt lösenord').fill(userPassword)
    await page.getByRole('button', { name: 'Logga in' }).click()
    await page.waitForURL('**/dashboard', { timeout: 10_000 })
    await navigateTo(page, '/recipes')

    // Switch to "Lagg till" tab
    await page.getByText('Lägg till', { exact: true }).click()

    // Choose the "Fyll i sjalv" (manual) method
    await page.getByText('Fyll i själv').click()

    // Now the RecipeEditForm should be visible — fill it in
    const uniqueSuffix = Date.now()
    createdRecipeName = `E2E Testreceptet ${uniqueSuffix}`

    // Fill recipe name
    await page.getByPlaceholder('Namn på receptet').fill(createdRecipeName)

    // Fill servings — clear existing value first, then type new value
    const servingsInput = page.locator('.servings-field input')
    await servingsInput.fill('2')

    // Fill the first ingredient row (one empty row is pre-created for manual mode)
    const ingredientRows = page.locator('.ingredient-row')
    const firstIngredient = ingredientRows.first()
    await firstIngredient.locator('input[placeholder="Ingrediens"]').fill('Pasta')
    await firstIngredient.locator('input[placeholder="Mängd"]').fill('400')
    await firstIngredient.locator('input[placeholder="Enhet"]').fill('g')

    // Add a second ingredient
    await page.getByText('+ Lägg till ingrediens').click()
    const secondIngredient = ingredientRows.nth(1)
    await secondIngredient.locator('input[placeholder="Ingrediens"]').fill('Grädde')
    await secondIngredient.locator('input[placeholder="Mängd"]').fill('2')
    await secondIngredient.locator('input[placeholder="Enhet"]').fill('dl')

    // Fill the first instruction (one empty row is pre-created for manual mode)
    const instructionRows = page.locator('.instruction-row')
    await instructionRows.first().locator('input[placeholder="Steg 1"]').fill('Koka pastan')

    // Add a second instruction
    await page.getByText('+ Lägg till steg').click()
    await instructionRows.nth(1).locator('input[placeholder="Steg 2"]').fill('Blanda med grädde')

    // Fill tags
    await page.getByPlaceholder('pasta, italienskt, snabb').fill('e2e-test, pasta')

    // Click "Spara recept"
    await page.getByRole('button', { name: 'Spara recept' }).click()

    // Should see success toast
    await expect(page.getByText('Receptet har sparats!')).toBeVisible({ timeout: 10_000 })

    // Success screen should show recipe name and "Visa mina recept" button
    await expect(page.getByText('Recept sparat!')).toBeVisible()
    await expect(page.getByText(createdRecipeName)).toBeVisible()

    // Click "Visa mina recept" to go back to the list
    await page.getByRole('button', { name: 'Visa mina recept' }).click()

    // Wait for the recipe list to load and verify the created recipe appears
    await expect(page.locator('.recipe-card').first()).toBeVisible({ timeout: 15_000 })

    // Search for the created recipe to confirm it exists
    await page.getByPlaceholder('Sök recept...').fill(createdRecipeName)
    await expect(page.getByText(createdRecipeName)).toBeVisible({ timeout: 10_000 })
  })

  test('delete recipe via API — removed from list', async ({ page }) => {
    // Login and navigate to recipes
    await page.goto('/login')
    await page.getByPlaceholder('din@email.se').fill(userEmail)
    await page.getByPlaceholder('Ditt lösenord').fill(userPassword)
    await page.getByRole('button', { name: 'Logga in' }).click()
    await page.waitForURL('**/dashboard', { timeout: 10_000 })

    // Get recipes list via API to find our created recipe
    const listRes = await apiCall(page, 'GET', '/recipes')
    expect(listRes.status).toBe(200)

    const recipes = (listRes.data as { recipes: { id: string; name: string }[] }).recipes
    const target = recipes.find((r) => r.name === createdRecipeName)
    expect(target).toBeTruthy()

    // Delete via API
    const deleteRes = await apiCall(page, 'DELETE', `/recipes/${target!.id}`)
    expect(deleteRes.status).toBe(204)

    // Navigate to recipes and verify the deleted recipe is gone
    await navigateTo(page, '/recipes')
    await expect(page.locator('.recipe-card').first()).toBeVisible({ timeout: 15_000 })

    // Search for the deleted recipe — should not be found
    await page.getByPlaceholder('Sök recept...').fill(createdRecipeName)

    // Should show "Inga traffar" empty state since no recipe matches the unique name
    await expect(page.getByText('Inga träffar')).toBeVisible({ timeout: 10_000 })
  })

  test('AI parse recipe — skip in E2E (external API dependency)', async () => {
    // AI recipe parsing depends on the Claude API, which is unreliable for automated
    // E2E testing (slow responses, rate limits, API key availability). Skip in E2E suite.
    test.skip(true, 'AI parse depends on external Claude API — skipped in E2E')
  })
})
