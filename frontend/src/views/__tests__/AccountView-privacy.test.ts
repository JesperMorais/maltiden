import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AccountView from '../AccountView.vue'
import { useUserStore } from '@/stores/user'

vi.mock('@/api/account.api', () => ({
  exportMyData: vi.fn(),
  deleteMyAccount: vi.fn(),
}))

import { deleteMyAccount } from '@/api/account.api'

function mountView() {
  return mount(AccountView, {
    global: {
      stubs: { BackLink: true, RouterLink: { template: '<a><slot /></a>' } },
    },
  })
}

describe('AccountView privacy actions', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  async function openDeleteForm() {
    const wrapper = mountView()
    const start = wrapper.findAll('button').find((b) => b.text().includes('Radera mitt konto'))
    expect(start).toBeTruthy()
    await start!.trigger('click')
    return wrapper
  }

  it('requires a password before calling the API', async () => {
    const wrapper = await openDeleteForm()
    await wrapper.find('form.delete-confirm').trigger('submit')
    await flushPromises()

    expect(deleteMyAccount).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Ange ditt lösenord')
  })

  it('deletes with the password and logs out on success', async () => {
    const logout = vi.spyOn(useUserStore(), 'logout').mockImplementation(() => {})
    vi.mocked(deleteMyAccount).mockResolvedValue()

    const wrapper = await openDeleteForm()
    await wrapper.find('#delete-password-input').setValue('hemligt123')
    await wrapper.find('form.delete-confirm').trigger('submit')
    await flushPromises()

    expect(deleteMyAccount).toHaveBeenCalledWith('hemligt123')
    expect(logout).toHaveBeenCalled()
  })

  it('shows a clear error on a wrong password and stays logged in', async () => {
    const logout = vi.spyOn(useUserStore(), 'logout').mockImplementation(() => {})
    vi.mocked(deleteMyAccount).mockRejectedValue({
      response: { status: 403, data: { error: 'invalid_password' } },
    })

    const wrapper = await openDeleteForm()
    await wrapper.find('#delete-password-input').setValue('fel')
    await wrapper.find('form.delete-confirm').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('Fel lösenord')
    expect(logout).not.toHaveBeenCalled()
  })
})
