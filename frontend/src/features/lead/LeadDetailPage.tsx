// Detail Lead screen: header (code/status/owner), 5 left-column data cards,
// and a sticky right-column follow-up timeline + composer.
import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, ExternalLink, Building2, MapPin, User, Wifi, Tag, Check, X as XIcon, Pencil } from 'lucide-react'
import { roleLabel } from '@/lib/roles'
import {
  useLeadDetail,
  useFollowUps,
  useCreateFollowUp,
  useUpdateLeadStatus,
  useSalesRoster,
  useReassignOwner,
} from './queries'
import { StatusPill, formatRelativeTime, STATUS_CONFIG, nextStage, earlierStages } from './shared'
import type { FollowUpResponse, UpdateLeadStatusInput } from './api'

// nextStage()/earlierStages() only ever return valid forward/backward pipeline
// stages, never BARU/HANDOFF_ODOO — so the cast to the mutation's status type is sound.
type PipelineStatus = UpdateLeadStatusInput['status']
import { useState } from 'react'
import { Modal } from '@/components/ui/modal'
import { showToast } from '@/hooks/useToast'
import { useAuth } from '@/auth/AuthContext'

const DT_ROW = 'flex items-start justify-between gap-4 border-b border-[#F5F8FC] py-[9px]'
const DT_ROW_LAST = 'flex items-start justify-between gap-4 py-[9px]'
const DT_K = 'pt-px text-[12.5px] text-[#94A3B8] shrink-0'
const DT_V = 'text-right text-[13.5px] font-medium text-[#334155]'

function DataRow({ label, value, last = false }: { label: string; value: React.ReactNode; last?: boolean }) {
  return (
    <div className={last ? DT_ROW_LAST : DT_ROW}>
      <span className={DT_K}>{label}</span>
      <span className={DT_V}>{value}</span>
    </div>
  )
}

function SectionCard({
  icon,
  title,
  children,
}: {
  icon: React.ReactNode
  title: string
  children: React.ReactNode
}) {
  return (
    <div className="overflow-hidden rounded-[16px] border border-[#E7EDF3] bg-white shadow-[0_1px_2px_rgba(15,23,42,0.04)]">
      <div className="flex items-center gap-[9px] border-b border-[#F1F5F9] px-[18px] py-[14px]">
        <span className="flex size-[28px] items-center justify-center rounded-[8px] bg-[#EEF3FC] text-[#1D4ED8]">
          {icon}
        </span>
        <span className="font-display text-[15px] font-bold">{title}</span>
      </div>
      <div className="px-[18px] py-[6px] pb-[14px]">{children}</div>
    </div>
  )
}

function TimelineEntry({ entry }: { entry: FollowUpResponse }) {
  const initials = (entry.created_by_name ?? '?')
    .trim()
    .split(/\s+/)
    .map((p) => p[0])
    .slice(0, 2)
    .join('')
    .toUpperCase()
  return (
    <div className="relative pb-[18px] pl-[34px]">
      <span className="absolute top-px left-0 flex size-[26px] items-center justify-center rounded-full border-2 border-white bg-[#EEF3FC] text-[9px] font-bold text-[#1D4ED8] shadow-[0_0_0_1px_#D3E0F7]">
        {initials}
      </span>
      <div className="mb-[3px] flex items-baseline justify-between gap-2.5">
        <span className="text-[12.5px] font-bold text-[#334155]">{entry.created_by_name}</span>
        <span className="shrink-0 font-mono text-[10.5px] text-[#94A3B8]">
          {entry.created_at ? formatRelativeTime(entry.created_at) : ''}
        </span>
      </div>
      <div className="rounded-[10px] border border-[#EEF2F7] bg-[#F7F9FC] px-3 py-[9px] text-[13px] leading-[1.55] text-[#475569]">
        {entry.note}
      </div>
    </div>
  )
}

