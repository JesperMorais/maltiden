import { defineConfig, devices } from '@playwright/test'

const E2E_PORT = 5174

export default defineConfig({
  testDir: './tests',
  testIgnore: process.env.CI ? ['**/e2e/**', '**/qa/**'] : ['**/e2e/**'],
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 2 : undefined,
  reporter: 'html',
  use: {
    baseURL: `http://localhost:${E2E_PORT}`,
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: {
    command: `npm run --prefix frontend dev:mock -- --port ${E2E_PORT}`,
    url: `http://localhost:${E2E_PORT}`,
    reuseExistingServer: !process.env.CI,
  },
})
