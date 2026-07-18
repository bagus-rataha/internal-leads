import { SimpleEntityList } from '@/components/masterdata/SimpleEntityList'
import { useServiceTypesAdmin, useCreateServiceType, useUpdateServiceType } from './queries'

export default function TypeLayananPage() {
  const { data: types, isLoading } = useServiceTypesAdmin()
  const createType = useCreateServiceType()
  const updateType = useUpdateServiceType()

  return (
    <SimpleEntityList
      title="Type Layanan"
      addLabel="Tambah Type"
      items={types as { id: string; name: string; is_active: boolean }[] | undefined}
      isLoading={isLoading}
      onCreate={(name) => createType.mutateAsync({ name })}
      onUpdate={(id, input) => updateType.mutateAsync({ id, input })}
      isSaving={createType.isPending || updateType.isPending}
    />
  )
}
