// Small text-only pill distinguishing a dashboard number that's bound to the
// selected date range ("Rentang") from one that's a snapshot of the current
// state regardless of the range filter ("Saat ini") - see
// .superpowers/specs/2026-09-18-dashboard-filter-clarity-design.md. Plain
// text, no icon - the user explicitly does not want an emoji/icon look here.
import { cn } from '@/lib/utils'

export function ScopeBadge({ type }: { type: 'range' | 'snapshot' }) {
  return (
    <span
      className={cn(
        'rounded-full px-2 py-0.5 text-[9.5px] font-bold tracking-[.03em] uppercase whitespace-nowrap',
        type === 'range' ? 'bg-[#EEF3FC] text-[#1D4ED8]' : 'bg-[#F1F5F9] text-[#64748B]'
      )}
    >
      {type === 'range' ? 'Rentang' : 'Saat ini'}
    </span>
  )
}
