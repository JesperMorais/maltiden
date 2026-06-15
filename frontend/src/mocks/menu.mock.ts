/**
 * Menu API Mock Data
 */

import type { Menu, MenuDay, GenerateMenuRequest, SaveMenuDay, MenuEconomy } from '@/api/menu.api'

const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

interface MockRecipe {
  id: string
  name: string
  emoji: string
  ingredients: { canonicalName: string; name: string }[]
}

const mockRecipes: MockRecipe[] = [
  {
    id: 'rec_1',
    name: 'Pasta Carbonara',
    emoji: '🍝',
    ingredients: [
      { canonicalName: 'pasta', name: 'Pasta' },
      { canonicalName: 'agg', name: 'Ägg' },
      { canonicalName: 'parmesan', name: 'Parmesan' },
      { canonicalName: 'lok', name: 'Gul lök' }
    ]
  },
  {
    id: 'rec_2',
    name: 'Kycklingwok',
    emoji: '🥘',
    ingredients: [
      { canonicalName: 'kyckling', name: 'Kycklingfilé' },
      { canonicalName: 'ris', name: 'Ris' },
      { canonicalName: 'lok', name: 'Gul lök' },
      { canonicalName: 'paprika', name: 'Paprika' }
    ]
  },
  {
    id: 'rec_3',
    name: 'Tacos',
    emoji: '🌮',
    ingredients: [
      { canonicalName: 'kottfars', name: 'Nötfärs' },
      { canonicalName: 'paprika', name: 'Paprika' },
      { canonicalName: 'lok', name: 'Gul lök' },
      { canonicalName: 'tortilla', name: 'Tortillabröd' }
    ]
  },
  {
    id: 'rec_4',
    name: 'Laxfilé med potatis',
    emoji: '🐟',
    ingredients: [
      { canonicalName: 'lax', name: 'Laxfilé' },
      { canonicalName: 'potatis', name: 'Potatis' },
      { canonicalName: 'smor', name: 'Smör' }
    ]
  },
  {
    id: 'rec_5',
    name: 'Köttfärssås',
    emoji: '🍖',
    ingredients: [
      { canonicalName: 'kottfars', name: 'Nötfärs' },
      { canonicalName: 'pasta', name: 'Pasta' },
      { canonicalName: 'lok', name: 'Gul lök' },
      { canonicalName: 'tomat', name: 'Krossade tomater' }
    ]
  },
  {
    id: 'rec_6',
    name: 'Vegetarisk curry',
    emoji: '🥗',
    ingredients: [
      { canonicalName: 'ris', name: 'Ris' },
      { canonicalName: 'paprika', name: 'Paprika' },
      { canonicalName: 'tomat', name: 'Krossade tomater' },
      { canonicalName: 'lok', name: 'Gul lök' }
    ]
  },
  {
    id: 'rec_7',
    name: 'Pizza',
    emoji: '🍕',
    ingredients: [
      { canonicalName: 'mjol', name: 'Vetemjöl' },
      { canonicalName: 'tomat', name: 'Krossade tomater' },
      { canonicalName: 'ost', name: 'Riven ost' }
    ]
  }
]

const recipeById = new Map(mockRecipes.map((r) => [r.id, r]))

/**
 * Derive a plausible ingredient economy from the chosen recipes:
 * shared ingredients (used by 2+ recipes), total ingredient refs and
 * distinct items to buy.
 */
function computeMockEconomy(recipeIds: string[]): MenuEconomy {
  const refs: { canonicalName: string; name: string }[] = []
  for (const id of recipeIds) {
    const recipe = recipeById.get(id)
    if (recipe) refs.push(...recipe.ingredients)
  }

  const counts = new Map<string, { name: string; recipeCount: number }>()
  for (const ref of refs) {
    const existing = counts.get(ref.canonicalName)
    if (existing) {
      existing.recipeCount += 1
    } else {
      counts.set(ref.canonicalName, { name: ref.name, recipeCount: 1 })
    }
  }

  const sharedIngredients = [...counts.entries()]
    .filter(([, v]) => v.recipeCount >= 2)
    .map(([canonicalName, v]) => ({ canonicalName, name: v.name, recipeCount: v.recipeCount }))
    .sort((a, b) => b.recipeCount - a.recipeCount)

  return {
    sharedIngredients,
    distinctItemsToBuy: counts.size,
    totalIngredientRefs: refs.length
  }
}

