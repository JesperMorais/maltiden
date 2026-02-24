import { test, expect } from '@playwright/test'
import { registerUser, loginUser, navigateTo, apiCall } from './helpers'

test.describe('Shopping List E2E', () => {
  test.describe.configure({ mode: 'serial' })

  // Shared state across serial tests
  let userEmail: string
  let userPassword: string

  test('setup: register + generate + save menu', async ({ page }) => {
    const { email, password } = await registerUser(page)
    userEmail = email
    userPassword = password

    // Navigate to menu generation page
    await navigateTo(page, '/menu/generate')
    await expect(page.getByRole('heading', { name: 'Generera veckomeny' })).toBeVisible({
      timeout: 10_000,
    })

    // Generate a menu
    await page.getByRole('button', { name: 'Generera nya' }).click()

    // Wait for save button to become enabled (animation complete)
    const saveButton = page.getByRole('button', { name: 'Spara meny' })
    await expect(saveButton).toBeVisible({ timeout: 30_000 })
    await expect(saveButton).toBeEnabled({ timeout: 15_000 })

    // Save the menu
    await saveButton.click()

    // Should redirect to dashboard
    await page.waitForURL('**/dashboard', { timeout: 10_000 })
    await expect(page).toHaveURL(/\/dashboard/)
  })

  test('view shopping list — categories and items visible', async ({ page }) => {
    await loginUser(page, userEmail, userPassword)

    await navigateTo(page, '/shopping-list')

    // Verify heading
    await expect(page.getByRole('heading', { name: 'Inköpslista' })).toBeVisible({
      timeout: 10_000,
    })

    // Verify progress text matches "X av Y varor"
    await expect(page.locator('.progress-card')).toBeVisible({ timeout: 15_000 })
    await expect(page.getByText(/\d+ av \d+ varor/)).toBeVisible()

    // At least one category heading
    const categoryHeadings = page.locator('.category-heading')
    const categoryCount = await categoryHeadings.count()
    expect(categoryCount).toBeGreaterThan(0)

    // Item rows > 0
    const itemRows = page.locator('.item-row')
    const itemCount = await itemRows.count()
    expect(itemCount).toBeGreaterThan(0)
  })

  test('progress card shows correct format', async ({ page }) => {
    await loginUser(page, userEmail, userPassword)
    await navigateTo(page, '/shopping-list')

    // Wait for list to load
    await expect(page.locator('.progress-card')).toBeVisible({ timeout: 15_000 })

    // Verify progress text format "X av Y varor"
    const progressText = page.locator('.progress-count')
    await expect(progressText).toHaveText(/\d+ av \d+ varor/)

    // Verify progress bar track exists
    await expect(page.locator('.progress-bar-track')).toBeVisible()
  })

  test('item rows are interactive — checkboxes clickable', async ({ page }) => {
    await loginUser(page, userEmail, userPassword)
    await navigateTo(page, '/shopping-list')

    // Wait for list to load
    await expect(page.locator('.progress-card')).toBeVisible({ timeout: 15_000 })

    // Verify checkbox elements are present and interactable
    const firstCheckbox = page.locator('.item-checkbox').first()
    await expect(firstCheckbox).toBeVisible()
    await expect(firstCheckbox).toBeEnabled()

    // Verify there are multiple items across categories
    const itemCount = await page.locator('.item-row').count()
    expect(itemCount).toBeGreaterThan(5)

    // Verify categories are structured correctly
    const categoryCount = await page.locator('.category-heading').count()
    expect(categoryCount).toBeGreaterThan(0)
  })
})
