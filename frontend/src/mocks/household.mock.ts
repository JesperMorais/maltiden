/**
 * Household API Mock Data
 */

import type { Household, MemberStatus, InviteResponse } from '@/api/household.api'

const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

const mockHousehold: Household = {
  id: 'hh_mock_123',
  name: 'Familjen Andersson',
  inviteCode: 'ABC123',
  members: [
    { id: 'usr_1', name: 'Anna', role: 'owner' },
    { id: 'usr_2', name: 'Erik', role: 'member' },
    { id: 'usr_3', name: 'Lisa', role: 'guest' }
  ]
}

const mockMemberStatuses: MemberStatus[] = [
  { id: 'usr_1', isEatingToday: true, wantsLunchBox: false },
  { id: 'usr_2', isEatingToday: true, wantsLunchBox: true },
  { id: 'usr_3', isEatingToday: false, wantsLunchBox: false }
]

export async function mockGetHousehold(): Promise<Household> {
  await delay(300)
  return { ...mockHousehold }
}

export async function mockUpdateHousehold(payload: { name: string }): Promise<{ ok: boolean }> {
  await delay(300)
  const name = payload.name.trim()
  if (!name) {
    throw { response: { status: 400, data: { error: 'household_name_required' } } }
  }
  if (name.length > 100) {
    throw { response: { status: 400, data: { error: 'household_name_too_long' } } }
  }
  mockHousehold.name = name
  return { ok: true }
}

export async function mockCreateInvite(): Promise<InviteResponse> {
  await delay(400)
  const code = Math.random().toString(36).substring(2, 8).toUpperCase()
  const expiresAt = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString()
  return { code, expiresAt }
}

export async function mockJoinHousehold(code: string): Promise<{ householdId: string }> {
  await delay(500)

  // Simulate invalid code
  if (code.length < 4) {
    throw { response: { status: 400, data: { error: 'invalid_code' } } }
  }

  return { householdId: 'hh_mock_123' }
}

export async function mockGetMemberStatus(): Promise<{ members: MemberStatus[] }> {
  await delay(200)
  return { members: [...mockMemberStatuses] }
}

export async function mockUpdateMemberStatus(
  memberId: string,
  status: Partial<{ isEatingToday: boolean; wantsLunchBox: boolean }>
): Promise<{ ok: boolean }> {
  await delay(300)

  // Update local mock data
  const member = mockMemberStatuses.find(m => m.id === memberId)
  if (member) {
    Object.assign(member, status)
  }

  return { ok: true }
}

export async function mockRemoveMember(memberId: string): Promise<{ ok: boolean }> {
  await delay(400)

  // Simulate can't remove owner
  if (memberId === 'usr_1') {
    throw { response: { status: 403, data: { error: 'cannot_remove' } } }
  }

  return { ok: true }
}
