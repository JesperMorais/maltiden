import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import type { Recipe } from '@/api/recipes.api'

// Mock the API + side-effect composables so the modal can mount in isolation.
const getRecipe = vi.fn()

vi.mock('@/api/recipes.api', () => ({
  getRecipe: (...args: unknown[]) => getRecipe(...args),
  updateRecipe: vi.fn(),
  deleteRecipe: vi.fn(),
}))

vi.mock('@/api/menu.api', () => ({
  getCurrentMenu: vi.fn(),
  saveMenu: vi.fn(),
}))

vi.mock('@/composables/useToast', () => ({
  useToast: () => ({ success: vi.fn(), error: vi.fn() }),
}))

vi.mock('@/composables/useFocusTrap', () => ({
  useFocusTrap: vi.fn(),
}))

import RecipeDetailModal from '../RecipeDetailModal.vue'

function makeRecipe(overrides: Partial<Recipe> = {}): Recipe {
  return {
    id: 'rec_1',
    name: 'Pasta Carbonara',
    servings: 4,
    emoji: '🍝',
    tags: ['pasta'],
    ingredients: [{ name: 'Spaghetti', amount: 400, unit: 'g' }],
    instructions: ['Koka pastan'],
    ...overrides,
  }
}

describe('RecipeDetailModal — per-serving nutrition', () => {
  beforeEach(() => {
    getRecipe.mockReset()
  })

  it('renders the nutrition block when the recipe has macros', async () => {
    getRecipe.mockResolvedValue(
      makeRecipe({ macros: { kcal: 620, protein: 28, carbs: 72, fat: 24 } }),
    )

    const wrapper = mount(RecipeDetailModal, { props: { recipeId: 'rec_1' } })
    await flushPromises()

    const section = document.querySelector('.nutrition-section')
    expect(section).not.toBeNull()
    const text = section!.textContent ?? ''
    expect(text).toContain('Näringsvärde per portion')
    expect(text).toContain('ca 620 kcal')
    expect(text).toContain('28 g protein')
    expect(text).toContain('72 g kolhydrater')
    expect(text).toContain('24 g fett')

    wrapper.unmount()
  })

  it('renders no nutrition block when the recipe has no macros', async () => {
    getRecipe.mockResolvedValue(makeRecipe())

    const wrapper = mount(RecipeDetailModal, { props: { recipeId: 'rec_1' } })
    await flushPromises()

    expect(document.querySelector('.nutrition-section')).toBeNull()

    wrapper.unmount()
  })
})
