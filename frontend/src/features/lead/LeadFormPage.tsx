// Create/edit lead form screen: 5 numbered cards (company, address, PIC,
// existing service, source + owner), reachable from the "Lead Baru" button
// on the list (create) or the "Edit Lead" button on the detail page (edit —
// mode is derived from the presence of a `:code` route param).
import { useEffect, useState, type FormEvent } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ChevronLeft, ChevronDown, Info, Check } from 'lucide-react'
import { useAuth } from '@/auth/AuthContext'
import { showToast } from '@/hooks/useToast'
import { useServiceTypes } from '@/features/reference/queries'
import { getLeadSources } from './refCache'
import { useCreateLead, useUpdateLead, useSalesRoster, useLeadDetail } from './queries'
import type { CreateLeadInput, LeadDetailResponse, LeadSourceResponse } from './api'
import {
  AddressFields,
  FIELD_LABEL,
  FIELD_INPUT,
  FIELD_INPUT_MONO,
  FIELD_SELECT,
  type AddressValue,
} from './AddressFields'

export interface LeadFormValues {
  company_name: string
  business_field: string
  website: string
  address: AddressValue
  rt: string
  rw: string
  street: string
  pic_name: string
  pic_position: string
  office_phone: string
  mobile_phone: string
  email: string
  service_type_id: string
  capacity_mbps: string
  existing_isp: string
  price: string
  other_services: string
  lead_source_id: string
  owner_id: string
}

export function blankForm(): LeadFormValues {
  return {
    company_name: '', business_field: '', website: '', address: {},
    rt: '', rw: '', street: '', pic_name: '', pic_position: '',
    office_phone: '', mobile_phone: '', email: '', service_type_id: '',
    capacity_mbps: '', existing_isp: '', price: '', other_services: '',
    lead_source_id: '', owner_id: '',
  }
}

function fromDetail(d: LeadDetailResponse): LeadFormValues {
  return {
    company_name: d.company_name ?? '',
    business_field: d.business_field ?? '',
    website: d.website ?? '',
    address: {
      province_id: d.province_id,
      city_id: d.city_id,
      district_id: d.district_id,
      village_id: d.village_id,
      zip_id: d.zip_id,
      zip_code: d.zip_code,
    },
    rt: d.rt ?? '',
    rw: d.rw ?? '',
    street: d.street ?? '',
    pic_name: d.pic_name ?? '',
    pic_position: d.pic_position ?? '',
    office_phone: d.office_phone ?? '',
    mobile_phone: d.mobile_phone ?? '',
    email: d.email ?? '',
    service_type_id: d.service_type_id ?? '',
    capacity_mbps: d.capacity_mbps != null ? String(d.capacity_mbps) : '',
    existing_isp: d.existing_isp ?? '',
    price: d.price != null ? String(d.price) : '',
    other_services: d.other_services ?? '',
    lead_source_id: d.lead_source_id ?? '',
    owner_id: d.owner_id ?? '',
  }
}

const s = (x: string) => (x.trim() === '' ? undefined : x.trim())
const n = (x: string) => {
  const t = x.trim()
  if (t === '') return undefined
  const num = Number(t)
  return Number.isFinite(num) ? num : undefined
}

// Build the API payload. `includeOwner` is false for SALES (owner is forced
// server-side to self; the field isn't even shown). Works for both Create
// and Update (Update's shape is the same, all-optional).
export function buildLeadPayload(v: LeadFormValues, includeOwner: boolean): CreateLeadInput {
  return {
    company_name: v.company_name.trim(),
    business_field: s(v.business_field),
    website: s(v.website),
    province_id: v.address.province_id,
    city_id: v.address.city_id,
    district_id: v.address.district_id,
    village_id: v.address.village_id,
    zip_id: v.address.zip_id,
    rt: s(v.rt),
    rw: s(v.rw),
    street: s(v.street),
    pic_name: s(v.pic_name),
    pic_position: s(v.pic_position),
    office_phone: s(v.office_phone),
    mobile_phone: s(v.mobile_phone),
    email: s(v.email),
    service_type_id: s(v.service_type_id),
    capacity_mbps: n(v.capacity_mbps),
    existing_isp: s(v.existing_isp),
    price: n(v.price),
    other_services: s(v.other_services),
    lead_source_id: s(v.lead_source_id),
    ...(includeOwner && v.owner_id ? { owner_id: v.owner_id } : {}),
  }
}

