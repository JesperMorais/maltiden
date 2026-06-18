/**
 * Menu API Mock Data
 */

import type { Menu, GenerateMenuRequest, GenerateMenuResponse, SaveMenuDay } from '@/api/menu.api'

const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

const mockRecipes = [
  { id: 'rec_1', name: 'Pasta Carbonara', emoji: '🍝' },
  { id: 'rec_2', name: 'Kycklingwok', emoji: '🥘' },
  { id: 'rec_3', name: 'Tacos', emoji: '🌮' },
  { id: 'rec_4', name: 'Laxfilé med potatis', emoji: '🐟' },
  { id: 'rec_5', name: 'Köttfärssås', emoji: '🍖' },
  { id: 'rec_6', name: 'Vegetarisk curry', emoji: '🥗' },
  { id: 'rec_7', name: 'Pizza', emoji: '🍕' }
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

  const days = []
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
        servings: request.servings + extraPortions
      })
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
