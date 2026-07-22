/**
 * Shopping List API Service
 * Handles shopping list operations
 */

import apiClient from './client'
import { USE_MOCKS } from '@/mocks'
import {
  mockGetShoppingList,
  mockToggleItem,
  mockAddCustomItem,
  mockDeleteCustomItem,
} from '@/mocks/shopping.mock'

export interface ShoppingItem {
  id: string
  name: string
  amount: number
  unit: string
  checked: boolean
  isCustom: boolean
}

export interface ShoppingCategory {
  name: string
  items: ShoppingItem[]
}

export interface ShoppingList {
  menuId: string
  categories: ShoppingCategory[]
}

export interface CreateCustomItemRequest {
  name: string
  unit?: string
  amount?: number
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
  menuId: string,
): Promise<{ ok: boolean }> {
  if (USE_MOCKS) {
    return mockToggleItem(itemId, checked)
  }

  const { data } = await apiClient.patch<{ ok: boolean }>(
    `/shopping-list/items/${itemId}`,
    { checked },
    { params: { menuId } },
  )
  return data
}

/**
 * Add a custom item to the shopping list
 */
export async function addCustomItem(
  menuId: string,
  item: CreateCustomItemRequest,
): Promise<ShoppingItem> {
  if (USE_MOCKS) {
    return mockAddCustomItem(menuId, item)
  }

  const { data } = await apiClient.post<ShoppingItem>('/shopping-list/items', item, {
    params: { menuId },
  })
  return { ...data, isCustom: true }
}

/**
 * Delete a custom item from the shopping list
 */
export async function deleteCustomItem(itemId: string): Promise<{ ok: boolean }> {
  if (USE_MOCKS) {
    return mockDeleteCustomItem(itemId)
  }

  const { data } = await apiClient.delete<{ ok: boolean }>(`/shopping-list/items/${itemId}`)
  return data
}
