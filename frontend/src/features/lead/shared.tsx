// Display bits shared between List and Detail - extracted here once Detail
// needed the same status pill/sales badge/relative-time formatting List
// already had, rather than duplicating them.
import { cn } from '@/lib/utils'

export const STATUS_CONFIG: Record<string, { label: string; className: string }> = {
  BARU: { label: 'Baru', className: 'bg-[#E0F2FE] text-[#0369A1]' },
  FOLLOW_UP: { label: 'Follow-up', className: 'bg-[#EEF3FC] text-[#1E3A8A]' },
  HANDOFF_ODOO: { label: 'Handoff Odoo', className: 'bg-[#DCFCE7] text-[#166534]' },
  LOST: { label: 'Hilang', className: 'bg-[#FEE2E2] text-[#B91C1C]' },
}

export const RELATIVE_TIME = new Intl.RelativeTimeFormat('id', { numeric: 'auto' })

export function formatRelativeTime(iso: string): string {
  const diffMs = new Date(iso).getTime() - Date.now()
  const diffMinutes = Math.round(diffMs / (1000 * 60))
  if (Math.abs(diffMinutes) < 60) return RELATIVE_TIME.format(diffMinutes, 'minute')
  const diffHours = Math.round(diffMinutes / 60)
  if (Math.abs(diffHours) < 24) return RELATIVE_TIME.format(diffHours, 'hour')
  return RELATIVE_TIME.format(Math.round(diffHours / 24), 'day')
}

export function StatusPill({ status }: { status?: string }) {
  const config = (status && STATUS_CONFIG[status]) || {
    label: status ?? '—',
    className: 'bg-muted text-muted-foreground',
  }
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium',
        config.className
      )}
    >
      <span className="size-1.5 rounded-full bg-current" />
      {config.label}
    </span>
  )
}

export function SalesBadge({ ownerName }: { ownerName?: string }) {
  const initial = ownerName ? ownerName.charAt(0).toUpperCase() : '?'
  return (
    <div className="flex items-center gap-2">
      <div className="flex size-[26px] shrink-0 items-center justify-center rounded-[7px] bg-[#EEF2F7] text-[10.5px] font-bold text-[#475569]">
        {initial}
      </div>
      {ownerName && <span className="text-sm">{ownerName}</span>}
    </div>
  )
}
