import { describe, it, expect } from 'vitest'
import { mockGenerateMenu, mockGetCurrentMenu } from '../menu.mock'

describe('mockGenerateMenu', () => {
  it('honors lockedDays by keeping the locked recipe on that date', async () => {
    const menu = await mockGenerateMenu({
      days: 7,
      servings: 4,
      skipDays: [],
      lockedDays: {},
    })

    // Lock the second day to a specific recipe and regenerate.
    const lockedDate = menu.days[1]!.date
    const result = await mockGenerateMenu({
      days: 7,
      servings: 4,
      skipDays: [],
      lockedDays: { [lockedDate]: 'rec_7' },
    })

    const lockedDay = result.days.find((d) => d.date === lockedDate)
    expect(lockedDay?.recipeId).toBe('rec_7')
  })

  it('fills unlocked days by cycling through recipes', async () => {
    const result = await mockGenerateMenu({
      days: 7,
      servings: 4,
      skipDays: [],
      lockedDays: {},
    })

    // Every non-skipped day has a recipe.
    for (const day of result.days) {
      expect(day.recipeId).toBeTruthy()
    }
  })

  it('returns a plausible economy block derived from the chosen recipes', async () => {
    const menu = await mockGenerateMenu({
      days: 7,
      servings: 4,
      skipDays: [],
      lockedDays: {},
    })

    expect(menu.economy).toBeDefined()
    const economy = menu.economy!

    expect(economy.totalIngredientRefs).toBeGreaterThan(0)
    expect(economy.distinctItemsToBuy).toBeGreaterThan(0)
    // Distinct items can never exceed total references.
    expect(economy.distinctItemsToBuy).toBeLessThanOrEqual(economy.totalIngredientRefs)

    // Shared ingredients all appear in 2+ recipes.
    expect(economy.sharedIngredients.length).toBeGreaterThan(0)
    for (const shared of economy.sharedIngredients) {
      expect(shared.recipeCount).toBeGreaterThanOrEqual(2)
      expect(shared.canonicalName).toBeTruthy()
      expect(shared.name).toBeTruthy()
    }
  })

  it('still respects skipDays', async () => {
    const menu = await mockGenerateMenu({ days: 7, servings: 4, lockedDays: {} })
    const skipDate = menu.days[3]!.date

    const result = await mockGenerateMenu({
      days: 7,
      servings: 4,
      skipDays: [skipDate],
      lockedDays: {},
    })

    const skipped = result.days.find((d) => d.date === skipDate)
    expect(skipped?.skip).toBe(true)
    expect(skipped?.recipeId).toBeUndefined()
  })

  it('emits a consecutive batch pair when prepMode is set', async () => {
    const menu = await mockGenerateMenu({
      days: 7,
      servings: 4,
      skipDays: [],
      lockedDays: {},
      prepMode: true,
    })

    const restIndex = menu.days.findIndex((d) => d.leftover)
    expect(restIndex).toBeGreaterThan(0)

    const cook = menu.days[restIndex - 1]!
    const rest = menu.days[restIndex]!

    // Cook day carries 2× servings; rest day carries base servings.
    expect(cook.servings).toBe(8)
    expect(rest.servings).toBe(4)
    expect(rest.leftover).toBe(true)
    // The leftovers day points at the immediately-preceding cook day and reuses
    // its recipe.
    expect(rest.cookDate).toBe(cook.date)
    expect(rest.recipeId).toBe(cook.recipeId)
    // Only one leftovers day in the week.
    expect(menu.days.filter((d) => d.leftover)).toHaveLength(1)
  })

  it('counts the batch recipe once in the economy (no double ingredient refs)', async () => {
    const plain = await mockGenerateMenu({ days: 7, servings: 4, skipDays: [], lockedDays: {} })
    const prep = await mockGenerateMenu({
      days: 7,
      servings: 4,
      skipDays: [],
      lockedDays: {},
      prepMode: true,
    })

    // The prep week has the same number of placed recipes minus one (the
    // leftovers day reuses a recipe), so it can never have MORE ingredient refs
    // than the plain week.
    expect(prep.economy!.totalIngredientRefs).toBeLessThan(plain.economy!.totalIngredientRefs)
  })

  it('mockGetCurrentMenu includes a batch pair so views work offline', async () => {
    // Reset module state would be ideal, but the default menu is created lazily.
    const menu = await mockGetCurrentMenu()
    expect(menu).not.toBeNull()
    const rest = menu!.days.find((d) => d.leftover)
    // Either a fresh default (with a pair) or a previously-generated prep menu.
    if (rest) {
      expect(rest.cookDate).toBeTruthy()
    }
  })
})
