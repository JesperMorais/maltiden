import { ref, computed } from 'vue'
import {
  getShoppingList,
  toggleItem,
  type ShoppingList,
  type ShoppingCategory,
} from '@/api/shopping.api'

export function useShoppingList() {
  const shoppingList = ref<ShoppingList | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  const categories = computed<ShoppingCategory[]>(
    () => shoppingList.value?.categories ?? []
  )

  const totalItems = computed(() =>
    categories.value.reduce((sum, cat) => sum + cat.items.length, 0)
  )

  const checkedItems = computed(() =>
    categories.value.reduce(
      (sum, cat) => sum + cat.items.filter((i) => i.checked).length,
      0
    )
  )

  const remainingItems = computed(() => totalItems.value - checkedItems.value)

  const progress = computed(() =>
    totalItems.value === 0 ? 0 : checkedItems.value / totalItems.value
  )

  async function fetchList(menuId?: string) {
    isLoading.value = true
    error.value = null
    try {
      shoppingList.value = await getShoppingList(menuId)
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Kunde inte hämta inköpslistan'
    } finally {
      isLoading.value = false
    }
  }

  async function toggle(itemId: string, checked: boolean) {
    // Optimistic update
    for (const cat of categories.value) {
      const item = cat.items.find((i) => i.id === itemId)
      if (item) {
        item.checked = checked
        break
      }
    }

    try {
      await toggleItem(itemId, checked)
    } catch {
      // Revert on failure
      for (const cat of categories.value) {
        const item = cat.items.find((i) => i.id === itemId)
        if (item) {
          item.checked = !checked
          break
        }
      }
    }
  }

  return {
    shoppingList,
    isLoading,
    error,
    categories,
    totalItems,
    checkedItems,
    remainingItems,
    progress,
    fetchList,
    toggle,
  }
}
