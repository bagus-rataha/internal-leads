// Sales team data-fetching. Envelope pattern matches every other features/*/api.ts
// (check res.ok, guard res.json, return body.data).
import { apiFetch } from '@/api/client'
import type { components } from '@/api/types'

export type TeamResponse = components['schemas']['dto.TeamResponse']

export async function fetchTeams(): Promise<TeamResponse[]> {
  const res = await apiFetch('/api/v1/teams')
  if (!res.ok) {
    throw new Error('Failed to load teams')
  }
  let body: { data?: TeamResponse[] }
  try {
    body = await res.json()
  } catch {
    throw new Error('Failed to load teams')
  }
  return body?.data ?? []
}

export type CreateTeamInput = components['schemas']['dto.CreateTeamInput']
export type UpdateTeamInput = components['schemas']['dto.UpdateTeamInput']

export async function createTeam(input: CreateTeamInput): Promise<TeamResponse> {
  const res = await apiFetch('/api/v1/teams', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (!res.ok) {
    throw new Error('Failed to save team')
  }
  let body: { data?: TeamResponse }
  try {
    body = await res.json()
  } catch {
    throw new Error('Failed to save team')
  }
  if (!body?.data) {
    throw new Error('Failed to save team')
  }
  return body.data
}

export async function updateTeam(id: string, input: UpdateTeamInput): Promise<TeamResponse> {
  const res = await apiFetch(`/api/v1/teams/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (!res.ok) {
    throw new Error('Failed to save team')
  }
  let body: { data?: TeamResponse }
  try {
    body = await res.json()
  } catch {
    throw new Error('Failed to save team')
  }
  if (!body?.data) {
    throw new Error('Failed to save team')
  }
  return body.data
}
