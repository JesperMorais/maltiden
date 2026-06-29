/**
 * Menu API Mock Data
 */

import type {
  Menu,
  GenerateMenuRequest,
  GenerateMenuResponse,
  MenuDay,
  Nutrition,
  SaveMenuDay,
} from '@/api/menu.api'

const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

interface MockRecipe {
  id: string
  name: string
  emoji: string
  nutrition: Nutrition
  ingredients: string[]
}

const mockRecipes: MockRecipe[] = [
  { id: 'rec_1', name: 'Pasta Carbonara', emoji: '🍝', nutrition: { calories: 650, proteinG: 28, carbsG: 70, fatG: 28 }, ingredients: ['pasta', 'ägg', 'bacon', 'parmesan', 'grädde', 'vitlök'] },
  { id: 'rec_2', name: 'Kycklingwok', emoji: '🥘', nutrition: { calories: 520, proteinG: 38, carbsG: 45, fatG: 18 }, ingredients: ['kyckling', 'paprika', 'lök', 'vitlök', 'sojasås', 'ris', 'ingefära'] },
  { id: 'rec_3', name: 'Tacos', emoji: '🌮', nutrition: { calories: 580, proteinG: 30, carbsG: 50, fatG: 26 }, ingredients: ['nötfärs', 'tortilla', 'paprika', 'lök', 'vitlök', 'majs', 'ost'] },
  { id: 'rec_4', name: 'Laxfilé med potatis', emoji: '🐟', nutrition: { calories: 610, proteinG: 42, carbsG: 40, fatG: 30 }, ingredients: ['lax', 'potatis', 'citron', 'dill', 'grädde'] },
  { id: 'rec_5', name: 'Köttfärssås', emoji: '🍖', nutrition: { calories: 700, proteinG: 35, carbsG: 55, fatG: 35 }, ingredients: ['nötfärs', 'pasta', 'lök', 'vitlök', 'krossade tomater', 'morot'] },
  { id: 'rec_6', name: 'Vegetarisk curry', emoji: '🥗', nutrition: { calories: 480, proteinG: 18, carbsG: 60, fatG: 16 }, ingredients: ['kikärtor', 'kokosmjölk', 'lök', 'vitlök', 'curry', 'ris', 'paprika'] },
  { id: 'rec_7', name: 'Pizza', emoji: '🍕', nutrition: { calories: 760, proteinG: 30, carbsG: 80, fatG: 32 }, ingredients: ['pizzadeg', 'krossade tomater', 'ost', 'lök', 'basilika', 'paprika'] }
]

/** Module-level state so shopping mock can read current servings */
let currentMockMenu: Menu | null = null

export function getMockMenuState(): Menu | null {
  return currentMockMenu
}

function getDateString(daysFromNow: number): string {
  const date = new Date()
  date.setDate(date.getDate() + daysFromNow)
  return date.toISOString().split('T')[0]!
}

export async function mockGenerateMenu(
  request: GenerateMenuRequest,
): Promise<GenerateMenuResponse> {
  await delay(1000) // Menu generation takes time

  const days: MenuDay[] = []
  const skipDays = new Set(request.skipDays || [])

  // Don't reuse recipes the caller asked to exclude (e.g. locked days on a
  // regenerate), so the same dish doesn't pile up across the week (#248).
  const excluded = new Set(request.excludeRecipeIds ?? [])
  const candidates = mockRecipes.filter((r) => !excluded.has(r.id))
  const pool = candidates.length > 0 ? candidates : mockRecipes

  let pick = 0
  for (let i = 0; i < request.days; i++) {
    const date = getDateString(i)

    if (skipDays.has(date)) {
      days.push({ date, skip: true, servings: 0 })
    } else {
      // Cycle through the candidate pool so each distinct recipe is used once
      // before any repeats — never the same dish twice while others are unused.
      const recipe = pool[pick % pool.length]!
      pick++
      const extraPortions = request.extraPortions?.[date] || 0
      days.push({
        date,
        recipeId: recipe.id,
        recipeName: recipe.name,
        emoji: recipe.emoji,
        servings: request.servings + extraPortions,
        nutrition: recipe.nutrition
      })
    }
  }

  // Simulate batch cooking (#248 Phase 2): pair the first cook-day with the
  // next eligible day, mirroring the backend's deterministic post-pass.
  if (request.prepMode) {
    const cook = days.find((d) => !d.skip && d.recipeId)
    if (cook) {
      const leftover = days.find((d) => !d.skip && d.recipeId && d.date > cook.date)
      if (leftover) {
        cook.prepMode = 'batch'
        cook.servings *= 2
        leftover.recipeId = cook.recipeId
        leftover.recipeName = cook.recipeName
        leftover.emoji = cook.emoji
        leftover.nutrition = cook.nutrition
        leftover.leftoverOf = cook.date
      }
    }
  }

  // AI arrangement (mock): surface a sample Swedish rationale so the feature is
  // visible in dev:mock. Wishes are never ignored in mock mode (AI "available").
  const rationale = request.arrange
    ? 'Veckan börjar lugnt med snabb vardagsmat och bygger upp mot helgens stora middag. Vi har samlat rätter som delar lök och pasta för en smidigare inköpslista.'
    : undefined

  const menu: GenerateMenuResponse = {
    id: 'menu_mock_' + Date.now(),
    days,
    sharedIngredients: computeMockShared(days),
    rationale,
    wishesIgnored: false,
    nutrition: averageMealMockNutrition(days),
  }
  currentMockMenu = menu
  return menu
}

