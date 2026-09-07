// Display bits shared between List and Detail - extracted here once Detail
// needed the same status pill/sales badge/relative-time formatting List
// already had, rather than duplicating them.
import { cn } from '@/lib/utils'

export const STATUS_CONFIG: Record<string, { label: string; className: string }> = {
  BARU: { label: 'Draft', className: 'bg-[#E0F2FE] text-[#0369A1]' },
  FOLLOW_UP: { label: 'Follow-up', className: 'bg-[#EEF3FC] text-[#1E3A8A]' },
  SURVEY: { label: 'Survey', className: 'bg-[#ECFDF5] text-[#047857]' },
  SALES_CONFIRMATION: { label: 'Sales Confirmation (SC)', className: 'bg-[#DCFCE7] text-[#15803D]' },
  REGISTRASI: { label: 'Registrasi', className: 'bg-[#D1FAE5] text-[#166534]' },
  INSTALASI: { label: 'Instalasi', className: 'bg-[#BBF7D0] text-[#166534]' },
  TRIAL: { label: 'Trial', className: 'bg-[#A7F3D0] text-[#14532D]' },
  INVOICE_BULANAN: { label: 'Invoice Bulanan', className: 'bg-[#16A34A] text-white' },
  LOST: { label: 'Hilang', className: 'bg-[#FEE2E2] text-[#B91C1C]' },
}

export const PIPELINE_ORDER = [
  'FOLLOW_UP', 'SURVEY', 'SALES_CONFIRMATION', 'REGISTRASI',
  'INSTALASI', 'TRIAL', 'INVOICE_BULANAN',
] as const

// Statuses offered as a filter choice in the UI - the full pipeline plus
// LOST. Shared by the lead list filter and the dashboard forecast-status
// filter.
export const STATUS_FILTER_OPTIONS: string[] = [
  'BARU', 'FOLLOW_UP', 'SURVEY', 'SALES_CONFIRMATION', 'REGISTRASI',
  'INSTALASI', 'TRIAL', 'INVOICE_BULANAN', 'LOST',
]

function chainIndex(status: string): number {
  return (PIPELINE_ORDER as readonly string[]).indexOf(status)
}

export function nextStage(status: string): string | null {
  const i = chainIndex(status)
  return i >= 0 && i < PIPELINE_ORDER.length - 1 ? PIPELINE_ORDER[i + 1] : null
}

export function earlierStages(status: string): string[] {
  const i = chainIndex(status)
  return i > 0 ? (PIPELINE_ORDER as readonly string[]).slice(0, i) : []
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
        'inline-flex items-center gap-1.5 whitespace-nowrap rounded-full px-2.5 py-1 text-xs font-medium',
        config.className
      )}
    >
      <span className="size-1.5 rounded-full bg-current" />
      {config.label}
    </span>
  )
}

export function SalesBadge({ ownerName }: { ownerName?: string }) {
  return <span className="text-sm">{ownerName ?? '—'}</span>
}
