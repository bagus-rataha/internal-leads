// Nested inside RouteGuard's already-authenticated tree - redirects
// non-admin roles away from Master Data screens. This is a UI convenience -
// the backend enforces the real boundary via RequireRole regardless of what
// this component does.
import type { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuth } from './AuthContext'

export function AdminRouteGuard({ children }: { children: ReactNode }) {
  const { user } = useAuth()

  if (user?.role !== 'ADMIN_SALES' && user?.role !== 'SU') {
    return <Navigate to="/leads" replace />
  }

  return <>{children}</>
}
