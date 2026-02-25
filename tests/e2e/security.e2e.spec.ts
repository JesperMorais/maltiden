import { test, expect } from '@playwright/test'
import { registerUserViaAPI, generateUniqueEmail } from './helpers'

const API = 'http://localhost:8080'

/** Authenticated fetch helper */
async function apiFetch(
  method: string,
  path: string,
  token: string,
  body?: Record<string, unknown>,
): Promise<{ status: number; data: Record<string, unknown> }> {
  const res = await fetch(`${API}${path}`, {
    method,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: body ? JSON.stringify(body) : undefined,
  })
  const data = await res.json().catch(() => ({}))
  return { status: res.status, data }
}

const testRecipe = {
  name: 'Security Test Recipe',
  servings: 4,
  ingredients: [{ name: 'Test', amount: 1, unit: 'st' }],
  instructions: ['Step 1'],
  tags: ['security-test'],
}

test.describe('Security: JWT Invalidation on Household Removal', () => {
  test('removed member gets 401 with stale JWT', async () => {
    // Register User A (household owner)
    const userA = await registerUserViaAPI({ name: 'Owner' })

    // Register User B
    const userB = await registerUserViaAPI({ name: 'Member' })

    // User A creates invite
    const inviteRes = await apiFetch('POST', '/households/invite', userA.token)
    expect(inviteRes.status).toBe(201)
    const inviteCode = (inviteRes.data as { code: string }).code

    // User B joins User A's household (save old token first)
    const oldTokenB = userB.token
    const joinRes = await apiFetch('POST', '/households/join', userB.token, {
      code: inviteCode,
    })
    expect(joinRes.status).toBe(200)

    // User B's old token should still work before removal (token_version unchanged)
    // User B needs to re-login to get a token with the new householdID,
    // but their old token_version is still valid
    const preRemoveRes = await apiFetch('GET', '/households/me', oldTokenB)
    expect(preRemoveRes.status).toBe(200)

    // User A removes User B from household
    const removeRes = await apiFetch(
      'DELETE',
      `/households/members/${userB.userId}`,
      userA.token,
    )
    expect(removeRes.status).toBe(200)

    // User B's old token should now be rejected (token_version incremented)
    const postRemoveRes = await apiFetch('GET', '/households/me', oldTokenB)
    expect(postRemoveRes.status).toBe(401)
  })
})

test.describe('Security: Recipe Ownership (Household-scoped)', () => {
  test('cross-household recipe edit blocked with 403', async () => {
    // User A creates a recipe
    const userA = await registerUserViaAPI({ name: 'Chef A' })
    const createRes = await apiFetch('POST', '/recipes', userA.token, testRecipe)
    expect(createRes.status).toBe(201)
    const recipeId = (createRes.data as { id: string }).id

    // User B (different household) tries to update it
    const userB = await registerUserViaAPI({ name: 'Chef B' })
    const updateRes = await apiFetch('PUT', `/recipes/${recipeId}`, userB.token, {
      ...testRecipe,
      name: 'Hacked Recipe',
    })
    expect(updateRes.status).toBe(403)
  })

  test('cross-household recipe delete blocked with 403', async () => {
    const userA = await registerUserViaAPI({ name: 'Chef A' })
    const createRes = await apiFetch('POST', '/recipes', userA.token, testRecipe)
    expect(createRes.status).toBe(201)
    const recipeId = (createRes.data as { id: string }).id

    const userB = await registerUserViaAPI({ name: 'Chef B' })
    const deleteRes = await apiFetch('DELETE', `/recipes/${recipeId}`, userB.token)
    expect(deleteRes.status).toBe(403)
  })

  test('owner can edit and delete their own recipe', async () => {
    const userA = await registerUserViaAPI({ name: 'Chef A' })
    const createRes = await apiFetch('POST', '/recipes', userA.token, testRecipe)
    expect(createRes.status).toBe(201)
    const recipeId = (createRes.data as { id: string }).id

    // Owner can update
    const updateRes = await apiFetch('PUT', `/recipes/${recipeId}`, userA.token, {
      ...testRecipe,
      name: 'Updated Recipe',
    })
    expect(updateRes.status).toBe(200)

    // Owner can delete
    const deleteRes = await apiFetch('DELETE', `/recipes/${recipeId}`, userA.token)
    expect(deleteRes.status).toBe(204)
  })

  test('seed recipes cannot be edited or deleted', async () => {
    const user = await registerUserViaAPI({ name: 'User' })

    // Get seed recipes
    const listRes = await fetch(`${API}/recipes`)
    const recipes = (await listRes.json()) as { recipes: { id: string; name: string }[] }
    expect(recipes.recipes.length).toBeGreaterThan(0)

    // Pick a seed recipe (one from migration 004)
    const seedRecipe = recipes.recipes.find((r) => r.name === 'Pasta Carbonara')
    expect(seedRecipe).toBeTruthy()

    // Try to update seed recipe
    const updateRes = await apiFetch('PUT', `/recipes/${seedRecipe!.id}`, user.token, {
      ...testRecipe,
      name: 'Hacked Seed',
    })
    expect(updateRes.status).toBe(403)

    // Try to delete seed recipe
    const deleteRes = await apiFetch('DELETE', `/recipes/${seedRecipe!.id}`, user.token)
    expect(deleteRes.status).toBe(403)
  })

  test('recipe visibility: own recipes + seed, not other household', async () => {
    const userA = await registerUserViaAPI({
      name: 'Chef A',
      email: generateUniqueEmail(),
    })
    const userB = await registerUserViaAPI({
      name: 'Chef B',
      email: generateUniqueEmail(),
    })

    // User A creates a unique recipe
    const uniqueName = `Private Recipe ${Date.now()}`
    const createRes = await apiFetch('POST', '/recipes', userA.token, {
      ...testRecipe,
      name: uniqueName,
    })
    expect(createRes.status).toBe(201)

    // User A can see their recipe via authenticated GET
    const listA = await apiFetch('GET', '/recipes', userA.token)
    expect(listA.status).toBe(200)
    const recipesA = (listA.data as { recipes: { name: string }[] }).recipes
    expect(recipesA.some((r) => r.name === uniqueName)).toBe(true)

    // User B should NOT see User A's private recipe
    const listB = await apiFetch('GET', '/recipes', userB.token)
    expect(listB.status).toBe(200)
    const recipesB = (listB.data as { recipes: { name: string }[] }).recipes
    expect(recipesB.some((r) => r.name === uniqueName)).toBe(false)

    // Both should see seed recipes
    expect(recipesA.some((r) => r.name === 'Pasta Carbonara')).toBe(true)
    expect(recipesB.some((r) => r.name === 'Pasta Carbonara')).toBe(true)
  })
})
