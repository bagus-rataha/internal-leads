import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  fetchUsers,
  createUser,
  updateUser,
  resetPassword,
  deactivateUser,
  type CreateUserInput,
  type UpdateUserInput,
  type ResetPasswordInput,
  type DeactivateUserInput,
} from './api'

export function useUsers(role?: string) {
  return useQuery({ queryKey: ['users', role ?? 'all'], queryFn: () => fetchUsers(role) })
}

export function useCreateUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateUserInput) => createUser(input),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['users'] }),
  })
}

export function useUpdateUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateUserInput }) => updateUser(id, input),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['users'] }),
  })
}

export function useResetPassword() {
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: ResetPasswordInput }) =>
      resetPassword(id, input),
  })
}

export function useDeactivateUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: DeactivateUserInput }) =>
      deactivateUser(id, input),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['users'] }),
  })
}
