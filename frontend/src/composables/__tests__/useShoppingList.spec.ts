import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useShoppingList } from '../useShoppingList'
import { getShoppingList, toggleItem, type ShoppingList } from '@/api/shopping.api'

vi.mock('@/api/shopping.api', () => ({
  getShoppingList: vi.fn(),
  toggleItem: vi.fn(),
}))

const toastError = vi.fn()
vi.mock('@/composables/useToast', () => ({
  useToast: () => ({ error: toastError }),
}))

function makeList(): ShoppingList {
  return {
    menuId: 'menu_1',
    categories: [
      {
        name: 'Frukt & grönt',
        items: [
          { id: 'item_1', name: 'Äpple', amount: 3, unit: 'st', checked: false },
          { id: 'item_2', name: 'Banan', amount: 1, unit: 'kg', checked: true },
        ],
      },
      {
        name: 'Mejeri',
        items: [{ id: 'item_3', name: 'Mjölk', amount: 1, unit: 'l', checked: false }],
      },
    ],
  }
}

describe('useShoppingList', () => {
  beforeEach(() => {
    vi.mocked(getShoppingList).mockReset()
    vi.mocked(toggleItem).mockReset()
    toastError.mockReset()
  })

  it('fetchList populates shoppingList/categories and computes totals', async () => {
    vi.mocked(getShoppingList).mockResolvedValue(makeList())
    const { fetchList, categories, totalItems, checkedItems } = useShoppingList()

    await fetchList('menu_1')

    expect(getShoppingList).toHaveBeenCalledWith('menu_1')
    expect(categories.value).toHaveLength(2)
    expect(totalItems.value).toBe(3)
    expect(checkedItems.value).toBe(1)
  })

  it('toggle flips item checked and recomputes checkedItems/progress', async () => {
    vi.mocked(getShoppingList).mockResolvedValue(makeList())
    vi.mocked(toggleItem).mockResolvedValue({ ok: true })
    const { fetchList, toggle, categories, checkedItems, progress } = useShoppingList()
    await fetchList('menu_1')

    await toggle('item_1', true)

    const item = categories.value[0]?.items.find((i) => i.id === 'item_1')
    expect(item?.checked).toBe(true)
    expect(checkedItems.value).toBe(2)
    expect(progress.value).toBeCloseTo(2 / 3)
  })

  it('toggle reverts checked flag when API rejects', async () => {
    vi.mocked(getShoppingList).mockResolvedValue(makeList())
    vi.mocked(toggleItem).mockRejectedValue(new Error('network error'))
    const { fetchList, toggle, categories } = useShoppingList()
    await fetchList('menu_1')

    await toggle('item_1', true)

    const item = categories.value[0]?.items.find((i) => i.id === 'item_1')
    expect(item?.checked).toBe(false)
    expect(toastError).toHaveBeenCalledWith('Kunde inte uppdatera varan. Försök igen.')
  })

  it('toggle no-ops when no menuId loaded', async () => {
    const { toggle } = useShoppingList()

    await toggle('item_1', true)

    expect(toggleItem).not.toHaveBeenCalled()
  })
})
