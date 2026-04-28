import { ref, computed, type Ref, type ComputedRef } from 'vue'
import {
  getShoppingList,
  toggleItem,
  addCustomItem as addCustomItemApi,
  deleteCustomItem as deleteCustomItemApi,
  type ShoppingList,
  type ShoppingCategory,
  type ShoppingItem,
} from '@/api/shopping.api'
import { useToast } from '@/composables/useToast'

export function useShoppingList() {
  const toast = useToast()
  const shoppingList = ref<ShoppingList | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Search
  const searchQuery: Ref<string> = ref('')
  const isSearchOpen: Ref<boolean> = ref(false)

  function toggleSearch() {
    isSearchOpen.value = !isSearchOpen.value
    if (!isSearchOpen.value) {
      searchQuery.value = ''
    }
  }

  // Category collapse
  const collapsedCategories: Ref<Set<string>> = ref(new Set())

  function toggleCategory(name: string) {
    const next = new Set(collapsedCategories.value)
    if (next.has(name)) {
      next.delete(name)
    } else {
      next.add(name)
    }
    collapsedCategories.value = next
  }

  // All categories (raw)
  const allCategories = computed<ShoppingCategory[]>(
    () => shoppingList.value?.categories ?? [],
  )

  // Separated lists — unchecked categories (filtered by search, empty removed)
  const uncheckedCategories: ComputedRef<ShoppingCategory[]> = computed(() => {
    const query = searchQuery.value.toLowerCase().trim()

    return allCategories.value
      .map((cat) => {
        let items = cat.items.filter((i) => !i.checked)
        if (query) {
          items = items.filter((i) => i.name.toLowerCase().includes(query))
        }
        return { name: cat.name, items }
      })
      .filter((cat) => cat.items.length > 0)
  })

  // Flat list of all checked items
  const checkedItems: ComputedRef<ShoppingItem[]> = computed(() =>
    allCategories.value.flatMap((cat) => cat.items.filter((i) => i.checked)),
  )

  const isCheckedCollapsed: Ref<boolean> = ref(true)

  // Stats
  const totalItems = computed(() =>
    allCategories.value.reduce((sum, cat) => sum + cat.items.length, 0),
  )

  const checkedItemCount = computed(() => checkedItems.value.length)

  const remainingItems = computed(() => totalItems.value - checkedItemCount.value)

  const progress = computed(() =>
    totalItems.value === 0 ? 0 : checkedItemCount.value / totalItems.value,
  )

  // Data fetching
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

  // Toggle item
  async function toggle(itemId: string, checked: boolean) {
    const menuId = shoppingList.value?.menuId
    if (!menuId) return

    // Optimistic update
    for (const cat of allCategories.value) {
      const item = cat.items.find((i) => i.id === itemId)
      if (item) {
        item.checked = checked
        break
      }
    }

    try {
      await toggleItem(itemId, checked, menuId)
    } catch {
      // Revert on failure
      for (const cat of allCategories.value) {
        const item = cat.items.find((i) => i.id === itemId)
        if (item) {
          item.checked = !checked
          break
        }
      }
      toast.error('Kunde inte uppdatera varan. Försök igen.')
    }
  }

  // Custom items
  async function addCustomItemToList(
    name: string,
    unit: string = 'st',
    amount: number = 1,
  ): Promise<void> {
    const menuId = shoppingList.value?.menuId
    if (!menuId) return
    if (!shoppingList.value) return

    const trimmedName = name.trim()
    if (!trimmedName) return

    // 1. Generate temp ID for the optimistic placeholder
    const tempId = `tmp_${crypto.randomUUID()}`
    const tempItem: ShoppingItem = {
      id: tempId,
      name: trimmedName,
      amount,
      unit,
      checked: false,
      isCustom: true,
    }

    // 2. Insert placeholder into 'Egna varor' immediately
    const cats = shoppingList.value.categories
    let egna = cats.find((c) => c.name === 'Egna varor')
    if (!egna) {
      egna = { name: 'Egna varor', items: [] }
      cats.push(egna)
    }
    egna.items.push(tempItem)

    try {
      // 3. Await the API call
      const created = await addCustomItemApi(menuId, {
        name: trimmedName,
        unit,
        amount,
      })

      // 4. Replace temp item with server-returned item (find by temp id, swap in place)
      const targetCat = shoppingList.value.categories.find(
        (c) => c.name === 'Egna varor',
      )
      if (targetCat) {
        const idx = targetCat.items.findIndex((i) => i.id === tempId)
        if (idx !== -1) {
          targetCat.items.splice(idx, 1, created)
        }
      }
    } catch {
      // 5. Roll back: remove the temp item, show toast
      const targetCat = shoppingList.value?.categories.find(
        (c) => c.name === 'Egna varor',
      )
      if (targetCat) {
        const idx = targetCat.items.findIndex((i) => i.id === tempId)
        if (idx !== -1) {
          targetCat.items.splice(idx, 1)
        }
      }
      toast.error('Kunde inte lägga till varan. Försök igen.')
    }
  }

  async function removeCustomItem(itemId: string): Promise<void> {
    const menuId = shoppingList.value?.menuId
    if (!menuId) return

    // Optimistic remove
    for (const cat of allCategories.value) {
      const idx = cat.items.findIndex((i) => i.id === itemId)
      if (idx !== -1) {
        cat.items.splice(idx, 1)
        break
      }
    }

    try {
      await deleteCustomItemApi(itemId)
    } catch {
      // Re-fetch on failure to restore state
      await fetchList(menuId)
      toast.error('Kunde inte ta bort varan. Försök igen.')
    }
  }

  return {
    shoppingList,
    isLoading,
    error,

    // Search
    searchQuery,
    isSearchOpen,
    toggleSearch,

    // Category collapse
    collapsedCategories,
    toggleCategory,

    // Separated lists
    uncheckedCategories,
    checkedItems,
    isCheckedCollapsed,

    // Stats
    totalItems,
    checkedItemCount,
    remainingItems,
    progress,

    // Actions
    fetchList,
    toggle,
    addCustomItem: addCustomItemToList,
    removeCustomItem,
  }
}
