// List Lead screen: renders whatever useLeads({}) returns (default params
// — no filters/pagination controls here, those come later). Table at
// lg: and up, stacked cards below it — same data, no horizontal scroll.
import { useEffect, useState, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { Clock, Search, CalendarIcon, X, ChevronLeft, ChevronRight, Plus, SlidersHorizontal } from 'lucide-react'
import type { DateRange } from 'react-day-picker'
import { useAuth } from '@/auth/AuthContext'
import { useLeads } from './useLeads'
import { getLeadSources } from './refCache'
import { fetchTeams, fetchSalesUsers, type LeadListParams, type LeadSourceResponse, type LeadStatus, type LeadResponse, type TeamResponse, type UserResponse } from './api'
import { STATUS_CONFIG, StatusPill, SalesBadge, formatRelativeTime } from './shared'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { cn } from '@/lib/utils'

const FILTER_LABEL_CLASSNAME = 'text-[10.5px] font-bold tracking-[.05em] text-[#94A3B8] uppercase'

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

function FollowUpInfo({ lead }: { lead: LeadResponse }) {
  return (
    <div className="flex flex-col items-start gap-1 lg:items-end">
      {lead.is_stale && (
        <span className="inline-flex items-center gap-[5px] rounded-full border border-[#FCE39A] bg-[#FEF3C7] px-[9px] py-[3px] text-[11px] font-bold text-[#92580A]">
          <Clock className="size-3" />
          Terlantar
        </span>
      )}
      <span
        className={cn(
          'font-mono text-[12.5px] font-medium',
          lead.is_stale ? 'text-[#92580A]' : 'text-[#64748B]'
        )}
      >
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

function LeadTable({
  items,
  showSales,
  pagination,
}: {
  items: LeadResponse[]
  showSales: boolean
  pagination: ReactNode
}) {
  const navigate = useNavigate()
  return (
    <div className="hidden overflow-hidden rounded-[14px] border border-[#E7EDF3] bg-white shadow-[0_1px_2px_rgba(15,23,42,0.04)] lg:block">
      <div className="overflow-x-auto">
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
              <tr
                key={lead.code}
                className="cursor-pointer hover:bg-[#F7F9FC]"
                onClick={() => navigate(`/leads/${lead.code}`)}
              >
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
      {pagination}
    </div>
  )
}

function LeadCards({
  items,
  showSales,
  pagination,
}: {
  items: LeadResponse[]
  showSales: boolean
  pagination: ReactNode
}) {
  const navigate = useNavigate()
  return (
    <div className="flex flex-col divide-y divide-[#F1F5F9] rounded-[14px] border border-[#E7EDF3] bg-white shadow-[0_1px_2px_rgba(15,23,42,0.04)] lg:hidden">
      {items.map((lead) => (
        <div
          key={lead.code}
          className="relative flex cursor-pointer flex-col gap-1.5 px-4 py-3 active:bg-[#F7F9FC]"
          onClick={() => navigate(`/leads/${lead.code}`)}
        >
          {lead.is_stale && (
            <span className="absolute top-2 bottom-2 left-0 w-[3px] rounded bg-amber-600" />
          )}
          <div className="flex items-start justify-between gap-2">
            <div className="min-w-0">
              <p className="font-mono text-xs text-primary">{lead.code}</p>
              <p className="truncate font-semibold">{lead.company_name}</p>
            </div>
            <StatusPill status={lead.status} />
          </div>
          <FollowUpInfo lead={lead} />
          {showSales && <SalesBadge ownerName={lead.owner_name} />}
        </div>
      ))}
      {pagination}
    </div>
  )
}

const STATUS_OPTIONS: LeadStatus[] = ['BARU', 'FOLLOW_UP', 'HANDOFF_ODOO', 'LOST']

const SELECT_CLASSNAME =
  'h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50'

type FilterValues = {
  status: LeadStatus | ''
  sourceId: string
  teamId: string
  ownerId: string
  dateRange?: DateRange
  staleOnly: boolean
}

const EMPTY_FILTERS: FilterValues = {
  status: '',
  sourceId: '',
  teamId: '',
  ownerId: '',
  dateRange: undefined,
  staleOnly: false,
}

function FilterBar({
  q,
  onQChange,
  draft,
  onDraftChange,
  sources,
  showTeamFilter,
  teams,
  showSalesFilter,
  salesUsers,
  onApply,
  onReset,
}: {
  q: string
  onQChange: (value: string) => void
  draft: FilterValues
  onDraftChange: (patch: Partial<FilterValues>) => void
  sources: LeadSourceResponse[]
  showTeamFilter: boolean
  teams: TeamResponse[]
  showSalesFilter: boolean
  salesUsers: UserResponse[]
  onApply: () => void
  onReset: () => void
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
            value={draft.status}
            onChange={(e) => onDraftChange({ status: e.target.value as LeadStatus | '' })}
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
            value={draft.sourceId}
            onChange={(e) => onDraftChange({ sourceId: e.target.value })}
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
              value={draft.teamId}
              onChange={(e) => onDraftChange({ teamId: e.target.value })}
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
              value={draft.ownerId}
              onChange={(e) => onDraftChange({ ownerId: e.target.value })}
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
              <span className={cn(!draft.dateRange?.from && !draft.dateRange?.to && 'text-muted-foreground')}>
                {formatDateRangeLabel(draft.dateRange)}
              </span>
              {(draft.dateRange?.from || draft.dateRange?.to) && (
                <X
                  className="ml-auto size-3.5 shrink-0 text-muted-foreground hover:text-foreground"
                  onClick={(e) => {
                    e.stopPropagation()
                    onDraftChange({ dateRange: undefined })
                  }}
                />
              )}
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0">
              <Calendar
                mode="range"
                selected={draft.dateRange}
                onSelect={(range) => onDraftChange({ dateRange: range })}
                defaultMonth={draft.dateRange?.from}
                numberOfMonths={1}
              />
            </PopoverContent>
          </Popover>
        </div>

        <Button
          variant={draft.staleOnly ? 'default' : 'outline'}
          size="sm"
          onClick={() => onDraftChange({ staleOnly: !draft.staleOnly })}
          className="lg:self-end"
        >
          Terlantar
        </Button>

        <div className="flex items-center justify-end gap-2 lg:basis-full lg:border-t lg:border-t-[#F1F5F9] lg:pt-3">
          <Button variant="outline" size="sm" onClick={onReset}>
            Reset
          </Button>
          <Button size="sm" onClick={onApply}>
            Terapkan
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

// Windowed pagination: all pages when there's few of them, otherwise first,
// last, current ±1 neighbor, and an ellipsis filling each gap.
function getPageItems(page: number, totalPages: number): (number | 'ellipsis')[] {
  if (totalPages <= 7) return Array.from({ length: totalPages }, (_, i) => i + 1)
  const items: (number | 'ellipsis')[] = [1]
  const start = Math.max(2, page - 1)
  const end = Math.min(totalPages - 1, page + 1)
  if (start > 2) items.push('ellipsis')
  for (let p = start; p <= end; p++) items.push(p)
  if (end < totalPages - 1) items.push('ellipsis')
  items.push(totalPages)
  return items
}

const PAGINATION_BUTTON_CLASSNAME =
  'flex size-[30px] shrink-0 items-center justify-center rounded-[7px] border text-[12px] font-semibold transition-colors disabled:cursor-not-allowed disabled:opacity-40'

function Pagination({
  page,
  limit,
  total,
  itemsShown,
  onPageChange,
}: {
  page: number
  limit: number
  total: number
  itemsShown: number
  onPageChange: (page: number) => void
}) {
  if (total === 0) return null
  const totalPages = Math.max(1, Math.ceil(total / limit))
  const pageItems = getPageItems(page, totalPages)
  return (
    <div className="flex items-center justify-between border-t border-t-[#F1F5F9] bg-[#FCFDFE] px-4 py-[11px]">
      <p className="text-[12px] text-[#94A3B8]">
        Menampilkan <strong className="text-[#475569]">{itemsShown}</strong> dari {total} lead
      </p>
      <div className="flex items-center gap-1.5">
        <button
          type="button"
          aria-label="Halaman sebelumnya"
          className={cn(PAGINATION_BUTTON_CLASSNAME, 'border-[#E2E8F0] bg-white')}
          disabled={page <= 1}
          onClick={() => onPageChange(page - 1)}
        >
          <ChevronLeft className="size-4 text-[#94A3B8]" />
        </button>
        {pageItems.map((item, i) =>
          item === 'ellipsis' ? (
            <span key={`ellipsis-${i}`} className="px-1 text-[12px] text-[#94A3B8]">
              …
            </span>
          ) : (
            <button
              key={item}
              type="button"
              className={cn(
                PAGINATION_BUTTON_CLASSNAME,
                item === page
                  ? 'border-[#1D4ED8] bg-[#1D4ED8] text-white'
                  : 'border-[#E2E8F0] bg-white text-[#475569] hover:bg-[#F7F9FC]'
              )}
              onClick={() => onPageChange(item)}
            >
              {item}
            </button>
          )
        )}
        <button
          type="button"
          aria-label="Halaman berikutnya"
          className={cn(PAGINATION_BUTTON_CLASSNAME, 'border-[#E2E8F0] bg-white')}
          disabled={page >= totalPages}
          onClick={() => onPageChange(page + 1)}
        >
          <ChevronRight className="size-4 text-[#94A3B8]" />
        </button>
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

  const pagination = (
    <Pagination
      page={page}
      limit={limit}
      total={total}
      itemsShown={items.length}
      onPageChange={onPageChange}
    />
  )

  return (
    <>
      <LeadTable items={items} showSales={showSales} pagination={pagination} />
      <LeadCards items={items} showSales={showSales} pagination={pagination} />
    </>
  )
}

export default function LeadListPage() {
  const navigate = useNavigate()
  const { user } = useAuth()
  const isAdmin = user?.role === 'ADMIN_SALES' || user?.role === 'SU'
  const showSalesFilter = user?.role !== 'SALES'

  const [retryNonce, setRetryNonce] = useState(0)
  const [filtersOpen, setFiltersOpen] = useState(false)

  const [q, setQ] = useState('')
  const [debouncedQ, setDebouncedQ] = useState('')
  const [applied, setApplied] = useState<FilterValues>(EMPTY_FILTERS)
  const [draft, setDraft] = useState<FilterValues>(EMPTY_FILTERS)

  function updateDraft(patch: Partial<FilterValues>) {
    setDraft((prev) => ({ ...prev, ...patch }))
  }

  function handleApplyFilters() {
    setApplied(draft)
    setPage(1)
    setFiltersOpen(false)
  }

  function handleResetFilters() {
    setDraft(EMPTY_FILTERS)
    setApplied(EMPTY_FILTERS)
    setPage(1)
  }
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

  // Search still auto-applies, so it still needs its own page reset.
  // Apply/Reset (handleApplyFilters/handleResetFilters above) reset the
  // page themselves when `applied` changes — no separate effect needed.
  useEffect(() => {
    setPage(1)
  }, [debouncedQ])

  const params: LeadListParams = {
    q: debouncedQ || undefined,
    status: applied.status || undefined,
    source_id: applied.sourceId || undefined,
    team_id: applied.teamId || undefined,
    owner_id: applied.ownerId || undefined,
    date_from: applied.dateRange?.from ? toLocalDateString(applied.dateRange.from) : undefined,
    date_to: applied.dateRange?.to ? toLocalDateString(applied.dateRange.to) : undefined,
    stale: applied.staleOnly ? 'true' : undefined,
    page,
  }

  const activeFilterCount = [
    q,
    applied.status,
    applied.sourceId,
    applied.teamId,
    applied.ownerId,
    applied.staleOnly,
    applied.dateRange?.from,
  ].filter(Boolean).length

  return (
    <div className="flex w-full max-w-full flex-col gap-4 p-4 lg:p-6">
      <div className="flex items-center justify-between">
        <h1 className="font-display text-xl font-extrabold">Lead</h1>
        <div className="flex items-center gap-2">
          <Button
            variant={filtersOpen ? 'default' : 'outline'}
            size="sm"
            onClick={() => setFiltersOpen((v) => !v)}
            className="gap-1.5"
          >
            <SlidersHorizontal className="size-4" />
            Filter
            {activeFilterCount > 0 && (
              <span className="ml-0.5 flex size-4 items-center justify-center rounded-full bg-white/20 text-[10px] font-bold">
                {activeFilterCount}
              </span>
            )}
          </Button>
          <Button
            size="sm"
            onClick={() => navigate('/leads/new')}
            className="gap-1.5 bg-[#1D4ED8] text-white shadow-[0_1px_2px_rgba(29,78,216,0.3)] hover:bg-[#1A45BE]"
          >
            <Plus className="size-4" />
            Lead Baru
          </Button>
        </div>
      </div>
      {filtersOpen && (
        <FilterBar
          q={q}
          onQChange={setQ}
          draft={draft}
          onDraftChange={updateDraft}
          sources={sources}
          showTeamFilter={isAdmin}
          teams={teams}
          showSalesFilter={showSalesFilter}
          salesUsers={salesUsers}
          onApply={handleApplyFilters}
          onReset={handleResetFilters}
        />
      )}
      <LeadListContent
        key={retryNonce}
        params={params}
        onRetry={() => setRetryNonce((n) => n + 1)}
        onPageChange={setPage}
      />
    </div>
  )
}
