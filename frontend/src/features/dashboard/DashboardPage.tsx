// Dashboard screen: filters (Rentang/Tim/Sales) drive every widget below via
// shared {date_from,date_to,team_id,owner_id} params - the funnel and the
// segments breakdowns ignore date_from/date_to internally (see their backend
// implementations) but still respect team_id/owner_id, which flow through
// the same params object.
import { useEffect, useState } from 'react'
import { CalendarIcon } from 'lucide-react'
import type { DateRange } from 'react-day-picker'
import { useAuth } from '@/auth/AuthContext'
import { roleLabel } from '@/lib/roles'
import { fetchTeams, type TeamResponse } from '@/features/team/api'
import { fetchUsers, LEAD_OWNER_ROLES, type UserResponse } from '@/features/user/api'
import { STATUS_CONFIG, STATUS_FILTER_OPTIONS } from '@/features/lead/shared'
import { Calendar } from '@/components/ui/calendar'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { cn } from '@/lib/utils'
import {
  useDashboardSummary,
  useDashboardActivity,
  useDashboardStaleLeads,
  useDashboardSalesActivity,
  useDashboardSegments,
} from './queries'
import { MetricCards, FunnelCard, ForecastCard } from './SummaryWidgets'
import { TrendChart } from './TrendChart'
import { StaleLeadsCard, SalesActivityCard } from './StaleAndSalesTables'
import { SegmentsSection } from './SegmentsSection'

function toLocalDateString(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// Calendar-day count between two Date objects, inclusive of both ends,
// ignoring time-of-day - Date.UTC with only y/m/d avoids DST/local-time
// artifacts a plain millisecond diff would introduce.
function daysBetweenInclusive(from: Date, to: Date): number {
  const utcFrom = Date.UTC(from.getFullYear(), from.getMonth(), from.getDate())
  const utcTo = Date.UTC(to.getFullYear(), to.getMonth(), to.getDate())
  return Math.round((utcTo - utcFrom) / 86400000) + 1
}

function startOfWeekMonday(date: Date): Date {
  const d = new Date(date)
  const day = d.getDay() // 0 = Sunday .. 6 = Saturday
  const diffFromMonday = day === 0 ? 6 : day - 1
  d.setDate(d.getDate() - diffFromMonday)
  return d
}

function startOfMonth(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), 1)
}

type RangePreset = 'today' | 'week' | 'month' | 'custom'

const MAX_CUSTOM_RANGE_DAYS = 90

const RANGE_PRESET_OPTIONS: { value: RangePreset; label: string }[] = [
  { value: 'today', label: 'Hari ini' },
  { value: 'week', label: 'Minggu ini' },
  { value: 'month', label: 'Bulan ini' },
  { value: 'custom', label: 'Custom' },
]

const DATE_RANGE_LABEL_FORMAT = new Intl.DateTimeFormat('id', { day: 'numeric', month: 'short', year: 'numeric' })

function formatCustomRangeLabel(range: DateRange | undefined): string {
  if (!range?.from && !range?.to) return 'Pilih tanggal'
  if (range?.from && range.to) return `${DATE_RANGE_LABEL_FORMAT.format(range.from)} – ${DATE_RANGE_LABEL_FORMAT.format(range.to)}`
  return DATE_RANGE_LABEL_FORMAT.format((range?.from ?? range?.to)!)
}

const SELECT_CLASSNAME =
  'h-9 rounded-lg border border-input bg-white px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50'
const FILTER_LABEL_CLASSNAME = 'text-[10.5px] font-bold tracking-[.05em] text-[#94A3B8] uppercase'
const PRESET_BUTTON_BASE = 'h-9 rounded-lg border px-3 text-sm font-semibold transition-colors'

