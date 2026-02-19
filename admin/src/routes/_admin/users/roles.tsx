import { useState, useMemo } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { useQueryClient } from '@tanstack/react-query'
import type { ColumnDef } from '@tanstack/react-table'
import type { Role, Permission, RolePermission, PermissionID } from '@modulacms/admin-sdk'
import { toast } from 'sonner'
import {
  ShieldCheck,
  MoreHorizontal,
  Plus,
  Pencil,
  Trash2,
  Check,
  X,
  Loader2,
  KeyRound,
} from 'lucide-react'
import { DataTable } from '@/components/data-table/data-table'
import { DataTableColumnHeader } from '@/components/data-table/column-header'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { EmptyState } from '@/components/shared/empty-state'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { sdk } from '@/lib/sdk'
import { queryKeys } from '@/lib/query-keys'
import {
  useRoles,
  useCreateRole,
  useUpdateRole,
  useDeleteRole,
  usePermissions,
  useCreatePermission,
  useUpdatePermission,
  useDeletePermission,
  useAllRolePermissions,
} from '@/queries/users'

export const Route = createFileRoute('/_admin/users/roles')({
  component: RolesPage,
})

function RolesPage() {
  const { data: roles, isLoading: rolesLoading } = useRoles()
  const { data: permissions, isLoading: permissionsLoading } = usePermissions()
  const { data: allRolePermissions, isLoading: rpLoading } = useAllRolePermissions()

  const rpByRole = useMemo(() => {
    if (!allRolePermissions) return new Map<string, RolePermission[]>()
    const map = new Map<string, RolePermission[]>()
    for (const rp of allRolePermissions) {
      const existing = map.get(rp.role_id)
      if (existing) {
        existing.push(rp)
      } else {
        map.set(rp.role_id, [rp])
      }
    }
    return map
  }, [allRolePermissions])

  const permissionsById = useMemo(() => {
    if (!permissions) return new Map<string, Permission>()
    const map = new Map<string, Permission>()
    for (const p of permissions) {
      map.set(p.permission_id, p)
    }
    return map
  }, [permissions])

  const isLoading = rolesLoading || permissionsLoading || rpLoading

  if (isLoading) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Roles & Permissions</h1>
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      <RolesSection
        roles={roles ?? []}
        permissions={permissions ?? []}
        allRolePermissions={allRolePermissions ?? []}
        rpByRole={rpByRole}
        permissionsById={permissionsById}
      />
      <PermissionsCatalog
        permissions={permissions ?? []}
        allRolePermissions={allRolePermissions ?? []}
      />
    </div>
  )
}

// ---------------------------------------------------------------------------
// Section 1: Roles
// ---------------------------------------------------------------------------

