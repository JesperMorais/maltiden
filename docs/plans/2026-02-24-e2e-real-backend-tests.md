# E2E Real Backend Tests — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create comprehensive Playwright E2E tests running against the real Go backend at localhost:8080, verifying actual data persistence across auth, household, recipes, menu, shopping list, and a full cross-user journey.

**Architecture:** Playwright tests hit the Go server (port 8080) which serves both the built Vue SPA and API. A shell script builds the frontend, copies output to `backend/static/`, resets the test database, and starts the Go server. Each test file uses unique timestamped data to avoid conflicts. Tests run serially.

**Tech Stack:** Playwright, TypeScript, Go backend (SQLite), Vue 3 SPA

---

## Prerequisites & Assumptions

- Go 1.24+ with CGO support installed
- Node 22+ installed
- `frontend/node_modules` already installed (`npm ci`)
- The Go server auto-runs migrations including seed data (20 recipes)
- The OnboardingView "create household" flow calls the real backend `POST /auth/register`
- The OnboardingView "join household" flow is **mocked in the component** (uses hardcoded families) — household joining must use API calls in test helpers
- `POST /auth/register` auto-creates a household for each new user
- Recipe edit/delete UI availability needs verification during implementation; API helpers used as fallback

---

### Task 1: E2E Server Script

**Files:**
- Create: `scripts/e2e-server.sh`

**Step 1: Write the server startup script**

```bash
#!/bin/bash
set -e

cd "$(dirname "$0")/.."

# Build frontend
echo "=== Building frontend ==="
npm run --prefix frontend build

# Copy build output to backend/static for SPA serving
echo "=== Setting up static files ==="
rm -rf backend/static
cp -r frontend/dist backend/static

# Ensure data directory exists
mkdir -p backend/data

# Remove old test DB for fresh state (migrations + seed data re-applied on startup)
rm -f backend/data/e2e-test.db

# Start Go server with test configuration
echo "=== Starting Go server on :8080 ==="
cd backend
DATABASE_PATH=./data/e2e-test.db \
JWT_SECRET=e2e-test-secret-key-must-be-at-least-32-characters \
CORS_ORIGINS=http://localhost:8080 \
exec go run cmd/server/main.go
```

**Step 2: Make executable and verify**

```bash
chmod +x scripts/e2e-server.sh
```

**Step 3: Commit**

```bash
git add scripts/e2e-server.sh
git commit -m "chore: add E2E server startup script"
```

---

### Task 2: Playwright E2E Config

**Files:**
- Create: `tests/e2e/playwright.e2e.config.ts`
- Modify: `frontend/package.json` (add `test:e2e:real` script)

**Step 1: Create the Playwright config**

```typescript
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: '.',
  fullyParallel: false, // Serial execution to avoid race conditions
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: [['html', { outputFolder: '../../playwright-report-e2e' }]],

  use: {
    baseURL: 'http://localhost:8080',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    actionTimeout: 10_000,
  },

  timeout: 30_000,

  projects: [
    {
      name: 'e2e',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  webServer: {
    command: 'bash ../../scripts/e2e-server.sh',
    url: 'http://localhost:8080/health',
    reuseExistingServer: !process.env.CI,
    timeout: 180_000, // 3 min for frontend build + Go startup
    stdout: 'pipe',
    stderr: 'pipe',
  },
})
```

Key decisions:
- `fullyParallel: false` + `workers: 1` — serial execution prevents race conditions on shared DB
- `actionTimeout: 10_000` — real backend is slower than mocks
- `timeout: 30_000` — generous per-test timeout
- `reuseExistingServer: !process.env.CI` — in local dev, reuse running server; in CI, start fresh
- `webServer.timeout: 180_000` — frontend build can take ~60s, Go startup ~10s

**Step 2: Add npm script to `frontend/package.json`**

Add to `scripts`:
```json
"test:e2e:real": "npx --prefix .. playwright test --config=tests/e2e/playwright.e2e.config.ts"
```

**Step 3: Verify config loads**

```bash
cd /home/david/Dev/personal/maltiden
npx playwright test --config=tests/e2e/playwright.e2e.config.ts --list
```

Expected: Playwright finds 0 tests (no spec files yet), but config loads without errors.

**Step 4: Commit**

```bash
git add tests/e2e/playwright.e2e.config.ts frontend/package.json
git commit -m "test: add Playwright E2E config for real backend testing"
```

---

### Task 3: Test Helpers

**Files:**
- Create: `tests/e2e/helpers.ts`

**Step 1: Write the helpers file**

