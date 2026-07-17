// List Lead screen: renders whatever useLeads({}) returns (default params
// — no filters/pagination controls here, those come later). Table at
// lg: and up, stacked cards below it — same data, no horizontal scroll.
import { useEffect, useState } from 'react'
import { Clock, Search, CalendarIcon, X } from 'lucide-react'
import type { DateRange } from 'react-day-picker'
import { useAuth } from '@/auth/AuthContext'
import { apiFetch } from '@/api/client'
import { useLeads } from './useLeads'
import { getLeadSources } from './refCache'
import type { LeadListParams, LeadSourceResponse, LeadStatus } from './api'
import type { components } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { cn } from '@/lib/utils'

type LeadResponse = components['schemas']['dto.LeadResponse']
type TeamResponse = components['schemas']['dto.TeamResponse']
type UserResponse = components['schemas']['dto.UserResponse']

// Team/sales dropdown data: a one-off reference fetch each, not reused
// elsewhere yet. Same envelope-unwrap pattern as fetchLeadSources in
// ./api.ts (check res.ok, guard res.json(), return body?.data ?? []) —
// ponytail: no session-cache wrapper like refCache.ts here, this is a
// single fetch-on-mount; add a cache if these end up reused elsewhere.
async function fetchTeams(): Promise<TeamResponse[]> {
  const res = await apiFetch('/api/v1/teams')
  if (!res.ok) {
    throw new Error('Failed to load teams')
  }
  let body: { data?: TeamResponse[] }
  try {
    body = await res.json()
  } catch {
    throw new Error('Failed to load teams')
  }
  return body?.data ?? []
}

async function fetchSalesUsers(): Promise<UserResponse[]> {
  const res = await apiFetch('/api/v1/users?role=SALES')
  if (!res.ok) {
    throw new Error('Failed to load sales users')
  }
  let body: { data?: UserResponse[] }
  try {
    body = await res.json()
  } catch {
    throw new Error('Failed to load sales users')
  }
  return body?.data ?? []
}

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
  BARU: { label: 'Baru', className: 'bg-[#E0F2FE] text-[#0369A1]' },
  FOLLOW_UP: { label: 'Follow-up', className: 'bg-[#EEF3FC] text-[#1E3A8A]' },
  HANDOFF_ODOO: { label: 'Handoff Odoo', className: 'bg-[#DCFCE7] text-[#166534]' },
  LOST: { label: 'Hilang', className: 'bg-[#FEE2E2] text-[#B91C1C]' },
}

const FILTER_LABEL_CLASSNAME = 'text-[10.5px] font-bold tracking-wide text-[#94A3B8] uppercase'

