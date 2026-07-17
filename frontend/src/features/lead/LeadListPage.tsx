// List Lead screen: renders whatever useLeads({}) returns (default params
// — no filters/pagination controls here, those come later). Table at
// lg: and up, stacked cards below it — same data, no horizontal scroll.
import { useState } from 'react'
import { Clock, Search } from 'lucide-react'
import { useAuth } from '@/auth/AuthContext'
import { useLeads } from './useLeads'
import type { components } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'

type LeadResponse = components['schemas']['dto.LeadResponse']

const RELATIVE_TIME = new Intl.RelativeTimeFormat('id', { numeric: 'auto' })

// One function covers minute/hour/day granularity — good enough for a
// "last follow-up" timestamp, no need for a full date library.
function formatRelativeTime(iso: string): string {
  const diffMs = new Date(iso).getTime() - Date.now()
  const diffMinutes = Math.round(diffMs / (1000 * 60))
  if (Math.abs(diffMinutes) < 60) return RELATIVE_TIME.format(diffMinutes, 'minute')
  const diffHours = Math.round(diffMinutes / 60)
  if (Math.abs(diffHours) < 24) return RELATIVE_TIME.format(diffHours, 'hour')
  return RELATIVE_TIME.format(Math.round(diffHours / 24), 'day')
}

const STATUS_CONFIG: Record<string, { label: string; className: string }> = {
  BARU: { label: 'Baru', className: 'bg-primary/10 text-primary' },
  FOLLOW_UP: { label: 'Follow-up', className: 'bg-amber-100 text-amber-800' },
  HANDOFF_ODOO: { label: 'Handoff Odoo', className: 'bg-green-100 text-green-800' },
  LOST: { label: 'Hilang', className: 'bg-muted text-muted-foreground' },
}

function StatusPill({ status }: { status?: string }) {
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

function SalesBadge({ ownerName }: { ownerName?: string }) {
  const initial = ownerName ? ownerName.charAt(0).toUpperCase() : '?'
  return (
    <div className="flex items-center gap-2">
      <div className="flex size-7 shrink-0 items-center justify-center rounded-full bg-primary font-display text-xs font-extrabold text-primary-foreground">
        {initial}
      </div>
      {ownerName && <span className="text-sm">{ownerName}</span>}
    </div>
  )
}

function FollowUpInfo({ lead }: { lead: LeadResponse }) {
  return (
    <div className="flex flex-col items-start gap-1 lg:items-end">
      {lead.is_stale && (
        <span className="inline-flex items-center gap-1 rounded-full bg-amber-100 px-2 py-0.5 text-xs font-medium text-amber-800">
          <Clock className="size-3" />
          Terlantar
        </span>
      )}
      <span className="text-xs text-muted-foreground">
        {lead.last_follow_up_at ? formatRelativeTime(lead.last_follow_up_at) : 'Belum ada follow-up'}
      </span>
    </div>
  )
}

function LoadingState() {
  return (
    <div className="flex flex-col gap-3">
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} className="h-14 w-full animate-pulse rounded-lg bg-muted" />
      ))}
    </div>
  )
}

function ErrorState({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <div className="flex flex-col items-center gap-3 rounded-xl border border-destructive/30 bg-destructive/5 px-6 py-10 text-center">
      <p className="text-sm font-medium text-destructive">{message}</p>
      <Button variant="outline" size="sm" onClick={onRetry}>
        Coba lagi
      </Button>
    </div>
  )
}

function EmptyState() {
  return (
    <div className="flex flex-col items-center gap-2 px-6 py-16 text-center">
      <Search className="size-8 text-muted-foreground" />
      <p className="font-medium">Tidak ada lead yang cocok</p>
      <p className="text-sm text-muted-foreground">Coba ubah filter atau kata kunci pencarian.</p>
    </div>
  )
}

