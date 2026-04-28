/**
 * Shopping List API Mock Data
 *
 * Builds a dynamic shopping list from mock recipes scaled by current menu servings.
 * When servings change (toggle eating / lunchbox), the amounts update accordingly.
 */

import type {
  ShoppingList,
  ShoppingCategory,
  ShoppingItem,
  CreateCustomItemRequest,
} from '@/api/shopping.api'
import { getMockMenuState } from './menu.mock'
import { mockRecipes } from './recipes.mock'

const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

/** Track checked state by ingredient name so it persists across list rebuilds */
const checkedByName = new Map<string, boolean>([
  ['Bacon', true],
  ['Riven ost', true],
  ['Ris', true],
  ['Lök', true],
  ['Sojasås', true],
])

/** Simple ingredient-to-category mapping for mock data */
const categoryMap: Record<string, string> = {
  Kycklingfilé: 'Kött & Fisk',
  Köttfärs: 'Kött & Fisk',
  Bacon: 'Kött & Fisk',
  Laxfilé: 'Kött & Fisk',
  Skinka: 'Kött & Fisk',
  Ägg: 'Mejeri',
  Parmesan: 'Mejeri',
  'Riven ost': 'Mejeri',
  Mozzarella: 'Mejeri',
  Kokosmjölk: 'Mejeri',
  Wokgrönsaker: 'Grönsaker',
  Sallad: 'Grönsaker',
  Tomat: 'Grönsaker',
  Lök: 'Grönsaker',
  Potatis: 'Grönsaker',
  Citron: 'Grönsaker',
  Dill: 'Grönsaker',
  Vitlök: 'Grönsaker',
  Spenat: 'Grönsaker',
  Champinjoner: 'Grönsaker',
  // Kryddor (mirrors backend categorization)
  Salt: 'Kryddor',
  Svartpeppar: 'Kryddor',
  Peppar: 'Kryddor',
  Oregano: 'Kryddor',
  Basilika: 'Kryddor',
  Timjan: 'Kryddor',
  Rosmarin: 'Kryddor',
  Paprikapulver: 'Kryddor',
  Chiliflakes: 'Kryddor',
  Kanel: 'Kryddor',
  Muskot: 'Kryddor',
}

const categoryOrder = ['Kött & Fisk', 'Mejeri', 'Grönsaker', 'Skafferi', 'Kryddor', 'Egna varor']

/** Build a deterministic UUID-format id for mocks (matches backend prefixedUUIDPattern). */
function mockCustomId(seq: number): string {
  const suffix = seq.toString(16).padStart(12, '0')
  return `citem_00000000-0000-4000-8000-${suffix}`
}

/** Custom items storage */
let customItems: ShoppingItem[] = [
  {
    id: mockCustomId(1),
    name: 'Hushållspapper',
    unit: 'st',
    amount: 2,
    checked: false,
    isCustom: true,
  },
  {
    id: mockCustomId(2),
    name: 'Diskmedel',
    unit: 'st',
    amount: 1,
    checked: false,
    isCustom: true,
  },
]

let nextCustomId = 3

function buildShoppingList(menuId: string): ShoppingList {
  const menu = getMockMenuState()

  // Aggregate ingredients across all menu days, scaled by servings
  const ingredients = new Map<string, { amount: number; unit: string }>()

  if (menu) {
    for (const day of menu.days) {
      if (day.skip || !day.recipeId) continue
      const recipe = mockRecipes.find((r) => r.id === day.recipeId)
      if (!recipe) continue

      const scale = day.servings / recipe.servings
      for (const ing of recipe.ingredients) {
        const existing = ingredients.get(ing.name)
        if (existing) {
          existing.amount = Math.round((existing.amount + ing.amount * scale) * 10) / 10
        } else {
          ingredients.set(ing.name, {
            amount: Math.round(ing.amount * scale * 10) / 10,
            unit: ing.unit,
          })
        }
      }
    }
  }

  // Group into categories
  const groups = new Map<string, ShoppingItem[]>()
  let idx = 0

  for (const [name, { amount, unit }] of ingredients) {
    const category = categoryMap[name] ?? 'Skafferi'
    let items = groups.get(category)
    if (!items) {
      items = []
      groups.set(category, items)
    }
    idx++
    // Spices collapse to name-only (shopping list just needs "buy salt",
    // not "3.6 st"); mirrors backend behaviour so mock mode looks the same.
    const isSpice = category === 'Kryddor'
    items.push({
      id: `item_${idx}`,
      name,
      amount: isSpice ? 0 : amount,
      unit: isSpice ? '' : unit,
      checked: checkedByName.get(name) ?? false,
      isCustom: false,
    })
  }

  // Add custom items
  if (customItems.length > 0) {
    groups.set('Egna varor', [...customItems])
  }

  const categories: ShoppingCategory[] = categoryOrder
    .filter((cat) => groups.has(cat))
    .map((cat) => ({ name: cat, items: groups.get(cat)! }))

  return { menuId, categories }
}

/** Keep a reference to the last built list so toggleItem can find items by ID */
let lastBuiltList: ShoppingList | null = null

export async function mockGetShoppingList(menuId?: string): Promise<ShoppingList> {
  await delay(200)
  const list = buildShoppingList(menuId ?? 'menu_current')
  lastBuiltList = list
  return JSON.parse(JSON.stringify(list))
}

export async function mockToggleItem(
  itemId: string,
  checked: boolean,
): Promise<{ ok: boolean }> {
  await delay(100)

  // Check custom items first
  const customItem = customItems.find((i) => i.id === itemId)
  if (customItem) {
    customItem.checked = checked
    return { ok: true }
  }

  // Find the ingredient name from the last built list
  if (lastBuiltList) {
    for (const category of lastBuiltList.categories) {
      const item = category.items.find((i) => i.id === itemId)
      if (item) {
        checkedByName.set(item.name, checked)
        break
      }
    }
  }

  return { ok: true }
}

export async function mockAddCustomItem(
  _menuId: string,
  req: CreateCustomItemRequest,
): Promise<ShoppingItem> {
  await delay(150)
  const item: ShoppingItem = {
    id: mockCustomId(nextCustomId++),
    name: req.name,
    unit: req.unit ?? 'st',
    amount: req.amount ?? 1,
    checked: false,
    isCustom: true,
  }
  customItems.push(item)
  return JSON.parse(JSON.stringify(item))
}

export async function mockDeleteCustomItem(itemId: string): Promise<{ ok: boolean }> {
  await delay(100)
  customItems = customItems.filter((i) => i.id !== itemId)
  return { ok: true }
}
