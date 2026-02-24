import { test, expect } from '@playwright/test'
import {
  registerUser,
  loginUser,
  generateUniqueName,
  apiCall,
  registerUserViaAPI,
} from './helpers'

test.describe('Cross-Flow E2E — Full 2-user journey with shared household', () => {
  test.describe.configure({ mode: 'serial' })

  // Shared state across serial tests
  let userAEmail: string
  let userAPassword: string
  const userAName = 'User A'
  let userBEmail: string
  let userBPassword: string
  let userBToken: string
  let inviteCode: string
  let recipeBName: string

  test('User A registers → household created', async ({ page }) => {
    const { email, password } = await registerUser(page, { name: userAName })
    userAEmail = email
    userAPassword = password

    // Backend auto-creates household named "<name>'s household"
    await expect(page.getByText(`${userAName}'s household`)).toBeVisible({ timeout: 10_000 })
  })

  test('User A adds 2 recipes via API', async ({ page }) => {
    await loginUser(page, userAEmail, userAPassword)

    const recipe1Name = generateUniqueName('CrossRecipeA1')
    const recipe2Name = generateUniqueName('CrossRecipeA2')

    const res1 = await apiCall(page, 'POST', '/recipes', {
      name: recipe1Name,
      servings: 4,
      ingredients: [
        { name: 'Pasta', amount: '400', unit: 'g' },
        { name: 'Grädde', amount: '2', unit: 'dl' },
      ],
      instructions: ['Koka pastan', 'Blanda med grädde'],
      tags: ['e2e-cross'],
    })
    expect(res1.status).toBe(201)

    const res2 = await apiCall(page, 'POST', '/recipes', {
      name: recipe2Name,
      servings: 2,
      ingredients: [
        { name: 'Ris', amount: '3', unit: 'dl' },
        { name: 'Kyckling', amount: '500', unit: 'g' },
      ],
      instructions: ['Koka riset', 'Stek kycklingen'],
      tags: ['e2e-cross'],
    })
    expect(res2.status).toBe(201)

    // Navigate to recipes page and verify both appear
    await page.goto('/recipes')
    await expect(page.locator('.recipe-card').first()).toBeVisible({ timeout: 15_000 })

    await page.getByPlaceholder('Sök recept...').fill('CrossRecipeA')
    await expect(page.getByText(recipe1Name)).toBeVisible({ timeout: 10_000 })
    await expect(page.getByText(recipe2Name)).toBeVisible({ timeout: 10_000 })
  })

  test('User A generates menu → saves', async ({ page }) => {
    await loginUser(page, userAEmail, userAPassword)

    await page.goto('/menu/generate')
    await expect(page.getByRole('heading', { name: 'Generera veckomeny' })).toBeVisible({
      timeout: 10_000,
    })

    // Click generate button from empty state
    await page.getByRole('button', { name: 'Generera meny' }).click()

    // Wait for save button to become enabled (animation complete)
    const saveButton = page.getByRole('button', { name: 'Spara meny' })
    await expect(saveButton).toBeVisible({ timeout: 30_000 })
    await expect(saveButton).toBeEnabled({ timeout: 15_000 })

    // Click save → redirect to dashboard
    await saveButton.click()
    await page.waitForURL('**/dashboard', { timeout: 10_000 })
    await expect(page).toHaveURL(/\/dashboard/)
  })

  test('User A views shopping list → checks off 3 items', async ({ page }) => {
    await loginUser(page, userAEmail, userAPassword)

    await page.goto('/shopping-list')

    // Wait for shopping list to load with progress text
    await expect(page.getByText(/av.*varor/)).toBeVisible({ timeout: 15_000 })

    // Check off 3 unchecked items
    const uncheckedCheckboxes = page.locator('.item-row:not(.checked) .item-checkbox')
    const availableCount = await uncheckedCheckboxes.count()
    expect(availableCount).toBeGreaterThanOrEqual(3)

    for (let i = 0; i < 3; i++) {
      // Always click the first unchecked item (list re-renders after each check)
      await page.locator('.item-row:not(.checked) .item-checkbox').first().click()
      // Small wait for optimistic update
      await page.waitForTimeout(300)
    }

    // Verify progress count shows at least 3 checked
    await expect(page.getByText(/[3-9]\d* av \d+ varor/)).toBeVisible({ timeout: 10_000 })
  })

  test('User A generates invite code', async ({ page }) => {
    await loginUser(page, userAEmail, userAPassword)

    const res = await apiCall(page, 'POST', '/households/invite')
    expect(res.status).toBe(201)
    expect(res.data.code).toBeTruthy()

    inviteCode = res.data.code as string
  })

  test('User B registers → joins with invite code', async () => {
    // Register User B via API (no UI needed)
    const userB = await registerUserViaAPI({ name: 'User B' })
    userBEmail = userB.email
    userBPassword = userB.password
    userBToken = userB.token

    // User B joins User A's household
    const joinRes = await fetch('http://localhost:8080/households/join', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${userBToken}`,
      },
      body: JSON.stringify({ code: inviteCode }),
    })

    expect(joinRes.ok).toBe(true)
    const joinData = await joinRes.json()
    expect(joinData.householdId).toBeTruthy()
  })

  test('User B sees same household, menu, shopping list', async ({ page }) => {
    await loginUser(page, userBEmail, userBPassword)

    // Should see User A's household name
    await expect(page.getByText(`${userAName}'s household`)).toBeVisible({ timeout: 10_000 })

    // Navigate to shopping list and verify checked items are shared
    await page.goto('/shopping-list')
    await expect(page.getByText(/av.*varor/)).toBeVisible({ timeout: 15_000 })

    // Progress should show >= 3 checked items (User A checked 3)
    await expect(page.getByText(/[3-9]\d* av \d+ varor/)).toBeVisible({ timeout: 10_000 })
  })

  test('User B adds recipe → User A can see it', async ({ page, browser }) => {
    // Login as User B and add a recipe via API
    await loginUser(page, userBEmail, userBPassword)

    recipeBName = generateUniqueName('CrossRecipeB')
    const res = await apiCall(page, 'POST', '/recipes', {
      name: recipeBName,
      servings: 3,
      ingredients: [
        { name: 'Lax', amount: '400', unit: 'g' },
        { name: 'Citron', amount: '1', unit: 'st' },
      ],
      instructions: ['Grilla laxen', 'Pressa citron över'],
      tags: ['e2e-cross-b'],
    })
    expect(res.status).toBe(201)

    // Open a new browser context for User A
    const userAContext = await browser.newContext()
    const userAPage = await userAContext.newPage()

    try {
      await loginUser(userAPage, userAEmail, userAPassword)
      await userAPage.goto('/recipes')
      await expect(userAPage.locator('.recipe-card').first()).toBeVisible({ timeout: 15_000 })

      // Search for User B's recipe
      await userAPage.getByPlaceholder('Sök recept...').fill(recipeBName)
      await expect(userAPage.getByText(recipeBName)).toBeVisible({ timeout: 10_000 })
    } finally {
      await userAContext.close()
    }
  })
})
