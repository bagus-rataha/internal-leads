// User data-fetching. Envelope pattern matches every other features/*/api.ts
// (check res.ok, guard res.json, return body.data).
import { apiFetch } from '@/api/client'
import type { components } from '@/api/types'

export type UserResponse = components['schemas']['dto.UserResponse']

export async function fetchSalesUsers(): Promise<UserResponse[]> {
  const res = await apiFetch('/api/v1/users?role=SALES')
  if (!res.ok) {
    throw new Error('Failed to load sales users')
  }
  let body: { data?: UserResponse[] }
  try {
    body = await res.json()
  } catch {
    throw new Error('Failed to load sales users')
  }
  return body?.data ?? []
}

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

export async function createUser(input: CreateUserInput): Promise<UserResponse> {
  const res = await apiFetch('/api/v1/users', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
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
