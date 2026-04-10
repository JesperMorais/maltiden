/**
 * Shopping List API Service
 * Handles shopping list operations
 */

import apiClient from './client'
import { USE_MOCKS } from '@/mocks'
import { mockGetShoppingList, mockToggleItem } from '@/mocks/shopping.mock'

export interface ShoppingItem {
  id: string
  name: string
  amount: number
  unit: string
  checked: boolean
}

export interface ShoppingCategory {
  name: string
  items: ShoppingItem[]
}

export interface ShoppingList {
  menuId: string
  categories: ShoppingCategory[]
}

/**
 * Get shopping list for a menu
 */
export async function getShoppingList(menuId?: string): Promise<ShoppingList> {
  if (USE_MOCKS) {
    return mockGetShoppingList(menuId)
  }

  const params = menuId ? { menuId } : {}
  const { data } = await apiClient.get<ShoppingList>('/shopping-list', { params })
  return data
}

/**
 * Toggle item checked status
 */
export async function toggleItem(
  itemId: string,
  checked: boolean,
  menuId?: string,
): Promise<{ ok: boolean }> {
  if (USE_MOCKS) {
    return mockToggleItem(itemId, checked)
  }

  const params = menuId ? { menuId } : {}
  const { data } = await apiClient.patch<{ ok: boolean }>(
    `/shopping-list/items/${itemId}`,
    { checked },
    { params },
  )
  return data
}