// date.toISOString() converts to UTC first — a user east of UTC clicking
// "today" on the calendar would have it saved as yesterday. Build the
// YYYY-MM-DD string from the Date's local components instead.
function toLocalDateString(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const DATE_RANGE_LABEL_FORMAT = new Intl.DateTimeFormat('id', {
  day: 'numeric',
  month: 'short',
  year: 'numeric',
})

function formatDateRangeLabel(range?: DateRange): string {
  if (!range?.from && !range?.to) return 'Rentang tanggal dibuat'
  if (range.from && range.to) {
    return `${DATE_RANGE_LABEL_FORMAT.format(range.from)} – ${DATE_RANGE_LABEL_FORMAT.format(range.to)}`
  }
  return DATE_RANGE_LABEL_FORMAT.format((range.from ?? range.to)!)
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
    <div className="hidden overflow-x-auto rounded-[14px] border border-[#E7EDF3] bg-white shadow-[0_1px_2px_rgba(15,23,42,0.04)] lg:block">
      <table className="w-full text-left text-sm">
        <thead className={cn('border-b bg-[#F7F9FC]', FILTER_LABEL_CLASSNAME)}>
          <tr>
            <th className="px-4 py-3">Kode</th>
            <th className="px-4 py-3">Perusahaan</th>
            <th className="px-4 py-3">Kota</th>
            <th className="px-4 py-3">PIC</th>
            <th className="px-4 py-3">Status</th>
            {showSales && <th className="px-4 py-3">Sales</th>}
            <th className="px-4 py-3 text-right">Follow-up terakhir</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-[#F1F5F9]">
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

const STATUS_OPTIONS: LeadStatus[] = ['BARU', 'FOLLOW_UP', 'HANDOFF_ODOO', 'LOST']

const SELECT_CLASSNAME =
  'h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50'

function FilterBar({
  q,
  onQChange,
  status,
  onStatusChange,
  sourceId,
  onSourceIdChange,
  staleOnly,
  onToggleStale,
  sources,
  showTeamFilter,
  teamId,
  onTeamIdChange,
  teams,
  showSalesFilter,
  ownerId,
  onOwnerIdChange,
  salesUsers,
  dateRange,
  onDateRangeChange,
}: {
  q: string
  onQChange: (value: string) => void
  status: LeadStatus | ''
  onStatusChange: (value: LeadStatus | '') => void
  sourceId: string
  onSourceIdChange: (value: string) => void
  staleOnly: boolean
  onToggleStale: () => void
  sources: LeadSourceResponse[]
  showTeamFilter: boolean
  teamId: string
  onTeamIdChange: (value: string) => void
  teams: TeamResponse[]
  showSalesFilter: boolean
  ownerId: string
  onOwnerIdChange: (value: string) => void
  salesUsers: UserResponse[]
  dateRange?: DateRange
  onDateRangeChange: (range?: DateRange) => void
}) {
  return (
    <Card className="rounded-[14px] shadow-[0_1px_2px_rgba(15,23,42,0.04)]">
      <CardContent className="flex flex-col gap-3 lg:flex-row lg:flex-wrap lg:items-end">
        <div className="flex flex-col gap-1 lg:min-w-[240px] lg:flex-1">
          <label className={FILTER_LABEL_CLASSNAME}>Cari</label>
          <div className="relative">
            <Search className="absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={q}
              onChange={(e) => onQChange(e.target.value)}
              placeholder="Cari kode lead atau nama perusahaan..."
              className="pl-8"
            />
          </div>
        </div>

        <div className="flex flex-col gap-1">
          <label className={FILTER_LABEL_CLASSNAME}>Status</label>
          <select
            value={status}
            onChange={(e) => onStatusChange(e.target.value as LeadStatus | '')}
            className={SELECT_CLASSNAME}
          >
            <option value="">Semua status</option>
            {STATUS_OPTIONS.map((option) => (
              <option key={option} value={option}>
                {STATUS_CONFIG[option].label}
              </option>
            ))}
          </select>
        </div>

        <div className="flex flex-col gap-1">
          <label className={FILTER_LABEL_CLASSNAME}>Sumber</label>
          <select
            value={sourceId}
            onChange={(e) => onSourceIdChange(e.target.value)}
            className={SELECT_CLASSNAME}
          >
            <option value="">Semua sumber</option>
            {sources.map((source, i) => (
              <option key={source.id ?? i} value={source.id ?? ''}>
                {source.name}
              </option>
            ))}
          </select>
        </div>

        {showTeamFilter && (
          <div className="flex flex-col gap-1">
            <label className={FILTER_LABEL_CLASSNAME}>Tim</label>
            <select
              value={teamId}
              onChange={(e) => onTeamIdChange(e.target.value)}
              className={SELECT_CLASSNAME}
            >
              <option value="">Semua tim</option>
              {teams.map((team, i) => (
                <option key={team.id ?? i} value={team.id ?? ''}>
                  {team.name}
                </option>
              ))}
            </select>
          </div>
        )}

        {showSalesFilter && (
          <div className="flex flex-col gap-1">
            <label className={FILTER_LABEL_CLASSNAME}>Sales</label>
            <select
              value={ownerId}
              onChange={(e) => onOwnerIdChange(e.target.value)}
              className={SELECT_CLASSNAME}
            >
              <option value="">Semua sales</option>
              {salesUsers.map((salesUser, i) => (
                <option key={salesUser.id ?? i} value={salesUser.id ?? ''}>
                  {salesUser.name}
                </option>
              ))}
            </select>
          </div>
        )}

        <div className="flex flex-col gap-1">
          <label className={FILTER_LABEL_CLASSNAME}>Rentang tanggal dibuat</label>
          <Popover>
            <PopoverTrigger
              className={cn(
                SELECT_CLASSNAME,
                'inline-flex items-center gap-1.5 text-left whitespace-nowrap'
              )}
            >
              <CalendarIcon className="size-3.5 shrink-0 text-muted-foreground" />
              <span className={cn(!dateRange?.from && !dateRange?.to && 'text-muted-foreground')}>
                {formatDateRangeLabel(dateRange)}
              </span>
              {(dateRange?.from || dateRange?.to) && (
                <X
                  className="ml-auto size-3.5 shrink-0 text-muted-foreground hover:text-foreground"
                  onClick={(e) => {
                    e.stopPropagation()
                    onDateRangeChange(undefined)
                  }}
                />
              )}
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0">
              <Calendar
                mode="range"
                selected={dateRange}
                onSelect={onDateRangeChange}
                defaultMonth={dateRange?.from}
                numberOfMonths={1}
              />
            </PopoverContent>
          </Popover>
        </div>

        <Button
          variant={staleOnly ? 'default' : 'outline'}
          size="sm"
          onClick={onToggleStale}
          className="lg:self-end"
        >
          Terlantar
        </Button>
      </CardContent>
    </Card>
  )
}

function Pagination({
  page,
  limit,
  total,
  onPrev,
  onNext,
}: {
  page: number
  limit: number
  total: number
  onPrev: () => void
  onNext: () => void
}) {
  if (total === 0) return null
  const start = (page - 1) * limit + 1
  const end = Math.min(page * limit, total)
  return (
    <div className="flex items-center justify-between gap-3 pt-1">
      <p className="text-sm text-muted-foreground">
        {start}–{end} dari {total}
      </p>
      <div className="flex gap-2">
        <Button variant="outline" size="sm" disabled={page <= 1} onClick={onPrev}>
          Sebelumnya
        </Button>
        <Button variant="outline" size="sm" disabled={end >= total} onClick={onNext}>
          Selanjutnya
        </Button>
      </div>
    </div>
  )
}

// Split out from the page component so "Coba lagi" can force a fresh
// useLeads() call by remounting this subtree (key={retryNonce} below) —
// useLeads has no refetch of its own and isn't this task's file to change.
// Filters/pagination state lives one level up in LeadListPage so a retry
// remount here doesn't reset them; params just flow in as a prop.
function LeadListContent({
  params,
  onRetry,
  onPageChange,
}: {
  params: LeadListParams
  onRetry: () => void
  onPageChange: (page: number) => void
}) {
  const { user } = useAuth()
  const { data, loading, error } = useLeads(params)
  const showSales = user?.role !== 'SALES'

  if (loading) return <LoadingState />
  if (error) return <ErrorState message={error} onRetry={onRetry} />

  const items = data?.items ?? []
  if (items.length === 0) return <EmptyState />

  const page = data?.page ?? params.page ?? 1
  const limit = data?.limit ?? params.limit ?? items.length
  const total = data?.total ?? items.length

  return (
    <>
      <LeadTable items={items} showSales={showSales} />
      <LeadCards items={items} showSales={showSales} />
      <Pagination
        page={page}
        limit={limit}
        total={total}
        onPrev={() => onPageChange(page - 1)}
        onNext={() => onPageChange(page + 1)}
      />
    </>
  )
}

export default function LeadListPage() {
  const { user } = useAuth()
  const isAdmin = user?.role === 'ADMIN_SALES' || user?.role === 'SU'
  const showSalesFilter = user?.role !== 'SALES'

  const [retryNonce, setRetryNonce] = useState(0)

  const [q, setQ] = useState('')
  const [debouncedQ, setDebouncedQ] = useState('')
  const [status, setStatus] = useState<LeadStatus | ''>('')
  const [sourceId, setSourceId] = useState('')
  const [teamId, setTeamId] = useState('')
  const [ownerId, setOwnerId] = useState('')
  const [dateRange, setDateRange] = useState<DateRange | undefined>(undefined)
  const [staleOnly, setStaleOnly] = useState(false)
  const [page, setPage] = useState(1)
  const [sources, setSources] = useState<LeadSourceResponse[]>([])
  const [teams, setTeams] = useState<TeamResponse[]>([])
  const [salesUsers, setSalesUsers] = useState<UserResponse[]>([])

  // Search box only joins the query params ~300ms after typing stops.
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedQ(q), 300)
    return () => clearTimeout(timer)
  }, [q])

  useEffect(() => {
    getLeadSources()
      .then(setSources)
      .catch(() => {
        // An empty dropdown on failure is an acceptable degradation — it
        // doesn't block the rest of the filter bar or the list itself.
      })
  }, [])

  // Team filter dropdown is admin-only — don't bother fetching it for a
  // role that will never see it rendered.
  useEffect(() => {
    if (!isAdmin) return
    fetchTeams()
      .then(setTeams)
      .catch(() => {
        // Same degrade-quietly approach as the sources fetch above.
      })
  }, [isAdmin])

  // Sales filter is hidden for SALES callers (and the backend wouldn't
  // scope it usefully for them anyway) — skip the fetch too.
  useEffect(() => {
    if (!showSalesFilter) return
    fetchSalesUsers()
      .then(setSalesUsers)
      .catch(() => {
        // Same degrade-quietly approach as the sources fetch above.
      })
  }, [showSalesFilter])

  const dateFrom = dateRange?.from ? toLocalDateString(dateRange.from) : ''
  const dateTo = dateRange?.to ? toLocalDateString(dateRange.to) : ''

  // A filter change invalidates whatever page the user was on — the old
  // page number may not even exist in the new, possibly-smaller result set.
  useEffect(() => {
    setPage(1)
  }, [debouncedQ, status, sourceId, teamId, ownerId, dateFrom, dateTo, staleOnly])

  const params: LeadListParams = {
    q: debouncedQ || undefined,
    status: status || undefined,
    source_id: sourceId || undefined,
    team_id: teamId || undefined,
    owner_id: ownerId || undefined,
    date_from: dateFrom || undefined,
    date_to: dateTo || undefined,
    stale: staleOnly ? 'true' : undefined,
    page,
  }

  return (
    <div className="flex w-full max-w-full flex-col gap-4 p-4 lg:p-6">
      <h1 className="font-display text-xl font-extrabold">Lead</h1>
      <FilterBar
        q={q}
        onQChange={setQ}
        status={status}
        onStatusChange={setStatus}
        sourceId={sourceId}
        onSourceIdChange={setSourceId}
        staleOnly={staleOnly}
        onToggleStale={() => setStaleOnly((v) => !v)}
        sources={sources}
        showTeamFilter={isAdmin}
        teamId={teamId}
        onTeamIdChange={setTeamId}
        teams={teams}
        showSalesFilter={showSalesFilter}
        ownerId={ownerId}
        onOwnerIdChange={setOwnerId}
        salesUsers={salesUsers}
        dateRange={dateRange}
        onDateRangeChange={setDateRange}
      />
      <LeadListContent
        key={retryNonce}
        params={params}
        onRetry={() => setRetryNonce((n) => n + 1)}
        onPageChange={setPage}
      />
    </div>
  )
}
