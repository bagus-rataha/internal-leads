// Create/edit lead form screen: 5 numbered cards (company, address, PIC,
// existing service, source + owner), reachable from the "Lead Baru" button
// on the list (create) or the "Edit Lead" button on the detail page (edit —
// mode is derived from the presence of a `:code` route param).
import { useEffect, useState, type FormEvent } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ChevronLeft, ChevronDown, Info, Check } from 'lucide-react'
import { useAuth } from '@/auth/AuthContext'
import { showToast } from '@/hooks/useToast'
import { roleLabel } from '@/lib/roles'
import { useServiceTypes, useLeadSources } from '@/features/reference/queries'
import { useCreateLead, useUpdateLead, useSalesRoster, useLeadDetail } from './queries'
import type { CreateLeadInput, LeadDetailResponse } from './api'
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
  forecast_mrr: string
  other_services: string
  lead_source_id: string
  owner_id: string
}

export function blankForm(defaultOwnerId?: string): LeadFormValues {
  return {
    company_name: '', business_field: '', website: '', address: {},
    rt: '', rw: '', street: '', pic_name: '', pic_position: '',
    office_phone: '', mobile_phone: '', email: '', service_type_id: '',
    capacity_mbps: '', existing_isp: '', price: '', forecast_mrr: '', other_services: '',
    lead_source_id: '', owner_id: defaultOwnerId ?? '',
  }
}

