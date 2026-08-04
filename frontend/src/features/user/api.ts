// User data-fetching. Envelope pattern matches every other features/*/api.ts
// (check res.ok, guard res.json, return body.data).
import { apiFetch } from '@/api/client'
import type { components } from '@/api/types'

export type UserResponse = components['schemas']['dto.UserResponse']

// Roles eligible to own a lead: both SALES and LEADER can be a lead's
// owner_id (see resolveOwnerID's role rule), so any owner-picking dropdown
// (dashboard filters, lead list, reassign) fetches this combined roster
// rather than SALES alone.
export const LEAD_OWNER_ROLES = 'SALES,LEADER'

export type CreateUserInput = components['schemas']['dto.CreateUserInput']
export type UpdateUserInput = components['schemas']['dto.UpdateUserInput']
export type ResetPasswordInput = components['schemas']['dto.ResetPasswordInput']

export async function fetchUsers(role?: string): Promise<UserResponse[]> {
  const qs = role ? `?role=${role}` : ''
  const res = await apiFetch(`/api/v1/users${qs}`)
  if (!res.ok) {
    throw new Error('Failed to load users')
  }
  let body: { data?: UserResponse[] }
  try {
    body = await res.json()
  } catch {
    throw new Error('Failed to load users')
  }
  return body?.data ?? []
}

// A 409 means the submitted email already belongs to another account.
// existingUserActive tells the caller whether that account could be
// reactivated instead of creating a duplicate.
export class EmailTakenError extends Error {
  existingUserId: string
  existingUserName: string
  existingUserActive: boolean
  constructor(existingUserId: string, existingUserName: string, existingUserActive: boolean) {
    super('Email already registered')
    this.existingUserId = existingUserId
    this.existingUserName = existingUserName
    this.existingUserActive = existingUserActive
  }
}

export async function createUser(input: CreateUserInput): Promise<UserResponse> {
  const res = await apiFetch('/api/v1/users', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (res.status === 409) {
    let data: {
      existing_user_id?: string
      existing_user_name?: string
      existing_user_active?: boolean
    } = {}
    try {
      const body = await res.json()
      data = body?.data ?? {}
    } catch {
      // fall through with empty data
    }
    throw new EmailTakenError(
      data.existing_user_id ?? '',
      data.existing_user_name ?? '',
      data.existing_user_active ?? false
    )
  }
  if (!res.ok) {
    let message = 'Failed to save user'
    try {
      const body = await res.json()
      if (body?.message) message = body.message
    } catch {
      // keep default message
    }
    throw new Error(message)
  }
  let body: { data?: UserResponse }
  try {
    body = await res.json()
  } catch {
    throw new Error('Failed to save user')
  }
  if (!body?.data) {
    throw new Error('Failed to save user')
  }
  return body.data
}

export async function updateUser(id: string, input: UpdateUserInput): Promise<UserResponse> {
  const res = await apiFetch(`/api/v1/users/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (!res.ok) {
    throw new Error('Failed to save user')
  }
  let body: { data?: UserResponse }
  try {
    body = await res.json()
  } catch {
    throw new Error('Failed to save user')
  }
  if (!body?.data) {
    throw new Error('Failed to save user')
  }
  return body.data
}

export async function resetPassword(id: string, input: ResetPasswordInput): Promise<void> {
  const res = await apiFetch(`/api/v1/users/${id}/reset-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (!res.ok) {
    throw new Error('Failed to reset password')
  }
}

export type DeactivateUserInput = components['schemas']['dto.DeactivateUserInput']

// A 422 with active_lead_count means the user still owns active leads and a
// reassign_to_user_id is required - callers branch on this via `.status`.
export class ActiveLeadsError extends Error {
  activeLeadCount: number
  constructor(activeLeadCount: number) {
    super('User has active leads')
    this.activeLeadCount = activeLeadCount
  }
}

export async function deactivateUser(id: string, input: DeactivateUserInput): Promise<void> {
  const res = await apiFetch(`/api/v1/users/${id}/deactivate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (res.status === 422) {
    let count = 0
    try {
      const body = await res.json()
      count = body?.data?.active_lead_count ?? 0
    } catch {
      // fall through with count 0
    }
    throw new ActiveLeadsError(count)
  }
  if (!res.ok) {
    throw new Error('Failed to deactivate user')
  }
}
