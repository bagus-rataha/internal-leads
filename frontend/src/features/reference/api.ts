// Reference data fetchers for the address cascade + service types. Same
// fetch/envelope contract as features/lead/api.ts (check res.ok, guard
// res.json, return body.data). These refs are static seed data.
import { apiFetch } from '@/api/client'
import type { components } from '@/api/types'

export type ProvinceResponse = components['schemas']['dto.ProvinceResponse']
export type CityResponse = components['schemas']['dto.CityResponse']
export type DistrictResponse = components['schemas']['dto.DistrictResponse']
export type VillageResponse = components['schemas']['dto.VillageResponse']
export type ServiceTypeResponse = components['schemas']['dto.ServiceTypeResponse']

async function getList<T>(path: string, failMsg: string): Promise<T[]> {
  const res = await apiFetch(path)
  if (!res.ok) throw new Error(failMsg)
  let body: { data?: T[] }
  try {
    body = await res.json()
  } catch {
    throw new Error(failMsg)
  }
  return body?.data ?? []
}

export function fetchProvinces(): Promise<ProvinceResponse[]> {
  return getList<ProvinceResponse>('/api/v1/refs/provinces', 'Failed to load provinces')
}

export function fetchCities(provinceId: number): Promise<CityResponse[]> {
  return getList<CityResponse>(`/api/v1/refs/cities?province_id=${provinceId}`, 'Failed to load cities')
}

export function fetchDistricts(cityId: number): Promise<DistrictResponse[]> {
  return getList<DistrictResponse>(`/api/v1/refs/districts?city_id=${cityId}`, 'Failed to load districts')
}

// No `q` — the caller shows all villages for the district as plain options
// (a kecamatan holds few villages; the backend caps the result at 50).
export function fetchVillages(districtId: number): Promise<VillageResponse[]> {
  return getList<VillageResponse>(`/api/v1/refs/villages?district_id=${districtId}`, 'Failed to load villages')
}

export function fetchServiceTypes(): Promise<ServiceTypeResponse[]> {
  return getList<ServiceTypeResponse>('/api/v1/refs/service-types', 'Failed to load service types')
}
