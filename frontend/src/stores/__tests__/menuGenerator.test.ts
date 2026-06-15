import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useMenuGeneratorStore } from '../menuGenerator'
import { generateMenu } from '@/api/menu.api'
import type { Menu } from '@/api/menu.api'

// vi.mock is hoisted; keep factories free of external references.
vi.mock('@/api/menu.api', () => ({
  generateMenu: vi.fn(),
  saveMenu: vi.fn().mockResolvedValue(undefined),
}))

vi.mock('@/api/client', () => ({
  default: { get: vi.fn(), post: vi.fn(), put: vi.fn() },
}))

vi.mock('@/composables/useToast', () => ({
  useToast: () => ({ success: vi.fn(), error: vi.fn() }),
}))

const mockedGenerate = vi.mocked(generateMenu)

function menuWithShared(shared: Menu['sharedIngredients']): Menu {
  // 7-day week so it maps cleanly onto the initialized draft.
  const days = Array.from({ length: 7 }, (_, i) => ({
    date: `2026-01-0${i + 1}`,
    recipeId: `rec_${i}`,
    recipeName: `Recept ${i}`,
    emoji: '🍽️',
    servings: 4,
  }))
  return { id: 'menu_test', days, sharedIngredients: shared }
}

describe('menuGenerator store — shared ingredients', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('exposes shared ingredients from the generate response', async () => {
    mockedGenerate.mockResolvedValue(
      menuWithShared([
        { name: 'Kyckling', recipeCount: 3 },
        { name: 'Ris', recipeCount: 2 },
      ]),
    )

    const store = useMenuGeneratorStore()
    store.initializeWeek(new Date('2026-01-01'))
    await store.generateInitialMenu()

    expect(store.sharedIngredientsCount).toBe(2)
    expect(store.sharedIngredients[0]).toEqual({ name: 'Kyckling', recipeCount: 3 })
  })

  it('defaults to an empty list when the response omits shared ingredients', async () => {
    mockedGenerate.mockResolvedValue(menuWithShared(undefined))

    const store = useMenuGeneratorStore()
    store.initializeWeek(new Date('2026-01-01'))
    await store.generateInitialMenu()

    expect(store.sharedIngredientsCount).toBe(0)
  })

  it('clears shared ingredients on clearDraft', async () => {
    mockedGenerate.mockResolvedValue(menuWithShared([{ name: 'Lök', recipeCount: 2 }]))

    const store = useMenuGeneratorStore()
    store.initializeWeek(new Date('2026-01-01'))
    await store.generateInitialMenu()
    expect(store.sharedIngredientsCount).toBe(1)

    store.clearDraft()
    expect(store.sharedIngredientsCount).toBe(0)
  })
})