function RequiredMark() {
  return <span className="text-[#DC2626]">*</span>
}

function OptTag() {
  return (
    <span className="ml-1.5 rounded-[5px] bg-[#EEF2F7] px-[7px] py-[1px] text-[10px] font-semibold text-[#94A3B8]">
      opsional
    </span>
  )
}

function SelectChevron() {
  return (
    <ChevronDown className="pointer-events-none absolute top-[13px] right-3 size-4 text-[#94A3B8]" />
  )
}

function FormCard({
  number,
  title,
  badgeMuted,
  extra,
  children,
}: {
  number: number
  title: string
  badgeMuted?: boolean
  extra?: React.ReactNode
  children: React.ReactNode
}) {
  return (
    <div className="mb-4 rounded-[16px] border border-[#E7EDF3] bg-white p-[20px_22px] shadow-[0_1px_2px_rgba(15,23,42,0.04)]">
      <div className="mb-4 flex items-center gap-[9px]">
        <span
          className={
            'flex size-6 items-center justify-center rounded-[7px] font-mono text-[12px] font-bold text-white ' +
            (badgeMuted ? 'bg-[#94A3B8]' : 'bg-[#1D4ED8]')
          }
        >
          {number}
        </span>
        <span className="font-display text-[16px] font-bold">{title}</span>
        {extra}
      </div>
      {children}
    </div>
  )
}

