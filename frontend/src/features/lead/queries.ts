// TanStack Query hooks for lead detail data: fetches plus mutations that
// each just need "refetch detail after success". List's own useLeads.ts
// stays hand-rolled; not retrofitted here.
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  fetchLeadDetail,
  fetchFollowUps,
  createFollowUp,
  updateLeadStatus,
  reassignOwner,
  fetchSalesUsers,
  createLead,
  updateLead,
  type UpdateLeadStatusInput,
  type CreateLeadInput,
  type UpdateLeadInput,
} from './api'

export function useLeadDetail(code: string) {
  return useQuery({ queryKey: ['lead', code], queryFn: () => fetchLeadDetail(code) })
}

export function useFollowUps(code: string) {
  return useQuery({ queryKey: ['lead', code, 'followups'], queryFn: () => fetchFollowUps(code) })
}

// enabled: false until the reassign UI is actually opened - no reason to
// fetch the sales roster before anyone asks to reassign.
export function useSalesRoster(enabled: boolean) {
  return useQuery({ queryKey: ['users', 'SALES'], queryFn: fetchSalesUsers, enabled })
}

export function useCreateFollowUp(code: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (note: string) => createFollowUp(code, note),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['lead', code] })
      queryClient.invalidateQueries({ queryKey: ['lead', code, 'followups'] })
    },
  })
}

export function useUpdateLeadStatus(code: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: UpdateLeadStatusInput) => updateLeadStatus(code, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['lead', code] })
    },
  })
}

export function useReassignOwner(code: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (ownerId: string) => reassignOwner(code, ownerId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['lead', code] })
    },
  })
}

export function useCreateLead() {
  return useMutation({
    mutationFn: (input: CreateLeadInput) => createLead(input),
  })
}

export function useUpdateLead(code: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: UpdateLeadInput) => updateLead(code, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['lead', code] })
    },
  })
}
