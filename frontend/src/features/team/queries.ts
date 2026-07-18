import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fetchTeams, createTeam, updateTeam, type CreateTeamInput, type UpdateTeamInput } from './api'

export function useTeams() {
  return useQuery({ queryKey: ['teams'], queryFn: fetchTeams })
}

export function useCreateTeam() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateTeamInput) => createTeam(input),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['teams'] }),
  })
}

export function useUpdateTeam() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateTeamInput }) => updateTeam(id, input),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['teams'] }),
  })
}
