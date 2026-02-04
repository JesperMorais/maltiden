/**
 * Shopping List API Mock Data
 */

import type { ShoppingList } from '@/api/shopping.api'

const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

const mockShoppingList: ShoppingList = {
  menuId: 'menu_current',
  categories: [
    {
      name: 'Kött & Fisk',
      items: [
        { id: 'item_1', name: 'Kycklingfilé', amount: 500, unit: 'g', checked: false },
        { id: 'item_2', name: 'Köttfärs', amount: 900, unit: 'g', checked: false },
        { id: 'item_3', name: 'Bacon', amount: 200, unit: 'g', checked: true },
        { id: 'item_4', name: 'Laxfilé', amount: 600, unit: 'g', checked: false }
      ]
    },
    {
      name: 'Mejeri',
      items: [
        { id: 'item_5', name: 'Ägg', amount: 4, unit: 'st', checked: false },
        { id: 'item_6', name: 'Parmesan', amount: 100, unit: 'g', checked: false },
        { id: 'item_7', name: 'Riven ost', amount: 200, unit: 'g', checked: true }
      ]
    },
    {
      name: 'Grönsaker',
      items: [
        { id: 'item_8', name: 'Wokgrönsaker', amount: 400, unit: 'g', checked: false },
        { id: 'item_9', name: 'Sallad', amount: 1, unit: 'st', checked: false },
        { id: 'item_10', name: 'Tomat', amount: 3, unit: 'st', checked: false },
        { id: 'item_11', name: 'Lök', amount: 2, unit: 'st', checked: true },
        { id: 'item_12', name: 'Potatis', amount: 800, unit: 'g', checked: false }
      ]
    },
    {
      name: 'Skafferi',
      items: [
        { id: 'item_13', name: 'Spaghetti', amount: 400, unit: 'g', checked: false },
        { id: 'item_14', name: 'Ris', amount: 4, unit: 'dl', checked: true },
        { id: 'item_15', name: 'Krossade tomater', amount: 400, unit: 'g', checked: false },
        { id: 'item_16', name: 'Tacokrydda', amount: 1, unit: 'påse', checked: false },
        { id: 'item_17', name: 'Tacoskal', amount: 12, unit: 'st', checked: false },
        { id: 'item_18', name: 'Sojasås', amount: 1, unit: 'flaska', checked: true }
      ]
    }
  ]
}

export async function mockGetShoppingList(_menuId?: string): Promise<ShoppingList> {
  await delay(300)
  return JSON.parse(JSON.stringify(mockShoppingList)) // Deep clone
}

export async function mockToggleItem(itemId: string, checked: boolean): Promise<{ ok: boolean }> {
  await delay(200)

  // Update local mock data
  for (const category of mockShoppingList.categories) {
    const item = category.items.find(i => i.id === itemId)
    if (item) {
      item.checked = checked
      break
    }
  }

  return { ok: true }
}
