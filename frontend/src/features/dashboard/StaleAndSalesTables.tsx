// Lead Terlantar + Keaktifan Sales - both simple
// row-table cards, kept in one file since they're visually and structurally
// similar (small standalone tables, not shared state).
import { AlertTriangle, CheckCircle2 } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'
import type { DashboardStaleLeadsResponse, DashboardSalesActivityResponse } from './api'

const TH = 'px-3.5 py-2 text-left text-[10.5px] font-bold tracking-[.05em] text-[#94A3B8] uppercase'
const TH_R = TH.replace('text-left', 'text-right')

export function StaleLeadsCard({
  data,
  loading,
  showOwner,
}: {
  data?: DashboardStaleLeadsResponse
  loading: boolean
  showOwner: boolean
}) {
  const navigate = useNavigate()

  return (
    <Card className="overflow-hidden rounded-[16px] border-[#E7EDF3] shadow-[0_1px_2px_rgba(15,23,42,0.04)]">
      <CardContent className="p-0">
        <div className="flex items-center gap-2 px-[18px] pt-[15px] pb-1">
          <AlertTriangle className="size-4 text-[#B45309]" />
          <p className="font-display text-[15px] font-bold">Lead Terlantar</p>
        </div>
        <p className="px-[18px] pb-2 text-[11.5px] text-[#94A3B8]">Aktif tanpa follow-up terlama · klik baris untuk buka Detail Lead</p>

        {loading ? (
          <div className="mx-[18px] mb-4 h-24 animate-pulse rounded-lg bg-muted" />
        ) : !data || !data.items || data.items.length === 0 ? (
          <div className="flex items-center gap-2 px-[18px] pt-1 pb-6 text-[13px] text-[#94A3B8]">
            <CheckCircle2 className="size-[17px] text-[#16A34A]" />
            Tidak ada lead terlantar dalam cakupan ini. Kerja bagus.
          </div>
        ) : (
          <div className="overflow-x-auto px-2 pb-2">
            <table className="w-full min-w-[520px] border-collapse">
              <thead>
                <tr>
                  <th className={TH}>Kode</th>
                  <th className={TH}>Perusahaan</th>
                  {showOwner && <th className={TH}>Sales</th>}
                  <th className={TH_R}>Sejak Follow-up</th>
                </tr>
              </thead>
              <tbody>
                {data.items.map((row) => (
                  <tr
                    key={row.code}
                    className="cursor-pointer hover:bg-[#FFFCF5]"
                    onClick={() => navigate(`/leads/${row.code}`)}
                  >
                    <td className="border-t border-t-[#F1F5F9] px-3.5 py-2.5 font-mono text-xs font-semibold whitespace-nowrap text-[#1D4ED8]">
                      {row.code}
                    </td>
                    <td className="border-t border-t-[#F1F5F9] px-3.5 py-2.5 text-[13px] font-semibold text-[#0F172A]">{row.company_name}</td>
                    {showOwner && <td className="border-t border-t-[#F1F5F9] px-3.5 py-2.5 text-[12.5px] text-[#64748B]">{row.owner_name}</td>}
                    <td className="border-t border-t-[#F1F5F9] px-3.5 py-2.5 text-right whitespace-nowrap">
                      <span className={cn('font-mono text-[13px] font-bold', (row.days_since ?? 0) > 14 ? 'text-[#B91C1C]' : 'text-[#B45309]')}>
                        {row.days_since}
                      </span>{' '}
                      <span className="text-[11px] text-[#B45309]">hari</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export function SalesActivityCard({
  data,
  loading,
  showTeamColumn,
}: {
  data?: DashboardSalesActivityResponse
  loading: boolean
  showTeamColumn: boolean
}) {
  const navigate = useNavigate()

  return (
    <Card className="overflow-hidden rounded-[16px] border-[#E7EDF3] shadow-[0_1px_2px_rgba(15,23,42,0.04)]">
      <CardContent className="p-0">
        <div className="px-[18px] pt-[15px] pb-1">
          <p className="font-display text-[15px] font-bold">Keaktifan Sales</p>
          <p className="mt-0.5 text-[11.5px] text-[#94A3B8]">
            Untuk menemukan yang butuh bantuan — bukan peringkat. Baris disorot = perlu perhatian.
          </p>
        </div>
        {loading ? (
          <div className="mx-[18px] mb-4 h-32 animate-pulse rounded-lg bg-muted" />
        ) : (
          <div className="overflow-x-auto px-2 pt-3 pb-1.5">
            <table className="w-full min-w-[820px] border-collapse">
              <thead>
                <tr>
                  <th className={TH}>Sales</th>
                  {showTeamColumn && <th className={TH}>Tim</th>}
                  <th className={TH_R}>Lead Baru</th>
                  <th className={TH_R}>Follow-up</th>
                  <th className={TH_R}>Avg FU/Lead</th>
                  <th className={TH_R}>Terlantar</th>
                  <th className={TH_R}>Survey</th>
                  <th className={TH_R}>Forecast MRR</th>
                  <th className={TH_R}>Conv.</th>
                  <th className={TH}>Aktivitas Terakhir</th>
                </tr>
              </thead>
              <tbody>
                {(data?.items ?? []).map((row) => {
                  const initials = (row.name ?? '?')
                    .trim()
                    .split(/\s+/)
                    .map((p) => p[0])
                    .slice(0, 2)
                    .join('')
                    .toUpperCase()
                  return (
                    <tr
                      key={row.user_id}
                      className={cn('cursor-pointer hover:bg-[#F5F8FE]', row.attention_tag && 'bg-[#FFFCF5]')}
                      onClick={() => navigate(`/leads?owner_id=${row.user_id}`)}
                    >
                      <td className="border-t border-t-[#F1F5F9] px-3.5 py-2.5">
                        <div className="flex items-center gap-2.5">
                          <span className="flex size-7 shrink-0 items-center justify-center rounded-lg bg-[#EEF3FC] text-[11px] font-bold text-[#1D4ED8]">
                            {initials}
                          </span>
                          <div className="flex items-center gap-1.5 text-[13px] font-semibold text-[#0F172A]">
                            {row.name}
                            {row.attention_tag && (
                              <span className="rounded-full border border-[#FCE39A] bg-[#FEF3C7] px-1.5 py-px text-[9.5px] font-bold whitespace-nowrap text-[#92580A]">
                                {row.attention_tag}
                              </span>
                            )}
                          </div>
                        </div>
                      </td>
                      {showTeamColumn && <td className="border-t border-t-[#F1F5F9] px-3.5 py-2.5 text-xs text-[#64748B]">{row.team_name ?? '—'}</td>}
                      <td className="border-t border-t-[#F1F5F9] px-3.5 py-2.5 text-right font-mono text-[13px] text-[#334155]">{row.lead_baru}</td>
                      <td className="border-t border-t-[#F1F5F9] px-3.5 py-2.5 text-right font-mono text-[13px] text-[#334155]">{row.follow_up}</td>
                      <td className="border-t border-t-[#F1F5F9] px-3.5 py-2.5 text-right font-mono text-[13px] text-[#334155]">
                        {row.avg_fu_per_lead?.toFixed(1)}
                      </td>
                      <td
                        className={cn(
                          'border-t border-t-[#F1F5F9] px-3.5 py-2.5 text-right font-mono text-[13px] font-bold',
                          (row.terlantar ?? 0) > 0 ? 'text-[#B45309]' : 'text-[#94A3B8]'
                        )}
                      >
                        {row.terlantar}
                      </td>
                      <td className="border-t border-t-[#F1F5F9] px-3.5 py-2.5 text-right font-mono text-[13px] text-[#334155]">{row.survey}</td>
                      <td className="border-t border-t-[#F1F5F9] px-3.5 py-2.5 text-right font-mono text-[13px] text-[#334155]">
                        Rp {(row.forecast_mrr ?? 0).toLocaleString('id-ID')}
                      </td>
                      <td className="border-t border-t-[#F1F5F9] px-3.5 py-2.5 text-right font-mono text-[13px] text-[#334155]">
                        {row.conv_pct ?? '—'}{row.conv_pct !== null && row.conv_pct !== undefined ? '%' : ''}
                      </td>
                      <td className="border-t border-t-[#F1F5F9] px-3.5 py-2.5 text-[12.5px] text-[#64748B]">
                        {row.last_activity ? new Date(row.last_activity).toLocaleDateString('id') : '—'}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
