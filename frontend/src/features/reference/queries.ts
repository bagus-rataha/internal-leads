// TanStack Query hooks for reference data. Long staleTime — this is static
// seed data (provinces/cities/districts/villages/service types), refetching
// it during a session is wasted work. Cascading hooks stay disabled until
// their parent id exists.
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fetchLeadSources } from '@/features/lead/api'
import {
  fetchProvinces,
  fetchCities,
  fetchDistricts,
  fetchVillages,
  fetchServiceTypes,
  fetchLeadSourcesAdmin,
  createLeadSource,
  updateLeadSource,
  fetchServiceTypesAdmin,
  createServiceType,
  updateServiceType,
  type CreateLeadSourceInput,
  type UpdateLeadSourceInput,
  type CreateServiceTypeInput,
  type UpdateServiceTypeInput,
} from './api'

const STATIC = { staleTime: Infinity, gcTime: Infinity } as const

export function useProvinces() {
  return useQuery({ queryKey: ['refs', 'provinces'], queryFn: fetchProvinces, ...STATIC })
}

export function useCities(provinceId?: number) {
  return useQuery({
    queryKey: ['refs', 'cities', provinceId],
    queryFn: () => fetchCities(provinceId as number),
    enabled: provinceId != null,
    ...STATIC,
  })
}

export function useDistricts(cityId?: number) {
  return useQuery({
    queryKey: ['refs', 'districts', cityId],
    queryFn: () => fetchDistricts(cityId as number),
    enabled: cityId != null,
    ...STATIC,
  })
}

export function useVillages(districtId?: number) {
  return useQuery({
    queryKey: ['refs', 'villages', districtId],
    queryFn: () => fetchVillages(districtId as number),
    enabled: districtId != null,
    ...STATIC,
  })
}

export function useServiceTypes() {
  return useQuery({ queryKey: ['refs', 'service-types'], queryFn: fetchServiceTypes, ...STATIC })
}

export function useLeadSources() {
  return useQuery({ queryKey: ['refs', 'lead-sources'], queryFn: fetchLeadSources, ...STATIC })
}

export function useLeadSourcesAdmin() {
  return useQuery({ queryKey: ['lead-sources', 'admin'], queryFn: fetchLeadSourcesAdmin })
}

export function useCreateLeadSource() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateLeadSourceInput) => createLeadSource(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['lead-sources', 'admin'] })
      queryClient.invalidateQueries({ queryKey: ['refs', 'lead-sources'] })
    },
  })
}

export function useUpdateLeadSource() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateLeadSourceInput }) =>
      updateLeadSource(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['lead-sources', 'admin'] })
      queryClient.invalidateQueries({ queryKey: ['refs', 'lead-sources'] })
    },
  })
}

export function useServiceTypesAdmin() {
  return useQuery({ queryKey: ['service-types', 'admin'], queryFn: fetchServiceTypesAdmin })
}

export function useCreateServiceType() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateServiceTypeInput) => createServiceType(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['service-types', 'admin'] })
      queryClient.invalidateQueries({ queryKey: ['refs', 'service-types'] })
    },
  })
}

export function useUpdateServiceType() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateServiceTypeInput }) =>
      updateServiceType(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['service-types', 'admin'] })
      queryClient.invalidateQueries({ queryKey: ['refs', 'service-types'] })
    },
  })
}
