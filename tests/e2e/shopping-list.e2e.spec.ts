import { test, expect } from '@playwright/test'
import { registerUser, loginUser } from './helpers'

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
    await page.goto('/menu/generate')
    await expect(page.getByRole('heading', { name: 'Generera veckomeny' })).toBeVisible({
      timeout: 10_000,
    })

    // Generate a menu
    await page.getByRole('button', { name: 'Generera meny' }).click()

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

    await page.goto('/shopping-list')

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

  test('check off item — progress updates', async ({ page }) => {
    await loginUser(page, userEmail, userPassword)
    await page.goto('/shopping-list')

    // Wait for list to load
    await expect(page.locator('.progress-card')).toBeVisible({ timeout: 15_000 })

    // Capture initial progress text
    const progressText = page.locator('.progress-count')
    const initialText = await progressText.textContent()
    expect(initialText).toBeTruthy()

    // Click first unchecked checkbox
    const uncheckedCheckbox = page.locator('.item-row:not(.checked) .item-checkbox').first()
    await expect(uncheckedCheckbox).toBeVisible()
    await uncheckedCheckbox.click()

    // Wait for optimistic update
    await page.waitForTimeout(500)

    // Verify progress text changed
    const updatedText = await progressText.textContent()
    expect(updatedText).not.toBe(initialText)
  })

  test('reload — checked item persists', async ({ page }) => {
    await loginUser(page, userEmail, userPassword)
    await page.goto('/shopping-list')

    // Wait for list to load
    await expect(page.locator('.progress-card')).toBeVisible({ timeout: 15_000 })

    // Find first unchecked item and capture its name
    const uncheckedRow = page.locator('.item-row:not(.checked)').first()
    await expect(uncheckedRow).toBeVisible()
    const itemName = await uncheckedRow.locator('.item-name').textContent()
    expect(itemName).toBeTruthy()

    // Check it
    await uncheckedRow.locator('.item-checkbox').click()

    // Wait for API call to complete
    await page.waitForTimeout(1000)

    // Reload the page
    await page.reload()

    // Wait for list to reload
    await expect(page.locator('.progress-card')).toBeVisible({ timeout: 15_000 })

    // Find the item by name and verify it has .checked class
    const itemRow = page.locator('.item-row', { has: page.locator('.item-name', { hasText: itemName! }) })
    await expect(itemRow).toHaveClass(/checked/)
  })

  test('uncheck item — restored', async ({ page }) => {
    await loginUser(page, userEmail, userPassword)
    await page.goto('/shopping-list')

    // Wait for list to load
    await expect(page.locator('.progress-card')).toBeVisible({ timeout: 15_000 })

    // Find a checked item row
    const checkedRow = page.locator('.item-row.checked').first()
    await expect(checkedRow).toBeVisible()

    // Click its checkbox to uncheck
    await checkedRow.locator('.item-checkbox').click()

    // Wait for optimistic update
    await page.waitForTimeout(500)

    // Verify it lost the .checked class
    await expect(checkedRow).not.toHaveClass(/checked/)
  })
})
