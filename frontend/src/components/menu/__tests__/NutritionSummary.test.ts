import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import NutritionSummary from '../NutritionSummary.vue'
import type { WeeklyNutrition } from '@/stores/menuGenerator'

function makeNutrition(overrides: Partial<WeeklyNutrition> = {}): WeeklyNutrition {
  return {
    daysWithMacros: 4,
    avgKcal: 540,
    avgProtein: 95,
    ...overrides,
  }
}

describe('NutritionSummary', () => {
  it('renders nothing when no day has macros (nutrition is null)', () => {
    const wrapper = mount(NutritionSummary, { props: { nutrition: null } })
    expect(wrapper.find('.nutrition-summary').exists()).toBe(false)
  })

  it('shows average per-day protein and kcal when nutrition is present', () => {
    const wrapper = mount(NutritionSummary, {
      props: { nutrition: makeNutrition() },
    })
    const text = wrapper.find('.nutrition-summary').text()
    expect(text).toContain('ca 95 g')
    expect(text).toContain('ca 540 kcal')
  })

  it('shows the target comparison when a protein target is set', () => {
    const wrapper = mount(NutritionSummary, {
      props: { nutrition: makeNutrition({ avgProtein: 95 }), proteinTarget: 110 },
    })
    const text = wrapper.find('.nutrition-summary').text()
    expect(text).toContain('mål 110 g')
    // Target not reached → no "uppnått" badge.
    expect(wrapper.find('.reached-badge').exists()).toBe(false)
  })

  it('flags the target as reached when average meets or exceeds it', () => {
    const wrapper = mount(NutritionSummary, {
      props: { nutrition: makeNutrition({ avgProtein: 120 }), proteinTarget: 110 },
    })
    expect(wrapper.find('.reached-badge').exists()).toBe(true)
    expect(wrapper.find('.reached-badge').text()).toContain('Proteinmål uppnått')
  })

  it('omits the target text when no target is set', () => {
    const wrapper = mount(NutritionSummary, {
      props: { nutrition: makeNutrition(), proteinTarget: 0 },
    })
    expect(wrapper.find('.nutrition-summary').text()).not.toContain('mål')
  })
})
