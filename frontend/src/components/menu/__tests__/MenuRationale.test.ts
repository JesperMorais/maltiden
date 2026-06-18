import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import MenuRationale from '../MenuRationale.vue'

describe('MenuRationale', () => {
  it('renders nothing when rationale is empty', () => {
    const wrapper = mount(MenuRationale, { props: { rationale: '' } })
    expect(wrapper.find('.menu-rationale').exists()).toBe(false)
  })

  it('renders nothing when rationale is only whitespace', () => {
    const wrapper = mount(MenuRationale, { props: { rationale: '   ' } })
    expect(wrapper.find('.menu-rationale').exists()).toBe(false)
  })

  it('renders nothing when rationale is omitted', () => {
    const wrapper = mount(MenuRationale)
    expect(wrapper.find('.menu-rationale').exists()).toBe(false)
  })

  it('shows the rationale text when present', () => {
    const wrapper = mount(MenuRationale, {
      props: { rationale: 'Kycklingen från söndag återanvänds i onsdagens sallad.' },
    })
    expect(wrapper.find('.menu-rationale').exists()).toBe(true)
    expect(wrapper.text()).toContain('Kycklingen från söndag')
    expect(wrapper.text()).toContain('Därför funkar veckan')
  })
})
