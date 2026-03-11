import { test, expect } from '@playwright/test'
import { login, navigateTo } from './helpers'

test.describe('Recipes page', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await navigateTo(page, '/recipes')
    await expect(page).toHaveURL(/\/recipes/)
  })

  test('shows recipe page header', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Recept' })).toBeVisible({ timeout: 5000 })
  })

  test('shows page description', async ({ page }) => {
    await expect(page.getByText(/Hantera dina recept/)).toBeVisible({ timeout: 5000 })
  })

  test('shows tab controls', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Mina recept' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Lägg till' })).toBeVisible()
  })

  test('my recipes tab is active by default', async ({ page }) => {
    const myRecipesTab = page.getByRole('button', { name: 'Mina recept' })
    await expect(myRecipesTab).toHaveClass(/active/)
  })

  test('shows mock recipes in list', async ({ page }) => {
    await page.waitForTimeout(1500)
    const recipes = page.locator('[class*="recipe-card"], [class*="RecipeCard"]')
    await expect(recipes.first()).toBeVisible({ timeout: 5000 })
  })

  test('recipe cards contain recipe names', async ({ page }) => {
    await page.waitForTimeout(1500)
    // Mock data has these recipes
    const recipeNames = ['Pasta Carbonara', 'Kycklingwok', 'Tacos', 'Laxfilé med potatis', 'Köttfärssås']
    let found = 0
    for (const name of recipeNames) {
      if (await page.getByText(name).first().isVisible({ timeout: 2000 }).catch(() => false)) {
        found++
      }
    }
    expect(found).toBeGreaterThanOrEqual(1)
  })

  test('switches to add recipe tab', async ({ page }) => {
    await page.getByRole('button', { name: 'Lägg till' }).click()
    await page.waitForTimeout(500)
    await expect(page.getByText(/Hur vill du lägga till/i)).toBeVisible({ timeout: 3000 })
  })

  test('add tab becomes active when clicked', async ({ page }) => {
    const addTab = page.locator('.tab-button', { hasText: 'Lägg till' })
    await addTab.click()
    await page.waitForTimeout(300)
    await expect(addTab).toHaveClass(/active/)
  })

  test('switching back to list tab shows recipes again', async ({ page }) => {
    await page.getByRole('button', { name: 'Lägg till' }).click()
    await page.waitForTimeout(500)
    await page.getByRole('button', { name: 'Mina recept' }).click()
    await page.waitForTimeout(1500)
    const recipes = page.locator('[class*="recipe-card"], [class*="RecipeCard"]')
    await expect(recipes.first()).toBeVisible({ timeout: 5000 })
  })

  test('back button returns to dashboard', async ({ page }) => {
    await page.locator('.back-link').click()
    await expect(page).toHaveURL(/\/dashboard/)
  })

  test('back button shows Dashboard text', async ({ page }) => {
    await expect(page.locator('.back-link')).toContainText('Dashboard')
  })
})