/** Pantry staples excluded from shared-ingredient economy (mirrors backend). */
const MOCK_STAPLES = new Set(['salt', 'peppar', 'svartpeppar', 'olja', 'olivolja', 'smör'])

/**
 * Ingredients used by two or more distinct recipes in the week, most-shared
 * first — computed from the recipes' real ingredient lists (mirrors the
 * backend's computeSharedIngredients), so the count reflects actual overlap.
 */
function computeMockShared(days: MenuDay[]): Menu['sharedIngredients'] {
  const byId = new Map(mockRecipes.map((r) => [r.id, r]))
  const counts = new Map<string, number>()
  const seenRecipe = new Set<string>()
  for (const d of days) {
    if (d.skip || !d.recipeId || seenRecipe.has(d.recipeId)) continue
    seenRecipe.add(d.recipeId)
    const recipe = byId.get(d.recipeId)
    if (!recipe) continue
    for (const name of new Set(recipe.ingredients)) {
      if (MOCK_STAPLES.has(name.toLowerCase())) continue
      counts.set(name, (counts.get(name) ?? 0) + 1)
    }
  }
  return [...counts.entries()]
    .filter(([, n]) => n >= 2)
    .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0], 'sv'))
    .map(([name, recipeCount]) => ({ name, recipeCount }))
}

/**
 * Average per-meal (per-serving) macros across the week's meals — what a
 * typical plate looks like, not a whole-week total. Each non-skipped day with a
 * recipe counts as one meal (leftover days included); macros are per serving.
 */
function averageMealMockNutrition(days: MenuDay[]): Menu['nutrition'] {
  let calories = 0
  let proteinG = 0
  let carbsG = 0
  let fatG = 0
  let meals = 0
  let withData = 0
  for (const d of days) {
    if (d.skip || !d.recipeId) continue
    meals++
    if (!d.nutrition) continue
    calories += d.nutrition.calories ?? 0
    proteinG += d.nutrition.proteinG ?? 0
    carbsG += d.nutrition.carbsG ?? 0
    fatG += d.nutrition.fatG ?? 0
    withData++
  }
  if (withData === 0) return undefined
  return {
    calories: calories / withData,
    proteinG: proteinG / withData,
    carbsG: carbsG / withData,
    fatG: fatG / withData,
    partial: withData < meals,
  }
}

export async function mockSaveMenu(days: SaveMenuDay[]): Promise<Menu> {
  await delay(200)

  const recipeLookup = new Map(mockRecipes.map((r) => [r.id, r]))

  const menu: Menu = {
    id: currentMockMenu?.id ?? 'menu_saved_' + Date.now(),
    days: days.map((day) => {
      const recipe = day.recipeId ? recipeLookup.get(day.recipeId) : undefined
      return {
        date: day.date,
        recipeId: day.recipeId,
        recipeName: recipe?.name,
        emoji: recipe?.emoji,
        servings: day.servings,
        skip: day.skip,
        prepMode: day.prepMode,
        leftoverOf: day.leftoverOf,
        nutrition: recipe?.nutrition,
      }
    }),
  }

  currentMockMenu = menu
  return JSON.parse(JSON.stringify(menu))
}

export async function mockGetCurrentMenu(): Promise<Menu | null> {
  await delay(300)

  if (!currentMockMenu) {
    currentMockMenu = {
      id: 'menu_current',
      days: [
        { date: getDateString(0), recipeId: 'rec_1', recipeName: 'Pasta Carbonara', emoji: '🍝', servings: 4 },
        { date: getDateString(1), recipeId: 'rec_2', recipeName: 'Kycklingwok', emoji: '🥘', servings: 4 },
        { date: getDateString(2), recipeId: 'rec_3', recipeName: 'Tacos', emoji: '🌮', servings: 4 },
        { date: getDateString(3), recipeId: 'rec_4', recipeName: 'Laxfilé med potatis', emoji: '🐟', servings: 4 },
        { date: getDateString(4), recipeId: 'rec_5', recipeName: 'Köttfärssås', emoji: '🍖', servings: 4 },
        { date: getDateString(5), skip: true, servings: 0 },
        { date: getDateString(6), recipeId: 'rec_6', recipeName: 'Vegetarisk curry', emoji: '🥗', servings: 4 }
      ]
    }
  }

  return JSON.parse(JSON.stringify(currentMockMenu))
}
