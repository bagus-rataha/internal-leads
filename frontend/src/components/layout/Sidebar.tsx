// Presentational shell nav. Every destination is locked (its target
// screen doesn't exist yet) and DATA/Settings are gated to
// ADMIN_SALES/SU per the RBAC visibility rules — gating is a rendering
// decision here, the backend is what actually enforces access.
import type { LucideIcon } from 'lucide-react'
import type { ReactNode } from 'react'
import { LayoutDashboard, FileText, Users, Tag, Wrench, UserCog, Lock } from 'lucide-react'
import { useAuth } from '@/auth/AuthContext'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface SidebarProps {
  isOpen: boolean
  onClose: () => void
}

interface NavItemProps {
  icon: LucideIcon
  label: string
  locked?: boolean
}

// Locked items render as <div>, never <button>/<a> — a disabled-looking
// interactive element that can still be focused/clicked is worse than one
// that plainly isn't interactive.
function NavItem({ icon: Icon, label, locked }: NavItemProps) {
  if (locked) {
    return (
      <div
        aria-disabled="true"
        className="flex cursor-not-allowed items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm text-muted-foreground select-none"
      >
        <Icon className="size-4" />
        <span>{label}</span>
        <Lock className="ml-auto size-3.5" />
      </div>
    )
  }

  return (
    <button
      type="button"
      className="flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm font-medium hover:bg-muted"
    >
      <Icon className="size-4" />
      <span>{label}</span>
    </button>
  )
}

function NavGroup({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-1">
      <p className="px-2.5 text-xs font-semibold tracking-wide text-muted-foreground uppercase">
        {label}
      </p>
      <div className="flex flex-col gap-0.5">{children}</div>
    </div>
  )
}

export function Sidebar({ isOpen, onClose }: SidebarProps) {
  const { user, logout } = useAuth()
  const isAdmin = user?.role === 'ADMIN_SALES' || user?.role === 'SU'
  const initial = user?.name?.charAt(0).toUpperCase() ?? '?'

  return (
    <>
      {/* Mobile-only backdrop, click to close the drawer. */}
      {isOpen && (
        <div
          aria-hidden="true"
          onClick={onClose}
          className="fixed inset-0 z-30 bg-black/40 lg:hidden"
        />
      )}

      <aside
        className={cn(
          'fixed inset-y-0 left-0 z-40 flex h-screen w-[248px] flex-col border-r bg-background transition-transform',
          'lg:static lg:translate-x-0',
          isOpen ? 'translate-x-0' : '-translate-x-full'
        )}
      >
        <div className="flex items-center gap-2.5 border-b px-4 py-4">
          <div className="flex size-[34px] shrink-0 items-center justify-center rounded-lg bg-primary font-display font-extrabold text-primary-foreground">
            L
          </div>
          <div className="min-w-0">
            <p className="font-display leading-tight font-extrabold">LMS</p>
            <p className="truncate text-xs leading-tight text-muted-foreground">
              Sales ISP Korporat
            </p>
          </div>
        </div>

        <nav className="flex flex-col gap-4 overflow-y-auto px-3 py-4">
          <NavGroup label="Kerja Harian">
            <NavItem icon={LayoutDashboard} label="Dashboard" locked />
            <NavItem icon={FileText} label="Lead" locked />
          </NavGroup>

          {isAdmin && (
            <NavGroup label="DATA">
              <NavItem icon={Users} label="Sales Team" locked />
              <NavItem icon={Tag} label="Sumber Lead" locked />
              <NavItem icon={Wrench} label="Type Layanan" locked />
            </NavGroup>
          )}

          {isAdmin && (
            <NavGroup label="Settings">
              <NavItem icon={UserCog} label="User" locked />
            </NavGroup>
          )}
        </nav>

        <div className="mt-auto border-t px-4 py-4">
          <div className="flex items-center gap-2.5">
            <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary font-display font-extrabold text-sm text-primary-foreground">
              {initial}
            </div>
            <div className="min-w-0">
              <p className="truncate text-sm font-medium">{user?.name}</p>
              <p className="truncate text-xs text-muted-foreground">
                {user?.role} · {user?.team_id ?? '-'}
              </p>
            </div>
          </div>
          <Button variant="outline" size="sm" className="mt-3 w-full" onClick={() => logout()}>
            Logout
          </Button>
        </div>
      </aside>
    </>
  )
}
