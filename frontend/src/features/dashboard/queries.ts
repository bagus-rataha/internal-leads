import { useQuery } from '@tanstack/react-query'
import {
  fetchDashboardSummary,
  fetchDashboardActivity,
  fetchDashboardStaleLeads,
  fetchDashboardSalesActivity,
  fetchDashboardSegments,
  type DashboardQueryParams,
} from './api'

export function useDashboardSummary(params: DashboardQueryParams) {
  return useQuery({ queryKey: ['dashboard', 'summary', params], queryFn: () => fetchDashboardSummary(params) })
}

export function useDashboardActivity(params: DashboardQueryParams) {
  return useQuery({ queryKey: ['dashboard', 'activity', params], queryFn: () => fetchDashboardActivity(params) })
}

export function useDashboardStaleLeads(params: DashboardQueryParams) {
  return useQuery({ queryKey: ['dashboard', 'stale-leads', params], queryFn: () => fetchDashboardStaleLeads(params) })
}

// enabled: false for SALES (the endpoint 403s for them anyway) - mirrors
// LeadListPage's existing "skip the fetch for a role that won't see it".
export function useDashboardSalesActivity(params: DashboardQueryParams, enabled: boolean) {
  return useQuery({
    queryKey: ['dashboard', 'sales-activity', params],
    queryFn: () => fetchDashboardSalesActivity(params),
    enabled,
  })
}

export function useDashboardSegments(params: DashboardQueryParams) {
  return useQuery({ queryKey: ['dashboard', 'segments', params], queryFn: () => fetchDashboardSegments(params) })
}
