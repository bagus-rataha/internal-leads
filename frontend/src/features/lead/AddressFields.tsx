// Controlled cascading address selector: Provinsi -> Kota -> Kecamatan ->
// Kelurahan, plus the read-only Kode Pos derived from the chosen village.
// Isolated from the rest of the lead form because the cascade-reset +
// zip-derivation logic is the trickiest part and deserves its own review.
import type { ChangeEvent } from 'react'
import { ChevronDown } from 'lucide-react'
import { useProvinces, useCities, useDistricts, useVillages } from '@/features/reference/queries'

// Exact values from the Form mockup. Shared by AddressFields and LeadFormPage.
export const FIELD_LABEL = 'mb-[7px] block text-[12px] font-semibold text-[#334155]'
export const FIELD_INPUT =
  'h-[42px] w-full rounded-[9px] border border-[#CBD5E1] bg-white px-3 text-[13.5px] text-[#0F172A] outline-none focus:border-[#1D4ED8] focus:shadow-[0_0_0_3px_#DCE7FB]'
export const FIELD_INPUT_MONO = FIELD_INPUT + ' font-mono text-[13px]'
export const FIELD_SELECT =
  'h-[42px] w-full appearance-none rounded-[9px] border border-[#CBD5E1] bg-white pl-3 pr-[34px] text-[13.5px] text-[#0F172A] outline-none cursor-pointer focus:border-[#1D4ED8] focus:shadow-[0_0_0_3px_#DCE7FB]'
export const FIELD_SELECT_DISABLED =
  'h-[42px] w-full appearance-none rounded-[9px] border border-[#E2E8F0] bg-[#F7F9FC] pl-3 pr-[34px] text-[13.5px] text-[#94A3B8] outline-none cursor-not-allowed'

export interface AddressValue {
  province_id?: number
  city_id?: number
  district_id?: number
  village_id?: number
  zip_id?: number
  zip_code?: string
}

function SelectChevron() {
  return (
    <ChevronDown className="pointer-events-none absolute top-[13px] right-3 size-4 text-[#94A3B8]" />
  )
}

type AddressField = 'province_id' | 'city_id' | 'district_id' | 'village_id'

export function AddressFields({
  value,
  onChange,
  errors,
  onBlurField,
}: {
  value: AddressValue
  onChange: (next: AddressValue) => void
  errors?: Partial<Record<AddressField, string>>
  onBlurField?: (field: AddressField) => void
}) {
  const { data: provinces = [] } = useProvinces()
  const { data: cities = [] } = useCities(value.province_id)
  const { data: districts = [] } = useDistricts(value.city_id)
  const { data: villages = [] } = useVillages(value.district_id)

  function onVillageChange(e: ChangeEvent<HTMLSelectElement>) {
    const id = Number(e.target.value) || undefined
    const v = villages.find((x) => x.id === id)
    onChange({
      province_id: value.province_id,
      city_id: value.city_id,
      district_id: value.district_id,
      village_id: id,
      zip_id: v?.zip?.id,
      zip_code: v?.zip?.code,
    })
  }

  return (
    <>
      <div>
        <label className={FIELD_LABEL}>
          Provinsi <span className="text-[#DC2626]">*</span>
        </label>
        <div className="relative">
          <select
            required
            value={value.province_id ?? ''}
            onChange={(e) =>
              onChange({ province_id: Number(e.target.value) || undefined })
            }
            onBlur={() => onBlurField?.('province_id')}
            className={
              errors?.province_id ? FIELD_SELECT + ' border-[#DC2626]' : FIELD_SELECT
            }
          >
            <option value="">Pilih provinsi…</option>
            {provinces.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
          <SelectChevron />
        </div>
        {errors?.province_id && (
          <p className="mt-1 text-[11px] text-[#DC2626]">{errors.province_id}</p>
        )}
      </div>

      <div>
        <label className={FIELD_LABEL}>
          Kota/Kabupaten <span className="text-[#DC2626]">*</span>
        </label>
        <div className="relative">
          <select
            required
            disabled={!value.province_id}
            value={value.city_id ?? ''}
            onChange={(e) =>
              onChange({
                province_id: value.province_id,
                city_id: Number(e.target.value) || undefined,
              })
            }
            onBlur={() => onBlurField?.('city_id')}
            className={
              (value.province_id ? FIELD_SELECT : FIELD_SELECT_DISABLED) +
              (errors?.city_id ? ' border-[#DC2626]' : '')
            }
          >
            <option value="">
              {value.province_id ? 'Pilih kota/kabupaten…' : 'Pilih provinsi dulu'}
            </option>
            {cities.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
          <SelectChevron />
        </div>
        {errors?.city_id && (
          <p className="mt-1 text-[11px] text-[#DC2626]">{errors.city_id}</p>
        )}
      </div>

      <div>
        <label className={FIELD_LABEL}>
          Kecamatan <span className="text-[#DC2626]">*</span>
        </label>
        <div className="relative">
          <select
            required
            disabled={!value.city_id}
            value={value.district_id ?? ''}
            onChange={(e) =>
              onChange({
                province_id: value.province_id,
                city_id: value.city_id,
                district_id: Number(e.target.value) || undefined,
              })
            }
            onBlur={() => onBlurField?.('district_id')}
            className={
              (value.city_id ? FIELD_SELECT : FIELD_SELECT_DISABLED) +
              (errors?.district_id ? ' border-[#DC2626]' : '')
            }
          >
            <option value="">{value.city_id ? 'Pilih kecamatan…' : 'Pilih kota dulu'}</option>
            {districts.map((d) => (
              <option key={d.id} value={d.id}>
                {d.name}
              </option>
            ))}
          </select>
          <SelectChevron />
        </div>
        {errors?.district_id && (
          <p className="mt-1 text-[11px] text-[#DC2626]">{errors.district_id}</p>
        )}
      </div>

      <div>
        <label className={FIELD_LABEL}>
          Kelurahan <span className="text-[#DC2626]">*</span>
        </label>
        <div className="relative">
          <select
            required
            disabled={!value.district_id}
            value={value.village_id ?? ''}
            onChange={onVillageChange}
            onBlur={() => onBlurField?.('village_id')}
            className={
              (value.district_id ? FIELD_SELECT : FIELD_SELECT_DISABLED) +
              (errors?.village_id ? ' border-[#DC2626]' : '')
            }
          >
            <option value="">
              {value.district_id ? 'Pilih kelurahan…' : 'Pilih kecamatan dulu'}
            </option>
            {villages.map((v) => (
              <option key={v.id} value={v.id}>
                {v.name}
              </option>
            ))}
          </select>
          <SelectChevron />
        </div>
        {errors?.village_id && (
          <p className="mt-1 text-[11px] text-[#DC2626]">{errors.village_id}</p>
        )}
      </div>
    </>
  )
}
