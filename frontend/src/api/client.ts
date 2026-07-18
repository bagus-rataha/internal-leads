// Fetch wrapper: attaches the in-memory access token, always sends
// credentials so the httpOnly refresh cookie goes along, and on a 401
// runs a single-flight refresh-then-retry before giving up.
//
// Access token lives ONLY in this module-level variable — never
// localStorage/sessionStorage (XSS risk). A page reload wipes it; the
// auth context's bootstrap effect re-derives it via
// /api/v1/auth/refresh on load.

let accessToken: string | null = null

export function setAccessToken(token: string | null) {
  accessToken = token
}

export function getAccessToken(): string | null {
  return accessToken
}

type AuthFailureCallback = () => void

const authFailureCallbacks: AuthFailureCallback[] = []

// The auth context subscribes here to redirect to /login on a hard
// auth failure — this module has no router access of its own.
export function onAuthFailure(callback: AuthFailureCallback) {
  authFailureCallbacks.push(callback)
}

// Single-flight guard: the backend rotates and invalidates the old
// refresh token on every call, so two concurrent 401s each triggering
// their own refresh would race and log the user out. All concurrent
// 401s await this same promise instead, and it's reset to null once
// the refresh settles so the next 401 storm can trigger a fresh one.
let refreshPromise: Promise<boolean> | null = null

// Fires auth-failure callbacks at most once per failed-refresh event
// (see doRefresh) rather than once per concurrent caller awaiting the
// shared refreshPromise.
function failAuth(): false {
  setAccessToken(null)
  for (const callback of authFailureCallbacks) callback()
  return false
}

async function doRefresh(): Promise<boolean> {
  try {
    const res = await fetch('/api/v1/auth/refresh', {
      method: 'POST',
      credentials: 'include',
    })
    if (!res.ok) return failAuth()

    const body = await res.json()
    const token = body?.data?.access_token
    if (!token) return failAuth()

    setAccessToken(token)
    return true
  } catch {
    return failAuth()
  }
}

export async function apiFetch(
  path: string,
  options: RequestInit = {},
  isRetry = false,
): Promise<Response> {
  const headers = new Headers(options.headers)
  if (accessToken) {
    headers.set('Authorization', `Bearer ${accessToken}`)
  }

  const response = await fetch(path, {
    ...options,
    headers,
    credentials: 'include',
  })

  if (response.status !== 401) {
    return response
  }

  // The retried request itself came back 401 (e.g. the user was
  // deactivated between refresh and retry) — force logout instead of
  // leaving a stale-but-still-set access token in place.
  if (isRetry) {
    failAuth()
    return response
  }

  if (!refreshPromise) {
    refreshPromise = doRefresh().finally(() => {
      refreshPromise = null
    })
  }
  const refreshed = await refreshPromise

  if (refreshed) {
    return apiFetch(path, options, true)
  }

  return response
}
