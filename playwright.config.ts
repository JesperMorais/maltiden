import { defineConfig, devices } from '@playwright/test'

const E2E_PORT = 5174

// Specs that should also run on a mobile viewport. CLAUDE.md describes Måltiden
// as a mobile-first MVP (375–414px), so the views users actually interact with
// on a phone get an extra mobile pass. Keep this list small to avoid doubling
// CI time — only the specs whose UX materially differs on mobile.
//
// NOTE: dashboard.spec.ts is intentionally excluded — the desktop dashboard
// uses a sidebar layout that's hidden behind the mobile bottom nav at <768px,
// so its element-visibility assertions don't apply. Mobile-specific dashboard
// behavior is covered by accessibility.spec.ts and responsive.spec.ts.
const MOBILE_SPECS = [
  '**/accessibility.spec.ts',
  '**/shopping-list.spec.ts',
  '**/recipes.spec.ts',
  '**/about.spec.ts',
]

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
    {
      name: 'mobile-chrome',
      use: { ...devices['Pixel 5'] },
      testMatch: MOBILE_SPECS,
    },
  ],
  webServer: {
    command: `npm run --prefix frontend dev:mock -- --port ${E2E_PORT}`,
    url: `http://localhost:${E2E_PORT}`,
    reuseExistingServer: !process.env.CI,
  },
})
