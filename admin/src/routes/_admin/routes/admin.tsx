import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import type { ColumnDef } from '@tanstack/react-table'
import type { AdminRoute, Slug, UserID } from '@modulacms/admin-sdk'
import { Shield, MoreHorizontal, Plus, Pencil, Trash2 } from 'lucide-react'
import { DataTable } from '@/components/data-table/data-table'
import { DataTableColumnHeader } from '@/components/data-table/column-header'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { EmptyState } from '@/components/shared/empty-state'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
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
import {
  useAdminRoutesList,
  useCreateAdminRoute,
  useUpdateAdminRoute,
  useDeleteAdminRoute,
} from '@/queries/routes'
import { useAuthContext } from '@/lib/auth'
import { formatDate } from '@/lib/utils'

export const Route = createFileRoute('/_admin/routes/admin')({
  component: AdminRoutesPage,
})

function AdminRoutesPage() {
  const { user } = useAuthContext()
  const { data: adminRoutes, isLoading } = useAdminRoutesList()
  const createAdminRoute = useCreateAdminRoute()
  const updateAdminRoute = useUpdateAdminRoute()
  const deleteAdminRoute = useDeleteAdminRoute()

  const [createOpen, setCreateOpen] = useState(false)
  const [createSlug, setCreateSlug] = useState('')
  const [createTitle, setCreateTitle] = useState('')
  const [createActive, setCreateActive] = useState(true)

  const [editOpen, setEditOpen] = useState(false)
  const [editTarget, setEditTarget] = useState<AdminRoute | null>(null)
  const [editSlug, setEditSlug] = useState('')
  const [editTitle, setEditTitle] = useState('')
  const [editActive, setEditActive] = useState(true)

  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<AdminRoute | null>(null)

  function handleCreate() {
    const now = new Date().toISOString()
    createAdminRoute.mutate(
      {
        slug: createSlug as Slug,
        title: createTitle,
        status: createActive ? 1 : 0,
        author_id: (user?.user_id ?? null) as UserID | null,
        date_created: now,
        date_modified: now,
      },
      {
        onSuccess: () => {
          setCreateOpen(false)
          setCreateSlug('')
          setCreateTitle('')
          setCreateActive(true)
        },
      },
    )
  }

  function openEdit(route: AdminRoute) {
    setEditTarget(route)
    setEditSlug(route.slug)
    setEditTitle(route.title)
    setEditActive(route.status === 1)
    setEditOpen(true)
  }

  function handleEdit() {
    if (!editTarget) return
    updateAdminRoute.mutate(
      {
        slug: editSlug as Slug,
        title: editTitle,
        status: editActive ? 1 : 0,
        author_id: editTarget.author_id,
        date_created: editTarget.date_created,
        date_modified: new Date().toISOString(),
        slug_2: editTarget.slug as Slug,
      },
      {
        onSuccess: () => {
          setEditOpen(false)
          setEditTarget(null)
        },
      },
    )
  }

  function handleDelete() {
    if (!deleteTarget) return
    deleteAdminRoute.mutate(deleteTarget.admin_route_id, {
      onSuccess: () => {
        setDeleteOpen(false)
        setDeleteTarget(null)
      },
    })
  }

  const columns: ColumnDef<AdminRoute>[] = [
    {
      accessorKey: 'slug',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Slug" />
      ),
      cell: ({ row }) => (
        <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-sm">
          {row.original.slug}
        </code>
      ),
    },
    {
      accessorKey: 'title',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Title" />
      ),
    },
    {
      accessorKey: 'status',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Status" />
      ),
      cell: ({ row }) => (
        <Badge variant={row.original.status === 1 ? 'default' : 'secondary'}>
          {row.original.status === 1 ? 'Active' : 'Inactive'}
        </Badge>
      ),
    },
    {
      accessorKey: 'date_created',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Created" />
      ),
      cell: ({ row }) => formatDate(row.original.date_created),
    },
    {
      id: 'actions',
      enableHiding: false,
      cell: ({ row }) => {
        const route = row.original
        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <MoreHorizontal className="h-4 w-4" />
                <span className="sr-only">Open menu</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => openEdit(route)}>
                <Pencil className="mr-2 h-4 w-4" />
                Edit
              </DropdownMenuItem>
              <DropdownMenuItem
                variant="destructive"
                onClick={() => {
                  setDeleteTarget(route)
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

  if (isLoading) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Admin Routes</h1>
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Admin Routes</h1>
          <p className="text-muted-foreground">
            Manage admin panel route configurations and access controls.
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Admin Route
        </Button>
      </div>

      {adminRoutes && adminRoutes.length > 0 ? (
        <DataTable
          columns={columns}
          data={adminRoutes}
          searchKey="title"
          searchPlaceholder="Search admin routes by title..."
        />
      ) : (
        <EmptyState
          icon={Shield}
          title="No admin routes yet"
          description="Create your first admin route to configure the panel."
          action={
            <Button onClick={() => setCreateOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Create Admin Route
            </Button>
          }
        />
      )}

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Admin Route</DialogTitle>
            <DialogDescription>
              Add a new admin panel route configuration.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="create-slug">Slug</Label>
              <Input
                id="create-slug"
                placeholder="e.g. /dashboard"
                value={createSlug}
                onChange={(e) => setCreateSlug(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="create-title">Title</Label>
              <Input
                id="create-title"
                placeholder="e.g. Dashboard"
                value={createTitle}
                onChange={(e) => setCreateTitle(e.target.value)}
              />
            </div>
            <div className="flex items-center gap-2">
              <Checkbox
                id="create-active"
                checked={createActive}
                onCheckedChange={(checked) => setCreateActive(checked === true)}
              />
              <Label htmlFor="create-active">Active</Label>
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setCreateOpen(false)}
              disabled={createAdminRoute.isPending}
            >
              Cancel
            </Button>
            <Button
              onClick={handleCreate}
              disabled={!createSlug.trim() || !createTitle.trim() || createAdminRoute.isPending}
            >
              {createAdminRoute.isPending ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Admin Route</DialogTitle>
            <DialogDescription>
              Update the admin route slug, title, and status.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="edit-slug">Slug</Label>
              <Input
                id="edit-slug"
                value={editSlug}
                onChange={(e) => setEditSlug(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-title">Title</Label>
              <Input
                id="edit-title"
                value={editTitle}
                onChange={(e) => setEditTitle(e.target.value)}
              />
            </div>
            <div className="flex items-center gap-2">
              <Checkbox
                id="edit-active"
                checked={editActive}
                onCheckedChange={(checked) => setEditActive(checked === true)}
              />
              <Label htmlFor="edit-active">Active</Label>
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setEditOpen(false)}
              disabled={updateAdminRoute.isPending}
            >
              Cancel
            </Button>
            <Button
              onClick={handleEdit}
              disabled={!editSlug.trim() || !editTitle.trim() || updateAdminRoute.isPending}
            >
              {updateAdminRoute.isPending ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title="Delete Admin Route"
        description={`Are you sure you want to delete the admin route "${deleteTarget?.title}"? This action cannot be undone.`}
        onConfirm={handleDelete}
        loading={deleteAdminRoute.isPending}
        variant="destructive"
      />
    </div>
  )
}
