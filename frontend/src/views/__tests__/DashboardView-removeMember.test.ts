import { describe, it, expect, vi, beforeEach } from 'vitest'
import { flushPromises } from '@vue/test-utils'

function makeHandler(
  dashboardData: { household: { members: { id: string }[] } } | null,
  removeMemberImpl: () => Promise<unknown>,
) {
  const toast = { success: vi.fn(), error: vi.fn() }

  async function handleRemoveMember(memberId: string) {
    if (!confirm('Vill du ta bort den här medlemmen från hushållet?')) return

    const household = dashboardData?.household
    if (!household) return

    const snapshot = [...household.members]
    household.members = household.members.filter((m) => m.id !== memberId)

    try {
      await removeMemberImpl()
      toast.success('Medlemmen har tagits bort.')
    } catch {
      household.members = snapshot
      toast.error('Kunde inte ta bort medlemmen. Försök igen.')
    }
  }

  return { handleRemoveMember, toast }
}

describe('handleRemoveMember', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    window.confirm = vi.fn()
  })

  it('does nothing when confirm is cancelled', async () => {
    vi.mocked(window.confirm).mockReturnValue(false)
    const data = { household: { members: [{ id: 'usr_1' }, { id: 'usr_2' }] } }
    const { handleRemoveMember, toast } = makeHandler(data, vi.fn().mockResolvedValue({ ok: true }))

    await handleRemoveMember('usr_1')
    await flushPromises()

    expect(data.household.members).toHaveLength(2)
    expect(toast.success).not.toHaveBeenCalled()
    expect(toast.error).not.toHaveBeenCalled()
  })

  it('optimistically removes member before API resolves', async () => {
    vi.mocked(window.confirm).mockReturnValue(true)
    let resolveApi!: () => void
    const apiPromise = new Promise<void>((res) => { resolveApi = res })
    const data = { household: { members: [{ id: 'usr_1' }, { id: 'usr_2' }] } }
    const { handleRemoveMember } = makeHandler(data, () => apiPromise)

    const pending = handleRemoveMember('usr_1')
    // Before API resolves, member should already be gone
    expect(data.household.members).toHaveLength(1)
    expect(data.household.members[0]?.id).toBe('usr_2')

    resolveApi()
    await pending
  })

  it('shows success toast after successful removal', async () => {
    vi.mocked(window.confirm).mockReturnValue(true)
    const data = { household: { members: [{ id: 'usr_1' }] } }
    const { handleRemoveMember, toast } = makeHandler(data, vi.fn().mockResolvedValue({ ok: true }))

    await handleRemoveMember('usr_1')
    await flushPromises()

    expect(data.household.members).toHaveLength(0)
    expect(toast.success).toHaveBeenCalledWith('Medlemmen har tagits bort.')
    expect(toast.error).not.toHaveBeenCalled()
  })

  it('rolls back and shows error toast when API fails', async () => {
    vi.mocked(window.confirm).mockReturnValue(true)
    const data = { household: { members: [{ id: 'usr_1' }, { id: 'usr_2' }] } }
    const { handleRemoveMember, toast } = makeHandler(
      data,
      vi.fn().mockRejectedValue(new Error('network error')),
    )

    await handleRemoveMember('usr_1')
    await flushPromises()

    expect(data.household.members).toHaveLength(2)
    expect(toast.error).toHaveBeenCalledWith('Kunde inte ta bort medlemmen. Försök igen.')
    expect(toast.success).not.toHaveBeenCalled()
  })
})
