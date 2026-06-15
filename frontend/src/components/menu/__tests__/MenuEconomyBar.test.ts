import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import MenuEconomyBar from '../MenuEconomyBar.vue'
import type { MenuEconomy } from '@/api/menu.api'

function makeEconomy(overrides: Partial<MenuEconomy> = {}): MenuEconomy {
  return {
    sharedIngredients: [
      { canonicalName: 'lok', name: 'Gul lök', recipeCount: 3 },
      { canonicalName: 'pasta', name: 'Pasta', recipeCount: 2 },
    ],
    distinctItemsToBuy: 8,
    totalIngredientRefs: 13,
    ...overrides,
  }
}

describe('MenuEconomyBar', () => {
  it('renders one chip per shared ingredient with name and ×count', () => {
    const wrapper = mount(MenuEconomyBar, { props: { economy: makeEconomy() } })

    const chips = wrapper.findAll('.ingredient-chip')
    expect(chips).toHaveLength(2)

    expect(chips[0]!.find('.chip-name').text()).toBe('Gul lök')
    expect(chips[0]!.find('.chip-count').text()).toBe('×3')
    expect(chips[1]!.find('.chip-name').text()).toBe('Pasta')
    expect(chips[1]!.find('.chip-count').text()).toBe('×2')
  })

  it('renders the "färre varor att köpa" value (totalIngredientRefs - distinctItemsToBuy)', () => {
    const wrapper = mount(MenuEconomyBar, { props: { economy: makeEconomy() } })

    const savings = wrapper.find('.savings-row').text()
    // 13 - 8 = 5 färre varor; 8 unika varor
    expect(savings).toContain('5')
    expect(savings).toContain('färre varor att köpa')
    expect(savings).toContain('8')
    expect(savings).toContain('unika varor')
  })

  it('clamps the overlap saved at zero when refs are fewer than distinct items', () => {
    const wrapper = mount(MenuEconomyBar, {
      props: {
        economy: makeEconomy({ distinctItemsToBuy: 10, totalIngredientRefs: 4 }),
      },
    })
    const strong = wrapper.find('.savings-indicator strong')
    expect(strong.text()).toBe('0')
  })

  it('renders nothing when economy is null', () => {
    const wrapper = mount(MenuEconomyBar, { props: { economy: null } })
    expect(wrapper.find('.economy-bar').exists()).toBe(false)
  })

  it('renders nothing when sharedIngredients is empty', () => {
    const wrapper = mount(MenuEconomyBar, {
      props: { economy: makeEconomy({ sharedIngredients: [] }) },
    })
    expect(wrapper.find('.economy-bar').exists()).toBe(false)
  })
})
