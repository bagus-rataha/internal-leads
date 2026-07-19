// Daily lead-baru vs follow-up counts per day across the selected
// range. Recharts chart - no hand-rolled SVG.
import { ComposedChart, Area, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import { Card, CardContent } from '@/components/ui/card'
import type { DashboardActivityResponse } from './api'

const DATE_LABEL_FORMAT = new Intl.DateTimeFormat('id', { day: 'numeric', month: 'short' })

function formatBucketLabel(isoDate: string): string {
  return DATE_LABEL_FORMAT.format(new Date(isoDate + 'T00:00:00'))
}

export function TrendChart({ activity, loading }: { activity?: DashboardActivityResponse; loading: boolean }) {
  const buckets = (activity?.buckets ?? []).map((b) => ({ ...b, label: formatBucketLabel(b.date ?? '') }))

  return (
    <Card className="overflow-hidden rounded-[16px] border-[#E7EDF3] shadow-[0_1px_2px_rgba(15,23,42,0.04)]">
      <CardContent className="p-0">
        <div className="flex flex-wrap items-start justify-between gap-3 px-[18px] pt-[15px] pb-1">
          <div>
            <p className="font-display text-[15px] font-bold">Tren Aktivitas</p>
            <p className="mt-0.5 text-[11.5px] text-[#94A3B8]">Ritme kerja — lead baru vs follow-up sepanjang periode</p>
          </div>
          <div className="flex gap-3.5">
            <span className="inline-flex items-center gap-1.5 text-[11.5px] font-semibold text-[#475569]">
              <span className="h-[3px] w-3 rounded-sm bg-[#1D4ED8]" />
              Lead baru
            </span>
            <span className="inline-flex items-center gap-1.5 text-[11.5px] font-semibold text-[#475569]">
              <span className="h-[3px] w-3 rounded-sm bg-[#F59E0B]" />
              Follow-up
            </span>
          </div>
        </div>
        <div className="h-[220px] px-3 pb-4">
          {loading ? (
            <div className="h-full w-full animate-pulse rounded-lg bg-muted" />
          ) : (
            <ResponsiveContainer width="100%" height="100%">
              <ComposedChart data={buckets} margin={{ top: 10, right: 12, left: 0, bottom: 0 }}>
                <CartesianGrid stroke="#F1F5F9" vertical={false} />
                <XAxis
                  dataKey="label"
                  tick={{ fontSize: 10, fill: '#94A3B8' }}
                  interval="preserveStartEnd"
                  axisLine={{ stroke: '#E2E8F0' }}
                  tickLine={false}
                />
                <YAxis tick={{ fontSize: 10, fill: '#B4C0CE' }} axisLine={false} tickLine={false} allowDecimals={false} />
                <Tooltip
                  contentStyle={{ background: '#0F172A', border: 'none', borderRadius: 9, color: '#fff', fontSize: 12 }}
                  labelStyle={{ color: '#94A3B8', fontSize: 10.5 }}
                />
                <Area type="monotone" dataKey="lead_baru" stroke="none" fill="rgba(29,78,216,.08)" isAnimationActive={false} />
                <Line
                  type="monotone"
                  dataKey="lead_baru"
                  name="Lead baru"
                  stroke="#1D4ED8"
                  strokeWidth={2.5}
                  dot={{ r: 3.2, fill: '#1D4ED8', stroke: '#fff', strokeWidth: 1.5 }}
                />
                <Line
                  type="monotone"
                  dataKey="follow_up"
                  name="Follow-up"
                  stroke="#F59E0B"
                  strokeWidth={2.5}
                  dot={{ r: 3.2, fill: '#F59E0B', stroke: '#fff', strokeWidth: 1.5 }}
                />
              </ComposedChart>
            </ResponsiveContainer>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