export default function LeadFormPage() {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { code } = useParams<{ code?: string }>()
  const mode = code ? 'edit' : 'create'
  const [values, setValues] = useState<LeadFormValues>(blankForm())
  const [sources, setSources] = useState<LeadSourceResponse[]>([])
  const { data: serviceTypes = [] } = useServiceTypes()
  const showOwnerField = user?.role !== 'SALES'
  const { data: salesRoster = [] } = useSalesRoster(showOwnerField)
  // Both mutations are declared unconditionally — hooks can't be conditional.
  // Only the one matching `mode` is ever fired in handleSubmit.
  const createMutation = useCreateLead()
  const updateMutation = useUpdateLead(code ?? '')
  const detail = useLeadDetail(code ?? '')

  useEffect(() => {
    getLeadSources()
      .then(setSources)
      .catch(() => {
        // An empty dropdown on failure is an acceptable degradation — same
        // pattern as LeadListPage's FilterBar.
      })
  }, [])

  // Once the edit-mode detail has loaded (keyed on `id` so it only runs once
  // per lead, not on every refetch/invalidation): bounce terminal leads back
  // to their detail page (defense-in-depth — the Edit button is already
  // hidden for them, this covers a direct URL visit), otherwise seed the form.
  useEffect(() => {
    if (mode !== 'edit' || !detail.data) return
    if (detail.data.status === 'HANDOFF_ODOO' || detail.data.status === 'LOST') {
      navigate('/leads/' + code, { replace: true })
      return
    }
    setValues(fromDetail(detail.data))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [detail.data?.id])

  function set<K extends keyof LeadFormValues>(key: K, value: LeadFormValues[K]) {
    setValues((v) => ({ ...v, [key]: value }))
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    const includeOwner = showOwnerField
    const payload = buildLeadPayload(values, includeOwner)
    if (mode === 'edit') {
      updateMutation.mutate(payload, {
        onSuccess: () => {
          showToast('Perubahan tersimpan')
          navigate('/leads/' + code)
        },
        onError: () => {
          showToast('Gagal menyimpan perubahan. Periksa data lalu coba lagi.')
        },
      })
      return
    }
    createMutation.mutate(payload, {
      onSuccess: (lead) => {
        showToast('Lead tersimpan · ' + lead.code)
        navigate('/leads/' + lead.code)
      },
      onError: () => {
        showToast('Gagal menyimpan lead. Periksa data alamat/referensi lalu coba lagi.')
      },
    })
  }

  if (mode === 'edit' && detail.isLoading) {
    return <div className="p-6 text-sm text-muted-foreground">Memuat…</div>
  }
  if (mode === 'edit' && (detail.isError || !detail.data)) {
    return (
      <div className="mx-auto max-w-[900px] p-4 lg:p-6">
        <button
          type="button"
          onClick={() => navigate('/leads')}
          className="-ml-2 mb-[14px] inline-flex items-center gap-1.5 rounded-[7px] px-2 py-1.5 text-[13px] font-semibold text-[#64748B] hover:bg-[#EEF2F7] hover:text-[#1D4ED8]"
        >
          <ChevronLeft className="size-4" />
          Kembali ke daftar lead
        </button>
        <div className="p-6 text-sm text-destructive">Lead tidak ditemukan.</div>
      </div>
    )
  }

  const initials = (user?.name ?? '?')
    .trim()
    .split(/\s+/)
    .map((p) => p[0])
    .slice(0, 2)
    .join('')
    .toUpperCase()

  return (
    <div className="mx-auto max-w-[900px] p-4 lg:p-6">
      <button
        type="button"
        onClick={() => navigate(mode === 'edit' ? '/leads/' + code : '/leads')}
        className="-ml-2 mb-[14px] inline-flex items-center gap-1.5 rounded-[7px] px-2 py-1.5 text-[13px] font-semibold text-[#64748B] hover:bg-[#EEF2F7] hover:text-[#1D4ED8]"
      >
        <ChevronLeft className="size-4" />
        {mode === 'edit' ? 'Kembali ke detail lead' : 'Kembali ke daftar lead'}
      </button>

      <h1 className="font-display text-[26px] font-extrabold tracking-[-.02em]">
        {mode === 'edit' ? 'Edit Lead' : 'Lead Baru'}
      </h1>
      <p className="mt-1.5 mb-5 flex items-center gap-1.5 text-[13px] text-[#64748B]">
        <Info className="size-[15px] shrink-0" />
        {mode === 'edit'
          ? 'Kode lead tidak dapat diubah. Perubahan tersimpan ke riwayat lead.'
          : 'Kode lead dihasilkan otomatis oleh sistem setelah tersimpan (mis. LD-2607-00XX).'}
      </p>

      <form onSubmit={handleSubmit}>
        <FormCard number={1} title="Data Perusahaan">
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-[2fr_1fr]">
            <div>
              <label className={FIELD_LABEL}>
                Nama Perusahaan <RequiredMark />
              </label>
              <input
                required
                value={values.company_name}
                onChange={(e) => set('company_name', e.target.value)}
                className={FIELD_INPUT}
              />
            </div>
            <div>
              <label className={FIELD_LABEL}>
                Bidang Usaha <RequiredMark />
              </label>
              <input
                required
                value={values.business_field}
                onChange={(e) => set('business_field', e.target.value)}
                className={FIELD_INPUT}
              />
            </div>
            <div className="lg:col-span-2">
              <label className={FIELD_LABEL}>
                Website <OptTag />
              </label>
              <input
                value={values.website}
                onChange={(e) => set('website', e.target.value)}
                className={FIELD_INPUT}
              />
            </div>
          </div>
        </FormCard>

        <FormCard number={2} title="Alamat">
          <p className="mb-4 text-[12.5px] text-[#94A3B8]">
            Pilih bertingkat: provinsi → kota/kabupaten → kecamatan → kelurahan. Kode pos terisi
            otomatis dari kelurahan yang dipilih.
          </p>
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <AddressFields
              value={values.address}
              onChange={(address) => setValues((v) => ({ ...v, address }))}
            />
          </div>
          <div className="mt-4 grid grid-cols-1 gap-4 lg:grid-cols-3">
            <div>
              <label className={FIELD_LABEL}>Kode Pos</label>
              <input
                readOnly
                value={values.address.zip_code ?? ''}
                placeholder="otomatis"
                className={FIELD_INPUT_MONO + ' bg-[#F7F9FC] text-[#475569]'}
              />
            </div>
            <div>
              <label className={FIELD_LABEL}>
                RT <OptTag />
              </label>
              <input
                value={values.rt}
                onChange={(e) => set('rt', e.target.value)}
                className={FIELD_INPUT_MONO}
              />
            </div>
            <div>
              <label className={FIELD_LABEL}>
                RW <OptTag />
              </label>
              <input
                value={values.rw}
                onChange={(e) => set('rw', e.target.value)}
                className={FIELD_INPUT_MONO}
              />
            </div>
          </div>
          <div className="mt-4">
            <label className={FIELD_LABEL}>
              Nama Jalan Lengkap <RequiredMark />
            </label>
            <input
              required
              value={values.street}
              onChange={(e) => set('street', e.target.value)}
              className={FIELD_INPUT}
            />
          </div>
        </FormCard>

        <FormCard number={3} title="PIC (Person in Charge)">
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <div>
              <label className={FIELD_LABEL}>
                Nama PIC <RequiredMark />
              </label>
              <input
                required
                value={values.pic_name}
                onChange={(e) => set('pic_name', e.target.value)}
                className={FIELD_INPUT}
              />
            </div>
            <div>
              <label className={FIELD_LABEL}>
                Jabatan <RequiredMark />
              </label>
              <input
                required
                value={values.pic_position}
                onChange={(e) => set('pic_position', e.target.value)}
                className={FIELD_INPUT}
              />
            </div>
            <div>
              <label className={FIELD_LABEL}>
                Telepon Kantor <OptTag />
              </label>
              <input
                value={values.office_phone}
                onChange={(e) => set('office_phone', e.target.value)}
                className={FIELD_INPUT_MONO}
              />
            </div>
            <div>
              <label className={FIELD_LABEL}>
                No. HP <RequiredMark />
              </label>
              <input
                required
                value={values.mobile_phone}
                onChange={(e) => set('mobile_phone', e.target.value)}
                className={FIELD_INPUT_MONO}
              />
            </div>
            <div className="lg:col-span-2">
              <label className={FIELD_LABEL}>
                Email <RequiredMark />
              </label>
              <input
                required
                type="email"
                value={values.email}
                onChange={(e) => set('email', e.target.value)}
                className={FIELD_INPUT}
              />
            </div>
          </div>
        </FormCard>

        <FormCard
          number={4}
          title="Layanan Eksisting"
          badgeMuted
          extra={
            <span className="ml-auto rounded-full bg-[#EEF2F7] px-[10px] py-[3px] text-[11px] font-semibold text-[#64748B]">
              Seluruh section opsional
            </span>
          }
        >
          <p className="mb-4 text-[12.5px] text-[#94A3B8]">
            Isi bila prospek sudah memakai layanan internet lain — sering belum diketahui saat
            lead pertama masuk.
          </p>
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <div>
              <label className={FIELD_LABEL}>Type Layanan</label>
              <div className="relative">
                <select
                  value={values.service_type_id}
                  onChange={(e) => set('service_type_id', e.target.value)}
                  className={FIELD_SELECT}
                >
                  <option value="">Pilih…</option>
                  {serviceTypes.map((st, i) => (
                    <option key={st.id ?? i} value={st.id ?? ''}>
                      {st.name}
                    </option>
                  ))}
                </select>
                <SelectChevron />
              </div>
            </div>
            <div>
              <label className={FIELD_LABEL}>Kapasitas (Mbps)</label>
              <input
                type="number"
                value={values.capacity_mbps}
                onChange={(e) => set('capacity_mbps', e.target.value)}
                className={FIELD_INPUT}
              />
            </div>
            <div>
              <label className={FIELD_LABEL}>ISP Eksisting</label>
              <input
                value={values.existing_isp}
                onChange={(e) => set('existing_isp', e.target.value)}
                className={FIELD_INPUT}
              />
            </div>
            <div>
              <label className={FIELD_LABEL}>Harga / bulan</label>
              <input
                type="number"
                value={values.price}
                onChange={(e) => set('price', e.target.value)}
                className={FIELD_INPUT_MONO}
              />
            </div>
            <div className="lg:col-span-2">
              <label className={FIELD_LABEL}>Layanan Lainnya</label>
              <input
                value={values.other_services}
                onChange={(e) => set('other_services', e.target.value)}
                className={FIELD_INPUT}
              />
            </div>
          </div>
        </FormCard>

        <FormCard number={5} title="Sumber Lead + Owner">
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <div className={showOwnerField ? '' : 'lg:col-span-2'}>
              <label className={FIELD_LABEL}>
                Sumber Lead <RequiredMark />
              </label>
              <div className="relative">
                <select
                  required
                  value={values.lead_source_id}
                  onChange={(e) => set('lead_source_id', e.target.value)}
                  className={FIELD_SELECT}
                >
                  <option value="">Pilih…</option>
                  {sources.map((source, i) => (
                    <option key={source.id ?? i} value={source.id ?? ''}>
                      {source.name}
                    </option>
                  ))}
                </select>
                <SelectChevron />
              </div>
            </div>

            {showOwnerField ? (
              <div>
                <label className={FIELD_LABEL}>
                  Owner Sales <RequiredMark />
                </label>
                <div className="relative">
                  <select
                    required
                    value={values.owner_id}
                    onChange={(e) => set('owner_id', e.target.value)}
                    className={FIELD_SELECT}
                  >
                    <option value="">Pilih…</option>
                    {salesRoster.map((salesUser, i) => (
                      <option key={salesUser.id ?? i} value={salesUser.id ?? ''}>
                        {salesUser.name}
                      </option>
                    ))}
                  </select>
                  <SelectChevron />
                </div>
              </div>
            ) : (
              <div>
                <label className={FIELD_LABEL}>Owner Sales</label>
                <div className="flex h-[42px] items-center gap-2 rounded-[9px] border border-[#E2E8F0] bg-[#F7F9FC] px-3">
                  <span className="flex size-6 items-center justify-center rounded-[6px] bg-[#EEF3FC] text-[10px] font-bold text-[#1D4ED8]">
                    {initials}
                  </span>
                  <span className="text-[13.5px] text-[#475569]">{user?.name} (Anda)</span>
                </div>
              </div>
            )}
          </div>
        </FormCard>

        <div className="flex justify-end gap-[10px]">
          <button
            type="button"
            onClick={() => navigate(mode === 'edit' ? '/leads/' + code : '/leads')}
            className="rounded-[8px] border border-[#CBD5E1] bg-white px-[18px] py-[9px] text-[14px] font-semibold text-[#334155] hover:bg-[#F7F9FC]"
          >
            Batal
          </button>
          <button
            type="submit"
            disabled={mode === 'edit' ? updateMutation.isPending : createMutation.isPending}
            className="inline-flex items-center gap-[7px] rounded-[8px] bg-[#1D4ED8] px-[18px] py-[9px] text-[14px] font-semibold text-white shadow-[0_1px_2px_rgba(29,78,216,0.3)] hover:bg-[#1A45BE] disabled:opacity-60"
          >
            <Check className="size-[15px]" />
            {mode === 'edit' ? 'Simpan Perubahan' : 'Simpan Lead'}
          </button>
        </div>
      </form>
    </div>
  )
}
