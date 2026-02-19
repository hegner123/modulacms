import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { RouteApprovalItem, HookApprovalItem } from '@modulacms/admin-sdk'
import { sdk } from '@/lib/sdk'
import { queryKeys } from '@/lib/query-keys'

// Plugins

export function usePlugins() {
  return useQuery({
    queryKey: queryKeys.plugins.list(),
    queryFn: () => sdk.plugins.list(),
  })
}

export function usePlugin(name: string) {
  return useQuery({
    queryKey: queryKeys.plugins.detail(name),
    queryFn: () => sdk.plugins.get(name),
    enabled: !!name,
  })
}

export function useReloadPlugin() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (name: string) => sdk.plugins.reload(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.plugins.all })
    },
  })
}

export function useEnablePlugin() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (name: string) => sdk.plugins.enable(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.plugins.all })
    },
  })
}

export function useDisablePlugin() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (name: string) => sdk.plugins.disable(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.plugins.all })
    },
  })
}

// Plugin Routes

export function usePluginRoutes() {
  return useQuery({
    queryKey: queryKeys.pluginRoutes.list(),
    queryFn: () => sdk.pluginRoutes.list(),
  })
}

export function useApprovePluginRoutes() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (routes: RouteApprovalItem[]) =>
      sdk.pluginRoutes.approve(routes),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.pluginRoutes.all })
    },
  })
}

export function useRevokePluginRoutes() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (routes: RouteApprovalItem[]) =>
      sdk.pluginRoutes.revoke(routes),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.pluginRoutes.all })
    },
  })
}

// Plugin Hooks

export function usePluginHooks() {
  return useQuery({
    queryKey: queryKeys.pluginHooks.list(),
    queryFn: () => sdk.pluginHooks.list(),
  })
}

export function useApprovePluginHooks() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (hooks: HookApprovalItem[]) =>
      sdk.pluginHooks.approve(hooks),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.pluginHooks.all })
    },
  })
}

export function useRevokePluginHooks() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (hooks: HookApprovalItem[]) =>
      sdk.pluginHooks.revoke(hooks),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.pluginHooks.all })
    },
  })
}
