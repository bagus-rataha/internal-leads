// Dashboard data-fetching: 5 endpoints sharing one
// {date_from,date_to,team_id,owner_id} query shape and one envelope-parsing
// helper - unlike other features/*/api.ts files, these 5 functions are
// near-identical enough (same file, same shape) that one shared helper is
// less code than 5 copy-pasted try/catch blocks, without introducing a
// cross-file abstraction.
import { apiFetch } from '@/api/client'
import type { components } from '@/api/types'

export type DashboardSummaryResponse = components['schemas']['dto.DashboardSummaryResponse']
export type DashboardActivityResponse = components['schemas']['dto.DashboardActivityResponse']
export type DashboardStaleLeadsResponse = components['schemas']['dto.DashboardStaleLeadsResponse']
export type DashboardSalesActivityResponse = components['schemas']['dto.DashboardSalesActivityResponse']
export type DashboardSegmentsResponse = components['schemas']['dto.DashboardSegmentsResponse']

export interface DashboardQueryParams {
  date_from: string
  date_to: string
  team_id?: string
  owner_id?: string
  status?: string
}

function buildDashboardQuery(params: DashboardQueryParams): string {
  const qs = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === '') continue
    qs.set(key, String(value))
  }
  return qs.toString()
}

async function fetchDashboard<T>(path: string, params: DashboardQueryParams, errorMessage: string): Promise<T> {
  const res = await apiFetch(`/api/v1/dashboard/${path}?${buildDashboardQuery(params)}`)
  if (!res.ok) {
    throw new Error(errorMessage)
  }
  let body: { data?: T }
  try {
    body = await res.json()
  } catch {
    throw new Error(errorMessage)
  }
  if (!body?.data) {
    throw new Error(errorMessage)
  }
  return body.data
}

export function fetchDashboardSummary(params: DashboardQueryParams) {
  return fetchDashboard<DashboardSummaryResponse>('summary', params, 'Failed to load dashboard summary')
}
export function fetchDashboardActivity(params: DashboardQueryParams) {
  return fetchDashboard<DashboardActivityResponse>('activity', params, 'Failed to load dashboard activity')
}
export function fetchDashboardStaleLeads(params: DashboardQueryParams) {
  return fetchDashboard<DashboardStaleLeadsResponse>('stale-leads', params, 'Failed to load stale leads')
}
export function fetchDashboardSalesActivity(params: DashboardQueryParams) {
  return fetchDashboard<DashboardSalesActivityResponse>('sales-activity', params, 'Failed to load sales activity')
}
export function fetchDashboardSegments(params: DashboardQueryParams) {
  return fetchDashboard<DashboardSegmentsResponse>('segments', params, 'Failed to load dashboard segments')
}
