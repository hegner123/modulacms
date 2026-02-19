import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { FieldID, CreateFieldParams, UpdateFieldParams } from '@modulacms/admin-sdk'
import { sdk } from '@/lib/sdk'
import { queryKeys } from '@/lib/query-keys'

export function useFields() {
  return useQuery({
    queryKey: queryKeys.fields.list(),
    queryFn: () => sdk.fields.list(),
  })
}

export function useField(id: string) {
  return useQuery({
    queryKey: queryKeys.fields.detail(id),
    queryFn: () => sdk.fields.get(id as FieldID),
    enabled: !!id,
  })
}

export function useCreateField() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: CreateFieldParams) => sdk.fields.create(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.fields.all })
    },
  })
}

export function useUpdateField() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: UpdateFieldParams) => sdk.fields.update(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.fields.all })
    },
  })
}

export function useDeleteField() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: FieldID) => sdk.fields.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.fields.all })
    },
  })
}