export default function LeadDetailPage() {
  const { code } = useParams<{ code: string }>()
  const navigate = useNavigate()
  const { user } = useAuth()
  const { data: lead, isLoading, error } = useLeadDetail(code!)
  const { data: followUps } = useFollowUps(code!)
  const createFollowUp = useCreateFollowUp(code!)
  const updateStatus = useUpdateLeadStatus(code!)
  const [draft, setDraft] = useState('')
  const [modal, setModal] = useState<'advance' | 'lost' | null>(null)
  const [lostReason, setLostReason] = useState('')
  const [reassignOpen, setReassignOpen] = useState(false)
  const { data: salesRoster } = useSalesRoster(reassignOpen)
  const [reassignTo, setReassignTo] = useState('')
  const reassign = useReassignOwner(code!)

  if (isLoading) {
    return <div className="p-6 text-sm text-muted-foreground">Memuat…</div>
  }
  if (error || !lead) {
    return <div className="p-6 text-sm text-destructive">Lead tidak ditemukan.</div>
  }

  const isLocked = lead.status === 'INVOICE_BULANAN' || lead.status === 'LOST' || lead.status === 'HANDOFF_ODOO'
  const canEdit = lead.status === 'BARU' || lead.status === 'FOLLOW_UP'
  const canFollowUp = !isLocked
  const isAdmin = user?.role === 'ADMIN_SALES' || user?.role === 'SU'
  const canReassign = isAdmin && !isLocked
  const advanceTarget = nextStage(lead.status ?? '')
  const backTargets = isAdmin ? earlierStages(lead.status ?? '') : []
  const showCreatedBy = lead.created_by_id !== lead.owner_id
  const websiteHref = lead.website
    ? /^https?:\/\//.test(lead.website)
      ? lead.website
      : `https://${lead.website}`
    : undefined
  const timeline = [...(followUps ?? [])].reverse() // newest-first, so a new entry appears right below the composer without scrolling

  function handleSubmitFollowUp() {
    const note = draft.trim()
    if (!note) return
    createFollowUp.mutate(note, {
      onSuccess: () => {
        setDraft('')
        showToast('Follow-up tercatat')
      },
      onError: () => showToast('Gagal menyimpan follow-up. Coba lagi.'),
    })
  }

  function handleAdvanceConfirm() {
    if (!advanceTarget) return
    updateStatus.mutate(
      { status: advanceTarget as PipelineStatus },
      {
        onSuccess: () => {
          setModal(null)
          showToast(`Lead maju ke ${STATUS_CONFIG[advanceTarget].label}`)
        },
        onError: () => showToast('Gagal mengubah status. Coba lagi.'),
      }
    )
  }

  function handleReassignConfirm() {
    if (!reassignTo) return
    reassign.mutate(reassignTo, {
      onSuccess: () => {
        setReassignOpen(false)
        showToast(`Owner dipindahkan`)
      },
      onError: () => showToast('Gagal memindahkan owner. Coba lagi.'),
    })
  }

  function handleLostConfirm() {
    const reason = lostReason.trim()
    if (!reason) return
    updateStatus.mutate(
      { status: 'LOST', lost_reason: reason },
      {
        onSuccess: () => {
          setModal(null)
          setLostReason('')
          showToast('Lead ditandai Lost')
        },
        onError: () => showToast('Gagal mengubah status. Coba lagi.'),
      }
    )
  }

  return (
    <div className="mx-auto max-w-[1220px] px-[20px] pt-[20px] pb-[60px] lg:px-[30px]">
      <button
        type="button"
        onClick={() => navigate('/leads')}
        className="-ml-2 mb-[14px] inline-flex items-center gap-1.5 rounded-[7px] px-2 py-1.5 text-[13px] font-semibold text-[#64748B] hover:bg-[#EEF2F7] hover:text-[#1D4ED8]"
      >
        <ArrowLeft className="size-4" />
        Kembali ke daftar lead
      </button>

      <div className="mb-[18px] rounded-[16px] border border-[#E7EDF3] bg-white p-[22px_24px] shadow-[0_1px_2px_rgba(15,23,42,0.04)]">
        <div className="flex flex-col gap-4 lg:flex-row lg:flex-wrap lg:items-start lg:justify-between lg:gap-6">
          <div className="min-w-0">
            <div className="mb-[7px] flex items-center gap-3">
              <span className="rounded-[7px] bg-[#1D4ED8] px-[11px] py-1 font-mono text-[13px] font-semibold tracking-[.02em] text-white">
                {lead.code}
              </span>
              <StatusPill status={lead.status} />
            </div>
            <h1 className="font-display text-[26px] font-extrabold tracking-[-.02em] leading-[1.15]">
              {lead.company_name}
            </h1>
            <div className="mt-[9px] flex flex-wrap items-center gap-3.5">
              <span className="inline-flex items-center gap-[7px] text-[12.5px] text-[#64748B]">
                <span className="flex size-[22px] items-center justify-center rounded-[6px] bg-[#EEF2F7] text-[9.5px] font-bold text-[#475569]">
                  {(lead.owner_name ?? '?').charAt(0).toUpperCase()}
                </span>
                Owner: <strong className="font-semibold text-[#334155]">{lead.owner_name}</strong>
                {lead.owner_team_name && <> · {lead.owner_team_name}</>}
              </span>
              {showCreatedBy && (
                <span className="text-[12px] text-[#94A3B8]">
                  · Diinput oleh {lead.created_by_name} untuk {lead.owner_name}
                </span>
              )}
            </div>
          </div>

          <div className="flex flex-wrap gap-[9px] lg:shrink-0">
            {canEdit && (
              <button
                type="button"
                onClick={() => navigate(`/leads/${lead.code}/edit`)}
                className="inline-flex items-center gap-[7px] rounded-[8px] border border-[#CBD5E1] bg-white px-[14px] py-2 text-[13px] font-semibold text-[#334155] hover:bg-[#F7F9FC]"
              >
                <Pencil className="size-[15px]" />
                Edit Lead
              </button>
            )}
            {canReassign && (
              <button
                type="button"
                onClick={() => {
                  setReassignTo(lead.owner_id ?? '')
                  setReassignOpen(true)
                }}
                className="inline-flex items-center gap-[5px] rounded-[7px] border border-[#CBD5E1] bg-white px-[10px] py-[3px] text-[11.5px] font-semibold text-[#334155] hover:bg-[#F7F9FC]"
              >
                Reassign
              </button>
            )}
            {!isLocked && (
              <button
                type="button"
                onClick={() => setModal('lost')}
                className="inline-flex items-center gap-[7px] rounded-[8px] bg-[#FEE2E2] px-[14px] py-2 text-[13px] font-semibold text-[#B91C1C] hover:bg-[#FECACA]"
              >
                <XIcon className="size-[15px]" />
                Tandai Lost
              </button>
            )}
            {advanceTarget && !isLocked && (
              <button
                type="button"
                onClick={() => setModal('advance')}
                className="inline-flex items-center gap-[7px] rounded-[8px] bg-[#1D4ED8] px-[15px] py-2 text-[13px] font-semibold text-white shadow-[0_1px_2px_rgba(29,78,216,0.3)] hover:bg-[#1A45BE]"
              >
                <Check className="size-[15px]" />
                Maju ke {STATUS_CONFIG[advanceTarget].label}
              </button>
            )}
            {backTargets.length > 0 && (
              <select
                aria-label="Koreksi status mundur"
                defaultValue=""
                onChange={(e) => {
                  const v = e.target.value
                  if (!v) return
                  updateStatus.mutate(
                    { status: v as PipelineStatus },
                    {
                      onSuccess: () => showToast(`Status dikoreksi ke ${STATUS_CONFIG[v].label}`),
                      onError: () => showToast('Gagal mengoreksi status. Coba lagi.'),
                    }
                  )
                  e.target.value = ''
                }}
                className="rounded-[8px] border border-[#CBD5E1] bg-white px-[10px] py-2 text-[13px] font-semibold text-[#334155] hover:bg-[#F7F9FC]"
              >
                <option value="">Koreksi mundur…</option>
                {backTargets.map((s) => (
                  <option key={s} value={s}>{STATUS_CONFIG[s].label}</option>
                ))}
              </select>
            )}
          </div>
        </div>

        {(isLocked || lead.status === 'LOST') && (
          <div className={lead.status === 'LOST'
            ? 'mt-4 flex gap-3 rounded-[12px] border border-[#FECACA] bg-[#FEE2E2] p-[13px_15px] text-[#B91C1C]'
            : 'mt-4 flex gap-3 rounded-[12px] border border-[#BBF7D0] bg-[#DCFCE7] p-[13px_15px] text-[#166534]'}>
            <div>
              <div className="text-[13px] font-bold">
                {lead.status === 'LOST' ? 'Lead dinyatakan Lost'
                  : lead.status === 'INVOICE_BULANAN' ? 'Pelanggan aktif — invoice bulanan berjalan'
                  : 'Lead sudah di-handoff ke Odoo (legacy)'}
              </div>
              <div className="mt-0.5 text-[12px] opacity-85">
                {lead.status === 'LOST' ? (lead.lost_reason ?? '')
                  : lead.status === 'INVOICE_BULANAN' ? 'Kelola langganan di Odoo.'
                  : 'Kelola selanjutnya di Odoo.'}
              </div>
            </div>
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 items-start gap-[18px] lg:grid-cols-[1.35fr_1fr]">
        <div className="flex flex-col gap-4">
          <SectionCard icon={<Building2 className="size-4" strokeWidth={1.8} />} title="Data Perusahaan">
            <DataRow label="Nama Perusahaan" value={lead.company_name} />
            <DataRow label="Bidang Usaha" value={lead.business_field ?? '—'} />
            <DataRow
              label="Website"
              last
              value={
                lead.website ? (
                  <a
                    href={websiteHref}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex items-center gap-[5px] font-medium text-[#1D4ED8]"
                  >
                    {lead.website}
                    <ExternalLink className="size-3" />
                  </a>
                ) : (
                  '—'
                )
              }
            />
          </SectionCard>

          <SectionCard icon={<MapPin className="size-4" strokeWidth={1.8} />} title="Alamat">
            <DataRow label="Alamat Lengkap" value={lead.street ? `${lead.street}, RT ${lead.rt ?? '-'}/RW ${lead.rw ?? '-'}` : '—'} />
            <DataRow label="Kel. / Kec." value={`${lead.village_name ?? '—'} · ${lead.district_name ?? '—'}`} />
            <DataRow label="Kota / Provinsi" value={`${lead.city_name ?? '—'}, ${lead.province_name ?? '—'}`} />
            <DataRow label="Kode Pos" last value={<span className="font-mono">{lead.zip_code ?? '—'}</span>} />
          </SectionCard>

          <SectionCard icon={<User className="size-4" strokeWidth={1.8} />} title="PIC (Person in Charge)">
            <DataRow label="Nama · Jabatan" value={`${lead.pic_name ?? '—'} · ${lead.pic_position ?? '—'}`} />
            <DataRow label="Telepon Kantor" value={<span className="font-mono">{lead.office_phone ?? '—'}</span>} />
            <DataRow
              label="No. HP"
              value={
                lead.mobile_phone ? (
                  <a href={`tel:${lead.mobile_phone}`} className="font-mono font-medium text-[#1D4ED8]">
                    {lead.mobile_phone}
                  </a>
                ) : (
                  '—'
                )
              }
            />
            <DataRow
              label="Email"
              last
              value={
                lead.email ? (
                  <a href={`mailto:${lead.email}`} className="font-medium text-[#1D4ED8]">
                    {lead.email}
                  </a>
                ) : (
                  '—'
                )
              }
            />
          </SectionCard>

          <SectionCard icon={<Wifi className="size-4" strokeWidth={1.8} />} title="Layanan Eksisting">
            {lead.service_type_id ? (
              <>
                <DataRow label="Type Layanan" value={lead.service_type_name ?? '—'} />
                <DataRow label="Kapasitas" value={<span className="font-mono">{lead.capacity_mbps ? `${lead.capacity_mbps} Mbps` : '—'}</span>} />
                <DataRow label="ISP Eksisting" value={lead.existing_isp ?? '—'} />
                <DataRow label="Harga" value={<span className="font-mono">{lead.price ? `Rp ${lead.price.toLocaleString('id-ID')}` : '—'}</span>} />
                <DataRow label="Layanan Lainnya" last value={lead.other_services ?? '—'} />
              </>
            ) : (
              <p className="py-3 text-[13px] text-[#94A3B8]">
                Belum diisi — sering belum diketahui saat lead pertama masuk.
              </p>
            )}
          </SectionCard>

          <SectionCard icon={<Tag className="size-4" strokeWidth={1.8} />} title="Sumber Lead">
            <DataRow
              label="Sumber"
              value={
                <span className="rounded-full border border-[#D3E0F7] bg-[#EEF3FC] px-[11px] py-[3px] text-[12.5px] font-semibold text-[#1E3A8A]">
                  {lead.lead_source_name ?? '—'}
                </span>
              }
            />
            <DataRow label="Forecast MRR" value={<span className="font-mono">{lead.forecast_mrr ? `Rp ${lead.forecast_mrr.toLocaleString('id-ID')}` : '—'}</span>} />
            <DataRow
              label="Diinput"
              last
              value={`${lead.created_at ? new Date(lead.created_at).toLocaleDateString('id-ID') : '—'} · oleh ${lead.created_by_name}`}
            />
          </SectionCard>
        </div>

        <div className="lg:sticky lg:top-0">
          <div className="overflow-hidden rounded-[16px] border border-[#E7EDF3] bg-white shadow-[0_1px_2px_rgba(15,23,42,0.04)]">
            <div className="flex items-center justify-between border-b border-[#F1F5F9] px-[18px] py-[14px]">
              <span className="font-display text-[15px] font-bold">Riwayat Follow-up</span>
              <span className="rounded-full bg-[#EEF2F7] px-[9px] py-[3px] text-[11px] font-semibold text-[#64748B]">
                {followUps?.length ?? 0} entri
              </span>
            </div>

            {canFollowUp && (
              <div className="border-b border-[#F1F5F9] bg-[#FAFCFF] px-[18px] py-[14px]">
                <label className="mb-1.5 block text-[11.5px] font-bold text-[#334155]">Catat follow-up baru</label>
                <textarea
                  value={draft}
                  onChange={(e) => setDraft(e.target.value)}
                  placeholder="mis. Telepon PIC, kirim proposal Dedicated 100 Mbps, minta jadwal survey…"
                  rows={3}
                  className="w-full resize-y rounded-[9px] border border-[#CBD5E1] bg-white px-3 py-2.5 text-[13px] leading-[1.5] text-[#0F172A] outline-none focus:border-[#1D4ED8] focus:ring-[3px] focus:ring-[#DCE7FB]"
                />
                <div className="mt-[9px] flex items-center justify-between">
                  <span className="text-[11px] text-[#94A3B8]">Tanggal terisi otomatis saat disimpan</span>
                  <button
                    type="button"
                    onClick={handleSubmitFollowUp}
                    disabled={!draft.trim() || createFollowUp.isPending}
                    className="inline-flex items-center gap-1.5 rounded-[8px] px-[15px] py-2 text-[13px] font-semibold text-white disabled:cursor-not-allowed"
                    style={{ background: draft.trim() ? '#1D4ED8' : '#C7D2E4' }}
                  >
                    Simpan
                  </button>
                </div>
              </div>
            )}

            <div className="lsa-scroll max-h-[calc(100vh-340px)] overflow-y-auto px-[18px] py-4">
              {timeline.length === 0 ? (
                <div className="px-3 py-6 text-center text-[13px] text-[#94A3B8]">
                  Belum ada follow-up. Mulai dari kolom di atas.
                </div>
              ) : (
                timeline.map((entry) => <TimelineEntry key={entry.id} entry={entry} />)
              )}
            </div>
          </div>
        </div>
      </div>

      <Modal open={modal === 'advance'} onClose={() => setModal(null)}>
        <div className="flex items-start gap-[14px]">
          <span className="flex size-[44px] shrink-0 items-center justify-center rounded-[11px] bg-[#EEF3FC] text-[#1D4ED8]">
            <Check className="size-[22px]" />
          </span>
          <div>
            <div className="font-display text-[19px] font-extrabold text-[#0F172A]">
              Majukan lead ke {advanceTarget && STATUS_CONFIG[advanceTarget].label}?
            </div>
            <div className="mt-1 text-[13px] leading-[1.55] text-[#64748B]">
              Lead <strong className="text-[#334155]">{lead.company_name}</strong> akan berpindah ke tahap{' '}
              <strong className="text-[#334155]">{advanceTarget && STATUS_CONFIG[advanceTarget].label}</strong>.
            </div>
          </div>
        </div>
        <div className="mt-5 flex justify-end gap-[10px]">
          <button type="button" onClick={() => setModal(null)} className="rounded-[8px] border border-[#CBD5E1] bg-white px-[18px] py-[9px] text-[14px] font-semibold text-[#334155] hover:bg-[#F7F9FC]">
            Batal
          </button>
          <button
            type="button"
            onClick={handleAdvanceConfirm}
            disabled={updateStatus.isPending}
            className="inline-flex items-center gap-[7px] rounded-[8px] bg-[#1D4ED8] px-[18px] py-[9px] text-[14px] font-semibold text-white shadow-[0_1px_2px_rgba(29,78,216,0.3)] hover:bg-[#1A45BE] disabled:opacity-60"
          >
            <Check className="size-[15px]" />
            Konfirmasi
          </button>
        </div>
      </Modal>

      <Modal open={modal === 'lost'} onClose={() => setModal(null)}>
        <div className="flex items-start gap-[14px]">
          <span className="flex size-[44px] shrink-0 items-center justify-center rounded-[11px] bg-[#FEE2E2] text-[#DC2626]">
            <XIcon className="size-[22px]" />
          </span>
          <div>
            <div className="font-display text-[19px] font-extrabold text-[#0F172A]">Tandai Lost</div>
            <div className="mt-1 text-[13px] leading-[1.55] text-[#64748B]">
              Lead <strong className="text-[#334155]">{lead.company_name}</strong> dinyatakan hilang. Status{' '}
              <strong className="text-[#334155]">terminal</strong>. Wajib mencatat alasan.
            </div>
          </div>
        </div>
        <label className="mt-[14px] mb-[7px] block text-[12px] font-semibold text-[#334155]">
          Alasan lost <span className="text-[#DC2626]">*</span>
        </label>
        <textarea
          value={lostReason}
          onChange={(e) => setLostReason(e.target.value)}
          placeholder="mis. Sudah kontrak dengan ISP lain / budget tidak tersedia"
          rows={3}
          className="w-full resize-y rounded-[9px] border border-[#CBD5E1] px-3 py-2.5 text-[13px] leading-[1.5] text-[#0F172A] outline-none focus:border-[#DC2626] focus:ring-[3px] focus:ring-[#FEE2E2]"
        />
        <div className="mt-[18px] flex justify-end gap-[10px]">
          <button type="button" onClick={() => setModal(null)} className="rounded-[8px] border border-[#CBD5E1] bg-white px-[18px] py-[9px] text-[14px] font-semibold text-[#334155] hover:bg-[#F7F9FC]">
            Batal
          </button>
          <button
            type="button"
            onClick={handleLostConfirm}
            disabled={!lostReason.trim() || updateStatus.isPending}
            className="inline-flex items-center gap-[7px] rounded-[8px] px-[18px] py-[9px] text-[14px] font-semibold text-white disabled:cursor-not-allowed"
            style={{ background: lostReason.trim() ? '#DC2626' : '#E7A9A9' }}
          >
            <XIcon className="size-[15px]" />
            Tandai Lost
          </button>
        </div>
      </Modal>

      <Modal open={reassignOpen} onClose={() => setReassignOpen(false)} maxWidth={420}>
        <div className="font-display text-[19px] font-extrabold text-[#0F172A]">Reassign Owner</div>
        <div className="mt-1 text-[13px] leading-[1.55] text-[#64748B]">
          Pindahkan kepemilikan lead <strong className="text-[#334155]">{lead.company_name}</strong> ke sales lain.
        </div>
        <label className="mt-4 mb-[7px] block text-[12px] font-semibold text-[#334155]">Owner baru</label>
        <select
          value={reassignTo}
          onChange={(e) => setReassignTo(e.target.value)}
          className="h-[42px] w-full rounded-[9px] border border-[#CBD5E1] px-3 text-[13.5px] text-[#0F172A] outline-none"
        >
          <option value="">Pilih owner…</option>
          {(salesRoster ?? []).map((u) => (
            <option key={u.id} value={u.id}>
              {u.name} ({roleLabel(u.role)})
            </option>
          ))}
        </select>
        <div className="mt-5 flex justify-end gap-[10px]">
          <button type="button" onClick={() => setReassignOpen(false)} className="rounded-[8px] border border-[#CBD5E1] bg-white px-[18px] py-[9px] text-[14px] font-semibold text-[#334155] hover:bg-[#F7F9FC]">
            Batal
          </button>
          <button
            type="button"
            onClick={handleReassignConfirm}
            disabled={!reassignTo || reassign.isPending}
            className="rounded-[8px] bg-[#1D4ED8] px-[18px] py-[9px] text-[14px] font-semibold text-white shadow-[0_1px_2px_rgba(29,78,216,0.3)] hover:bg-[#1A45BE] disabled:opacity-60"
          >
            Simpan
          </button>
        </div>
      </Modal>
    </div>
  )
}