function RolesSection({
  roles,
  permissions,
  allRolePermissions,
  rpByRole,
  permissionsById,
}: {
  roles: Role[]
  permissions: Permission[]
  allRolePermissions: RolePermission[]
  rpByRole: Map<string, RolePermission[]>
  permissionsById: Map<string, Permission>
}) {
  const queryClient = useQueryClient()
  const createRole = useCreateRole()
  const updateRole = useUpdateRole()
  const deleteRole = useDeleteRole()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<'create' | 'edit'>('create')
  const [dialogLabel, setDialogLabel] = useState('')
  const [dialogEditTarget, setDialogEditTarget] = useState<Role | null>(null)
  const [dialogChecked, setDialogChecked] = useState<Set<PermissionID>>(new Set())
  const [dialogSaving, setDialogSaving] = useState(false)

  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<Role | null>(null)

  function openCreate() {
    setDialogMode('create')
    setDialogLabel('')
    setDialogEditTarget(null)
    setDialogChecked(new Set())
    setDialogOpen(true)
  }

  function openEdit(role: Role) {
    setDialogMode('edit')
    setDialogLabel(role.label)
    setDialogEditTarget(role)
    const current = rpByRole.get(role.role_id) ?? []
    setDialogChecked(new Set(current.map((rp) => rp.permission_id)))
    setDialogOpen(true)
  }

  function togglePermission(pid: PermissionID) {
    setDialogChecked((prev) => {
      const next = new Set(prev)
      if (next.has(pid)) {
        next.delete(pid)
      } else {
        next.add(pid)
      }
      return next
    })
  }

  async function handleSave() {
    setDialogSaving(true)
    try {
      if (dialogMode === 'create') {
        const role = await createRole.mutateAsync({ label: dialogLabel })
        const results = await Promise.allSettled(
          Array.from(dialogChecked).map((pid) =>
            sdk.rolePermissions.create({ role_id: role.role_id, permission_id: pid }),
          ),
        )
        await queryClient.invalidateQueries({ queryKey: queryKeys.rolePermissions.all })
        const failures = results.filter((r) => r.status === 'rejected')
        if (failures.length > 0) {
          toast.error(`${failures.length} permission(s) failed to assign`)
          return
        }
        setDialogOpen(false)
      } else if (dialogEditTarget) {
        await updateRole.mutateAsync({
          role_id: dialogEditTarget.role_id,
          label: dialogLabel,
        })
        const currentPerms = allRolePermissions.filter(
          (rp) => rp.role_id === dialogEditTarget.role_id,
        )
        const toAdd = Array.from(dialogChecked).filter(
          (pid) => !currentPerms.some((rp) => rp.permission_id === pid),
        )
        const toRemove = currentPerms.filter(
          (rp) => !dialogChecked.has(rp.permission_id),
        )
        const results = await Promise.allSettled([
          ...toAdd.map((pid) =>
            sdk.rolePermissions.create({
              role_id: dialogEditTarget.role_id,
              permission_id: pid,
            }),
          ),
          ...toRemove.map((rp) => sdk.rolePermissions.remove(rp.id)),
        ])
        await queryClient.invalidateQueries({ queryKey: queryKeys.rolePermissions.all })
        const failures = results.filter((r) => r.status === 'rejected')
        if (failures.length > 0) {
          toast.error(`${failures.length} permission change(s) failed`)
          return
        }
        setDialogOpen(false)
      }
    } catch {
      // createRole/updateRole onError already toasts
    } finally {
      setDialogSaving(false)
    }
  }

  function handleDelete() {
    if (!deleteTarget) return
    deleteRole.mutate(deleteTarget.role_id, {
      onSuccess: () => {
        setDeleteOpen(false)
        setDeleteTarget(null)
      },
    })
  }

  const columns: ColumnDef<Role>[] = [
    {
      accessorKey: 'label',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Label" />
      ),
    },
    {
      id: 'permissions',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Permissions" />
      ),
      cell: ({ row }) => {
        const rps = rpByRole.get(row.original.role_id) ?? []
        if (rps.length === 0) {
          return <span className="text-muted-foreground text-sm">None</span>
        }
        return (
          <div className="flex flex-wrap gap-1">
            {rps.map((rp) => {
              const perm = permissionsById.get(rp.permission_id)
              return (
                <Badge key={rp.id} variant="secondary">
                  {perm?.label ?? rp.permission_id}
                </Badge>
              )
            })}
          </div>
        )
      },
    },
    {
      id: 'actions',
      enableHiding: false,
      cell: ({ row }) => {
        const role = row.original
        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <MoreHorizontal className="h-4 w-4" />
                <span className="sr-only">Open menu</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => openEdit(role)}>
                <Pencil className="mr-2 h-4 w-4" />
                Edit
              </DropdownMenuItem>
              <DropdownMenuItem
                variant="destructive"
                onClick={() => {
                  setDeleteTarget(role)
                  setDeleteOpen(true)
                }}
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        )
      },
    },
  ]

  return (
    <>
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Roles & Permissions</h1>
            <p className="text-muted-foreground">
              Define roles and assign permissions to control access across the
              admin panel.
            </p>
          </div>
          <Button onClick={openCreate}>
            <Plus className="mr-2 h-4 w-4" />
            Create Role
          </Button>
        </div>

        {roles.length > 0 ? (
          <DataTable
            columns={columns}
            data={roles}
            searchKey="label"
            searchPlaceholder="Search roles..."
          />
        ) : (
          <EmptyState
            icon={ShieldCheck}
            title="No roles yet"
            description="Create your first role to manage permissions."
            action={
              <Button onClick={openCreate}>
                <Plus className="mr-2 h-4 w-4" />
                Create Role
              </Button>
            }
          />
        )}
      </div>

      {/* Create / Edit Role Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {dialogMode === 'create' ? 'Create Role' : 'Edit Role'}
            </DialogTitle>
            <DialogDescription>
              {dialogMode === 'create'
                ? 'Define a new role and assign permissions.'
                : 'Update the role label and permissions.'}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="role-label">Label</Label>
              <Input
                id="role-label"
                placeholder="e.g. editor"
                value={dialogLabel}
                onChange={(e) => setDialogLabel(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label>Permissions</Label>
              {permissions.length > 0 ? (
                <div className="grid grid-cols-2 gap-2">
                  {permissions.map((p) => (
                    <label
                      key={p.permission_id}
                      className="flex items-center gap-2 rounded-md border px-3 py-2 text-sm cursor-pointer hover:bg-accent"
                    >
                      <Checkbox
                        checked={dialogChecked.has(p.permission_id)}
                        onCheckedChange={() => togglePermission(p.permission_id)}
                      />
                      {p.label}
                    </label>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">
                  No permissions defined yet. Add some in the Permissions Catalog
                  below.
                </p>
              )}
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDialogOpen(false)}
              disabled={dialogSaving}
            >
              Cancel
            </Button>
            <Button
              onClick={handleSave}
              disabled={!dialogLabel.trim() || dialogSaving}
            >
              {dialogSaving ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Saving...
                </>
              ) : dialogMode === 'create' ? (
                'Create'
              ) : (
                'Save'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title="Delete Role"
        description={`Are you sure you want to delete the role "${deleteTarget?.label}"? This action cannot be undone.`}
        onConfirm={handleDelete}
        loading={deleteRole.isPending}
        variant="destructive"
      />
    </>
  )
}

// ---------------------------------------------------------------------------
// Section 2: Permissions Catalog
// ---------------------------------------------------------------------------

function PermissionsCatalog({
  permissions,
  allRolePermissions,
}: {
  permissions: Permission[]
  allRolePermissions: RolePermission[]
}) {
  const createPermission = useCreatePermission()
  const updatePermission = useUpdatePermission()
  const deletePermission = useDeletePermission()

  const [newLabel, setNewLabel] = useState('')
  const [editingId, setEditingId] = useState<PermissionID | null>(null)
  const [editingLabel, setEditingLabel] = useState('')

  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<Permission | null>(null)

  function handleAdd() {
    if (!newLabel.trim()) return
    createPermission.mutate(
      { label: newLabel.trim() },
      {
        onSuccess: () => setNewLabel(''),
      },
    )
  }

  function startEdit(perm: Permission) {
    setEditingId(perm.permission_id)
    setEditingLabel(perm.label)
  }

  function cancelEdit() {
    setEditingId(null)
    setEditingLabel('')
  }

  function saveEdit() {
    if (!editingId || !editingLabel.trim()) return
    updatePermission.mutate(
      { permission_id: editingId, label: editingLabel.trim() },
      {
        onSuccess: () => {
          setEditingId(null)
          setEditingLabel('')
        },
      },
    )
  }

  function confirmDelete() {
    if (!deleteTarget) return
    deletePermission.mutate(deleteTarget.permission_id, {
      onSuccess: () => {
        setDeleteOpen(false)
        setDeleteTarget(null)
      },
    })
  }

  function countRolesUsing(permissionId: PermissionID): number {
    let count = 0
    for (const rp of allRolePermissions) {
      if (rp.permission_id === permissionId) count++
    }
    return count
  }

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Permissions Catalog</CardTitle>
          <CardDescription>
            Manage the available permissions that can be assigned to roles.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Add permission inline */}
          <div className="flex gap-2">
            <Input
              placeholder="New permission label..."
              value={newLabel}
              onChange={(e) => setNewLabel(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleAdd()
              }}
            />
            <Button
              onClick={handleAdd}
              disabled={!newLabel.trim() || createPermission.isPending}
              size="sm"
            >
              {createPermission.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <Plus className="mr-2 h-4 w-4" />
              )}
              Add
            </Button>
          </div>

          {/* Permissions list */}
          {permissions.length === 0 ? (
            <EmptyState
              icon={KeyRound}
              title="No permissions yet"
              description="Add your first permission to start building access control."
            />
          ) : (
            <div className="divide-y rounded-md border">
              {permissions.map((perm) => (
                <div
                  key={perm.permission_id}
                  className="flex items-center justify-between px-4 py-2"
                >
                  {editingId === perm.permission_id ? (
                    <div className="flex flex-1 items-center gap-2">
                      <Input
                        value={editingLabel}
                        onChange={(e) => setEditingLabel(e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') saveEdit()
                          if (e.key === 'Escape') cancelEdit()
                        }}
                        className="h-8"
                        autoFocus
                      />
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8"
                        onClick={saveEdit}
                        disabled={
                          !editingLabel.trim() || updatePermission.isPending
                        }
                      >
                        <Check className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8"
                        onClick={cancelEdit}
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </div>
                  ) : (
                    <>
                      <span className="text-sm">{perm.label}</span>
                      <div className="flex items-center gap-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-8 w-8"
                          onClick={() => startEdit(perm)}
                        >
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-8 w-8 text-destructive hover:text-destructive"
                          onClick={() => {
                            setDeleteTarget(perm)
                            setDeleteOpen(true)
                          }}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title="Delete Permission"
        description={
          deleteTarget
            ? (() => {
                const count = countRolesUsing(deleteTarget.permission_id)
                if (count > 0) {
                  return `This permission is assigned to ${count} role(s). Deleting it will remove it from all of them. Are you sure?`
                }
                return `Are you sure you want to delete the permission "${deleteTarget.label}"? This action cannot be undone.`
              })()
            : ''
        }
        onConfirm={confirmDelete}
        loading={deletePermission.isPending}
        variant="destructive"
      />
    </>
  )
}