export default function DashboardPage() {
  const { user } = useAuth()
  const isAdmin = user?.role === 'ADMIN_SALES' || user?.role === 'SU'
  const showSalesWidget = user?.role !== 'SALES'
  const showTeamFilter = isAdmin

  const [rangePreset, setRangePreset] = useState<RangePreset>('month')
  const [customRange, setCustomRange] = useState<DateRange | undefined>()
  const [teamId, setTeamId] = useState('')
  const [ownerId, setOwnerId] = useState('')
  const [status, setStatus] = useState('')
  const [teams, setTeams] = useState<TeamResponse[]>([])
  const [salesUsers, setSalesUsers] = useState<UserResponse[]>([])

  useEffect(() => {
    if (!showTeamFilter) return
    fetchTeams()
      .then(setTeams)
      .catch(() => {
        // An empty dropdown on failure is an acceptable degradation.
      })
  }, [showTeamFilter])

  useEffect(() => {
    if (!showSalesWidget) return
    fetchUsers(LEAD_OWNER_ROLES)
      .then(setSalesUsers)
      .catch(() => {
        // Same degrade-quietly approach.
      })
  }, [showSalesWidget])

  const today = new Date()
  let dateFrom: Date
  let dateTo: Date
  switch (rangePreset) {
    case 'today':
      dateFrom = today
      dateTo = today
      break
    case 'week':
      dateFrom = startOfWeekMonday(today)
      dateTo = today
      break
    case 'month':
      dateFrom = startOfMonth(today)
      dateTo = today
      break
    case 'custom':
      dateFrom = customRange?.from ?? today
      dateTo = customRange?.to ?? customRange?.from ?? today
      break
  }

  const rangeDays = daysBetweenInclusive(dateFrom, dateTo)
  const customIncomplete = rangePreset === 'custom' && (!customRange?.from || !customRange?.to)
  const customTooLong = rangePreset === 'custom' && rangeDays > MAX_CUSTOM_RANGE_DAYS
  const rangeValid = !customIncomplete && !customTooLong

  const params = {
    date_from: toLocalDateString(dateFrom),
    date_to: toLocalDateString(dateTo),
    team_id: teamId || undefined,
    owner_id: ownerId || undefined,
  }
  // Only Summary and SalesActivity read `status` server-side (see dashboard_service.go) -
  // the other three ignore it, so they stay on `params` to avoid refetching on every status change.
  const paramsWithStatus = { ...params, status: status || undefined }

  const summary = useDashboardSummary(paramsWithStatus, rangeValid)
  const activity = useDashboardActivity(params, rangeValid)
  const staleLeads = useDashboardStaleLeads(params, rangeValid)
  const salesActivity = useDashboardSalesActivity(paramsWithStatus, showSalesWidget && rangeValid)
  const segments = useDashboardSegments(params, rangeValid)

  const scopeLabel =
    user?.role === 'SALES'
      ? 'Ringkasan aktivitas Anda — dihitung dari lead & follow-up.'
      : user?.role === 'LEADER'
        ? `Ringkasan ${user.team_name ?? 'tim Anda'} — dapat dirinci per sales.`
        : 'Ringkasan semua tim — dapat dirinci per tim lalu per sales.'

  return (
    <div className="flex w-full max-w-full flex-col gap-4 p-4 lg:p-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="font-display text-xl font-extrabold lg:text-[28px]">Dashboard</h1>
          <p className="mt-1 text-sm text-muted-foreground">{scopeLabel}</p>
        </div>
        <div className="flex flex-wrap items-end gap-2.5">
          <div className="flex flex-col gap-1">
            <label className={FILTER_LABEL_CLASSNAME}>Rentang</label>
            <div className="flex items-center gap-1.5">
              {RANGE_PRESET_OPTIONS.map((o) => (
                <button
                  key={o.value}
                  type="button"
                  onClick={() => setRangePreset(o.value)}
                  className={cn(
                    PRESET_BUTTON_BASE,
                    rangePreset === o.value
                      ? 'border-[#1D4ED8] bg-[#EEF3FC] text-[#1D4ED8]'
                      : 'border-input bg-white text-[#334155] hover:border-[#BDD0F7]'
                  )}
                >
                  {o.label}
                </button>
              ))}
              {rangePreset === 'custom' && (
                <Popover>
                  <PopoverTrigger className={cn(SELECT_CLASSNAME, 'inline-flex items-center gap-1.5 text-left whitespace-nowrap')}>
                    <CalendarIcon className="size-3.5 shrink-0 text-muted-foreground" />
                    <span className={cn(!customRange?.from && !customRange?.to && 'text-muted-foreground')}>
                      {formatCustomRangeLabel(customRange)}
                    </span>
                  </PopoverTrigger>
                  <PopoverContent className="w-auto p-0">
                    <Calendar mode="range" selected={customRange} onSelect={setCustomRange} defaultMonth={customRange?.from} numberOfMonths={1} />
                  </PopoverContent>
                </Popover>
              )}
            </div>
            {customTooLong && (
              <p className="text-[11px] font-semibold text-[#B91C1C]">Rentang custom maksimal {MAX_CUSTOM_RANGE_DAYS} hari.</p>
            )}
          </div>
          <div className="flex flex-col gap-1">
            <label className={FILTER_LABEL_CLASSNAME}>Status (forecast)</label>
            <select value={status} onChange={(e) => setStatus(e.target.value)} className={SELECT_CLASSNAME}>
              <option value="">Semua status</option>
              {STATUS_FILTER_OPTIONS.map((value) => (
                <option key={value} value={value}>
                  {STATUS_CONFIG[value].label}
                </option>
              ))}
            </select>
          </div>
          {showTeamFilter && (
            <div className="flex flex-col gap-1">
              <label className={FILTER_LABEL_CLASSNAME}>Tim</label>
              <select value={teamId} onChange={(e) => setTeamId(e.target.value)} className={SELECT_CLASSNAME}>
                <option value="">Semua tim</option>
                {teams.map((t, i) => (
                  <option key={t.id ?? i} value={t.id ?? ''}>
                    {t.name}
                  </option>
                ))}
              </select>
            </div>
          )}
          {showSalesWidget && (
            <div className="flex flex-col gap-1">
              <label className={FILTER_LABEL_CLASSNAME}>Sales</label>
              <select value={ownerId} onChange={(e) => setOwnerId(e.target.value)} className={SELECT_CLASSNAME}>
                <option value="">Semua sales</option>
                {salesUsers.map((u, i) => (
                  <option key={u.id ?? i} value={u.id ?? ''}>
                    {u.name} ({roleLabel(u.role)})
                  </option>
                ))}
              </select>
            </div>
          )}
        </div>
      </div>

      <MetricCards summary={summary.data} loading={summary.isLoading} compareLabel={`vs ${rangeDays} hari sebelumnya`} />

      <ForecastCard
        summary={summary.data}
        loading={summary.isLoading}
        statusLabel={status ? (STATUS_CONFIG[status]?.label ?? status) : 'Semua status'}
      />

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-[360px_1fr]">
        <FunnelCard summary={summary.data} loading={summary.isLoading} />
        <TrendChart activity={activity.data} loading={activity.isLoading} />
      </div>

      <StaleLeadsCard data={staleLeads.data} loading={staleLeads.isLoading} showOwner={showSalesWidget} />

      {showSalesWidget && <SalesActivityCard data={salesActivity.data} loading={salesActivity.isLoading} showTeamColumn={isAdmin} />}

      <SegmentsSection data={segments.data} loading={segments.isLoading} />
    </div>
  )
}
