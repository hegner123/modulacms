import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type {
  UserID,
  RoleID,
  PermissionID,
  RolePermissionID,
  CreateUserParams,
  UpdateUserParams,
  CreateRoleParams,
  UpdateRoleParams,
  CreatePermissionParams,
  UpdatePermissionParams,
  CreateRolePermissionParams,
  CreateTokenParams,
  CreateSshKeyRequest,
} from '@modulacms/admin-sdk'
import { toast } from 'sonner'
import { sdk } from '@/lib/sdk'
import { queryKeys } from '@/lib/query-keys'

// Users

export function useUsers() {
  return useQuery({
    queryKey: queryKeys.users.list(),
    queryFn: () => sdk.users.list(),
  })
}

export function useUser(id: string) {
  return useQuery({
    queryKey: queryKeys.users.detail(id),
    queryFn: () => sdk.users.get(id as UserID),
    enabled: !!id,
  })
}

export function useCreateUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: CreateUserParams) => sdk.auth.register(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all })
    },
  })
}

export function useUpdateUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: UpdateUserParams) => sdk.users.update(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all })
    },
  })
}

export function useDeleteUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: UserID) => sdk.users.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all })
    },
  })
}

// Roles

export function useRoles() {
  return useQuery({
    queryKey: queryKeys.roles.list(),
    queryFn: () => sdk.roles.list(),
  })
}

export function useCreateRole() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: CreateRoleParams) => sdk.roles.create(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.roles.all })
    },
    onError: (err) => toast.error(err.message),
  })
}

export function useUpdateRole() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: UpdateRoleParams) => sdk.roles.update(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.roles.all })
    },
    onError: (err) => toast.error(err.message),
  })
}

export function useDeleteRole() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: RoleID) => sdk.roles.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.roles.all })
    },
    onError: (err) => toast.error(err.message),
  })
}

// Permissions

export function usePermissions() {
  return useQuery({
    queryKey: queryKeys.permissions.list(),
    queryFn: () => sdk.permissions.list(),
  })
}

export function useCreatePermission() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: CreatePermissionParams) => sdk.permissions.create(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.permissions.all })
    },
    onError: (err) => toast.error(err.message),
  })
}

export function useUpdatePermission() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: UpdatePermissionParams) => sdk.permissions.update(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.permissions.all })
    },
    onError: (err) => toast.error(err.message),
  })
}

export function useDeletePermission() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: PermissionID) => sdk.permissions.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.permissions.all })
    },
    onError: (err) => toast.error(err.message),
  })
}

// Role Permissions

export function useAllRolePermissions() {
  return useQuery({
    queryKey: queryKeys.rolePermissions.list(),
    queryFn: () => sdk.rolePermissions.list(),
  })
}

export function useCreateRolePermission() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: CreateRolePermissionParams) => sdk.rolePermissions.create(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rolePermissions.all })
    },
    onError: (err) => toast.error(err.message),
  })
}

export function useDeleteRolePermission() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: RolePermissionID) => sdk.rolePermissions.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rolePermissions.all })
    },
    onError: (err) => toast.error(err.message),
  })
}

// Tokens

export function useTokens() {
  return useQuery({
    queryKey: queryKeys.tokens.list(),
    queryFn: () => sdk.tokens.list(),
  })
}

export function useCreateToken() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: CreateTokenParams) => sdk.tokens.create(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.tokens.all })
    },
  })
}

export function useDeleteToken() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => sdk.tokens.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.tokens.all })
    },
  })
}

// SSH Keys

export function useSshKeys() {
  return useQuery({
    queryKey: queryKeys.sshKeys.list(),
    queryFn: () => sdk.sshKeys.list(),
  })
}

export function useCreateSshKey() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: CreateSshKeyRequest) => sdk.sshKeys.create(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sshKeys.all })
    },
  })
}

export function useDeleteSshKey() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => sdk.sshKeys.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sshKeys.all })
    },
  })
}
