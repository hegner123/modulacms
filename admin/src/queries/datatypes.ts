import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { DatatypeID, CreateDatatypeParams, UpdateDatatypeParams, CreateDatatypeFieldParams } from '@modulacms/admin-sdk'
import { sdk } from '@/lib/sdk'
import { queryKeys } from '@/lib/query-keys'

export function useDatatypes() {
  return useQuery({
    queryKey: queryKeys.datatypes.list(),
    queryFn: () => sdk.datatypes.list(),
  })
}

export function useDatatype(id: string) {
  return useQuery({
    queryKey: queryKeys.datatypes.detail(id),
    queryFn: () => sdk.datatypes.get(id as DatatypeID),
    enabled: !!id,
  })
}

export function useCreateDatatype() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: CreateDatatypeParams) => sdk.datatypes.create(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.datatypes.all })
    },
  })
}

export function useUpdateDatatype() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: UpdateDatatypeParams) => sdk.datatypes.update(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.datatypes.all })
    },
  })
}

export function useDeleteDatatype() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: DatatypeID) => sdk.datatypes.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.datatypes.all })
    },
  })
}

export function useDatatypeFields() {
  return useQuery({
    queryKey: queryKeys.datatypeFields.list(),
    queryFn: () => sdk.datatypeFields.list(),
  })
}

export function useDatatypeFieldsByDatatype(datatypeId: string) {
  const { data: allFields, ...rest } = useDatatypeFields()

  const filtered = allFields?.filter(
    (field) => field.datatype_id === datatypeId
  )

  return { data: filtered, ...rest }
}

export function useCreateDatatypeField() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: CreateDatatypeFieldParams) => sdk.datatypeFields.create(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.datatypeFields.all })
    },
  })
}

export function useDeleteDatatypeField() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => sdk.datatypeFields.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.datatypeFields.all })
    },
  })
}
