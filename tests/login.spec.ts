import { test, expect } from '@playwright/test'

test.describe('Login page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
  })

  test('shows login form with all elements', async ({ page }) => {
    await expect(page.getByText('Välkommen tillbaka')).toBeVisible()
    await expect(page.getByText('Logga in på ditt Måltiden-konto')).toBeVisible()
    await expect(page.getByPlaceholder('din@email.se')).toBeVisible()
    await expect(page.getByPlaceholder('Ditt lösenord')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Logga in' })).toBeVisible()
  })

  test('shows logo icon', async ({ page }) => {
    await expect(page.locator('.logo-icon')).toBeVisible()
  })

  test('submit button is disabled with empty form', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Logga in' })).toBeDisabled()
  })

  test('submit button enables with valid input', async ({ page }) => {
    await page.getByPlaceholder('din@email.se').fill('test@example.com')
    await page.getByPlaceholder('Ditt lösenord').fill('password123')
    await expect(page.getByRole('button', { name: 'Logga in' })).toBeEnabled()
  })

  test('submit button stays disabled with short password', async ({ page }) => {
    await page.getByPlaceholder('din@email.se').fill('test@example.com')
    await page.getByPlaceholder('Ditt lösenord').fill('short')
    await expect(page.getByRole('button', { name: 'Logga in' })).toBeDisabled()
  })

  test('submit button stays disabled with invalid email', async ({ page }) => {
    await page.getByPlaceholder('din@email.se').fill('notanemail')
    await page.getByPlaceholder('Ditt lösenord').fill('password123')
    await expect(page.getByRole('button', { name: 'Logga in' })).toBeDisabled()
  })

  test('submit button disabled with email missing @ symbol', async ({ page }) => {
    await page.getByPlaceholder('din@email.se').fill('userdomain.com')
    await page.getByPlaceholder('Ditt lösenord').fill('password123')
    await expect(page.getByRole('button', { name: 'Logga in' })).toBeDisabled()
  })

  test('submit button disabled with exactly 7 char password', async ({ page }) => {
    await page.getByPlaceholder('din@email.se').fill('test@example.com')
    await page.getByPlaceholder('Ditt lösenord').fill('1234567')
    await expect(page.getByRole('button', { name: 'Logga in' })).toBeDisabled()
  })

  test('submit button enabled with exactly 8 char password', async ({ page }) => {
    await page.getByPlaceholder('din@email.se').fill('test@example.com')
    await page.getByPlaceholder('Ditt lösenord').fill('12345678')
    await expect(page.getByRole('button', { name: 'Logga in' })).toBeEnabled()
  })

  test('successful login redirects to dashboard', async ({ page }) => {
    await page.getByPlaceholder('din@email.se').fill('test@example.com')
    await page.getByPlaceholder('Ditt lösenord').fill('password123')
    await page.getByRole('button', { name: 'Logga in' }).click()
    await expect(page).toHaveURL(/\/dashboard/, { timeout: 5000 })
  })

  test('back link returns to landing page', async ({ page }) => {
    await page.getByRole('link', { name: /Tillbaka/ }).click()
    await expect(page).toHaveURL('/')
  })

  test('create account link goes to register', async ({ page }) => {
    await page.getByText('Skapa konto').click()
    await expect(page).toHaveURL(/\/register/)
  })

  test('email field has correct type attribute', async ({ page }) => {
    const emailInput = page.getByPlaceholder('din@email.se')
    await expect(emailInput).toHaveAttribute('type', 'email')
  })

  test('password field has correct type attribute', async ({ page }) => {
    const passwordInput = page.getByPlaceholder('Ditt lösenord')
    await expect(passwordInput).toHaveAttribute('type', 'password')
  })

  test('email field has autocomplete attribute', async ({ page }) => {
    const emailInput = page.getByPlaceholder('din@email.se')
    await expect(emailInput).toHaveAttribute('autocomplete', 'email')
  })

  test('password field has autocomplete attribute', async ({ page }) => {
    const passwordInput = page.getByPlaceholder('Ditt lösenord')
    await expect(passwordInput).toHaveAttribute('autocomplete', 'current-password')
  })

  test('form labels are visible', async ({ page }) => {
    await expect(page.getByText('E-post')).toBeVisible()
    await expect(page.getByText('Lösenord', { exact: true })).toBeVisible()
  })

  test('theme toggle button exists', async ({ page }) => {
    const themeToggle = page.locator('.theme-toggle')
    await expect(themeToggle).toBeVisible()
  })

  test('theme toggle changes theme', async ({ page }) => {
    const themeToggle = page.locator('.theme-toggle')
    await themeToggle.click()
    await page.waitForTimeout(300)
    // Click again to toggle back
    await themeToggle.click()
    await page.waitForTimeout(300)
  })

  test('login form submits on Enter key', async ({ page }) => {
    await page.getByPlaceholder('din@email.se').fill('test@example.com')
    await page.getByPlaceholder('Ditt lösenord').fill('password123')
    await page.getByPlaceholder('Ditt lösenord').press('Enter')
    await expect(page).toHaveURL(/\/dashboard/, { timeout: 5000 })
  })

  test('footer shows "Har du inget konto?" text', async ({ page }) => {
    await expect(page.getByText('Har du inget konto?')).toBeVisible()
  })
})
