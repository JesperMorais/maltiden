import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import ForgotPasswordView from '../ForgotPasswordView.vue'

// Stub the auth API so we can observe (and control) calls without HTTP.
vi.mock('@/api/auth.api', () => ({
  forgotPassword: vi.fn(),
}))

import { forgotPassword } from '@/api/auth.api'

// Stubs for components that pull in router/theme/etc — keep tests focused on
// this view's behaviour.
const RouterLinkStub = defineComponent({
  props: ['to'],
  setup(_, { slots }) {
    return () => h('a', { href: '#' }, slots.default?.())
  },
})

function mountView() {
  return mount(ForgotPasswordView, {
    global: {
      stubs: {
        BackLink: true,
        BaseThemeToggle: true,
        BaseButton: defineComponent({
          props: ['disabled', 'loading', 'type'],
          setup(props, { slots }) {
            return () =>
              h(
                'button',
                { type: props.type, disabled: props.disabled || props.loading },
                slots.default?.(),
              )
          },
        }),
        RouterLink: RouterLinkStub,
      },
    },
  })
}

describe('ForgotPasswordView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('disables submit until a valid-looking email is entered', async () => {
    const wrapper = mountView()
    const button = wrapper.find('button[type="submit"]')
    expect((button.element as HTMLButtonElement).disabled).toBe(true)

    await wrapper.find('input[type="email"]').setValue('user@example.com')
    expect((button.element as HTMLButtonElement).disabled).toBe(false)
  })

  it('calls forgotPassword and shows the success state on submit', async () => {
    vi.mocked(forgotPassword).mockResolvedValueOnce(undefined)
    const wrapper = mountView()

    await wrapper.find('input[type="email"]').setValue('user@example.com')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(forgotPassword).toHaveBeenCalledWith('user@example.com')
    expect(wrapper.text()).toContain('Kolla din inkorg')
    expect(wrapper.text()).toContain('Om kontot finns')
  })

  it('still shows success even if the API rejects (no leak)', async () => {
    vi.mocked(forgotPassword).mockRejectedValueOnce(new Error('boom'))
    const wrapper = mountView()

    await wrapper.find('input[type="email"]').setValue('unknown@nowhere.com')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Kolla din inkorg')
  })

  it('trims whitespace from the email before submitting', async () => {
    vi.mocked(forgotPassword).mockResolvedValueOnce(undefined)
    const wrapper = mountView()

    await wrapper.find('input[type="email"]').setValue('  user@example.com  ')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(forgotPassword).toHaveBeenCalledWith('user@example.com')
  })
})
