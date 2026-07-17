// Data-fetching for the lead list: query building + envelope handling.
// Mirrors the fetch/envelope pattern already used in the auth context
// (check res.ok before parsing, guard res.json() in try/catch).
import { apiFetch } from '@/api/client'
import type { components } from '@/api/types'

export type PaginatedLeadResponse = components['schemas']['dto.PaginatedLeadResponse']
export type LeadSourceResponse = components['schemas']['dto.LeadSourceResponse']

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
