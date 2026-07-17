// Gate protected routes on auth state. Renders nothing protected while the
// bootstrap refresh (AuthContext) is still in flight — a reload with no
// token in memory must not flash protected content or redirect to /login
// before the refresh cookie has had a chance to prove the session is valid.
import type { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuth } from './AuthContext'

export function RouteGuard({ children }: { children: ReactNode }) {
  const { user, isLoading } = useAuth()

  if (isLoading) {
    return <div>Memuat...</div>
  }

  if (!user) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}
