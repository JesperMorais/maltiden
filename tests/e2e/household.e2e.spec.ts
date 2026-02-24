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

  // Shared state across serial tests
  let ownerEmail: string
  let ownerPassword: string
  let ownerName: string
  let inviteCode: string

  test('new user auto-creates household visible on dashboard', async ({ page }) => {
    ownerName = 'Household Owner'
    const { email, password } = await registerUser(page, {
      name: ownerName,
      householdName: 'Ignored by API', // backend derives name from user name
    })
    ownerEmail = email
    ownerPassword = password

    // Backend creates household named "<name>'s household"
    await expect(page.getByText(`${ownerName}'s household`)).toBeVisible({ timeout: 10_000 })

    // The "Hushallet" widget section should be visible
    await expect(page.getByText('Hushållet')).toBeVisible({ timeout: 10_000 })
  })

  test('generate invite code via API', async ({ page }) => {
    await loginUser(page, ownerEmail, ownerPassword)

    // Use the API helper to create an invite code
    const res = await apiCall(page, 'POST', '/households/invite')

    expect(res.status).toBe(201)
    expect(res.data.code).toBeTruthy()
    expect(typeof res.data.code).toBe('string')
    expect(res.data.expiresAt).toBeTruthy()

    inviteCode = res.data.code as string
  })

  test('second user joins with invite code and appears in household', async ({ page }) => {
    // Register user B via API (fast, no UI needed)
    const userB = await registerUserViaAPI({ name: 'Joining Member' })

    // User B joins the owner's household via API
    const joinRes = await fetch('http://localhost:8080/households/join', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${userB.token}`,
      },
      body: JSON.stringify({ code: inviteCode }),
    })

    expect(joinRes.ok).toBe(true)
    const joinData = await joinRes.json()
    expect(joinData.householdId).toBeTruthy()

    // Now login as owner and verify user B appears in the household widget
    await loginUser(page, ownerEmail, ownerPassword)

    // Wait for the household widget members list to load
    await expect(page.getByText('Hushållet')).toBeVisible({ timeout: 10_000 })

    // Verify both owner and the joined member are visible
    await expect(page.getByText(ownerName)).toBeVisible({ timeout: 10_000 })
    await expect(page.getByText('Joining Member')).toBeVisible({ timeout: 10_000 })

    // Member count should now show 2 persons
    await expect(page.getByText('2 personer')).toBeVisible({ timeout: 10_000 })
  })

  test('join with invalid code returns error', async () => {
    // Register user C via API
    const userC = await registerUserViaAPI({ name: 'Invalid Joiner' })

    // Attempt to join with a bogus code
    const joinRes = await fetch('http://localhost:8080/households/join', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${userC.token}`,
      },
      body: JSON.stringify({ code: 'INVALID-CODE-999' }),
    })

    // Backend returns 400 with "invalid_code" for non-existent codes
    expect(joinRes.ok).toBe(false)
    expect(joinRes.status).toBe(400)

    const errorData = await joinRes.json()
    expect(errorData.error).toBe('invalid_code')
  })

  test('owner badge visible on dashboard', async ({ page }) => {
    await loginUser(page, ownerEmail, ownerPassword)

    // Wait for dashboard to load
    await expect(page.getByText('Hushållet')).toBeVisible({ timeout: 10_000 })

    // The owner badge should be visible in both the header and the household widget
    // DashboardHeader shows role badge with text "Agare" for owner role
    const ownerBadges = page.getByText('Ägare')
    await expect(ownerBadges.first()).toBeVisible({ timeout: 10_000 })
  })
})