/** Module-level state so shopping mock can read current servings */
let currentMockMenu: Menu | null = null

export function getMockMenuState(): Menu | null {
  return currentMockMenu
}

function getDateString(daysFromNow: number): string {
  const date = new Date()
  date.setDate(date.getDate() + daysFromNow)
  // Local calendar date (not UTC) so mock dates match the store's
  // local-anchored placeholder dates and the backend's local day.
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export async function mockGenerateMenu(request: GenerateMenuRequest): Promise<Menu> {
  await delay(1000) // Menu generation takes time

  const days: MenuDay[] = []
  const skipDays = new Set(request.skipDays || [])
  const lockedDays = request.lockedDays ?? {}
  const chosenRecipeIds: string[] = []

  // Cursor used to fill unlocked days by cycling through the recipe list.
  let cycle = 0

  for (let i = 0; i < request.days; i++) {
    const date = getDateString(i)

    if (skipDays.has(date)) {
      days.push({ date, skip: true, servings: 0 })
      continue
    }

    const extraPortions = request.extraPortions?.[date] || 0

    // Honor locked days: keep the exact recipe the client locked.
    const lockedRecipeId = lockedDays[date]
    const recipe = lockedRecipeId
      ? (recipeById.get(lockedRecipeId) ?? mockRecipes[cycle++ % mockRecipes.length]!)
      : mockRecipes[cycle++ % mockRecipes.length]!

    chosenRecipeIds.push(recipe.id)
    days.push({
      date,
      recipeId: recipe.id,
      recipeName: recipe.name,
      emoji: recipe.emoji,
      servings: request.servings + extraPortions
    })
  }

  // Prep mode: place ONE batchable pair across two consecutive days. The first
  // day cooks 2× servings; the second reuses leftovers (base servings,
  // leftover: true, cookDate pointing at the first day).
  if (request.prepMode) {
    for (let i = 0; i < days.length - 1; i++) {
      const cook = days[i]!
      const rest = days[i + 1]!
      // Both days must carry a (non-skipped) recipe to form a batch pair.
      if (cook.recipeId && rest.recipeId && !cook.skip && !rest.skip) {
        cook.servings = request.servings * 2
        rest.recipeId = cook.recipeId
        rest.recipeName = cook.recipeName
        rest.emoji = cook.emoji
        rest.servings = request.servings
        rest.leftover = true
        rest.cookDate = cook.date
        break
      }
    }
  }

  // Economy counts each cooked recipe once — leftovers days reuse a cook day's
  // recipe and contribute no additional ingredient references.
  const cookedRecipeIds = request.prepMode
    ? days.filter((d) => d.recipeId && !d.leftover).map((d) => d.recipeId!)
    : chosenRecipeIds

  const menu: Menu = {
    id: 'menu_mock_' + Date.now(),
    days,
    economy: computeMockEconomy(cookedRecipeIds)
  }
  currentMockMenu = menu
  return menu
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
        leftover: day.leftover,
        cookDate: day.cookDate,
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
        // Batch pair: a cook day (2× servings) followed by a leftovers day.
        { date: getDateString(1), recipeId: 'rec_2', recipeName: 'Kycklingwok', emoji: '🥘', servings: 8 },
        { date: getDateString(2), recipeId: 'rec_2', recipeName: 'Kycklingwok', emoji: '🥘', servings: 4, leftover: true, cookDate: getDateString(1) },
        { date: getDateString(3), recipeId: 'rec_4', recipeName: 'Laxfilé med potatis', emoji: '🐟', servings: 4 },
        { date: getDateString(4), recipeId: 'rec_5', recipeName: 'Köttfärssås', emoji: '🍖', servings: 4 },
        { date: getDateString(5), skip: true, servings: 0 },
        { date: getDateString(6), recipeId: 'rec_6', recipeName: 'Vegetarisk curry', emoji: '🥗', servings: 4 }
      ]
    }
  }

  return JSON.parse(JSON.stringify(currentMockMenu))
}
