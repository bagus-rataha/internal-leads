// Metric cards + conversion funnel, both sourced from
// GET /dashboard/summary. The backend only sends {value, change_pct} per
// card (ARCHITECTURE.md's zero-state contract) - this file owns the exact
// three-way wording ("—" / "baru" / arrow+percent), never computes a raw
// percentage itself.
import { AlertTriangle } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'
import type { DashboardSummaryResponse } from './api'

function metricDelta(value: number, changePct: number | null | undefined, goodUp: boolean) {
  if (changePct === null || changePct === undefined) {
    return value === 0
      ? { text: '—', className: 'text-[#94A3B8] bg-[#F1F5F9]' }
      : { text: 'baru', className: 'text-[#1D4ED8] bg-[#EEF3FC]' }
  }
  if (changePct === 0) {
    return { text: '→ 0%', className: 'text-[#64748B] bg-[#EEF2F7]' }
  }
  const isIncrease = changePct > 0
  const good = goodUp ? isIncrease : !isIncrease
  return {
    text: `${isIncrease ? '↑' : '↓'} ${Math.abs(changePct)}%`,
    className: good ? 'text-[#166534] bg-[#DCFCE7]' : 'text-[#B91C1C] bg-[#FEE2E2]',
  }
}

function MetricCard({
  label,
  value,
  changePct,
  goodUp,
  compareLabel,
  onClick,
  danger,
}: {
  label: string
  value: number
  changePct: number | null | undefined
  goodUp: boolean
  compareLabel: string
  onClick: () => void
  danger?: boolean
}) {
  const delta = metricDelta(value, changePct, goodUp)
  return (
    <button
      onClick={onClick}
      className={cn(
        'rounded-[15px] border p-4 text-left shadow-[0_1px_2px_rgba(15,23,42,0.04)] transition-colors',
        danger ? 'border-[#FCE39A] bg-[#FEF9EC]' : 'border-[#E7EDF3] bg-white hover:border-[#BDD0F7]'
      )}
    >
      <div className="mb-3 flex items-center justify-between gap-2">
        <span className={cn('flex items-center gap-1.5 text-[12.5px] font-semibold', danger ? 'text-[#92580A]' : 'text-[#64748B]')}>
          {danger && <AlertTriangle className="size-3.5" />}
          {label}
        </span>
        <span className={cn('rounded-full px-2.5 py-0.5 text-[11px] font-bold whitespace-nowrap', delta.className)}>{delta.text}</span>
      </div>
      <div
        className={cn(
          'font-display text-[30px] leading-none font-extrabold',
          danger ? (value > 0 ? 'text-[#B45309]' : 'text-[#166534]') : 'text-[#0F172A]'
        )}
      >
        {value}
      </div>
      <div className={cn('mt-[7px] text-[11px]', danger ? 'font-semibold text-[#B45309]' : 'text-[#94A3B8]')}>{compareLabel}</div>
    </button>
  )
}

export function MetricCards({
  summary,
  loading,
  compareLabel,
}: {
  summary?: DashboardSummaryResponse
  loading: boolean
  compareLabel: string
}) {
  const navigate = useNavigate()

  if (loading || !summary) {
    return (
      <div className="grid grid-cols-2 gap-3.5 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="h-[110px] animate-pulse rounded-[15px] bg-muted" />
        ))}
      </div>
    )
  }

  return (
    <div className="grid grid-cols-2 gap-3.5 lg:grid-cols-4">
      <MetricCard
        label="Lead Baru"
        value={summary.lead_baru?.value ?? 0}
        changePct={summary.lead_baru?.change_pct}
        goodUp
        compareLabel={compareLabel}
        onClick={() => navigate('/leads?status=BARU')}
      />
      <MetricCard
        label="Follow-up Ditulis"
        value={summary.follow_up?.value ?? 0}
        changePct={summary.follow_up?.change_pct}
        goodUp
        compareLabel={compareLabel}
        onClick={() => navigate('/leads')}
      />
      <MetricCard
        label="Handoff ke Odoo"
        value={summary.handoff?.value ?? 0}
        changePct={summary.handoff?.change_pct}
        goodUp
        compareLabel={compareLabel}
        onClick={() => navigate('/leads?status=HANDOFF_ODOO')}
      />
      <MetricCard
        label="Lead Terlantar"
        value={summary.terlantar?.value ?? 0}
        changePct={summary.terlantar?.change_pct}
        goodUp={false}
        compareLabel="Aktif tanpa follow-up >7 hari"
        onClick={() => navigate('/leads?stale=true')}
        danger
      />
    </div>
  )
}

const FUNNEL_SHADE = ['#1D4ED8', '#4F74E0', '#8FA9EE']

export function FunnelCard({ summary, loading }: { summary?: DashboardSummaryResponse; loading: boolean }) {
  if (loading || !summary) {
    return <div className="h-[260px] animate-pulse rounded-[16px] bg-muted" />
  }
  const stages = summary.funnel?.stages ?? []

  return (
    <Card className="overflow-hidden rounded-[16px] border-[#E7EDF3] shadow-[0_1px_2px_rgba(15,23,42,0.04)]">
      <CardContent className="p-0">
        <div className="px-[18px] pt-[15px] pb-1">
          <p className="font-display text-[15px] font-bold">Funnel Konversi</p>
          <p className="mt-0.5 text-[11.5px] text-[#94A3B8]">Baru → Follow-up → Handoff Odoo</p>
        </div>
        <div className="px-[18px] pt-[14px] pb-[18px]">
          {stages.map((stage, i) => (
            <div key={stage.name} className="mb-[11px]">
              <div className="mb-[5px] flex items-baseline justify-between text-xs">
                <span className="font-semibold text-[#475569]">{stage.name}</span>
                <span className="whitespace-nowrap">
                  <strong className="font-mono text-[13px] font-bold text-[#0F172A]">{stage.count}</strong>{' '}
                  <span className="font-mono text-[11.5px] text-[#94A3B8]">· {stage.pct ?? '—'}%</span>
                </span>
              </div>
              <div className="h-6 overflow-hidden rounded-[9px] bg-[#EEF2F7]">
                <div
                  className="h-full rounded-[9px] transition-[width]"
                  style={{ width: `${stage.pct ?? 0}%`, background: FUNNEL_SHADE[i] ?? FUNNEL_SHADE[2] }}
                />
              </div>
            </div>
          ))}
          <div className="mt-3.5 flex flex-wrap gap-2 border-t border-t-[#F1F5F9] pt-3.5 text-[11.5px]">
            <span className="rounded-lg border border-[#D3E0F7] bg-[#EEF3FC] px-2.5 py-1 text-[#475569]">
              Baru→FU <strong className="text-[#1D4ED8]">{summary.funnel?.baru_to_fu_pct ?? '—'}%</strong>
            </span>
            <span className="rounded-lg border border-[#D3E0F7] bg-[#EEF3FC] px-2.5 py-1 text-[#475569]">
              FU→Handoff <strong className="text-[#1D4ED8]">{summary.funnel?.fu_to_handoff_pct ?? '—'}%</strong>
            </span>
            <span className="rounded-lg border border-[#FECACA] bg-[#FEE2E2] px-2.5 py-1 text-[#B91C1C]">
              Lost (keluar) <strong>{summary.funnel?.lost_count ?? 0}</strong> · {summary.funnel?.lost_pct ?? '—'}%
            </span>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
