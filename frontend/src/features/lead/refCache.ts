// Session-lifetime cache for lead sources: reference data that's static
// per the frontend's stated rule for this kind of dropdown data (cache
// for the session, don't refetch on every filter-panel mount). No
// TanStack Query in this app — a memoized module-level promise is the
// whole mechanism needed.
import { fetchLeadSources, type LeadSourceResponse } from './api'

// ponytail: plain in-memory memo, no TTL/invalidation — reference data
// only changes via backend seed/admin action, and a page reload already
// clears this along with the rest of app state.
let cached: Promise<LeadSourceResponse[]> | null = null

export function getLeadSources(): Promise<LeadSourceResponse[]> {
  if (!cached) {
    cached = fetchLeadSources().catch((err) => {
      // Don't poison the cache with a failed request — let the next
      // caller retry instead of being stuck with a rejected promise for
      // the rest of the session.
      cached = null
      throw err
    })
  }
  return cached
}
