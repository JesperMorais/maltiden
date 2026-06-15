import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import MenuDayCard from '../MenuDayCard.vue'
import type { DraftMenuDay } from '@/stores/menuGenerator'

function makeDay(overrides: Partial<DraftMenuDay> = {}): DraftMenuDay {
  return {
    date: '2026-06-17',
    dayName: 'Onsdag',
    dayShort: 'Ons',
    recipeId: 'rec_1',
    recipeName: 'Pasta Carbonara',
    emoji: '🍝',
    servings: 4,
    ...overrides,
  }
}

describe('MenuDayCard — prep mode', () => {
  it('renders a normal day with no prep badge and a lock button', () => {
    const wrapper = mount(MenuDayCard, {
      props: { day: makeDay(), isLocked: false },
    })

    expect(wrapper.find('.prep-badge').exists()).toBe(false)
    expect(wrapper.find('.menu-day-card').classes()).not.toContain('cook-day')
    expect(wrapper.find('.menu-day-card').classes()).not.toContain('leftover')
    expect(wrapper.find('.lock-button').exists()).toBe(true)
    expect(wrapper.find('.recipe-servings').text()).toBe('4 portioner')
  })

  it('renders the cook-day badge and 2× servings', () => {
    const wrapper = mount(MenuDayCard, {
      props: { day: makeDay({ servings: 8, isCookDay: true }), isLocked: false },
    })

    const badge = wrapper.find('.prep-badge-cook')
    expect(badge.exists()).toBe(true)
    expect(badge.text()).toContain('Lagas (2 dagar)')
    expect(wrapper.find('.menu-day-card').classes()).toContain('cook-day')
    // 2× servings come straight from day.servings.
    expect(wrapper.find('.recipe-servings').text()).toBe('8 portioner')
    // A cook day is still lockable.
    expect(wrapper.find('.lock-button').exists()).toBe(true)
  })

  it('renders the leftovers badge and hides the lock button', () => {
    const wrapper = mount(MenuDayCard, {
      props: {
        day: makeDay({ leftover: true, cookDate: '2026-06-16' }),
        isLocked: false,
      },
    })

    const badge = wrapper.find('.prep-badge-leftover')
    expect(badge.exists()).toBe(true)
    expect(badge.text()).toContain('Rester')
    expect(wrapper.find('.menu-day-card').classes()).toContain('leftover')
    expect(wrapper.find('.prep-badge-cook').exists()).toBe(false)
    // Leftovers days follow their cook day → no independent lock control.
    expect(wrapper.find('.lock-button').exists()).toBe(false)
  })

  it('treats a day that is both leftover and (spuriously) isCookDay as leftover only', () => {
    const wrapper = mount(MenuDayCard, {
      props: {
        day: makeDay({ leftover: true, isCookDay: true, cookDate: '2026-06-16' }),
        isLocked: false,
      },
    })
    expect(wrapper.find('.prep-badge-leftover').exists()).toBe(true)
    expect(wrapper.find('.prep-badge-cook').exists()).toBe(false)
  })
})
