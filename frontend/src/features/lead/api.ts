// Data-fetching for the lead list: query building + envelope handling.
// Mirrors the fetch/envelope pattern already used in the auth context
// (check res.ok before parsing, guard res.json() in try/catch).
import { apiFetch, getAccessToken } from '@/api/client'
import type { components } from '@/api/types'

export type { TeamResponse } from '@/features/team/api'
export type { UserResponse } from '@/features/user/api'
export { fetchTeams } from '@/features/team/api'
export { fetchSalesUsers } from '@/features/user/api'

export type PaginatedLeadResponse = components['schemas']['dto.PaginatedLeadResponse']
export type LeadSourceResponse = components['schemas']['dto.LeadSourceResponse']
export type LeadResponse = components['schemas']['dto.LeadResponse']

// The list endpoint's swagger annotation never declared query params, so
// there's no generated type for them — hand-written here from the
// handler's actual c.Query(...) keys.
export type LeadStatus = 'BARU' | 'FOLLOW_UP' | 'HANDOFF_ODOO' | 'LOST'
export type LeadSort = 'code' | '-code' | 'company_name' | '-company_name'

export interface LeadListParams {
  q?: string
  status?: LeadStatus | ''
  source_id?: string
  team_id?: string
  owner_id?: string
  province_id?: string
  city_id?: string
  date_from?: string
  date_to?: string
  follow_up_from?: string
  follow_up_to?: string
  stale?: 'true' | 'false'
  sort?: LeadSort | ''
  page?: number
  limit?: number
}

// Only real values make it into the query string — the backend treats an
// omitted key differently from an explicit empty one (e.g. `status=`
// would be a literal, meaningless filter rather than "no filter").
export function buildLeadQuery(params: LeadListParams): string {
  const qs = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === '') continue
    qs.set(key, String(value))
  }
  return qs.toString()
}

export async function fetchLeads(params: LeadListParams): Promise<PaginatedLeadResponse> {
  const qs = buildLeadQuery(params)
  const res = await apiFetch('/api/v1/leads?' + qs)
  if (!res.ok) {
    throw new Error('Failed to load leads')
  }

  let body: { data?: PaginatedLeadResponse }
  try {
    body = await res.json()
  } catch {
    throw new Error('Failed to load leads')
  }
  if (!body?.data) {
    throw new Error('Failed to load leads')
  }
  return body.data
}

export async function fetchLeadSources(): Promise<LeadSourceResponse[]> {
  const res = await apiFetch('/api/v1/refs/lead-sources')
  if (!res.ok) {
    throw new Error('Failed to load lead sources')
  }

  let body: { data?: LeadSourceResponse[] }
  try {
    body = await res.json()
  } catch {
    throw new Error('Failed to load lead sources')
  }
  return body?.data ?? []
}

export type LeadDetailResponse = components['schemas']['dto.LeadDetailResponse']
export type FollowUpResponse = components['schemas']['dto.FollowUpResponse']

export async function fetchLeadDetail(code: string): Promise<LeadDetailResponse> {
  const res = await apiFetch(`/api/v1/leads/${code}`)
  if (!res.ok) {
    throw new Error('Failed to load lead')
  }
  let body: { data?: LeadDetailResponse }
  try {
    body = await res.json()
  } catch {
    throw new Error('Failed to load lead')
  }
  if (!body?.data) {
    throw new Error('Failed to load lead')
  }
  return body.data
}

export async function fetchFollowUps(code: string): Promise<FollowUpResponse[]> {
  const res = await apiFetch(`/api/v1/leads/${code}/followups`)
  if (!res.ok) {
    throw new Error('Failed to load follow-ups')
  }
  let body: { data?: FollowUpResponse[] }
  try {
    body = await res.json()
  } catch {
    throw new Error('Failed to load follow-ups')
  }
  return body?.data ?? []
}

