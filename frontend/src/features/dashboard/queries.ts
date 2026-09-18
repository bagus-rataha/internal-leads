import { useQuery } from '@tanstack/react-query'
import {
  fetchDashboardSummary,
  fetchDashboardActivity,
  fetchDashboardStaleLeads,
  fetchDashboardSalesActivity,
  fetchDashboardSegments,
  type DashboardQueryParams,
} from './api'

export function useDashboardSummary(params: DashboardQueryParams, enabled = true) {
  return useQuery({ queryKey: ['dashboard', 'summary', params], queryFn: () => fetchDashboardSummary(params), enabled })
}

export function useDashboardActivity(params: DashboardQueryParams, enabled = true) {
  return useQuery({ queryKey: ['dashboard', 'activity', params], queryFn: () => fetchDashboardActivity(params), enabled })
}

export function useDashboardStaleLeads(params: DashboardQueryParams, enabled = true) {
  return useQuery({ queryKey: ['dashboard', 'stale-leads', params], queryFn: () => fetchDashboardStaleLeads(params), enabled })
}

// enabled: false for SALES (the endpoint 403s for them anyway) - mirrors
// LeadListPage's existing "skip the fetch for a role that won't see it".
// The caller (DashboardPage) ANDs this with its own range-validity check.
export function useDashboardSalesActivity(params: DashboardQueryParams, enabled: boolean) {
  return useQuery({
    queryKey: ['dashboard', 'sales-activity', params],
    queryFn: () => fetchDashboardSalesActivity(params),
    enabled,
  })
}

export function useDashboardSegments(params: DashboardQueryParams, enabled = true) {
  return useQuery({ queryKey: ['dashboard', 'segments', params], queryFn: () => fetchDashboardSegments(params), enabled })
}
