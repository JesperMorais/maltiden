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
  await page.evaluate((p) => {
    // Access the Vue Router instance from the app
    const app = (document.querySelector('#app') as any)?.__vue_app__
    const router = app?.config?.globalProperties?.$router
    if (router) {
      router.push(p)
    } else {
      // Fallback: use history.pushState + popstate to trigger router
      window.history.pushState({}, '', p)
      window.dispatchEvent(new PopStateEvent('popstate'))
    }
  }, path)
  await page.waitForURL(new RegExp(path.replace(/\//g, '\\/')), { timeout: 5000 })
}