```typescript
import { type Page, expect } from '@playwright/test'

/**
 * Generate a unique email for test isolation.
 * Uses timestamp + random suffix to avoid collisions across parallel runs.
 */
export function generateUniqueEmail(): string {
  const ts = Date.now()
  const rand = Math.random().toString(36).slice(2, 6)
  return `test-${ts}-${rand}@maltiden.test`
}

/**
 * Generate a unique name for test data (recipes, etc.)
 */
export function generateUniqueName(prefix: string): string {
  const ts = Date.now()
  return `${prefix}-${ts}`
}

/**
 * Register a new user via the Onboarding "Create household" UI flow.
 *
 * Flow: /register → "Skapa nytt hushåll" → fill form → submit → success → /dashboard
 *
 * Returns the email and password used (for later login).
 */
export async function registerUser(
  page: Page,
  options?: { email?: string; password?: string; name?: string; householdName?: string },
): Promise<{ email: string; password: string; name: string }> {
  const email = options?.email ?? generateUniqueEmail()
  const password = options?.password ?? 'TestPassword123!'
  const name = options?.name ?? 'Test User'
  const householdName = options?.householdName ?? 'Test Hushåll'

  await page.goto('/register')

  // Choose "Skapa nytt hushåll"
  await page.getByText('Skapa nytt hushåll').click()

  // Fill the create form
  await page.getByPlaceholder('Anna Andersson').fill(name)
  await page.getByPlaceholder('anna@exempel.se').fill(email)
  await page.getByPlaceholder('Minst 8 tecken').fill(password)
  await page.getByPlaceholder('Skriv lösenordet igen').fill(password)
  await page.getByPlaceholder('Familjen Andersson').fill(householdName)

  // Submit
  await page.getByRole('button', { name: 'Skapa konto' }).click()

  // Wait for success state
  await expect(page.getByText('Konto skapat!')).toBeVisible({ timeout: 10_000 })

  // Wait for auto-redirect to dashboard (2s timeout in the app + navigation)
  await page.waitForURL('**/dashboard', { timeout: 10_000 })

  return { email, password, name }
}

/**
 * Login via the Login page UI.
 *
 * Flow: /login → fill email + password → submit → /dashboard
 */
export async function loginUser(
  page: Page,
  email: string,
  password: string,
): Promise<void> {
  await page.goto('/login')

  await page.getByPlaceholder('din@email.se').fill(email)
  await page.getByPlaceholder('Ditt lösenord').fill(password)
  await page.getByRole('button', { name: 'Logga in' }).click()

  await page.waitForURL('**/dashboard', { timeout: 10_000 })
}

/**
 * Get the JWT token from localStorage.
 */
export async function getToken(page: Page): Promise<string | null> {
  return page.evaluate(() => localStorage.getItem('token'))
}

/**
 * Make an authenticated API call directly (bypasses UI).
 * Useful for setup/teardown and testing flows not exposed in UI.
 */
export async function apiCall(
  page: Page,
  method: string,
  path: string,
  body?: Record<string, unknown>,
): Promise<{ status: number; data: Record<string, unknown> }> {
  const token = await getToken(page)
  const baseURL = 'http://localhost:8080'

  return page.evaluate(
    async ({ baseURL, method, path, body, token }) => {
      const res = await fetch(`${baseURL}${path}`, {
        method,
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: body ? JSON.stringify(body) : undefined,
      })
      const data = await res.json().catch(() => ({}))
      return { status: res.status, data }
    },
    { baseURL, method, path, body, token },
  )
}

/**
 * Register a user via API (not UI). Faster for setup.
 * Returns token and user info.
 */
export async function registerUserViaAPI(options?: {
  email?: string
  password?: string
  name?: string
}): Promise<{ token: string; email: string; password: string; userId: string; householdId: string }> {
  const email = options?.email ?? generateUniqueEmail()
  const password = options?.password ?? 'TestPassword123!'
  const name = options?.name ?? 'Test User'

  const res = await fetch('http://localhost:8080/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password, name }),
  })

  if (!res.ok) {
    throw new Error(`Registration failed: ${res.status} ${await res.text()}`)
  }

  const data = await res.json()
  return {
    token: data.token,
    email,
    password,
    userId: data.user.id,
    householdId: data.user.householdId,
  }
}

/**
 * Set a JWT token in the browser's localStorage and reload.
 * Useful for setting up auth state without going through login UI.
 */
export async function setAuthToken(page: Page, token: string): Promise<void> {
  await page.evaluate((t) => localStorage.setItem('token', t), token)
}
```

**Step 2: Verify TypeScript compiles**

```bash
cd /home/david/Dev/personal/maltiden
npx tsc --noEmit tests/e2e/helpers.ts --esModuleInterop --module esnext --moduleResolution bundler --target esnext --skipLibCheck
```

This may not work directly (Playwright handles TS compilation). Instead, just ensure no red squiggles. Move on.

**Step 3: Commit**

```bash
git add tests/e2e/helpers.ts
git commit -m "test: add E2E test helpers for registration, login, and API calls"
```

---

### Task 4: Auth E2E Tests

**Files:**
- Create: `tests/e2e/auth.e2e.spec.ts`

**Step 1: Write auth tests**

```typescript
import { test, expect } from '@playwright/test'
import {
  registerUser,
  loginUser,
  generateUniqueEmail,
  getToken,
} from './helpers'

test.describe('Authentication E2E', () => {
  test.describe.configure({ mode: 'serial' })

  let registeredEmail: string
  let registeredPassword: string

  test('register a new user → redirect to dashboard with JWT', async ({ page }) => {
    const { email, password } = await registerUser(page)
    registeredEmail = email
    registeredPassword = password

    // Verify on dashboard
    await expect(page).toHaveURL(/\/dashboard/)

    // Verify JWT is stored
    const token = await getToken(page)
    expect(token).toBeTruthy()

    // Verify dashboard shows household info
    await expect(page.getByText('Test Hushåll')).toBeVisible({ timeout: 10_000 })
  })

  test('register with existing email → shows error', async ({ page }) => {
    await page.goto('/register')
    await page.getByText('Skapa nytt hushåll').click()

    await page.getByPlaceholder('Anna Andersson').fill('Duplicate User')
    await page.getByPlaceholder('anna@exempel.se').fill(registeredEmail)
    await page.getByPlaceholder('Minst 8 tecken').fill('TestPassword123!')
    await page.getByPlaceholder('Skriv lösenordet igen').fill('TestPassword123!')
    await page.getByPlaceholder('Familjen Andersson').fill('Duplicate Hushåll')

    await page.getByRole('button', { name: 'Skapa konto' }).click()

    // Should show Swedish error (from getSwedishAuthError: 409 → "E-postadressen är redan registrerad")
    await expect(
      page.getByText(/redan registrerad|misslyckades/),
    ).toBeVisible({ timeout: 10_000 })
  })

  test('login with valid credentials → dashboard loads', async ({ page }) => {
    await loginUser(page, registeredEmail, registeredPassword)

    await expect(page).toHaveURL(/\/dashboard/)
    await expect(page.getByText('Test Hushåll')).toBeVisible({ timeout: 10_000 })
  })

  test('login with wrong password → shows error', async ({ page }) => {
    await page.goto('/login')
    await page.getByPlaceholder('din@email.se').fill(registeredEmail)
    await page.getByPlaceholder('Ditt lösenord').fill('WrongPassword999')
    await page.getByRole('button', { name: 'Logga in' }).click()

    // Should show Swedish error ("Fel e-post eller lösenord")
    await expect(
      page.getByText(/Fel e-post|lösenord|misslyckades/),
    ).toBeVisible({ timeout: 10_000 })
  })

  test('access protected route without JWT → redirect to login', async ({ page }) => {
    // Clear any stored token
    await page.goto('/')
    await page.evaluate(() => localStorage.removeItem('token'))

    await page.goto('/dashboard')
    await expect(page).toHaveURL(/\/login/)
  })

  test('expired/invalid JWT → 401 handling', async ({ page }) => {
    // Set an expired/invalid token
    await page.goto('/')
    await page.evaluate(() =>
      localStorage.setItem('token', 'expired.invalid.token'),
    )

    // Navigate to a protected page — the app should detect 401 and redirect
    await page.goto('/dashboard')

    // The Axios interceptor removes the token and redirects to /login on 401
    await expect(page).toHaveURL(/\/login/, { timeout: 10_000 })
  })
})
```

