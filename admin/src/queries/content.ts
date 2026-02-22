import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type {
  Slug,
  AdminContentID,
  ContentID,
  ContentFieldID,
  AdminTreeResponse,
  CreateAdminContentDataParams,
  UpdateAdminContentDataParams,
  CreateAdminContentFieldParams,
  UpdateAdminContentFieldParams,
  CreateContentDataParams,
  UpdateContentDataParams,
  CreateContentFieldParams,
  UpdateContentFieldParams,
  ReorderContentDataParams,
} from '@modulacms/admin-sdk'
import { sdk, cms } from '@/lib/sdk'
import { queryKeys } from '@/lib/query-keys'

export function useAdminRoutes() {
  return useQuery({
    queryKey: queryKeys.adminRoutes.ordered(),
    queryFn: () => sdk.adminRoutes.listOrdered(),
  })
}

export function useAdminTree(slug: string) {
  return useQuery({
    queryKey: queryKeys.adminTree.bySlug(slug),
    queryFn: () => sdk.adminTree.get(slug as Slug) as Promise<AdminTreeResponse>,
    enabled: !!slug,
  })
}

export function useTree(slug: string) {
  return useQuery({
    queryKey: queryKeys.tree.bySlug(slug),
    queryFn: () => cms.getPage(slug, { format: 'raw' }),
    enabled: !!slug,
  })
}

export function useAdminContentData() {
  return useQuery({
    queryKey: queryKeys.adminContentData.list(),
    queryFn: () => sdk.adminContentData.list(),
  })
}

export function useCreateAdminContentData() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: CreateAdminContentDataParams) => sdk.adminContentData.create(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.adminContentData.all })
      queryClient.invalidateQueries({ queryKey: queryKeys.adminTree.all })
    },
  })
}

export function useUpdateAdminContentData() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: UpdateAdminContentDataParams) => sdk.adminContentData.update(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.adminContentData.all })
      queryClient.invalidateQueries({ queryKey: queryKeys.adminTree.all })
    },
  })
}

export function useDeleteAdminContentData() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: AdminContentID) => sdk.adminContentData.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.adminContentData.all })
      queryClient.invalidateQueries({ queryKey: queryKeys.adminTree.all })
    },
  })
}

export function useAdminContentFields() {
  return useQuery({
    queryKey: queryKeys.adminContentFields.list(),
    queryFn: () => sdk.adminContentFields.list(),
  })
}

export function useCreateAdminContentField() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: CreateAdminContentFieldParams) => sdk.adminContentFields.create(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.adminContentFields.all })
      queryClient.invalidateQueries({ queryKey: queryKeys.adminTree.all })
    },
  })
}

export function useUpdateAdminContentField() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: UpdateAdminContentFieldParams) => sdk.adminContentFields.update(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.adminContentFields.all })
      queryClient.invalidateQueries({ queryKey: queryKeys.adminTree.all })
    },
  })
}

// Public content data

export function useContentData() {
  return useQuery({
    queryKey: queryKeys.contentData.list(),
    queryFn: () => sdk.contentData.list(),
  })
}

export function useCreateContentData() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: CreateContentDataParams) => sdk.contentData.create(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.contentData.all })
      queryClient.invalidateQueries({ queryKey: queryKeys.tree.all })
    },
  })
}

export function useUpdateContentData() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: UpdateContentDataParams) => sdk.contentData.update(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.contentData.all })
      queryClient.invalidateQueries({ queryKey: queryKeys.tree.all })
    },
  })
}

export function useDeleteContentData() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: ContentID) => sdk.contentData.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.contentData.all })
      queryClient.invalidateQueries({ queryKey: queryKeys.tree.all })
    },
  })
}

export function useReorderContentData() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: ReorderContentDataParams) => sdk.contentData.reorder(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.contentData.all })
      queryClient.invalidateQueries({ queryKey: queryKeys.tree.all })
    },
  })
}

export function useContentFields() {
  return useQuery({
    queryKey: queryKeys.contentFields.list(),
    queryFn: () => sdk.contentFields.list(),
  })
}

export function useCreateContentField() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: CreateContentFieldParams) => sdk.contentFields.create(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.contentFields.all })
      queryClient.invalidateQueries({ queryKey: queryKeys.tree.all })
    },
  })
}

export function useUpdateContentField() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: UpdateContentFieldParams) => sdk.contentFields.update(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.contentFields.all })
      queryClient.invalidateQueries({ queryKey: queryKeys.tree.all })
    },
  })
}

export function useDeleteContentField() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: ContentFieldID) => sdk.contentFields.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.contentFields.all })
      queryClient.invalidateQueries({ queryKey: queryKeys.tree.all })
    },
  })
}
