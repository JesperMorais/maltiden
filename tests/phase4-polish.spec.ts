import { test, expect } from '@playwright/test'
import { login, navigateTo } from './helpers'

test.describe('Phase 4: Polish & Delight', () => {
  test.describe('GenerateMenuSkeleton component', () => {
    test.beforeEach(async ({ page }) => {
      await login(page)
      await page.getByRole('button', { name: /Generera meny/ }).click()
      await expect(page).toHaveURL(/\/menu\/generate/)
    })

    test('shows empty state initially instead of skeleton', async ({ page }) => {
      const emptyState = page.locator('.empty-state')
      await expect(emptyState).toBeVisible({ timeout: 5000 })
    })

    test('skeleton is not visible when menu has not been generated', async ({ page }) => {
      const skeleton = page.locator('.generate-menu-skeleton')
      await expect(skeleton).not.toBeVisible()
    })

    test('page renders without errors', async ({ page }) => {
      await expect(page.getByText('Generera veckomeny')).toBeVisible({ timeout: 5000 })
      await expect(page.getByText('Måndag - Fredag')).toBeVisible()
    })
  })

  test.describe('EmptyState entrance animation', () => {
    test.beforeEach(async ({ page }) => {
      await login(page)
      await page.getByRole('button', { name: /Generera meny/ }).click()
      await expect(page).toHaveURL(/\/menu\/generate/)
    })

    test('empty state element exists with expected structure', async ({ page }) => {
      const emptyState = page.locator('.empty-state')
      await expect(emptyState).toBeVisible({ timeout: 5000 })
    })

    test('empty state contains generate button', async ({ page }) => {
      const generateBtn = page.locator('.empty-state .generate-button')
      await expect(generateBtn).toBeVisible({ timeout: 5000 })
    })

    test('empty state shows feature list', async ({ page }) => {
      await expect(page.getByText('5 måltider (Måndag-Fredag)')).toBeVisible({ timeout: 5000 })
      await expect(page.getByText('Slumpmässiga recept', { exact: true })).toBeVisible()
      await expect(page.getByText('Lås dagar du vill behålla', { exact: true })).toBeVisible()
    })

    test('empty state shows title', async ({ page }) => {
      await expect(page.getByText('Skapa din veckomeny')).toBeVisible({ timeout: 5000 })
    })
  })

  test.describe('RecipeDetailModal skeleton', () => {
    test.beforeEach(async ({ page }) => {
      await login(page)
      // Navigate to recipes via dashboard quick action
      await page.getByRole('button', { name: /Recept/ }).first().click()
      await expect(page).toHaveURL(/\/recipes/, { timeout: 5000 })
      await page.waitForTimeout(2000)
    })

    test('clicking a recipe card opens detail modal with content', async ({ page }) => {
      const recipeCard = page.locator('[class*="recipe-card"]').first()
      await expect(recipeCard).toBeVisible({ timeout: 8000 })
      await recipeCard.click()

      const modal = page.getByRole('dialog')
      await expect(modal).toBeVisible({ timeout: 8000 })

      // Verify it shows actual recipe content (not skeleton blocks)
      const recipeTitle = modal.locator('.recipe-title')
      await expect(recipeTitle).toBeVisible({ timeout: 5000 })
    })

    test('recipe detail modal shows ingredients section', async ({ page }) => {
      const recipeCard = page.locator('[class*="recipe-card"]').first()
      await expect(recipeCard).toBeVisible({ timeout: 8000 })
      await recipeCard.click()

      const modal = page.getByRole('dialog')
      await expect(modal).toBeVisible({ timeout: 8000 })

      await expect(modal.getByText('Ingredienser')).toBeVisible({ timeout: 5000 })
    })

    test('recipe detail modal shows instructions section', async ({ page }) => {
      const recipeCard = page.locator('[class*="recipe-card"]').first()
      await expect(recipeCard).toBeVisible({ timeout: 8000 })
      await recipeCard.click()

      const modal = page.getByRole('dialog')
      await expect(modal).toBeVisible({ timeout: 8000 })

      await expect(modal.getByText('Instruktioner')).toBeVisible({ timeout: 5000 })
    })

    test('recipe detail modal has close button', async ({ page }) => {
      const recipeCard = page.locator('[class*="recipe-card"]').first()
      await expect(recipeCard).toBeVisible({ timeout: 8000 })
      await recipeCard.click()

      const modal = page.getByRole('dialog')
      await expect(modal).toBeVisible({ timeout: 8000 })

      await expect(modal.getByRole('button', { name: 'Stäng' })).toBeVisible()
    })
  })

  test.describe('GenerateMenuView skeleton wiring', () => {
    test.beforeEach(async ({ page }) => {
      await login(page)
      await page.getByRole('button', { name: /Generera meny/ }).click()
      await expect(page).toHaveURL(/\/menu\/generate/)
    })

    test('skeleton is not visible when showing empty state', async ({ page }) => {
      // Empty state should be visible
      await expect(page.locator('.empty-state')).toBeVisible({ timeout: 5000 })

      // Skeleton should NOT be visible
      await expect(page.locator('.generate-menu-skeleton')).not.toBeVisible()
    })

    test('menu grid is not visible in empty state', async ({ page }) => {
      await expect(page.locator('.empty-state')).toBeVisible({ timeout: 5000 })
      await expect(page.locator('.menu-grid')).not.toBeVisible()
    })

    test('page header renders correctly', async ({ page }) => {
      await expect(page.getByText('Generera veckomeny')).toBeVisible({ timeout: 5000 })
      await expect(page.getByText('Måndag - Fredag')).toBeVisible()
      await expect(page.getByText(/Skapa en meny för 5 dagar/)).toBeVisible()
    })

    test('content container exists', async ({ page }) => {
      const container = page.locator('.content-container')
      await expect(container).toBeVisible({ timeout: 5000 })
    })
  })
})
