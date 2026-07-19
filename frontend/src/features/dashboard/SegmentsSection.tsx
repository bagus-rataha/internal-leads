// Lead source, region penetration, competitor intel, and business field breakdowns,
// all from one GET /dashboard/segments response.
import { useState } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import type { DashboardSegmentsResponse } from './api'

const RUPIAH_FORMAT = new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 })
function formatRupiah(v: number | null | undefined): string {
  return v == null ? '—' : RUPIAH_FORMAT.format(v)
}

function convColor(pct: number | null | undefined): string {
  if (pct === null || pct === undefined) return '#94A3B8'
  return pct >= 40 ? '#166534' : pct >= 20 ? '#1D4ED8' : '#B91C1C'
}

function SegmentTable({
  title,
  subtitle,
  colLabel,
  rows,
  convNote,
}: {
  title: string
  subtitle: string
  colLabel: string
  rows: { name: string; count: number; conversion_pct?: number | null; warn?: boolean }[]
  convNote: boolean
}) {
  return (
    <Card className="overflow-hidden rounded-[16px] border-[#E7EDF3] shadow-[0_1px_2px_rgba(15,23,42,0.04)]">
      <CardContent className="p-0">
        <div className="px-[18px] pt-[15px] pb-1.5">
          <p className="font-display text-[15px] font-bold">{title}</p>
          <p className="mt-0.5 text-[11.5px] text-[#94A3B8]">{subtitle}</p>
        </div>
        <table className="w-full border-collapse">
          <thead>
            <tr>
              <th className="px-[18px] py-1.5 text-left text-[10px] font-bold tracking-[.05em] text-[#94A3B8] uppercase">{colLabel}</th>
              <th className="px-2 py-1.5 text-right text-[10px] font-bold tracking-[.05em] text-[#94A3B8] uppercase">Lead</th>
              <th className="px-[18px] py-1.5 text-right text-[10px] font-bold tracking-[.05em] text-[#94A3B8] uppercase">Konversi</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.name}>
                <td className="border-t border-t-[#F1F5F9] px-[18px] py-2.5 text-[12.5px] font-medium text-[#334155]">
                  <span className="inline-flex items-center gap-1.5">
                    {row.name}
                    {row.warn && (
                      <span className="rounded-full border border-[#FCE39A] bg-[#FEF3C7] px-1.5 py-px text-[9px] font-bold text-[#92580A]">
                        konversi rendah
                      </span>
                    )}
                  </span>
                </td>
                <td className="border-t border-t-[#F1F5F9] px-2 py-2.5 text-right font-mono text-[12.5px] font-semibold text-[#0F172A]">{row.count}</td>
                <td className="border-t border-t-[#F1F5F9] px-[18px] py-2.5 text-right font-mono text-[12.5px] font-semibold" style={{ color: convColor(row.conversion_pct) }}>
                  {row.conversion_pct ?? '—'}{row.conversion_pct !== null && row.conversion_pct !== undefined ? '%' : ''}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {convNote && (
          <div className="border-t border-t-[#F1F5F9] px-[18px] py-2.5 text-[11px] text-[#94A3B8]">
            Konversi belum tersedia — belum ada handoff dalam cakupan ini.
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function ExpandableTable({
  title,
  subtitle,
  rows,
}: {
  title: string
  subtitle: string
  rows: { name: string; count: number }[]
}) {
  const [expanded, setExpanded] = useState(false)
  const visible = expanded ? rows : rows.slice(0, 5)
  const rest = rows.slice(5)

  return (
    <Card className="overflow-hidden rounded-[16px] border-[#E7EDF3] shadow-[0_1px_2px_rgba(15,23,42,0.04)]">
      <CardContent className="p-0">
        <div className="px-[18px] pt-[15px] pb-1.5">
          <p className="font-display text-[15px] font-bold">{title}</p>
          <p className="mt-0.5 text-[11.5px] text-[#94A3B8]">{subtitle}</p>
        </div>
        <table className="mt-1.5 w-full border-collapse">
          <tbody>
            {visible.map((row) => (
              <tr key={row.name}>
                <td className="border-t border-t-[#F1F5F9] px-[18px] py-2 text-[12.5px] font-medium text-[#334155]">{row.name}</td>
                <td className="border-t border-t-[#F1F5F9] px-[18px] py-2 text-right font-mono text-[12.5px] font-semibold text-[#0F172A]">{row.count}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {rest.length > 0 && (
          <div
            className="cursor-pointer border-t border-t-[#F1F5F9] px-[18px] py-2.5 text-[11.5px] font-semibold text-[#1D4ED8] hover:bg-[#F5F8FE]"
            onClick={() => setExpanded((v) => !v)}
          >
            {expanded ? 'Tampilkan lebih sedikit' : `+ ${rest.length} lainnya (${rest.reduce((a, r) => a + r.count, 0)} lead)`}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export function SegmentsSection({ data, loading }: { data?: DashboardSegmentsResponse; loading: boolean }) {
  if (loading || !data) {
    return (
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="h-[220px] animate-pulse rounded-[16px] bg-muted" />
        ))}
      </div>
    )
  }

  const convNote = !data.any_handoff
  const regions = data.regions ?? []

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-4">
        <div className="shrink-0">
          <p className="font-display text-[17px] font-extrabold text-[#0F172A]">Segmen &amp; pasar</p>
          <p className="mt-px text-xs text-[#94A3B8]">Untuk evaluasi berkala — bukan kerja harian.</p>
        </div>
        <div className="h-px flex-1 bg-[#E2E8F0]" />
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <SegmentTable title="Sumber Lead" subtitle="Volume tinggi + konversi rendah = masalah, bukan prestasi" colLabel="Sumber" rows={(data.sources ?? []) as { name: string; count: number; conversion_pct?: number | null; warn?: boolean }[]} convNote={convNote} />
        <ExpandableTable title="Bidang Usaha" subtitle="Segmen paling responsif · 5 teratas" rows={(data.business_fields ?? []) as { name: string; count: number }[]} />
      </div>

      <div className="grid grid-cols-1 items-start gap-4 lg:grid-cols-2">
        <Card className="overflow-hidden rounded-[16px] border-[#E7EDF3] shadow-[0_1px_2px_rgba(15,23,42,0.04)]">
          <CardContent className="p-0">
            <div className="px-[18px] pt-[15px] pb-2">
              <p className="font-display text-[15px] font-bold">Intel Kompetitor</p>
              <p className="mt-0.5 text-[11.5px] text-[#94A3B8]">Harga rata-rata per Mbps &amp; ISP eksisting calon pelanggan</p>
            </div>
            <div className="grid grid-cols-3 gap-2.5 px-[18px] pb-2">
              {[
                ['Rata-rata / Mbps', data.competitor?.avg_price_per_mbps],
                ['Dedicated / Mbps', data.competitor?.avg_price_per_mbps_dedicated],
                ['Broadband / Mbps', data.competitor?.avg_price_per_mbps_broadband],
              ].map(([label, value]) => (
                <div key={label as string} className="rounded-[11px] border border-[#EEF2F7] bg-[#F7F9FC] p-2.5">
                  <div className="text-[10.5px] font-semibold text-[#94A3B8]">{label}</div>
                  <div className="mt-[3px] font-mono text-[15px] font-bold text-[#0F172A]">{formatRupiah(value as number | null | undefined)}</div>
                </div>
              ))}
            </div>
            <ExpandableTable title="" subtitle="" rows={(data.isps ?? []) as { name: string; count: number }[]} />
          </CardContent>
        </Card>

        <Card className="overflow-hidden rounded-[16px] border-[#E7EDF3] shadow-[0_1px_2px_rgba(15,23,42,0.04)]">
          <CardContent className="p-0">
            <div className="px-[18px] pt-[15px] pb-2">
              <p className="font-display text-[15px] font-bold">Penetrasi Wilayah</p>
              <p className="mt-0.5 text-[11.5px] text-[#94A3B8]">Provinsi → kota · lead &amp; handoff</p>
            </div>
            <table className="w-full border-collapse">
              <thead>
                <tr>
                  <th className="px-[18px] py-1.5 text-left text-[10px] font-bold tracking-[.05em] text-[#94A3B8] uppercase">Wilayah</th>
                  <th className="px-2 py-1.5 text-right text-[10px] font-bold tracking-[.05em] text-[#94A3B8] uppercase">Lead</th>
                  <th className="px-[18px] py-1.5 text-right text-[10px] font-bold tracking-[.05em] text-[#94A3B8] uppercase">Handoff</th>
                </tr>
              </thead>
              <tbody>
                {regions.map((row, i) => (
                  <tr key={`${row.level}-${row.name}-${i}`} className={row.level === 'province' ? 'bg-[#F7F9FC]' : ''}>
                    <td
                      className="border-t border-t-[#F1F5F9] px-[18px] py-2"
                      style={row.level === 'city' ? { paddingLeft: 32, fontSize: 12.5, color: '#475569', fontWeight: 500 } : { fontFamily: "'Plus Jakarta Sans'", fontSize: 13.5, fontWeight: 700, color: '#0F172A' }}
                    >
                      {row.name}
                    </td>
                    <td className="border-t border-t-[#F1F5F9] px-2 py-2 text-right font-mono text-[12.5px] font-semibold text-[#334155]">{row.lead_count}</td>
                    <td className="border-t border-t-[#F1F5F9] px-[18px] py-2 text-right font-mono text-[12.5px] font-semibold" style={{ color: (row.handoff_count ?? 0) > 0 ? '#166534' : '#CBD5E1' }}>
                      {row.handoff_count}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
