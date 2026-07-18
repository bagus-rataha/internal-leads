import { SimpleEntityList } from '@/components/masterdata/SimpleEntityList'
import { useTeams, useCreateTeam, useUpdateTeam } from './queries'

export default function SalesTeamPage() {
  const { data: teams, isLoading } = useTeams()
  const createTeam = useCreateTeam()
  const updateTeam = useUpdateTeam()

  return (
    <SimpleEntityList
      title="Sales Team"
      addLabel="Tambah Tim"
      items={teams as { id: string; name: string; is_active: boolean }[] | undefined}
      isLoading={isLoading}
      onCreate={(name) => createTeam.mutateAsync({ name })}
      onUpdate={(id, input) => updateTeam.mutateAsync({ id, input })}
      isSaving={createTeam.isPending || updateTeam.isPending}
    />
  )
}