function LeadTable({ items, showSales }: { items: LeadResponse[]; showSales: boolean }) {
  return (
    <div className="hidden overflow-x-auto rounded-xl border lg:block">
      <table className="w-full text-left text-sm">
        <thead className="border-b bg-muted/50 text-xs text-muted-foreground uppercase">
          <tr>
            <th className="px-4 py-3 font-medium">Kode</th>
            <th className="px-4 py-3 font-medium">Perusahaan</th>
            <th className="px-4 py-3 font-medium">Kota</th>
            <th className="px-4 py-3 font-medium">PIC</th>
            <th className="px-4 py-3 font-medium">Status</th>
            {showSales && <th className="px-4 py-3 font-medium">Sales</th>}
            <th className="px-4 py-3 text-right font-medium">Follow-up terakhir</th>
          </tr>
        </thead>
        <tbody className="divide-y">
          {items.map((lead) => (
            <tr key={lead.code}>
              <td className="relative px-4 py-3">
                {lead.is_stale && (
                  <span className="absolute top-2 bottom-2 left-0 w-[3px] rounded bg-amber-600" />
                )}
                <span className="font-mono text-primary">{lead.code}</span>
              </td>
              <td className="px-4 py-3">
                <p className="font-semibold">{lead.company_name}</p>
                {lead.business_field && (
                  <p className="text-xs text-muted-foreground">{lead.business_field}</p>
                )}
              </td>
              <td className="px-4 py-3">{lead.city_name ?? '—'}</td>
              <td className="px-4 py-3">
                <p>{lead.pic_name}</p>
                {lead.pic_position && (
                  <p className="text-xs text-muted-foreground">{lead.pic_position}</p>
                )}
              </td>
              <td className="px-4 py-3">
                <StatusPill status={lead.status} />
              </td>
              {showSales && (
                <td className="px-4 py-3">
                  <SalesBadge ownerName={lead.owner_name} />
                </td>
              )}
              <td className="px-4 py-3 text-right">
                <FollowUpInfo lead={lead} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function LeadCards({ items, showSales }: { items: LeadResponse[]; showSales: boolean }) {
  return (
    <div className="flex flex-col gap-3 lg:hidden">
      {items.map((lead) => (
        <Card key={lead.code} className="relative">
          {lead.is_stale && (
            <span className="absolute top-3 bottom-3 left-0 w-[3px] rounded bg-amber-600" />
          )}
          <CardContent className="flex flex-col gap-2">
            <div className="flex items-start justify-between gap-2">
              <span className="font-mono text-primary">{lead.code}</span>
              <StatusPill status={lead.status} />
            </div>
            <div>
              <p className="font-semibold">{lead.company_name}</p>
              {lead.business_field && (
                <p className="text-xs text-muted-foreground">{lead.business_field}</p>
              )}
            </div>
            <div className="grid grid-cols-2 gap-2 text-sm">
              <div>
                <p className="text-xs text-muted-foreground">Kota</p>
                <p>{lead.city_name ?? '—'}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">PIC</p>
                <p>{lead.pic_name}</p>
                {lead.pic_position && (
                  <p className="text-xs text-muted-foreground">{lead.pic_position}</p>
                )}
              </div>
            </div>
            {showSales && (
              <div>
                <p className="text-xs text-muted-foreground">Sales</p>
                <SalesBadge ownerName={lead.owner_name} />
              </div>
            )}
            <FollowUpInfo lead={lead} />
          </CardContent>
        </Card>
      ))}
    </div>
  )
}

// Split out from the page component so "Coba lagi" can force a fresh
// useLeads() call by remounting this subtree (key={retryNonce} below) —
// useLeads has no refetch of its own and isn't this task's file to change.
function LeadListContent({ onRetry }: { onRetry: () => void }) {
  const { user } = useAuth()
  const { data, loading, error } = useLeads({})
  const showSales = user?.role !== 'SALES'

  if (loading) return <LoadingState />
  if (error) return <ErrorState message={error} onRetry={onRetry} />

  const items = data?.items ?? []
  if (items.length === 0) return <EmptyState />

  return (
    <>
      <LeadTable items={items} showSales={showSales} />
      <LeadCards items={items} showSales={showSales} />
    </>
  )
}

export default function LeadListPage() {
  const [retryNonce, setRetryNonce] = useState(0)

  return (
    <div className="flex w-full max-w-full flex-col gap-4 p-4 lg:p-6">
      <h1 className="font-display text-xl font-extrabold">Lead</h1>
      <LeadListContent key={retryNonce} onRetry={() => setRetryNonce((n) => n + 1)} />
    </div>
  )
}
