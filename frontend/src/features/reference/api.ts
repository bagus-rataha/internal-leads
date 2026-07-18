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
export type LeadSourceAdminResponse = components['schemas']['dto.LeadSourceAdminResponse']
export type CreateLeadSourceInput = components['schemas']['dto.CreateLeadSourceInput']
export type UpdateLeadSourceInput = components['schemas']['dto.UpdateLeadSourceInput']
export type ServiceTypeAdminResponse = components['schemas']['dto.ServiceTypeAdminResponse']
export type CreateServiceTypeInput = components['schemas']['dto.CreateServiceTypeInput']
export type UpdateServiceTypeInput = components['schemas']['dto.UpdateServiceTypeInput']

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

async function postJSON<TInput, TOut>(path: string, input: TInput, failMsg: string): Promise<TOut> {
  const res = await apiFetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (!res.ok) throw new Error(failMsg)
  let body: { data?: TOut }
  try {
    body = await res.json()
  } catch {
    throw new Error(failMsg)
  }
  if (!body?.data) throw new Error(failMsg)
  return body.data
}

async function patchJSON<TInput, TOut>(path: string, input: TInput, failMsg: string): Promise<TOut> {
  const res = await apiFetch(path, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (!res.ok) throw new Error(failMsg)
  let body: { data?: TOut }
  try {
    body = await res.json()
  } catch {
    throw new Error(failMsg)
  }
  if (!body?.data) throw new Error(failMsg)
  return body.data
}

export function fetchLeadSourcesAdmin(): Promise<LeadSourceAdminResponse[]> {
  return getList<LeadSourceAdminResponse>('/api/v1/lead-sources', 'Failed to load lead sources')
}

export function createLeadSource(input: CreateLeadSourceInput): Promise<LeadSourceAdminResponse> {
  return postJSON('/api/v1/lead-sources', input, 'Failed to save lead source')
}

export function updateLeadSource(id: string, input: UpdateLeadSourceInput): Promise<LeadSourceAdminResponse> {
  return patchJSON(`/api/v1/lead-sources/${id}`, input, 'Failed to save lead source')
}

export function fetchServiceTypesAdmin(): Promise<ServiceTypeAdminResponse[]> {
  return getList<ServiceTypeAdminResponse>('/api/v1/service-types', 'Failed to load service types')
}

export function createServiceType(input: CreateServiceTypeInput): Promise<ServiceTypeAdminResponse> {
  return postJSON('/api/v1/service-types', input, 'Failed to save service type')
}

export function updateServiceType(id: string, input: UpdateServiceTypeInput): Promise<ServiceTypeAdminResponse> {
  return patchJSON(`/api/v1/service-types/${id}`, input, 'Failed to save service type')
}
