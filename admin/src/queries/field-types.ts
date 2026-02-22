import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type {
  FieldTypeID,
  AdminFieldTypeID,
  CreateFieldTypeParams,
  UpdateFieldTypeParams,
  CreateAdminFieldTypeParams,
  UpdateAdminFieldTypeParams,
} from '@modulacms/admin-sdk'
import { toast } from 'sonner'
import { sdk } from '@/lib/sdk'
import { queryKeys } from '@/lib/query-keys'

// Field Types

export function useFieldTypes() {
  return useQuery({
    queryKey: queryKeys.fieldTypes.list(),
    queryFn: () => sdk.fieldTypes.list(),
  })
}

export function useCreateFieldType() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: CreateFieldTypeParams) => sdk.fieldTypes.create(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.fieldTypes.all })
    },
    onError: (err) => toast.error(err.message),
  })
}

export function useUpdateFieldType() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: UpdateFieldTypeParams) => sdk.fieldTypes.update(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.fieldTypes.all })
    },
    onError: (err) => toast.error(err.message),
  })
}

export function useDeleteFieldType() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: FieldTypeID) => sdk.fieldTypes.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.fieldTypes.all })
    },
    onError: (err) => toast.error(err.message),
  })
}

// Admin Field Types

export function useAdminFieldTypes() {
  return useQuery({
    queryKey: queryKeys.adminFieldTypes.list(),
    queryFn: () => sdk.adminFieldTypes.list(),
  })
}

export function useCreateAdminFieldType() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: CreateAdminFieldTypeParams) => sdk.adminFieldTypes.create(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.adminFieldTypes.all })
    },
    onError: (err) => toast.error(err.message),
  })
}

export function useUpdateAdminFieldType() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: UpdateAdminFieldTypeParams) => sdk.adminFieldTypes.update(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.adminFieldTypes.all })
    },
    onError: (err) => toast.error(err.message),
  })
}

export function useDeleteAdminFieldType() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: AdminFieldTypeID) => sdk.adminFieldTypes.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.adminFieldTypes.all })
    },
    onError: (err) => toast.error(err.message),
  })
}