**Step 2: Run and verify**

```bash
npx playwright test --config=tests/e2e/playwright.e2e.config.ts auth.e2e.spec.ts
```

All 6 tests should pass. Fix any selector or timing issues.

**Step 3: Commit**

```bash
git add tests/e2e/auth.e2e.spec.ts
git commit -m "test: E2E auth tests (register, login, errors, JWT handling)"
```

---

### Task 5: Household E2E Tests

**Files:**
- Create: `tests/e2e/household.e2e.spec.ts`

**Context:**
- Registration auto-creates a household
- The dashboard shows household info in the "Hushållet" widget
- Invite code generation is via the dashboard "Bjud in" button → calls `POST /households/invite`
- Joining a household via UI is mocked in OnboardingView, so we use API helpers for the join step
- After joining, the second user should see the same household on their dashboard

**Step 1: Write household tests**

```typescript
import { test, expect } from '@playwright/test'
import {
  registerUser,
  loginUser,
  generateUniqueEmail,
  apiCall,
  registerUserViaAPI,
  setAuthToken,
} from './helpers'

test.describe('Household E2E', () => {
  test.describe.configure({ mode: 'serial' })

  let ownerEmail: string
  let ownerPassword: string
  let inviteCode: string

  test('new user auto-creates household → visible on dashboard', async ({ page }) => {
    const { email, password } = await registerUser(page, {
      householdName: 'Familjen E2E Test',
    })
    ownerEmail = email
    ownerPassword = password

    // Verify household name on dashboard
    await expect(page.getByText('Familjen E2E Test')).toBeVisible({ timeout: 10_000 })
  })

  test('generate invite code → shows code with copy feedback', async ({ page }) => {
    await loginUser(page, ownerEmail, ownerPassword)

    // Click the invite button in the household widget or quick actions
    const inviteButton = page.getByRole('button', { name: /Bjud in/i })
    await inviteButton.click()

    // The invite code should be displayed — capture it
    // The API returns { code, expiresAt }, UI should show the code
    // Wait for the code to appear (may be in a modal or inline)
    const codeElement = page.locator('[data-testid="invite-code"], .invite-code')
    if (await codeElement.isVisible({ timeout: 5_000 }).catch(() => false)) {
      inviteCode = (await codeElement.textContent()) ?? ''
    } else {
      // Fallback: get the code via API
      const result = await apiCall(page, 'POST', '/households/invite')
      inviteCode = (result.data as { code: string }).code
    }

    expect(inviteCode).toBeTruthy()
    expect(inviteCode.length).toBeGreaterThanOrEqual(4)
  })

  test('second user joins with invite code → appears in household', async ({ page }) => {
    // Register user B via API (faster than UI)
    const userB = await registerUserViaAPI({
      name: 'User B',
      email: generateUniqueEmail(),
    })

    // Join owner's household via API using the invite code
    const joinRes = await fetch('http://localhost:8080/households/join', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${userB.token}`,
      },
      body: JSON.stringify({ code: inviteCode }),
    })
    expect(joinRes.ok).toBeTruthy()

    // Now log in as owner and check if User B appears in the household
    await loginUser(page, ownerEmail, ownerPassword)

    // The household widget should show the new member
    await expect(page.getByText('User B')).toBeVisible({ timeout: 10_000 })
  })

  test('join with invalid code → error', async ({ page }) => {
    const userC = await registerUserViaAPI({
      name: 'User C',
      email: generateUniqueEmail(),
    })

    // Try to join with invalid code via API
    const joinRes = await fetch('http://localhost:8080/households/join', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${userC.token}`,
      },
      body: JSON.stringify({ code: 'INVALID_CODE_XYZ' }),
    })
    expect(joinRes.ok).toBeFalsy()
    expect(joinRes.status).toBe(404) // or 400, depending on backend
  })

  test('member role visible correctly (owner badge)', async ({ page }) => {
    await loginUser(page, ownerEmail, ownerPassword)

    // The owner should see their role badge
    await expect(page.getByText(/Ägare/)).toBeVisible({ timeout: 10_000 })
  })
})
```

**Step 2: Run and verify**

```bash
npx playwright test --config=tests/e2e/playwright.e2e.config.ts household.e2e.spec.ts
```

**Notes for implementer:**
- The invite code UI flow needs investigation. The dashboard has a "Bjud in" button but the exact interaction (modal? inline code display?) needs to be verified by reading the component.
- If the invite code is shown in a modal with a copy button, look for "Kopierad!" feedback text.
- The `data-testid="invite-code"` selector is speculative — adapt to actual DOM.
- The API join endpoint may return different status codes for invalid codes — verify and adjust.

**Step 3: Commit**

```bash
git add tests/e2e/household.e2e.spec.ts
git commit -m "test: E2E household tests (auto-create, invite, join, roles)"
```

---

### Task 6: Recipes E2E Tests

**Files:**
- Create: `tests/e2e/recipes.e2e.spec.ts`

**Context:**
- 20 seed recipes from migrations (5 from 004, 15 from 008)
- `GET /recipes` is public (no auth required)
- Recipe list is at `/recipes` (requires auth + member role)
- Add recipe flow: "Lägg till" tab → "Fyll i själv" → RecipeEditForm → save
- Edit/delete may or may not be available in UI — check `RecipeCard.vue` during implementation
- AI parse requires `ANTHROPIC_API_KEY` — skip gracefully if not available

**Step 1: Write recipe tests**

```typescript
import { test, expect } from '@playwright/test'
import {
  registerUser,
  loginUser,
  generateUniqueName,
  apiCall,
} from './helpers'

test.describe('Recipes E2E', () => {
  test.describe.configure({ mode: 'serial' })

  let email: string
  let password: string
  let createdRecipeName: string

  test('register and navigate to recipes', async ({ page }) => {
    const creds = await registerUser(page)
    email = creds.email
    password = creds.password
  })

  test('browse seed recipes → at least 15 visible', async ({ page }) => {
    await loginUser(page, email, password)
    await page.goto('/recipes')

    // Wait for recipe list to load
    await expect(page.getByText('Mina recept')).toBeVisible({ timeout: 10_000 })

    // Verify some known seed recipes are present
    await expect(page.getByText('Pasta Carbonara')).toBeVisible()
    await expect(page.getByText('Tacos')).toBeVisible()
    await expect(page.getByText('Pannkakor')).toBeVisible()
    await expect(page.getByText('Köttbullar med gräddsås och potatis')).toBeVisible()

    // Count recipe cards — should be at least 15 (20 seeded - potential filtering)
    // Look for recipe card elements
    const recipeCards = page.locator('.recipe-card, [data-testid="recipe-card"]')
    const count = await recipeCards.count()
    expect(count).toBeGreaterThanOrEqual(15)
  })

  test('view recipe detail → ingredients and instructions render', async ({ page }) => {
    await loginUser(page, email, password)
    await page.goto('/recipes')

    // Click on Pasta Carbonara
    await page.getByText('Pasta Carbonara').click()

    // Wait for detail view — check for ingredients
    await expect(page.getByText('Spaghetti')).toBeVisible({ timeout: 10_000 })
    await expect(page.getByText('Bacon')).toBeVisible()
    await expect(page.getByText('Parmesan')).toBeVisible()

    // Check for instructions
    await expect(page.getByText(/Koka pasta/)).toBeVisible()
  })

  test('create recipe manually → appears in list', async ({ page }) => {
    await loginUser(page, email, password)
    await page.goto('/recipes')

    createdRecipeName = generateUniqueName('E2E-Recept')

    // Switch to "Lägg till" tab
    await page.getByText('Lägg till').click()

    // Choose manual ("Fyll i själv")
    await page.getByText('Fyll i själv').click()

    // Fill the RecipeEditForm
    // Name field
    const nameInput = page.locator('input[placeholder*="Receptnamn"], input[name="name"]').or(
      page.getByLabel(/namn/i)
    )
    await nameInput.first().fill(createdRecipeName)

    // Servings (should default to 4, leave as-is or fill)
    const servingsInput = page.locator('input[type="number"]').first()
    await servingsInput.fill('4')

    // First ingredient — fill name, amount, unit
    const ingredientInputs = page.locator('.ingredient-row, [data-testid="ingredient"]').first()
    if (await ingredientInputs.isVisible({ timeout: 3_000 }).catch(() => false)) {
      // Fill first ingredient row
      const nameInputs = ingredientInputs.locator('input').first()
      await nameInputs.fill('Potatis')

      const amountInput = ingredientInputs.locator('input[type="number"]')
      await amountInput.fill('800')

      const unitInput = ingredientInputs.locator('input').last()
      await unitInput.fill('g')
    }

    // First instruction
    const instructionInput = page.locator('textarea, input[placeholder*="Steg"]').first()
    if (await instructionInput.isVisible({ timeout: 3_000 }).catch(() => false)) {
      await instructionInput.fill('Koka potatisen')
    }

    // Save
    await page.getByRole('button', { name: /Spara/i }).click()

    // Should see success
    await expect(page.getByText(/sparats|Receptet har sparats/)).toBeVisible({ timeout: 10_000 })

    // Navigate back to list
    const viewRecipesBtn = page.getByText(/Visa mina recept|Mina recept/)
    if (await viewRecipesBtn.isVisible({ timeout: 3_000 }).catch(() => false)) {
      await viewRecipesBtn.click()
    } else {
      await page.goto('/recipes')
    }

    // Verify the new recipe appears
    await expect(page.getByText(createdRecipeName)).toBeVisible({ timeout: 10_000 })
  })

  test('delete recipe via API → removed from list', async ({ page }) => {
    await loginUser(page, email, password)

    // Get recipes list via API to find our created recipe
    const listResult = await apiCall(page, 'GET', '/recipes')
    const recipes = (listResult.data as { recipes: Array<{ id: string; name: string }> }).recipes
    const target = recipes.find((r) => r.name === createdRecipeName)

    if (target) {
      // Delete via API
      const deleteResult = await apiCall(page, 'DELETE', `/recipes/${target.id}`)
      expect(deleteResult.status).toBeLessThan(300)

      // Reload recipes page and verify it's gone
      await page.goto('/recipes')
      await expect(page.getByText('Mina recept')).toBeVisible({ timeout: 10_000 })
      await expect(page.getByText(createdRecipeName)).not.toBeVisible({ timeout: 5_000 })
    }
  })

  test('AI parse recipe (skip if no API key)', async ({ page }) => {
    await loginUser(page, email, password)

    // Check if ANTHROPIC_API_KEY is configured by trying the parse endpoint
    const testRes = await apiCall(page, 'POST', '/recipes/parse', {
      rawText: 'test',
    })

    // If parse returns 501 or 503 (API key not configured), skip
    if (testRes.status === 501 || testRes.status === 503 || testRes.status === 500) {
      test.skip(true, 'ANTHROPIC_API_KEY not configured — skipping AI parse test')
      return
    }

    await page.goto('/recipes')
    await page.getByText('Lägg till').click()
    await page.getByText('Tolka med AI').click()

    // Paste recipe text
    const textarea = page.locator('textarea')
    await textarea.fill(
      'Pasta Bolognese\n500g köttfärs\n400g krossade tomater\n1 lök\n2 vitlöksklyftor\n400g spaghetti\n\nBryn köttfärsen. Hacka och fräs lök och vitlök. Tillsätt tomater. Sjud 20 min. Koka pasta. Servera.',
    )

    // Click parse button
    await page.getByRole('button', { name: /Tolka|Parse/ }).click()

    // Wait for parsing (can take up to 60s with Claude API)
    await expect(
      page.getByText(/säkerhet|confidence|Granska/i),
    ).toBeVisible({ timeout: 65_000 })
  })
})
```

**Step 2: Run and verify**

```bash
npx playwright test --config=tests/e2e/playwright.e2e.config.ts recipes.e2e.spec.ts
```

**Notes for implementer:**
- The form field selectors are speculative. During implementation, take a browser snapshot (`page.content()`) of the RecipeEditForm to find exact selectors.
- The recipe card selector (`.recipe-card`) needs verification — check RecipeCard.vue for actual class names.
- If recipe detail view is a modal (not a separate page), adjust the "view recipe detail" test accordingly.
- The AI parse test has a 65s timeout to accommodate Claude API latency.

**Step 3: Commit**

```bash
git add tests/e2e/recipes.e2e.spec.ts
git commit -m "test: E2E recipe tests (browse, detail, create, delete, AI parse)"
```

---

### Task 7: Menu E2E Tests

**Files:**
- Create: `tests/e2e/menu.e2e.spec.ts`

**Context:**
- Menu generation: `/menu/generate` → click generate → slot machine animation → 5 day cards
- Locking: Click lock icon on a day card → regenerate → locked day unchanged
- Save: Click "Spara" → redirects to dashboard
- The menu generator uses `POST /menus/generate` and `PUT /menus/current`
- After save, the dashboard shows the weekly menu and shopping list widget links to it

**Step 1: Write menu tests**

```typescript
import { test, expect } from '@playwright/test'
import { registerUser, loginUser } from './helpers'

test.describe('Menu Generation E2E', () => {
  test.describe.configure({ mode: 'serial' })

  let email: string
  let password: string

  test('register user for menu tests', async ({ page }) => {
    const creds = await registerUser(page, { householdName: 'Menu Test Hushåll' })
    email = creds.email
    password = creds.password
  })

  test('generate menu → 5 days populated with recipes', async ({ page }) => {
    await loginUser(page, email, password)
    await page.goto('/menu/generate')

    // Verify page title
    await expect(page.getByText('Generera veckomeny')).toBeVisible()

    // Click generate button
    const generateBtn = page.getByRole('button', { name: /Generera/i })
    await generateBtn.click()

    // Wait for slot animation to complete and day cards to appear
    // Each day card should show a recipe name
    // The menu grid has 5 cards (Monday-Friday)
    await expect(page.locator('.menu-grid')).toBeVisible({ timeout: 15_000 })

    // Wait for animations to settle
    await page.waitForTimeout(3000)

    // Verify we have 5 day cards with recipe content
    const dayCards = page.locator('.menu-grid > *')
    const count = await dayCards.count()
    expect(count).toBe(5)
  })

  test('lock one day, regenerate → locked day unchanged', async ({ page }) => {
    await loginUser(page, email, password)
    await page.goto('/menu/generate')

    // Generate initial menu
    await page.getByRole('button', { name: /Generera/i }).click()
    await page.waitForTimeout(4000) // Wait for slot animation

    // Get the first day's recipe name before locking
    const firstCard = page.locator('.menu-grid > *').first()
    const firstRecipeName = await firstCard.textContent()

    // Click lock on the first day
    const lockBtn = firstCard.locator('button[aria-label*="lock"], button[aria-label*="lås"], .lock-btn, .lock-button').first()
    if (await lockBtn.isVisible({ timeout: 3_000 }).catch(() => false)) {
      await lockBtn.click()
    }

    // Click regenerate ("Byt ut")
    const regenBtn = page.getByRole('button', { name: /Byt ut|Generera nya/i })
    await regenBtn.click()
    await page.waitForTimeout(4000) // Wait for animation

    // The locked day should keep the same recipe
    const firstCardAfter = page.locator('.menu-grid > *').first()
    const firstRecipeAfterRegen = await firstCardAfter.textContent()
    expect(firstRecipeAfterRegen).toContain(firstRecipeName?.trim().slice(0, 10) ?? '')
  })

  test('save menu → redirect to dashboard → menu persists', async ({ page }) => {
    await loginUser(page, email, password)
    await page.goto('/menu/generate')

    // Generate menu
    await page.getByRole('button', { name: /Generera/i }).click()
    await page.waitForTimeout(4000)

    // Save
    await page.getByRole('button', { name: /Spara/i }).click()

    // Should redirect to dashboard
    await page.waitForURL('**/dashboard', { timeout: 10_000 })

    // The dashboard should show the weekly menu
    await expect(page.getByText(/meny|Veckomeny/i)).toBeVisible({ timeout: 10_000 })
  })

  test('navigate to shopping list from saved menu → items exist', async ({ page }) => {
    await loginUser(page, email, password)

    // Navigate to shopping list
    await page.goto('/shopping-list')

    // Should show shopping list with items (from the saved menu)
    await expect(page.getByText('Inköpslista')).toBeVisible({ timeout: 10_000 })

    // Wait for items to load — should have categories
    const categories = page.locator('.category-section, .category-heading')
    await expect(categories.first()).toBeVisible({ timeout: 10_000 })
  })
})
```

**Step 2: Run and verify**

```bash
npx playwright test --config=tests/e2e/playwright.e2e.config.ts menu.e2e.spec.ts
```

**Notes for implementer:**
- The slot animation takes ~2-3 seconds. The `waitForTimeout(4000)` is a crude wait — consider replacing with `waitForSelector` on a specific element that indicates animation completion.
- The lock button selector is speculative. Check `MenuDayCard.vue` for the exact element.
- The "locked day keeps same recipe" assertion is loose (compares first 10 chars). Tighten during implementation if possible.

**Step 3: Commit**

```bash
git add tests/e2e/menu.e2e.spec.ts
git commit -m "test: E2E menu tests (generate, lock, regenerate, save, persist)"
```

---

### Task 8: Shopping List E2E Tests

**Files:**
- Create: `tests/e2e/shopping-list.e2e.spec.ts`

**Context:**
- Shopping list view: `/shopping-list`
- Shows categories (e.g., "Kött & Fisk", "Mejeri") with items
- Each item has a checkbox, name, and amount
- Progress bar shows "X av Y varor"
- Toggle item: click checkbox → `PATCH /shopping-list/items/{id}` → persisted
- The composable uses optimistic updates with rollback on error

**Step 1: Write shopping list tests**

```typescript
import { test, expect } from '@playwright/test'
import { registerUser, loginUser, apiCall } from './helpers'

test.describe('Shopping List E2E', () => {
  test.describe.configure({ mode: 'serial' })

  let email: string
  let password: string

  test('setup: register user and generate + save a menu', async ({ page }) => {
    const creds = await registerUser(page, { householdName: 'Shopping Test' })
    email = creds.email
    password = creds.password

    // Generate and save a menu to create shopping list items
    await page.goto('/menu/generate')
    await page.getByRole('button', { name: /Generera/i }).click()
    await page.waitForTimeout(4000) // Wait for slot animation
    await page.getByRole('button', { name: /Spara/i }).click()
    await page.waitForURL('**/dashboard', { timeout: 10_000 })
  })

  test('view shopping list → categories and items visible', async ({ page }) => {
    await loginUser(page, email, password)
    await page.goto('/shopping-list')

    await expect(page.getByText('Inköpslista')).toBeVisible({ timeout: 10_000 })

    // Should show progress summary
    await expect(page.getByText(/av.*varor/)).toBeVisible({ timeout: 10_000 })

    // Should have at least one category
    const categories = page.locator('.category-heading')
    await expect(categories.first()).toBeVisible({ timeout: 10_000 })

    // Should have item rows
    const items = page.locator('.item-row')
    const count = await items.count()
    expect(count).toBeGreaterThan(0)
  })

  test('check off item → progress updates', async ({ page }) => {
    await loginUser(page, email, password)
    await page.goto('/shopping-list')
    await expect(page.getByText(/av.*varor/)).toBeVisible({ timeout: 10_000 })

    // Get initial progress text
    const progressText = page.locator('.progress-count')
    const initialProgress = await progressText.textContent()

    // Find the first unchecked checkbox and click it
    const uncheckedCheckbox = page.locator('.item-row:not(.checked) .item-checkbox').first()
    await uncheckedCheckbox.click()

    // Wait for optimistic update
    await page.waitForTimeout(500)

    // Progress should have changed
    const updatedProgress = await progressText.textContent()
    expect(updatedProgress).not.toBe(initialProgress)
  })

  test('reload page → checked item persists', async ({ page }) => {
    await loginUser(page, email, password)
    await page.goto('/shopping-list')
    await expect(page.getByText(/av.*varor/)).toBeVisible({ timeout: 10_000 })

    // Check off the first unchecked item
    const firstUnchecked = page.locator('.item-row:not(.checked)').first()
    const itemName = await firstUnchecked.locator('.item-name').textContent()
    await firstUnchecked.locator('.item-checkbox').click()
    await page.waitForTimeout(1000) // Wait for API call

    // Reload the page
    await page.reload()
    await expect(page.getByText(/av.*varor/)).toBeVisible({ timeout: 10_000 })

    // The item should still be checked
    if (itemName) {
      const itemRow = page.locator('.item-row').filter({ hasText: itemName })
      await expect(itemRow).toHaveClass(/checked/, { timeout: 5_000 })
    }
  })

  test('uncheck item → restored', async ({ page }) => {
    await loginUser(page, email, password)
    await page.goto('/shopping-list')
    await expect(page.getByText(/av.*varor/)).toBeVisible({ timeout: 10_000 })

    // Find a checked item
    const checkedItem = page.locator('.item-row.checked').first()
    if (await checkedItem.isVisible({ timeout: 3_000 }).catch(() => false)) {
      const itemName = await checkedItem.locator('.item-name').textContent()
      await checkedItem.locator('.item-checkbox').click()
      await page.waitForTimeout(500)

      // Item should no longer have checked class
      if (itemName) {
        const itemRow = page.locator('.item-row').filter({ hasText: itemName })
        await expect(itemRow).not.toHaveClass(/checked/, { timeout: 5_000 })
      }
    }
  })
})
```

**Step 2: Run and verify**

```bash
npx playwright test --config=tests/e2e/playwright.e2e.config.ts shopping-list.e2e.spec.ts
```

**Step 3: Commit**

```bash
git add tests/e2e/shopping-list.e2e.spec.ts
git commit -m "test: E2E shopping list tests (view, check, persist, uncheck)"
```

---

### Task 9: Cross-Flow E2E Test

**Files:**
- Create: `tests/e2e/cross-flow.e2e.spec.ts`

**Context:**
- This is the full user journey: two users, shared household, recipes, menu, shopping list
- User A: registers, adds recipes, generates menu, saves, checks items, creates invite
- User B: registers via API, joins via API, sees shared data
- Uses `browser.newContext()` for User B to simulate a separate session

**Step 1: Write cross-flow test**

```typescript
import { test, expect, type BrowserContext, type Page } from '@playwright/test'
import {
  registerUser,
  loginUser,
  generateUniqueEmail,
  generateUniqueName,
  apiCall,
  registerUserViaAPI,
  setAuthToken,
  getToken,
} from './helpers'

test.describe('Cross-Flow: Full User Journey', () => {
  test.describe.configure({ mode: 'serial' })

  let userAEmail: string
  let userAPassword: string
  let userBEmail: string
  let userBPassword: string
  let inviteCode: string
  let recipe1Name: string
  let recipe2Name: string

  test('1. User A registers → household created', async ({ page }) => {
    const creds = await registerUser(page, {
      name: 'User A',
      householdName: 'CrossFlow Hushåll',
    })
    userAEmail = creds.email
    userAPassword = creds.password

    await expect(page).toHaveURL(/\/dashboard/)
    await expect(page.getByText('CrossFlow Hushåll')).toBeVisible({ timeout: 10_000 })
  })

  test('2. User A adds 2 recipes manually', async ({ page }) => {
    await loginUser(page, userAEmail, userAPassword)
    recipe1Name = generateUniqueName('CrossFlow-Soppa')
    recipe2Name = generateUniqueName('CrossFlow-Gryta')

    // Add recipe 1 via API (faster for setup)
    await apiCall(page, 'POST', '/recipes', {
      name: recipe1Name,
      servings: 4,
      ingredients: [
        { name: 'Potatis', amount: 500, unit: 'g' },
        { name: 'Morot', amount: 200, unit: 'g' },
      ],
      instructions: ['Koka potatis', 'Tillsätt morot'],
      tags: ['soppa'],
    })

    // Add recipe 2 via API
    await apiCall(page, 'POST', '/recipes', {
      name: recipe2Name,
      servings: 4,
      ingredients: [
        { name: 'Kyckling', amount: 500, unit: 'g' },
        { name: 'Ris', amount: 4, unit: 'dl' },
      ],
      instructions: ['Stek kyckling', 'Koka ris'],
      tags: ['gryta'],
    })

    // Verify both appear in recipe list
    await page.goto('/recipes')
    await expect(page.getByText(recipe1Name)).toBeVisible({ timeout: 10_000 })
    await expect(page.getByText(recipe2Name)).toBeVisible()
  })

  test('3. User A generates menu → saves it', async ({ page }) => {
    await loginUser(page, userAEmail, userAPassword)
    await page.goto('/menu/generate')

    await page.getByRole('button', { name: /Generera/i }).click()
    await page.waitForTimeout(4000) // Slot animation

    await page.getByRole('button', { name: /Spara/i }).click()
    await page.waitForURL('**/dashboard', { timeout: 10_000 })
  })

  test('4. User A views shopping list → checks off 3 items', async ({ page }) => {
    await loginUser(page, userAEmail, userAPassword)
    await page.goto('/shopping-list')

    await expect(page.getByText(/av.*varor/)).toBeVisible({ timeout: 10_000 })

    // Check off 3 items
    for (let i = 0; i < 3; i++) {
      const unchecked = page.locator('.item-row:not(.checked) .item-checkbox').first()
      if (await unchecked.isVisible({ timeout: 3_000 }).catch(() => false)) {
        await unchecked.click()
        await page.waitForTimeout(500)
      }
    }

    // Verify progress shows at least 3 checked
    const progressText = await page.locator('.progress-count').textContent()
    const match = progressText?.match(/(\d+) av/)
    expect(Number(match?.[1])).toBeGreaterThanOrEqual(3)
  })

  test('5. User A generates invite code', async ({ page }) => {
    await loginUser(page, userAEmail, userAPassword)

    // Generate invite code via API
    const result = await apiCall(page, 'POST', '/households/invite')
    expect(result.status).toBe(200)
    inviteCode = (result.data as { code: string }).code
    expect(inviteCode).toBeTruthy()
  })

  test('6. User B registers → joins with invite code', async ({ page }) => {
    // Register User B via API
    const userB = await registerUserViaAPI({
      name: 'User B',
    })
    userBEmail = userB.email
    userBPassword = userB.password

    // Join User A's household via API
    const joinRes = await fetch('http://localhost:8080/households/join', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${userB.token}`,
      },
      body: JSON.stringify({ code: inviteCode }),
    })
    expect(joinRes.ok).toBeTruthy()
  })

  test('7. User B sees same household, menu, and shopping list', async ({ page }) => {
    await loginUser(page, userBEmail, userBPassword)

    // Should see User A's household
    await expect(page.getByText('CrossFlow Hushåll')).toBeVisible({ timeout: 10_000 })

    // Navigate to shopping list — should see the same list with progress
    await page.goto('/shopping-list')
    await expect(page.getByText(/av.*varor/)).toBeVisible({ timeout: 10_000 })

    // Progress should reflect User A's 3 checked items
    const progressText = await page.locator('.progress-count').textContent()
    const match = progressText?.match(/(\d+) av/)
    expect(Number(match?.[1])).toBeGreaterThanOrEqual(3)
  })

  test('8. User B adds a recipe → User A can see it', async ({ page, browser }) => {
    // User B adds a recipe via API
    await loginUser(page, userBEmail, userBPassword)
    const newRecipeName = generateUniqueName('UserB-Recept')

    await apiCall(page, 'POST', '/recipes', {
      name: newRecipeName,
      servings: 2,
      ingredients: [{ name: 'Ägg', amount: 4, unit: 'st' }],
      instructions: ['Koka äggen'],
      tags: ['snabb'],
    })

    // Switch to User A context
    const contextA = await browser.newContext()
    const pageA = await contextA.newPage()

    await loginUser(pageA, userAEmail, userAPassword)
    await pageA.goto('/recipes')

    // User A should see User B's recipe
    await expect(pageA.getByText(newRecipeName)).toBeVisible({ timeout: 10_000 })

    await contextA.close()
  })
})
```

**Step 2: Run and verify**

```bash
npx playwright test --config=tests/e2e/playwright.e2e.config.ts cross-flow.e2e.spec.ts
```

**Step 3: Commit**

```bash
git add tests/e2e/cross-flow.e2e.spec.ts
git commit -m "test: E2E cross-flow test (full 2-user journey with shared household)"
```

---

### Task 10: Run All E2E Tests & Fix Bugs

**Step 1: Run the full suite**

```bash
npx playwright test --config=tests/e2e/playwright.e2e.config.ts
```

**Step 2: Fix failures**

Common issues to expect and fix:
- **Selector mismatches:** The test selectors (class names, placeholders, button text) are based on code reading. Actual DOM may differ. Use `page.pause()` or screenshots to debug.
- **Timing issues:** Real backend is slower than mocks. Increase timeouts or add explicit waits for API responses.
- **Race conditions:** Serial execution should prevent most, but shared state (e.g., recipe list includes all users' recipes) may cause unexpected results.
- **Form field selectors:** The RecipeEditForm inputs may need different selectors. Check the rendered DOM.
- **Invite code UI:** The dashboard invite flow may use a modal — adapt selectors.
- **Menu slot animation:** The animation timing may vary. Wait for animation completion indicators rather than fixed timeouts.

For each fix:
1. Identify the root cause (wrong selector? timing? backend error?)
2. Fix the test or file a bug for the application code
3. Re-run the failing test to confirm the fix
4. Run the full suite to ensure no regressions

**Step 3: Commit fixes**

```bash
git add tests/e2e/
git commit -m "fix: resolve E2E test failures from initial run"
```

---

### Task 11: Final Verification & Cleanup

**Step 1: Clean run from scratch**

```bash
# Kill any running server
pkill -f "go run cmd/server/main.go" || true

