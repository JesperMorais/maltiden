import { expect, type Page } from '@playwright/test'

/** Log in and wait for dashboard to load */
export async function login(page: Page) {
  await page.goto('/login')
  await page.getByPlaceholder('din@email.se').fill('test@example.com')
  await page.getByPlaceholder('Ditt lösenord').fill('password123')
  await page.getByRole('button', { name: 'Logga in' }).click()
  await expect(page).toHaveURL(/\/dashboard/, { timeout: 5000 })
  await page.waitForSelector('.dashboard-content', { timeout: 8000 })
}

/** Navigate via Vue Router (preserves Pinia state) */
export async function navigateTo(page: Page, path: string) {
  // Click a link to trigger Vue Router navigation
  await page.evaluate((p) => {
    const link = document.createElement('a')
    link.href = p
    link.style.display = 'none'
    document.body.appendChild(link)
    link.click()
    link.remove()
  }, path)
  await page.waitForURL(new RegExp(path.replace(/\//g, '\\/')), { timeout: 5000 })
}
