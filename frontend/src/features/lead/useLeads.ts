// Hand-rolled data hook for the lead list — no TanStack Query in this
// app, just useState/useEffect matching the rest of the codebase's style.
import { useEffect, useState } from 'react'
import { fetchLeads, type LeadListParams, type PaginatedLeadResponse } from './api'

interface UseLeadsResult {
  data: PaginatedLeadResponse | null
  loading: boolean
  error: string | null
}

export function useLeads(params: LeadListParams): UseLeadsResult {
  const [data, setData] = useState<PaginatedLeadResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Serialized so a fresh object literal from the caller on every render
  // doesn't retrigger this effect — only an actual value change should
  // cause a refetch.
  const paramsKey = JSON.stringify(params)

  useEffect(() => {
    // Race guard: if params change again before this fetch resolves, the
    // stale response must not overwrite state meant for the newer params.
    let cancelled = false

    setLoading(true)
    setError(null)

    fetchLeads(params)
      .then((result) => {
        if (cancelled) return
        setData(result)
      })
      .catch((err) => {
        if (cancelled) return
        setError(err instanceof Error ? err.message : 'Failed to load leads')
      })
      .finally(() => {
        if (cancelled) return
        setLoading(false)
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- paramsKey is the stable dep
  }, [paramsKey])

  return { data, loading, error }
}