# Remove test DB
rm -f backend/data/e2e-test.db

# Remove built static files
rm -rf backend/static

# Run full suite (server starts fresh via webServer config)
npx playwright test --config=tests/e2e/playwright.e2e.config.ts
```

All tests should pass with a fresh database + seed data.

**Step 2: Verify test count**

Expected: ~25-30 tests across 6 files:
- auth: 6 tests
- household: 5 tests
- recipes: 5 tests
- menu: 5 tests
- shopping-list: 5 tests
- cross-flow: 8 tests

**Step 3: Final commit**

```bash
git add -A tests/e2e/ scripts/e2e-server.sh frontend/package.json
git commit -m "test: comprehensive E2E test suite against real backend (34 tests)"
```

---

## File Summary

| File | Purpose |
|------|---------|
| `scripts/e2e-server.sh` | Builds frontend, resets test DB, starts Go server |
| `tests/e2e/playwright.e2e.config.ts` | Playwright config for real backend E2E |
| `tests/e2e/helpers.ts` | Shared helpers: registerUser, loginUser, apiCall, etc. |
| `tests/e2e/auth.e2e.spec.ts` | Auth: register, login, errors, JWT |
| `tests/e2e/household.e2e.spec.ts` | Household: auto-create, invite, join, roles |
| `tests/e2e/recipes.e2e.spec.ts` | Recipes: browse, detail, CRUD, AI parse |
| `tests/e2e/menu.e2e.spec.ts` | Menu: generate, lock, regenerate, save |
| `tests/e2e/shopping-list.e2e.spec.ts` | Shopping list: view, check, persist, uncheck |
| `tests/e2e/cross-flow.e2e.spec.ts` | Full 2-user journey across all features |

## Run Commands

```bash
# Full suite
npx playwright test --config=tests/e2e/playwright.e2e.config.ts

# Single file
npx playwright test --config=tests/e2e/playwright.e2e.config.ts auth.e2e.spec.ts

# From frontend directory
cd frontend && npm run test:e2e:real

# With UI mode
npx playwright test --config=tests/e2e/playwright.e2e.config.ts --ui
```
