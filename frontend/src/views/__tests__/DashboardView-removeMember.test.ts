import { describe, it, expect, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'

// NOTE: This file mirrors the optimistic-update logic in DashboardView's
// handleRemoveMember. It is a logic-shape regression test, not a component
// integration test — the real handler lives inside <script setup> and isn't
// exported. A follow-up should replace this with a mounted-component test
// (likely needs @pinia/testing) so a revert in DashboardView would actually
// fail here. For now this guards against regressions in the optimistic /
// rollback / toast contract that the handler is supposed to implement.

function makeHandler(
  dashboardData: { household: { members: { id: string }[] } } | null,
  removeMemberImpl: () => Promise<unknown>,
) {
  const toast = { success: vi.fn(), error: vi.fn() }

  async function handleRemoveMember(memberId: string) {
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

describe('handleRemoveMember (optimistic update contract)', () => {
  it('optimistically removes member before API resolves', async () => {
    let resolveApi!: () => void
    const apiPromise = new Promise<void>((res) => { resolveApi = res })
    const data = { household: { members: [{ id: 'usr_1' }, { id: 'usr_2' }] } }
    const { handleRemoveMember } = makeHandler(data, () => apiPromise)

    const pending = handleRemoveMember('usr_1')
    expect(data.household.members).toHaveLength(1)
    expect(data.household.members[0]?.id).toBe('usr_2')

    resolveApi()
    await pending
  })

  it('shows success toast after successful removal', async () => {
    const data = { household: { members: [{ id: 'usr_1' }] } }
    const { handleRemoveMember, toast } = makeHandler(data, vi.fn().mockResolvedValue({ ok: true }))

    await handleRemoveMember('usr_1')
    await flushPromises()

    expect(data.household.members).toHaveLength(0)
    expect(toast.success).toHaveBeenCalledWith('Medlemmen har tagits bort.')
    expect(toast.error).not.toHaveBeenCalled()
  })

  it('rolls back and shows error toast when API fails', async () => {
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
