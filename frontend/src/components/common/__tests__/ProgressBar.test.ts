import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import ProgressBar from '../ProgressBar.vue'

describe('ProgressBar', () => {
  it('renders when active is true', () => {
    const wrapper = mount(ProgressBar, {
      props: { progress: 50, active: true },
    })

    expect(wrapper.find('.progress-bar-container').exists()).toBe(true)
  })

  it('does not render when active is false', () => {
    const wrapper = mount(ProgressBar, {
      props: { progress: 50, active: false },
    })

    expect(wrapper.find('.progress-bar-container').exists()).toBe(false)
  })

  it('defaults active to true', () => {
    const wrapper = mount(ProgressBar, {
      props: { progress: 30 },
    })

    expect(wrapper.find('.progress-bar-container').exists()).toBe(true)
  })

  it('sets fill width based on progress prop', () => {
    const wrapper = mount(ProgressBar, {
      props: { progress: 75, active: true },
    })

    const fill = wrapper.find('.progress-bar-fill')
    expect(fill.attributes('style')).toContain('width: 75%')
  })

  it('clamps width at 100%', () => {
    const wrapper = mount(ProgressBar, {
      props: { progress: 150, active: true },
    })

    const fill = wrapper.find('.progress-bar-fill')
    expect(fill.attributes('style')).toContain('width: 100%')
  })

  it('shows 0% width for progress 0', () => {
    const wrapper = mount(ProgressBar, {
      props: { progress: 0, active: true },
    })

    const fill = wrapper.find('.progress-bar-fill')
    expect(fill.attributes('style')).toContain('width: 0%')
  })

  it('updates width reactively when progress changes', async () => {
    const wrapper = mount(ProgressBar, {
      props: { progress: 20, active: true },
    })

    let fill = wrapper.find('.progress-bar-fill')
    expect(fill.attributes('style')).toContain('width: 20%')

    await wrapper.setProps({ progress: 80 })

    fill = wrapper.find('.progress-bar-fill')
    expect(fill.attributes('style')).toContain('width: 80%')
  })

  it('has track and fill elements', () => {
    const wrapper = mount(ProgressBar, {
      props: { progress: 50, active: true },
    })

    expect(wrapper.find('.progress-bar-track').exists()).toBe(true)
    expect(wrapper.find('.progress-bar-fill').exists()).toBe(true)
  })
})
