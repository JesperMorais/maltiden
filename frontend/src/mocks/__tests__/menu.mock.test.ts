import { describe, it, expect } from 'vitest'
import { mockGenerateMenu } from '../menu.mock'

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
})
