// Dashboard screen: filters (Rentang/Tim/Sales) drive every widget below via
// shared {date_from,date_to,team_id,owner_id} params - the funnel and the
// segments breakdowns ignore date_from/date_to internally (see their backend
// implementations) but still respect team_id/owner_id, which flow through
// the same params object.
import { useEffect, useState } from 'react'
import { useAuth } from '@/auth/AuthContext'
import { roleLabel } from '@/lib/roles'
import { fetchTeams, type TeamResponse } from '@/features/team/api'
import { fetchUsers, LEAD_OWNER_ROLES, type UserResponse } from '@/features/user/api'
import { STATUS_CONFIG } from '@/features/lead/shared'
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

const RANGE_OPTIONS = [
  { value: 7, label: '7 hari terakhir' },
  { value: 30, label: '30 hari terakhir' },
  { value: 90, label: '90 hari terakhir' },
] as const

const SELECT_CLASSNAME =
  'h-9 rounded-lg border border-input bg-white px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50'
const FILTER_LABEL_CLASSNAME = 'text-[10.5px] font-bold tracking-[.05em] text-[#94A3B8] uppercase'

export default function DashboardPage() {
  const { user } = useAuth()
  const isAdmin = user?.role === 'ADMIN_SALES' || user?.role === 'SU'
  const showSalesWidget = user?.role !== 'SALES'
  const showTeamFilter = isAdmin

  const [rangeDays, setRangeDays] = useState(30)
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

  const dateTo = new Date()
  const dateFrom = new Date(dateTo)
  dateFrom.setDate(dateFrom.getDate() - (rangeDays - 1))

  const params = {
    date_from: toLocalDateString(dateFrom),
    date_to: toLocalDateString(dateTo),
    team_id: teamId || undefined,
    owner_id: ownerId || undefined,
    status: status || undefined,
  }

  const summary = useDashboardSummary(params)
  const activity = useDashboardActivity(params)
  const staleLeads = useDashboardStaleLeads(params)
  const salesActivity = useDashboardSalesActivity(params, showSalesWidget)
  const segments = useDashboardSegments(params)

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
            <select value={rangeDays} onChange={(e) => setRangeDays(Number(e.target.value))} className={SELECT_CLASSNAME}>
              {RANGE_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>
                  {o.label}
                </option>
              ))}
            </select>
          </div>
          <div className="flex flex-col gap-1">
            <label className={FILTER_LABEL_CLASSNAME}>Status (forecast)</label>
            <select value={status} onChange={(e) => setStatus(e.target.value)} className={SELECT_CLASSNAME}>
              <option value="">Semua status</option>
              {Object.entries(STATUS_CONFIG).map(([value, cfg]) => (
                <option key={value} value={value}>
                  {cfg.label}
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
