// TanStack Query hooks for reference data. Long staleTime — this is static
// seed data (provinces/cities/districts/villages/service types), refetching
// it during a session is wasted work. Cascading hooks stay disabled until
// their parent id exists.
import { useQuery } from '@tanstack/react-query'
import {
  fetchProvinces,
  fetchCities,
  fetchDistricts,
  fetchVillages,
  fetchServiceTypes,
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
