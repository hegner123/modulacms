import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type {
  RouteID,
  AdminRouteID,
  CreateRouteParams,
  UpdateRouteParams,
  CreateAdminRouteParams,
  UpdateAdminRouteParams,
} from '@modulacms/admin-sdk'
import { sdk } from '@/lib/sdk'
import { queryKeys } from '@/lib/query-keys'

// Public routes

export function useRoutes() {
  return useQuery({
    queryKey: queryKeys.routes.list(),
    queryFn: () => sdk.routes.list(),
  })
}

export function useCreateRoute() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: CreateRouteParams) => sdk.routes.create(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.routes.all })
    },
  })
}

export function useUpdateRoute() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: UpdateRouteParams) => sdk.routes.update(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.routes.all })
    },
  })
}

export function useDeleteRoute() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: RouteID) => sdk.routes.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.routes.all })
    },
  })
}

// Admin routes

export function useAdminRoutesList() {
  return useQuery({
    queryKey: queryKeys.adminRoutes.list(),
    queryFn: () => sdk.adminRoutes.list(),
  })
}

export function useCreateAdminRoute() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: CreateAdminRouteParams) => sdk.adminRoutes.create(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.adminRoutes.all })
    },
  })
}

export function useUpdateAdminRoute() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: UpdateAdminRouteParams) => sdk.adminRoutes.update(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.adminRoutes.all })
    },
  })
}

export function useDeleteAdminRoute() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: AdminRouteID) => sdk.adminRoutes.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.adminRoutes.all })
    },
  })
}
