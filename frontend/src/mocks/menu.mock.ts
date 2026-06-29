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

const mockRecipes: { id: string; name: string; emoji: string; nutrition: Nutrition }[] = [
  { id: 'rec_1', name: 'Pasta Carbonara', emoji: '🍝', nutrition: { calories: 650, proteinG: 28, carbsG: 70, fatG: 28 } },
  { id: 'rec_2', name: 'Kycklingwok', emoji: '🥘', nutrition: { calories: 520, proteinG: 38, carbsG: 45, fatG: 18 } },
  { id: 'rec_3', name: 'Tacos', emoji: '🌮', nutrition: { calories: 580, proteinG: 30, carbsG: 50, fatG: 26 } },
  { id: 'rec_4', name: 'Laxfilé med potatis', emoji: '🐟', nutrition: { calories: 610, proteinG: 42, carbsG: 40, fatG: 30 } },
  { id: 'rec_5', name: 'Köttfärssås', emoji: '🍖', nutrition: { calories: 700, proteinG: 35, carbsG: 55, fatG: 35 } },
  { id: 'rec_6', name: 'Vegetarisk curry', emoji: '🥗', nutrition: { calories: 480, proteinG: 18, carbsG: 60, fatG: 16 } },
  { id: 'rec_7', name: 'Pizza', emoji: '🍕', nutrition: { calories: 760, proteinG: 30, carbsG: 80, fatG: 32 } }
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

  for (let i = 0; i < request.days; i++) {
    const date = getDateString(i)
    const recipe = mockRecipes[i % mockRecipes.length]!

    if (skipDays.has(date)) {
      days.push({ date, skip: true, servings: 0 })
    } else {
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
    // Illustrative shared-ingredient economy so the UX is exercisable in mock mode.
    sharedIngredients: [
      { name: 'Lök', recipeCount: 3 },
      { name: 'Vitlök', recipeCount: 3 },
      { name: 'Grädde', recipeCount: 2 },
    ],
    rationale,
    wishesIgnored: false,
    nutrition: weeklyMockNutrition(days),
  }
  currentMockMenu = menu
  return menu
}

/** Sum cooked days' per-serving macros × servings, excluding leftover days. */
function weeklyMockNutrition(days: MenuDay[]): Menu['nutrition'] {
  let calories = 0
  let proteinG = 0
  let carbsG = 0
  let fatG = 0
  let partial = false
  let counted = 0
  for (const d of days) {
    if (d.skip || !d.recipeId || d.leftoverOf) continue
    if (!d.nutrition) {
      partial = true
      continue
    }
    calories += (d.nutrition.calories ?? 0) * d.servings
    proteinG += (d.nutrition.proteinG ?? 0) * d.servings
    carbsG += (d.nutrition.carbsG ?? 0) * d.servings
    fatG += (d.nutrition.fatG ?? 0) * d.servings
    counted++
  }
  if (counted === 0) return undefined
  return { calories, proteinG, carbsG, fatG, partial }
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