export async function createFollowUp(code: string, note: string): Promise<FollowUpResponse> {
  const res = await apiFetch(`/api/v1/leads/${code}/followups`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ note }),
  })
  if (!res.ok) {
    throw new Error('Failed to save follow-up')
  }
  let body: { data?: FollowUpResponse }
  try {
    body = await res.json()
  } catch {
    throw new Error('Failed to save follow-up')
  }
  if (!body?.data) {
    throw new Error('Failed to save follow-up')
  }
  return body.data
}

export interface UpdateLeadStatusInput {
  status: 'HANDOFF_ODOO' | 'LOST'
  lost_reason?: string
}

export async function updateLeadStatus(code: string, input: UpdateLeadStatusInput): Promise<LeadResponse> {
  const res = await apiFetch(`/api/v1/leads/${code}/status`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (!res.ok) {
    throw new Error('Failed to update lead status')
  }
  let body: { data?: LeadResponse }
  try {
    body = await res.json()
  } catch {
    throw new Error('Failed to update lead status')
  }
  if (!body?.data) {
    throw new Error('Failed to update lead status')
  }
  return body.data
}

export async function reassignOwner(code: string, ownerId: string): Promise<LeadResponse> {
  const res = await apiFetch(`/api/v1/leads/${code}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ owner_id: ownerId }),
  })
  if (!res.ok) {
    throw new Error('Failed to reassign owner')
  }
  let body: { data?: LeadResponse }
  try {
    body = await res.json()
  } catch {
    throw new Error('Failed to reassign owner')
  }
  if (!body?.data) {
    throw new Error('Failed to reassign owner')
  }
  return body.data
}

export type CreateLeadInput = components['schemas']['dto.CreateLeadInput']
export type UpdateLeadInput = components['schemas']['dto.UpdateLeadInput']

export async function createLead(input: CreateLeadInput): Promise<LeadResponse> {
  const res = await apiFetch('/api/v1/leads', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (!res.ok) {
    throw new Error('Failed to save lead')
  }
  let body: { data?: LeadResponse }
  try {
    body = await res.json()
  } catch {
    throw new Error('Failed to save lead')
  }
  if (!body?.data) {
    throw new Error('Failed to save lead')
  }
  return body.data
}

export async function updateLead(code: string, input: UpdateLeadInput): Promise<LeadResponse> {
  const res = await apiFetch(`/api/v1/leads/${code}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (!res.ok) {
    throw new Error('Failed to save changes')
  }
  let body: { data?: LeadResponse }
  try {
    body = await res.json()
  } catch {
    throw new Error('Failed to save changes')
  }
  if (!body?.data) {
    throw new Error('Failed to save changes')
  }
  return body.data
}

export class ExportTooManyRowsError extends Error {
  rowCount: number
  constructor(rowCount: number) {
    super('too many rows')
    this.rowCount = rowCount
  }
}

// Uses a raw fetch, not apiFetch: a wrong export password also returns 401,
// and apiFetch treats any 401 as an expired session - on retry-after-refresh
// it would still be 401 (the password is still wrong) and apiFetch force-
// logs-out the whole app. This endpoint's 401 must stay a plain request
// failure, never a session event.
export async function exportLeads(params: LeadListParams, password: string): Promise<{ blob: Blob; filename: string }> {
  const qs = buildLeadQuery(params)
  const res = await fetch('/api/v1/leads/export?' + qs, {
    headers: {
      Authorization: `Bearer ${getAccessToken() ?? ''}`,
      'X-Export-Password': password,
    },
  })

  if (res.status === 401) {
    throw new Error('invalid credentials')
  }
  if (res.status === 422) {
    let rowCount = 0
    try {
      const body = await res.json()
      rowCount = body?.data?.row_count ?? 0
    } catch {
      // rowCount stays 0 - the toast still explains to narrow filters
    }
    throw new ExportTooManyRowsError(rowCount)
  }
  if (!res.ok) {
    throw new Error('Failed to export leads')
  }

  const disposition = res.headers.get('Content-Disposition') ?? ''
  const filename = disposition.match(/filename="([^"]+)"/)?.[1] ?? 'leads-export.xlsx'
  const blob = await res.blob()
  return { blob, filename }
}

export function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}
