import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import type { ColumnDef } from '@tanstack/react-table'
import type { Route as RouteType, Slug } from '@modulacms/admin-sdk'
import { Globe, MoreHorizontal, Plus, Pencil, Trash2 } from 'lucide-react'
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
import { useRoutes, useCreateRoute, useUpdateRoute, useDeleteRoute } from '@/queries/routes'
import { useAuthContext } from '@/lib/auth'
import { formatDate } from '@/lib/utils'

export const Route = createFileRoute('/_admin/routes/')({
  component: PublicRoutesPage,
})

function PublicRoutesPage() {
  const { user } = useAuthContext()
  const { data: routes, isLoading } = useRoutes()
  const createRoute = useCreateRoute()
  const updateRoute = useUpdateRoute()
  const deleteRoute = useDeleteRoute()

  const [createOpen, setCreateOpen] = useState(false)
  const [createSlug, setCreateSlug] = useState('')
  const [createTitle, setCreateTitle] = useState('')
  const [createActive, setCreateActive] = useState(true)

  const [editOpen, setEditOpen] = useState(false)
  const [editTarget, setEditTarget] = useState<RouteType | null>(null)
  const [editSlug, setEditSlug] = useState('')
  const [editTitle, setEditTitle] = useState('')
  const [editActive, setEditActive] = useState(true)

  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<RouteType | null>(null)

  function handleCreate() {
    const now = new Date().toISOString()
    createRoute.mutate(
      {
        slug: createSlug as Slug,
        title: createTitle,
        status: createActive ? 1 : 0,
        author_id: user?.user_id ?? null,
        date_created: now,
        date_modified: now,
      } as Parameters<typeof createRoute.mutate>[0],
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

  function openEdit(route: RouteType) {
    setEditTarget(route)
    setEditSlug(route.slug)
    setEditTitle(route.title)
    setEditActive(route.status === 1)
    setEditOpen(true)
  }

  function handleEdit() {
    if (!editTarget) return
    updateRoute.mutate(
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
    deleteRoute.mutate(deleteTarget.route_id, {
      onSuccess: () => {
        setDeleteOpen(false)
        setDeleteTarget(null)
      },
    })
  }

  const columns: ColumnDef<RouteType>[] = [
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
        <h1 className="text-2xl font-bold">Public Routes</h1>
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Public Routes</h1>
          <p className="text-muted-foreground">
            Configure the public-facing URL routes for your site content.
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Route
        </Button>
      </div>

      {routes && routes.length > 0 ? (
        <DataTable
          columns={columns}
          data={routes}
          searchKey="title"
          searchPlaceholder="Search routes by title..."
        />
      ) : (
        <EmptyState
          icon={Globe}
          title="No routes yet"
          description="Create your first route to map a URL to content."
          action={
            <Button onClick={() => setCreateOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Create Route
            </Button>
          }
        />
      )}

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Route</DialogTitle>
            <DialogDescription>
              Add a new public-facing URL route.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="create-slug">Slug</Label>
              <Input
                id="create-slug"
                placeholder="e.g. /about"
                value={createSlug}
                onChange={(e) => setCreateSlug(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="create-title">Title</Label>
              <Input
                id="create-title"
                placeholder="e.g. About Page"
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
              disabled={createRoute.isPending}
            >
              Cancel
            </Button>
            <Button
              onClick={handleCreate}
              disabled={!createSlug.trim() || !createTitle.trim() || createRoute.isPending}
            >
              {createRoute.isPending ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Route</DialogTitle>
            <DialogDescription>
              Update the route slug, title, and status.
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
              disabled={updateRoute.isPending}
            >
              Cancel
            </Button>
            <Button
              onClick={handleEdit}
              disabled={!editSlug.trim() || !editTitle.trim() || updateRoute.isPending}
            >
              {updateRoute.isPending ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title="Delete Route"
        description={`Are you sure you want to delete the route "${deleteTarget?.title}"? This action cannot be undone.`}
        onConfirm={handleDelete}
        loading={deleteRoute.isPending}
        variant="destructive"
      />
    </div>
  )
}
