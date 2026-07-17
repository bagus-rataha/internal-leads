// Auth state for the whole app: current user, bootstrap-in-progress flag,
// login/logout. Access token itself lives in api/client.ts (in-memory
// only) — this context only calls setAccessToken, never stores the
// token itself.
import { createContext, useContext, useEffect, useRef, useState, type ReactNode } from 'react'
import { apiFetch, setAccessToken, onAuthFailure } from '@/api/client'
import type { components } from '@/api/types'

type UserResponse = components['schemas']['dto.UserResponse']
type TokenResponse = components['schemas']['dto.TokenResponse']

interface AuthContextValue {
  user: UserResponse | null
  isLoading: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<UserResponse | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  const bootstrapped = useRef(false)

  useEffect(() => {
    // StrictMode dev double-invoke guard: without this, bootstrap() and
    // onAuthFailure() each fire twice back-to-back (no cleanup on this
    // effect), racing two /auth/refresh calls against the same
    // not-yet-rotated cookie. Guard ensures both run exactly once per mount.
    if (bootstrapped.current) return
    bootstrapped.current = true

    // A failed background refresh (e.g. a 401 mid-session whose retry-after-
    // refresh also fails) fires this from client.ts — clear user so
    // RouteGuard redirects, not just on bootstrap.
    onAuthFailure(() => setUser(null))

    // Bootstrap: a page load/reload has no access token in memory (it never
    // survives a reload by design). Try the refresh cookie directly — NOT
    // through apiFetch, which only refreshes in reaction to a 401 from some
    // other request; there is no prior request to react to here.
    async function bootstrap() {
      try {
        const res = await fetch('/api/v1/auth/refresh', {
          method: 'POST',
          credentials: 'include',
        })
        if (!res.ok) {
          setUser(null)
          return
        }

        const body = await res.json()
        const token: string | undefined = body?.data?.access_token
        if (!token) {
          setUser(null)
          return
        }
        setAccessToken(token)

        const meRes = await apiFetch('/api/v1/users/me')
        if (!meRes.ok) {
          setUser(null)
          return
        }
        const meBody = await meRes.json()
        setUser(meBody?.data ?? null)
      } catch {
        setUser(null)
      } finally {
        setIsLoading(false)
      }
    }

    bootstrap()
  }, [])

  async function login(email: string, password: string) {
    const res = await apiFetch('/api/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
    const body = await res.json()
    if (!res.ok) {
      throw new Error(body?.message || 'Login gagal')
    }

    const data: TokenResponse | undefined = body?.data
    if (!data?.access_token || !data?.user) {
      throw new Error('Respons login tidak valid')
    }
    setAccessToken(data.access_token)
    setUser(data.user)
  }

  async function logout() {
    try {
      await apiFetch('/api/v1/auth/logout', { method: 'POST' })
    } catch {
      // best-effort — client-side state is cleared regardless
    }
    setAccessToken(null)
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, isLoading, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return ctx
}
