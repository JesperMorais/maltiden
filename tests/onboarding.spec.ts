import { test, expect } from '@playwright/test'

test.describe('Onboarding page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/register')
    await page.waitForTimeout(500)
  })

  test('shows welcome header', async ({ page }) => {
    await expect(page.getByText('Välkommen till Måltiden')).toBeVisible({ timeout: 5000 })
    await expect(page.getByText('Hur vill du komma igång?')).toBeVisible()
  })

  test('shows logo icon', async ({ page }) => {
    await expect(page.locator('.logo-icon')).toBeVisible()
  })

  test('shows two choices: join or create household', async ({ page }) => {
    await expect(page.getByText(/gå med/i).first()).toBeVisible({ timeout: 5000 })
    await expect(page.getByText(/skapa/i).first()).toBeVisible()
  })

  test('shows back link to landing', async ({ page }) => {
    const backLink = page.getByRole('link', { name: /Tillbaka/ })
    await expect(backLink).toBeVisible()
    await expect(backLink).toHaveAttribute('href', '/')
  })

  test('back link navigates to landing', async ({ page }) => {
    await page.getByRole('link', { name: /Tillbaka/ }).click()
    await expect(page).toHaveURL('/')
  })

  test('shows theme toggle', async ({ page }) => {
    await expect(page.locator('.theme-toggle')).toBeVisible()
  })

  test('join household flow: shows invite code input', async ({ page }) => {
    await page.getByText(/gå med/i).first().click()
    await page.waitForTimeout(500)
    const codeInput = page.locator('input').first()
    await expect(codeInput).toBeVisible({ timeout: 3000 })
  })

  test('join household flow: code input has placeholder', async ({ page }) => {
    await page.getByText(/gå med/i).first().click()
    await page.waitForTimeout(500)
    await expect(page.getByPlaceholder(/ABC123/)).toBeVisible({ timeout: 3000 })
  })

  test('join household flow: continue button disabled with empty code', async ({ page }) => {
    await page.getByText(/gå med/i).first().click()
    await page.waitForTimeout(500)
    const continueBtn = page.getByRole('button', { name: /Fortsätt/ })
    await expect(continueBtn).toBeDisabled()
  })

  test('join household flow: continue button enabled with valid code', async ({ page }) => {
    await page.getByText(/gå med/i).first().click()
    await page.waitForTimeout(500)
    await page.getByPlaceholder(/ABC123/).fill('TEST99')
    const continueBtn = page.getByRole('button', { name: /Fortsätt/ })
    await expect(continueBtn).toBeEnabled()
  })

  test('join household flow: valid code shows welcome message', async ({ page }) => {
    await page.getByText(/gå med/i).first().click()
    await page.waitForTimeout(500)
    await page.getByPlaceholder(/ABC123/).fill('ABC123')
    await page.getByRole('button', { name: /Fortsätt/ }).click()
    // Should show the matched family name
    await expect(page.getByText('Familjen Andersson')).toBeVisible({ timeout: 5000 })
  })

  test('join household flow: invalid code shows error', async ({ page }) => {
    await page.getByText(/gå med/i).first().click()
    await page.waitForTimeout(500)
    await page.getByPlaceholder(/ABC123/).fill('WRONG1')
    await page.getByRole('button', { name: /Fortsätt/ }).click()
    await expect(page.getByText(/hittades inte/i)).toBeVisible({ timeout: 5000 })
  })

  test('join household flow: shows member-or-guest choice after valid code', async ({ page }) => {
    await page.getByText(/gå med/i).first().click()
    await page.waitForTimeout(500)
    await page.getByPlaceholder(/ABC123/).fill('ABC123')
    await page.getByRole('button', { name: /Fortsätt/ }).click()
    // Wait for welcome step and then auto-advance
    await expect(page.getByText(/Hur vill du gå med/i)).toBeVisible({ timeout: 5000 })
  })

  test('join household flow: member choice shows form', async ({ page }) => {
    await page.getByText(/gå med/i).first().click()
    await page.waitForTimeout(500)
    await page.getByPlaceholder(/ABC123/).fill('ABC123')
    await page.getByRole('button', { name: /Fortsätt/ }).click()
    await expect(page.getByText(/Hur vill du gå med/i)).toBeVisible({ timeout: 5000 })

    // Click "Bli medlem"
    await page.getByText('Bli medlem').click()
    await page.waitForTimeout(500)

    // Should show member registration form
    await expect(page.getByText('Skapa ditt medlemskonto')).toBeVisible({ timeout: 3000 })
    await expect(page.getByPlaceholder('Anna Andersson')).toBeVisible()
    await expect(page.getByPlaceholder('anna@exempel.se')).toBeVisible()
  })

  test('join household flow: guest choice shows simple form', async ({ page }) => {
    await page.getByText(/gå med/i).first().click()
    await page.waitForTimeout(500)
    await page.getByPlaceholder(/ABC123/).fill('ABC123')
    await page.getByRole('button', { name: /Fortsätt/ }).click()
    await expect(page.getByText(/Hur vill du gå med/i)).toBeVisible({ timeout: 5000 })

    // Click "Gå med som gäst"
    await page.getByText('Gå med som gäst').first().click()
    await page.waitForTimeout(500)

    // Should show simple guest form (just name)
    await expect(page.getByText('Gå med som gäst').first()).toBeVisible({ timeout: 3000 })
    await expect(page.getByPlaceholder('Anna')).toBeVisible()
  })

  test('create household flow shows registration form', async ({ page }) => {
    await page.getByText('Skapa nytt hushåll').click()
    await page.waitForTimeout(800)
    const inputs = page.locator('.inline-form input')
    await expect(inputs.first()).toBeVisible({ timeout: 3000 })
  })

  test('create household flow: all form fields present', async ({ page }) => {
    await page.getByText('Skapa nytt hushåll').click()
    await page.waitForTimeout(800)

    await expect(page.getByPlaceholder('Anna Andersson').first()).toBeVisible({ timeout: 3000 })
    await expect(page.getByPlaceholder('anna@exempel.se').first()).toBeVisible()
    await expect(page.getByPlaceholder('Minst 8 tecken').first()).toBeVisible()
    await expect(page.getByPlaceholder('Skriv lösenordet igen').first()).toBeVisible()
    await expect(page.getByPlaceholder(/Familjen Andersson/).first()).toBeVisible()
  })

  test('create household flow: submit button disabled with empty form', async ({ page }) => {
    await page.getByText('Skapa nytt hushåll').click()
    await page.waitForTimeout(800)
    await expect(page.getByRole('button', { name: 'Skapa konto' })).toBeDisabled()
  })

  test('create household flow: password mismatch shows error', async ({ page }) => {
    await page.getByText('Skapa nytt hushåll').click()
    await page.waitForTimeout(800)

    await page.getByPlaceholder('Anna Andersson').first().fill('Test User')
    await page.getByPlaceholder('anna@exempel.se').first().fill('test@test.se')
    await page.getByPlaceholder('Minst 8 tecken').first().fill('password123')
    await page.getByPlaceholder('Skriv lösenordet igen').first().fill('different')

    await expect(page.getByText('Lösenorden matchar inte')).toBeVisible({ timeout: 3000 })
  })

  test('create household flow: submit enabled with valid form', async ({ page }) => {
    await page.getByText('Skapa nytt hushåll').click()
    const form = page.locator('form.form-card').filter({ hasText: 'Skapa konto' })
    await expect(form).toBeVisible({ timeout: 5000 })
    // Wait for form-slide CSS transition to complete before filling inputs
    await expect(page.locator('.form-slide-enter-active')).toHaveCount(0, { timeout: 3000 })

    await form.getByPlaceholder('Anna Andersson').fill('Test User')
    await form.getByPlaceholder('anna@exempel.se').fill('test@test.se')
    await form.getByPlaceholder('Minst 8 tecken').fill('password123')
    await form.getByPlaceholder('Skriv lösenordet igen').fill('password123')
    await form.getByPlaceholder(/Familjen Andersson/).fill('Testfamiljen')

    // Re-dispatch input events to ensure Vue v-model picks up all values in CI
    await form.evaluate((formEl) => {
      formEl.querySelectorAll('input').forEach((input) => {
        input.dispatchEvent(new Event('input', { bubbles: true }))
      })
    })

    await expect(page.getByRole('button', { name: 'Skapa konto' })).toBeEnabled({ timeout: 5000 })
  })

  test('create household flow: successful registration shows success', async ({ page }) => {
    await page.getByText('Skapa nytt hushåll').click()
    const form = page.locator('form.form-card').filter({ hasText: 'Skapa konto' })
    await expect(form).toBeVisible({ timeout: 5000 })
    // Wait for form-slide CSS transition to complete before filling inputs
    await expect(page.locator('.form-slide-enter-active')).toHaveCount(0, { timeout: 3000 })

    await form.getByPlaceholder('Anna Andersson').fill('Test User')
    await form.getByPlaceholder('anna@exempel.se').fill('test@test.se')
    await form.getByPlaceholder('Minst 8 tecken').fill('password123')
    await form.getByPlaceholder('Skriv lösenordet igen').fill('password123')
    await form.getByPlaceholder(/Familjen Andersson/).fill('Testfamiljen')

    // Re-dispatch input events to ensure Vue v-model picks up all values in CI
    await form.evaluate((formEl) => {
      formEl.querySelectorAll('input').forEach((input) => {
        input.dispatchEvent(new Event('input', { bubbles: true }))
      })
    })

    await page.getByRole('button', { name: 'Skapa konto' }).click()
    await expect(page.getByText('Konto skapat!')).toBeVisible({ timeout: 10000 })
  })

  test('change choice resets back to initial state', async ({ page }) => {
    // Select join
    await page.getByText(/gå med/i).first().click()
    await page.waitForTimeout(500)

    // Click change choice
    await page.getByText(/Ändra val/).click()
    await page.waitForTimeout(500)

    // Should show both choices again
    await expect(page.getByText(/gå med/i).first()).toBeVisible()
    await expect(page.getByText(/skapa/i).first()).toBeVisible()
  })

  test('shows terms and privacy links in create form', async ({ page }) => {
    await page.getByText('Skapa nytt hushåll').click()
    await page.waitForTimeout(800)
    await expect(page.getByText(/villkor/)).toBeVisible({ timeout: 3000 })
    await expect(page.getByText(/integritetspolicy/)).toBeVisible()
  })

  test('shows benefits comparison in join flow', async ({ page }) => {
    await page.getByText(/gå med/i).first().click()
    await page.waitForTimeout(500)
    await page.getByPlaceholder(/ABC123/).fill('ABC123')
    await page.getByRole('button', { name: /Fortsätt/ }).click()
    await expect(page.getByText(/Hur vill du gå med/i)).toBeVisible({ timeout: 5000 })

    // Should show benefits comparison
    await expect(page.getByText(/Som medlem får du/)).toBeVisible()
    await expect(page.getByText(/Som gäst kan du/)).toBeVisible()
  })
})
