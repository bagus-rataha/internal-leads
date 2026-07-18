import { SimpleEntityList } from '@/components/masterdata/SimpleEntityList'
import { useLeadSourcesAdmin, useCreateLeadSource, useUpdateLeadSource } from './queries'

export default function LeadSourcePage() {
  const { data: sources, isLoading } = useLeadSourcesAdmin()
  const createSource = useCreateLeadSource()
  const updateSource = useUpdateLeadSource()

  return (
    <SimpleEntityList
      title="Sumber Lead"
      addLabel="Tambah Sumber"
      items={sources as { id: string; name: string; is_active: boolean }[] | undefined}
      isLoading={isLoading}
      onCreate={(name) => createSource.mutateAsync({ name })}
      onUpdate={(id, input) => updateSource.mutateAsync({ id, input })}
      isSaving={createSource.isPending || updateSource.isPending}
    />
  )
}