// Editing a lead loads the stored phone as `62<national>` (or possibly a
// legacy `0<national>`/`+62<national>`) — strip it back to the bare national
// digits so the input shows only what the user is meant to type.
function stripPhonePrefix(raw: string | undefined | null): string {
  const v = (raw ?? '').trim()
  if (v.startsWith('+62')) return v.slice(3)
  if (v.startsWith('62')) return v.slice(2)
  if (v.startsWith('0')) return v.slice(1)
  return v
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
    office_phone: stripPhonePrefix(d.office_phone),
    mobile_phone: stripPhonePrefix(d.mobile_phone),
    email: d.email ?? '',
    service_type_id: d.service_type_id ?? '',
    capacity_mbps: d.capacity_mbps != null ? String(d.capacity_mbps) : '',
    existing_isp: d.existing_isp ?? '',
    price: d.price != null ? String(d.price) : '',
    forecast_mrr: d.forecast_mrr != null ? String(d.forecast_mrr) : '',
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

// National digits -> full stored phone (`62<national>`); undefined when empty.
const normalizePhone = (national: string) => {
  const v = national.trim()
  return v === '' ? undefined : '62' + v
}

// Bare domain -> schemed URL; undefined when empty. Already-schemed input
// passes through untouched (the lenient WEBSITE_RE accepts either).
const normalizeWebsite = (website: string) => {
  const v = website.trim()
  if (v === '') return undefined
  return /^https?:\/\//i.test(v) ? v : 'https://' + v
}

// Build the API payload. `includeOwner` is false for SALES (owner is forced
// server-side to self; the field isn't even shown). Works for both Create
// and Update (Update's shape is the same, all-optional).
export function buildLeadPayload(v: LeadFormValues, includeOwner: boolean): CreateLeadInput {
  return {
    company_name: v.company_name.trim(),
    business_field: s(v.business_field),
    website: normalizeWebsite(v.website),
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
    office_phone: normalizePhone(v.office_phone),
    mobile_phone: normalizePhone(v.mobile_phone),
    email: s(v.email),
    service_type_id: s(v.service_type_id),
    capacity_mbps: n(v.capacity_mbps),
    existing_isp: s(v.existing_isp),
    price: n(v.price),
    forecast_mrr: n(v.forecast_mrr),
    other_services: s(v.other_services),
    lead_source_id: s(v.lead_source_id),
    ...(includeOwner && v.owner_id ? { owner_id: v.owner_id } : {}),
  }
}

// ---- Validation ---------------------------------------------------------

type AddressField = 'province_id' | 'city_id' | 'district_id' | 'village_id'
type FieldName = keyof LeadFormValues | AddressField

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const WEBSITE_RE = /^(https?:\/\/)?([a-z0-9-]+\.)+[a-z]{2,}(\/\S*)?$/i
const RTRW_RE = /^\d{1,4}$/

const PRESET_OTHER = 'Lainnya'

// Researched via web search (not the sales team's own field data) - see
// .superpowers/specs/2026-08-26-forecast-followup-owner-isp-preset-design.md
// section 3. Expanded from 20 to 33 with 13 more researched via web search
// on 2026-08-27. Still flagged for the product owner's own sanity-check.
const ISP_PRESETS = [
  'Telkom Indonesia (IndiHome/IndiBiz/Astinet)',
  'Biznet',
  'Lintasarta',
  'Indosat Business (IOH)',
  'XL Axiata Business',
  'Iconnet (PLN Icon Plus)',
  'MyRepublic',
  'Moratelindo (Oxygen.id Business)',
  'CBN',
  'iForte',
  'First Media',
  'MNC Play',
  'Indonet',
  'Primacom (PRIMALINKnet)',
  'Telkomsel Enterprise',
  'Starlink Business',
  'Corbec Communication',
  'Melvar Lintasbuana',
  'GTN (Graha Teknologi Nusantara)',
  'Skynindo',
  'ION Network',
  'Intimedia (Sarana Intimedia Telematika)',
  'WOWNET',
  'Padi Technology',
  'D~NET',
  'GMedia',
  'PRIMADONA Net',
  'DataComm',
  'Fibernet',
  'NAP Info',
  'SKINET (Sumber Koneksi Indonesia)',
  'Fiberstar',
  'PSN (Pasifik Satelit Nusantara)',
] as const

// Sourced from KBLI 2025 (BPS's official business-field classification, 21
// top-level categories), curated to 19, then expanded to 35 by breaking the
// top-level KBLI 2020 categories down to golongan-pokok (2-digit division)
// granularity - e.g. "Manufaktur" split into ~10 industry-specific rows.
// NOTE: this breakdown was NOT independently re-verified against the
// official BPS PDF (a scanned, non-text-searchable image file) - sanity
// check against real field data before trusting it.
const BUSINESS_FIELD_PRESETS = [
  'Pertanian, Kehutanan & Perikanan',
  'Pertambangan & Penggalian',
  'Industri Makanan & Minuman',
  'Industri Tekstil, Pakaian & Alas Kaki',
  'Industri Kayu, Kertas & Percetakan',
  'Industri Kimia & Farmasi',
  'Industri Karet, Plastik & Barang Galian Bukan Logam',
  'Industri Logam Dasar & Barang Logam',
  'Industri Komputer, Elektronik & Optik',
  'Industri Peralatan Listrik',
  'Industri Mesin & Perlengkapan',
  'Industri Kendaraan Bermotor & Alat Angkutan Lain',
  'Industri Furnitur & Manufaktur Lainnya',
  'Listrik, Gas & Energi',
  'Pengelolaan Air, Limbah & Daur Ulang',
  'Konstruksi',
  'Perdagangan Besar (Grosir/Distributor)',
  'Perdagangan Eceran (Retail)',
  'Transportasi & Pergudangan (Logistik)',
  'Akomodasi (Hotel/Penginapan)',
  'Makanan & Minuman (Restoran/F&B)',
  'Telekomunikasi',
  'Pemrograman, Konsultansi & Aktivitas Komputer',
  'Penerbitan, Media & Penyiaran',
  'Jasa Keuangan & Perbankan',
  'Asuransi & Dana Pensiun',
  'Real Estat & Properti',
  'Jasa Profesional, Ilmiah & Teknis (Konsultan/Hukum/Akuntansi)',
  'Jasa Persewaan & Sewa Guna Usaha',
  'Jasa Ketenagakerjaan, Agen Perjalanan & Penunjang Usaha',
  'Administrasi Pemerintahan & Pertahanan',
  'Pendidikan',
  'Kesehatan & Aktivitas Sosial',
  'Kesenian, Hiburan & Rekreasi',
  'Jasa Lainnya',
] as const

// Derives a preset dropdown's selection from the actual stored value: a
// preset match selects that preset, any other non-empty value (including
// legacy free-text data from before this dropdown existed) selects
// "Lainnya" with the text fallback pre-filled, and an empty value selects
// nothing. Shared by both existing_isp and business_field - same logic,
// different preset arrays.
function presetSelectionFor(value: string, presets: readonly string[]): string {
  if (value === '') return ''
  return presets.includes(value) ? value : PRESET_OTHER
}

function phoneError(national: string, required: boolean): string | undefined {
  const v = national.trim()
  if (!v) return required ? 'Wajib diisi' : undefined
  if (!/^\d+$/.test(v)) return 'Hanya angka, tanpa spasi/tanda'
  if (v.startsWith('0')) return 'Tanpa 0 di depan'
  if (v.length < 8 || v.length > 13) return 'Nomor tidak valid'
  return undefined
}

// Order matters: it's also the top-to-bottom DOM order used to find the
// first errored field to scroll to on a blocked submit.
const ALL_FIELDS: FieldName[] = [
  'company_name',
  'business_field',
  'website',
  'province_id',
  'city_id',
  'district_id',
  'village_id',
  'rt',
  'rw',
  'street',
  'pic_name',
  'pic_position',
  'office_phone',
  'mobile_phone',
  'email',
  'capacity_mbps',
  'existing_isp',
  'price',
  'forecast_mrr',
  'lead_source_id',
  'owner_id',
]

function validateField(
  name: FieldName,
  values: LeadFormValues,
  showOwner: boolean,
  ispSelection: string,
): string | undefined {
  switch (name) {
    case 'company_name':
    case 'business_field':
    case 'street':
    case 'pic_name':
    case 'pic_position':
      return values[name].trim() === '' ? 'Wajib diisi' : undefined
    case 'existing_isp':
      return ispSelection === PRESET_OTHER && values.existing_isp.trim() === '' ? 'Wajib diisi' : undefined
    case 'lead_source_id':
      return values.lead_source_id === '' ? 'Wajib dipilih' : undefined
    case 'owner_id':
      return showOwner && values.owner_id === '' ? 'Wajib dipilih' : undefined
    case 'province_id':
    case 'city_id':
    case 'district_id':
    case 'village_id':
      return values.address[name] == null ? 'Wajib dipilih' : undefined
    case 'email': {
      const v = values.email.trim()
      if (!v) return 'Wajib diisi'
      return EMAIL_RE.test(v) ? undefined : 'Format email tidak valid'
    }
    case 'mobile_phone':
      return phoneError(values.mobile_phone, true)
    case 'office_phone':
      return phoneError(values.office_phone, false)
    case 'website': {
      const v = values.website.trim()
      if (!v) return undefined
      return WEBSITE_RE.test(v) ? undefined : 'Format website tidak valid'
    }
    case 'rt':
    case 'rw': {
      const v = values[name].trim()
      if (!v) return undefined
      return RTRW_RE.test(v) ? undefined : 'Hanya angka (maks 4 digit)'
    }
    case 'capacity_mbps': {
      const v = values.capacity_mbps.trim()
      if (!v) return undefined
      const num = Number(v)
      return Number.isInteger(num) && num > 0 ? undefined : 'Harus lebih dari 0'
    }
    case 'price': {
      const v = values.price.trim()
      if (!v) return undefined
      const num = Number(v)
      return Number.isFinite(num) && num >= 0 ? undefined : 'Tidak boleh negatif'
    }
    case 'forecast_mrr': {
      const v = values.forecast_mrr.trim()
      if (!v) return undefined
      const num = Number(v)
      return Number.isFinite(num) && num >= 0 ? undefined : 'Tidak boleh negatif'
    }
    default:
      return undefined
  }
}

function validateAll(
  values: LeadFormValues,
  showOwner: boolean,
  ispSelection: string,
): Record<string, string> {
  const errs: Record<string, string> = {}
  for (const f of ALL_FIELDS) {
    const err = validateField(f, values, showOwner, ispSelection)
    if (err) errs[f] = err
  }
  return errs
}

// The four address ids don't have their own DOM ids (AddressFields is a
// sibling task's file, not touched here) — scroll to the whole address
// block instead of the exact select.
const ADDRESS_FIELD_SET = new Set<FieldName>([
  'province_id',
  'city_id',
  'district_id',
  'village_id',
])
const scrollTargetId = (field: FieldName) =>
  ADDRESS_FIELD_SET.has(field) ? 'address-fields' : field

// Prefix chip group: input's left corners flattened, no left border (the
// prefix chip supplies it), right side keeps normal field rounding.
const PREFIXED_INPUT = FIELD_INPUT_MONO
  .replace('w-full', 'flex-1 min-w-0')
  .replace('rounded-[9px]', 'rounded-r-[9px] rounded-l-none') + ' border-l-0'
const INPUT_PREFIX_CHIP =
  'flex h-[42px] shrink-0 items-center rounded-l-[9px] border border-r-0 border-[#CBD5E1] bg-[#F1F5F9] px-3 font-mono text-[13px] text-[#64748B]'

const errClass = (base: string, hasError: boolean) => (hasError ? base + ' border-[#DC2626]' : base)

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
  const [values, setValues] = useState<LeadFormValues>(() => blankForm(mode === 'create' && user?.role === 'LEADER' ? user.id : undefined))
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [touched, setTouched] = useState<Record<string, boolean>>({})
  const [ispSelection, setIspSelection] = useState(() => presetSelectionFor(values.existing_isp, ISP_PRESETS))
  const [bizFieldSelection, setBizFieldSelection] = useState(() =>
    presetSelectionFor(values.business_field, BUSINESS_FIELD_PRESETS),
  )
  const { data: sources = [] } = useLeadSources()
  const { data: serviceTypes = [] } = useServiceTypes()
  const showOwnerField = user?.role !== 'SALES'
  const { data: salesRoster = [] } = useSalesRoster(showOwnerField)
  // Both mutations are declared unconditionally — hooks can't be conditional.
  // Only the one matching `mode` is ever fired in handleSubmit.
  const createMutation = useCreateLead()
  const updateMutation = useUpdateLead(code ?? '')
  const detail = useLeadDetail(code ?? '')

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
    const next = fromDetail(detail.data)
    setValues(next)
    setIspSelection(presetSelectionFor(next.existing_isp, ISP_PRESETS))
    setBizFieldSelection(presetSelectionFor(next.business_field, BUSINESS_FIELD_PRESETS))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [detail.data?.id])

  function set<K extends keyof LeadFormValues>(key: K, value: LeadFormValues[K]) {
    setValues((v) => ({ ...v, [key]: value }))
  }

  // Validates on first blur, then live on every change while touched.
  function blur(field: keyof LeadFormValues) {
    setTouched((t) => ({ ...t, [field]: true }))
    setErrors((e) => ({ ...e, [field]: validateField(field, values, showOwnerField, ispSelection) ?? '' }))
  }

  function change<K extends keyof LeadFormValues>(key: K, value: LeadFormValues[K]) {
    set(key, value)
    if (touched[key]) {
      const next = { ...values, [key]: value }
      setErrors((e) => ({ ...e, [key]: validateField(key, next, showOwnerField, ispSelection) ?? '' }))
    }
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    const errs = validateAll(values, showOwnerField, ispSelection)
    setErrors(errs)
    setTouched((t) => {
      const next = { ...t }
      for (const f of ALL_FIELDS) next[f] = true
      return next
    })
    if (Object.keys(errs).length > 0) {
      const firstField = ALL_FIELDS.find((f) => errs[f])
      if (firstField) {
        document.getElementById(scrollTargetId(firstField))?.scrollIntoView({ block: 'center' })
      }
      return
    }
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

      <form onSubmit={handleSubmit} noValidate>
        <FormCard number={1} title="Data Perusahaan">
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-[2fr_1fr]">
            <div>
              <label className={FIELD_LABEL}>
                Nama Perusahaan <RequiredMark />
              </label>
              <input
                id="company_name"
                required
                value={values.company_name}
                onChange={(e) => change('company_name', e.target.value)}
                onBlur={() => blur('company_name')}
                placeholder="mis. PT Sinar Baja Elektrik"
                className={errClass(FIELD_INPUT, touched.company_name && !!errors.company_name)}
              />
              {touched.company_name && errors.company_name && (
                <p className="mt-1 text-[11px] text-[#DC2626]">{errors.company_name}</p>
              )}
            </div>
            <div>
              <label className={FIELD_LABEL}>
                Bidang Usaha <RequiredMark />
              </label>
              <div className="relative">
                <select
                  id="business_field"
                  required
                  value={bizFieldSelection}
                  onChange={(e) => {
                    const v = e.target.value
                    setBizFieldSelection(v)
                    if (v !== PRESET_OTHER) change('business_field', v)
                  }}
                  onBlur={() => blur('business_field')}
                  className={errClass(FIELD_SELECT, touched.business_field && !!errors.business_field)}
                >
                  <option value="">Pilih…</option>
                  {BUSINESS_FIELD_PRESETS.map((bf) => (
                    <option key={bf} value={bf}>
                      {bf}
                    </option>
                  ))}
                  <option value={PRESET_OTHER}>{PRESET_OTHER}</option>
                </select>
                <SelectChevron />
              </div>
              {bizFieldSelection === PRESET_OTHER && (
                <input
                  value={values.business_field}
                  onChange={(e) => change('business_field', e.target.value)}
                  onBlur={() => blur('business_field')}
                  placeholder="Tulis bidang usaha"
                  className={
                    errClass(FIELD_INPUT, touched.business_field && !!errors.business_field) + ' mt-2'
                  }
                />
              )}
              {touched.business_field && errors.business_field && (
                <p className="mt-1 text-[11px] text-[#DC2626]">{errors.business_field}</p>
              )}
            </div>
            <div className="lg:col-span-2">
              <label className={FIELD_LABEL}>
                Website <OptTag />
              </label>
              <input
                id="website"
                value={values.website}
                onChange={(e) => change('website', e.target.value)}
                onBlur={() => blur('website')}
                placeholder="https://"
                className={errClass(FIELD_INPUT, touched.website && !!errors.website)}
              />
              {touched.website && errors.website && (
                <p className="mt-1 text-[11px] text-[#DC2626]">{errors.website}</p>
              )}
            </div>
          </div>
        </FormCard>

        <FormCard number={2} title="Alamat">
          <p className="mb-4 text-[12.5px] text-[#94A3B8]">
            Pilih bertingkat: provinsi → kota/kabupaten → kecamatan → kelurahan. Kode pos terisi
            otomatis dari kelurahan yang dipilih.
          </p>
          <div id="address-fields" className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <AddressFields
              value={values.address}
              onChange={(address) => {
                setValues((v) => ({ ...v, address }))
                setErrors((e) => {
                  const next = { ...e }
                  for (const f of ['province_id', 'city_id', 'district_id', 'village_id'] as const) {
                    if (touched[f]) {
                      next[f] =
                        validateField(f, { ...values, address }, showOwnerField, ispSelection) ?? ''
                    }
                  }
                  return next
                })
              }}
              errors={{
                province_id: touched.province_id ? errors.province_id : undefined,
                city_id: touched.city_id ? errors.city_id : undefined,
                district_id: touched.district_id ? errors.district_id : undefined,
                village_id: touched.village_id ? errors.village_id : undefined,
              }}
              onBlurField={(f) => {
                setTouched((t) => ({ ...t, [f]: true }))
                setErrors((e) => ({ ...e, [f]: validateField(f, values, showOwnerField, ispSelection) ?? '' }))
              }}
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
                id="rt"
                inputMode="numeric"
                maxLength={4}
                value={values.rt}
                onChange={(e) => change('rt', e.target.value)}
                onBlur={() => blur('rt')}
                placeholder="001"
                className={errClass(FIELD_INPUT_MONO, touched.rt && !!errors.rt)}
              />
              {touched.rt && errors.rt && (
                <p className="mt-1 text-[11px] text-[#DC2626]">{errors.rt}</p>
              )}
            </div>
            <div>
              <label className={FIELD_LABEL}>
                RW <OptTag />
              </label>
              <input
                id="rw"
                inputMode="numeric"
                maxLength={4}
                value={values.rw}
                onChange={(e) => change('rw', e.target.value)}
                onBlur={() => blur('rw')}
                placeholder="005"
                className={errClass(FIELD_INPUT_MONO, touched.rw && !!errors.rw)}
              />
              {touched.rw && errors.rw && (
                <p className="mt-1 text-[11px] text-[#DC2626]">{errors.rw}</p>
              )}
            </div>
          </div>
          <div className="mt-4">
            <label className={FIELD_LABEL}>
              Nama Jalan Lengkap <RequiredMark />
            </label>
            <input
              id="street"
              required
              value={values.street}
              onChange={(e) => change('street', e.target.value)}
              onBlur={() => blur('street')}
              placeholder="mis. Jl. Rungkut Industri Raya No. 12, Blok B"
              className={errClass(FIELD_INPUT, touched.street && !!errors.street)}
            />
            {touched.street && errors.street && (
              <p className="mt-1 text-[11px] text-[#DC2626]">{errors.street}</p>
            )}
          </div>
        </FormCard>

        <FormCard number={3} title="PIC (Person in Charge)">
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <div>
              <label className={FIELD_LABEL}>
                Nama PIC <RequiredMark />
              </label>
              <input
                id="pic_name"
                required
                value={values.pic_name}
                onChange={(e) => change('pic_name', e.target.value)}
                onBlur={() => blur('pic_name')}
                placeholder="mis. Andi Wijaya"
                className={errClass(FIELD_INPUT, touched.pic_name && !!errors.pic_name)}
              />
              {touched.pic_name && errors.pic_name && (
                <p className="mt-1 text-[11px] text-[#DC2626]">{errors.pic_name}</p>
              )}
            </div>
            <div>
              <label className={FIELD_LABEL}>
                Jabatan <RequiredMark />
              </label>
              <input
                id="pic_position"
                required
                value={values.pic_position}
                onChange={(e) => change('pic_position', e.target.value)}
                onBlur={() => blur('pic_position')}
                placeholder="mis. IT Manager"
                className={errClass(FIELD_INPUT, touched.pic_position && !!errors.pic_position)}
              />
              {touched.pic_position && errors.pic_position && (
                <p className="mt-1 text-[11px] text-[#DC2626]">{errors.pic_position}</p>
              )}
            </div>
            <div>
              <label className={FIELD_LABEL}>
                Telepon Kantor <OptTag />
              </label>
              <div className="flex">
                <span className={INPUT_PREFIX_CHIP}>+62</span>
                <input
                  id="office_phone"
                  inputMode="numeric"
                  maxLength={13}
                  value={values.office_phone}
                  onChange={(e) => change('office_phone', e.target.value)}
                  onBlur={() => blur('office_phone')}
                  placeholder="2112345678"
                  className={errClass(PREFIXED_INPUT, touched.office_phone && !!errors.office_phone)}
                />
              </div>
              {touched.office_phone && errors.office_phone && (
                <p className="mt-1 text-[11px] text-[#DC2626]">{errors.office_phone}</p>
              )}
            </div>
            <div>
              <label className={FIELD_LABEL}>
                No. HP <RequiredMark />
              </label>
              <div className="flex">
                <span className={INPUT_PREFIX_CHIP}>+62</span>
                <input
                  id="mobile_phone"
                  required
                  inputMode="numeric"
                  maxLength={13}
                  value={values.mobile_phone}
                  onChange={(e) => change('mobile_phone', e.target.value)}
                  onBlur={() => blur('mobile_phone')}
                  placeholder="81234567890"
                  className={errClass(PREFIXED_INPUT, touched.mobile_phone && !!errors.mobile_phone)}
                />
              </div>
              {touched.mobile_phone && errors.mobile_phone && (
                <p className="mt-1 text-[11px] text-[#DC2626]">{errors.mobile_phone}</p>
              )}
            </div>
            <div className="lg:col-span-2">
              <label className={FIELD_LABEL}>
                Email <RequiredMark />
              </label>
              <input
                id="email"
                required
                type="email"
                value={values.email}
                onChange={(e) => change('email', e.target.value)}
                onBlur={() => blur('email')}
                placeholder="mis. andi@perusahaan.co.id"
                className={errClass(FIELD_INPUT, touched.email && !!errors.email)}
              />
              {touched.email && errors.email && (
                <p className="mt-1 text-[11px] text-[#DC2626]">{errors.email}</p>
              )}
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
                id="capacity_mbps"
                type="number"
                value={values.capacity_mbps}
                onChange={(e) => change('capacity_mbps', e.target.value)}
                onBlur={() => blur('capacity_mbps')}
                placeholder="50"
                className={errClass(FIELD_INPUT, touched.capacity_mbps && !!errors.capacity_mbps)}
              />
              {touched.capacity_mbps && errors.capacity_mbps && (
                <p className="mt-1 text-[11px] text-[#DC2626]">{errors.capacity_mbps}</p>
              )}
            </div>
            <div>
              <label className={FIELD_LABEL}>ISP Eksisting</label>
              <div className="relative">
                <select
                  value={ispSelection}
                  onChange={(e) => {
                    const v = e.target.value
                    setIspSelection(v)
                    if (v !== PRESET_OTHER) change('existing_isp', v)
                  }}
                  onBlur={() => blur('existing_isp')}
                  className={errClass(FIELD_SELECT, touched.existing_isp && !!errors.existing_isp)}
                >
                  <option value="">Pilih…</option>
                  {ISP_PRESETS.map((isp) => (
                    <option key={isp} value={isp}>
                      {isp}
                    </option>
                  ))}
                  <option value={PRESET_OTHER}>{PRESET_OTHER}</option>
                </select>
                <SelectChevron />
              </div>
              {ispSelection === PRESET_OTHER && (
                <input
                  value={values.existing_isp}
                  onChange={(e) => change('existing_isp', e.target.value)}
                  onBlur={() => blur('existing_isp')}
                  placeholder="Tulis nama ISP"
                  className={errClass(FIELD_INPUT, touched.existing_isp && !!errors.existing_isp) + ' mt-2'}
                />
              )}
              {touched.existing_isp && errors.existing_isp && (
                <p className="mt-1 text-[11px] text-[#DC2626]">{errors.existing_isp}</p>
              )}
            </div>
            <div>
              <label className={FIELD_LABEL}>Harga / bulan</label>
              <div className="flex">
                <span className={INPUT_PREFIX_CHIP}>Rp</span>
                <input
                  id="price"
                  type="number"
                  value={values.price}
                  onChange={(e) => change('price', e.target.value)}
                  onBlur={() => blur('price')}
                  placeholder="3500000"
                  className={errClass(PREFIXED_INPUT, touched.price && !!errors.price)}
                />
              </div>
              {touched.price && errors.price && (
                <p className="mt-1 text-[11px] text-[#DC2626]">{errors.price}</p>
              )}
            </div>
            <div>
              <label className={FIELD_LABEL}>Forecast Pendapatan (MRR)</label>
              <div className="flex">
                <span className={INPUT_PREFIX_CHIP}>Rp</span>
                <input
                  id="forecast_mrr"
                  type="number"
                  value={values.forecast_mrr}
                  onChange={(e) => change('forecast_mrr', e.target.value)}
                  onBlur={() => blur('forecast_mrr')}
                  placeholder="5000000"
                  className={errClass(PREFIXED_INPUT, touched.forecast_mrr && !!errors.forecast_mrr)}
                />
              </div>
              {touched.forecast_mrr && errors.forecast_mrr && (
                <p className="mt-1 text-[11px] text-[#DC2626]">{errors.forecast_mrr}</p>
              )}
            </div>
            <div className="lg:col-span-2">
              <label className={FIELD_LABEL}>Layanan Lainnya</label>
              <input
                value={values.other_services}
                onChange={(e) => set('other_services', e.target.value)}
                placeholder="mis. IP publik, cloud PABX"
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
                  id="lead_source_id"
                  required
                  value={values.lead_source_id}
                  onChange={(e) => change('lead_source_id', e.target.value)}
                  onBlur={() => blur('lead_source_id')}
                  className={errClass(
                    FIELD_SELECT,
                    touched.lead_source_id && !!errors.lead_source_id,
                  )}
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
              {touched.lead_source_id && errors.lead_source_id && (
                <p className="mt-1 text-[11px] text-[#DC2626]">{errors.lead_source_id}</p>
              )}
            </div>

            {showOwnerField ? (
              <div>
                <label className={FIELD_LABEL}>
                  Owner Sales <RequiredMark />
                </label>
                <div className="relative">
                  <select
                    id="owner_id"
                    required
                    value={values.owner_id}
                    onChange={(e) => change('owner_id', e.target.value)}
                    onBlur={() => blur('owner_id')}
                    className={errClass(FIELD_SELECT, touched.owner_id && !!errors.owner_id)}
                  >
                    <option value="">Pilih…</option>
                    {salesRoster.map((salesUser, i) => (
                      <option key={salesUser.id ?? i} value={salesUser.id ?? ''}>
                        {salesUser.name} ({roleLabel(salesUser.role)})
                      </option>
                    ))}
                  </select>
                  <SelectChevron />
                </div>
                {touched.owner_id && errors.owner_id && (
                  <p className="mt-1 text-[11px] text-[#DC2626]">{errors.owner_id}</p>
                )}
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
