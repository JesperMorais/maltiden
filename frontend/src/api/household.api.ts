/**
 * Household API Service
 * Handles household management, invites, and member operations
 */

import apiClient from './client'
import { USE_MOCKS } from '@/mocks'
import {
  mockGetHousehold,
  mockCreateInvite,
  mockJoinHousehold,
  mockGetMemberStatus,
  mockUpdateMemberStatus,
  mockRemoveMember,
  mockUpdateHousehold
} from '@/mocks/household.mock'
import type { UserRole } from './types/dashboard.types'

export interface HouseholdMember {
  id: string
  name: string
  role: UserRole
  avatarUrl?: string
}

export interface Household {
  id: string
  name: string
  inviteCode: string
  members: HouseholdMember[]
}

export interface MemberStatus {
  id: string
  isEatingToday: boolean
  wantsLunchBox: boolean
}

export interface InviteResponse {
  code: string
  expiresAt: string
}

/**
 * Get current user's household
 */
export async function getHousehold(): Promise<Household> {
  if (USE_MOCKS) {
    return mockGetHousehold()
  }

  const { data } = await apiClient.get<Household>('/households/me')
  return data
}

/**
 * Update the current user's household (name for now)
 */
export async function updateHousehold(payload: { name: string }): Promise<{ ok: boolean }> {
  if (USE_MOCKS) {
    return mockUpdateHousehold(payload)
  }

  const { data } = await apiClient.patch<{ ok: boolean }>('/households/me', payload)
  return data
}

/**
 * Create a new invite code
 */
export async function createInvite(): Promise<InviteResponse> {
  if (USE_MOCKS) {
    return mockCreateInvite()
  }

  const { data } = await apiClient.post<InviteResponse>('/households/invite', {})
  return data
}

/**
 * Join a household with invite code
 */
export async function joinHousehold(code: string): Promise<{ householdId: string }> {
  if (USE_MOCKS) {
    return mockJoinHousehold(code)
  }

  const { data } = await apiClient.post<{ householdId: string }>('/households/join', { code })
  return data
}

/**
 * Get all members' eating status
 */
export async function getMemberStatus(): Promise<{ members: MemberStatus[] }> {
  if (USE_MOCKS) {
    return mockGetMemberStatus()
  }

  const { data } = await apiClient.get<{ members: MemberStatus[] }>('/households/members/status')
  return data
}

/**
 * Update a member's status (eating today, wants lunch box)
 */
export async function updateMemberStatus(
  memberId: string,
  status: Partial<{ isEatingToday: boolean; wantsLunchBox: boolean }>
): Promise<{ ok: boolean }> {
  if (USE_MOCKS) {
    return mockUpdateMemberStatus(memberId, status)
  }

  const { data } = await apiClient.patch<{ ok: boolean }>(`/households/members/${memberId}/status`, status)
  return data
}

/**
 * Remove a member from household
 */
export async function removeMember(memberId: string): Promise<{ ok: boolean }> {
  if (USE_MOCKS) {
    return mockRemoveMember(memberId)
  }

  const { data } = await apiClient.delete<{ ok: boolean }>(`/households/members/${memberId}`)
  return data
}
