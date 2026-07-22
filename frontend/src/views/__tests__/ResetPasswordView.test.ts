import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import ResetPasswordView from '../ResetPasswordView.vue'

vi.mock('@/api/auth.api', () => ({
  resetPassword: vi.fn(),
}))

const pushMock = vi.fn()
const routeRef: { query: Record<string, string> } = { query: { token: 'tok-abc' } }

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => routeRef,
  RouterLink: defineComponent({
    props: ['to'],
    setup(_, { slots }) {
      return () => h('a', { href: '#' }, slots.default?.())
    },
  }),
}))

const successToast = vi.fn()
vi.mock('@/composables/useToast', () => ({
  useToast: () => ({ success: successToast, error: vi.fn(), info: vi.fn(), warning: vi.fn() }),
}))

import { resetPassword } from '@/api/auth.api'

function mountView() {
  return mount(ResetPasswordView, {
    global: {
      stubs: {
        BackLink: true,
        BaseThemeToggle: true,
        PasswordStrength: true,
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
      },
    },
  })
}

describe('ResetPasswordView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    routeRef.query = { token: 'tok-abc' }
  })

  it('shows an error if the token is missing from the query', async () => {
    routeRef.query = {}
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.text()).toContain('ogiltig eller saknas')
  })

  it('disables submit until both passwords match and meet strength', async () => {
    const wrapper = mountView()
    const button = wrapper.find('button[type="submit"]')
    expect((button.element as HTMLButtonElement).disabled).toBe(true)

    const inputs = wrapper.findAll('input[type="password"]')
    await inputs[0]!.setValue('Strongpass1')
    await inputs[1]!.setValue('Strongpass1')
    expect((button.element as HTMLButtonElement).disabled).toBe(false)
  })

  it('keeps submit disabled if passwords do not match', async () => {
    const wrapper = mountView()
    const inputs = wrapper.findAll('input[type="password"]')
    await inputs[0]!.setValue('Strongpass1')
    await inputs[1]!.setValue('Different123')

    const button = wrapper.find('button[type="submit"]')
    expect((button.element as HTMLButtonElement).disabled).toBe(true)
    expect(wrapper.text()).toContain('matchar inte')
  })

  it('calls resetPassword and redirects to /login on success', async () => {
    vi.mocked(resetPassword).mockResolvedValueOnce(undefined)
    const wrapper = mountView()

    const inputs = wrapper.findAll('input[type="password"]')
    await inputs[0]!.setValue('Strongpass1')
    await inputs[1]!.setValue('Strongpass1')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(resetPassword).toHaveBeenCalledWith('tok-abc', 'Strongpass1')
    expect(successToast).toHaveBeenCalledWith('Lösenord uppdaterat')
    expect(pushMock).toHaveBeenCalledWith('/login')
  })

  it('shows a friendly message when the token has expired', async () => {
    vi.mocked(resetPassword).mockRejectedValueOnce({
      response: { status: 400, data: { error: 'expired_reset_token' } },
    })
    const wrapper = mountView()

    const inputs = wrapper.findAll('input[type="password"]')
    await inputs[0]!.setValue('Strongpass1')
    await inputs[1]!.setValue('Strongpass1')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('gått ut')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows a friendly message when the token is invalid', async () => {
    vi.mocked(resetPassword).mockRejectedValueOnce({
      response: { status: 400, data: { error: 'invalid_reset_token' } },
    })
    const wrapper = mountView()

    const inputs = wrapper.findAll('input[type="password"]')
    await inputs[0]!.setValue('Strongpass1')
    await inputs[1]!.setValue('Strongpass1')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('ogiltig')
  })
})
