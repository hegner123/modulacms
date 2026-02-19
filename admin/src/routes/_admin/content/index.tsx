import { useState, useMemo } from 'react'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import type { ColumnDef } from '@tanstack/react-table'
import type { Route as RouteType, Slug, DatatypeID, UserID, ContentStatus } from '@modulacms/admin-sdk'
import { FileText, MoreHorizontal, Plus, Trash2 } from 'lucide-react'
import { useRoutes, useCreateRoute, useDeleteRoute } from '@/queries/routes'
import { useDatatypes } from '@/queries/datatypes'
import { useContentData, useCreateContentData, useDeleteContentData, useContentFields, useDeleteContentField } from '@/queries/content'
import { useAuthContext } from '@/lib/auth'
import { DataTable } from '@/components/data-table/data-table'
import { DataTableColumnHeader } from '@/components/data-table/column-header'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { EmptyState } from '@/components/shared/empty-state'
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { formatDate } from '@/lib/utils'

export const Route = createFileRoute('/_admin/content/')({
  component: ContentPage,
})

function ContentPage() {
  const navigate = useNavigate()
  const { user } = useAuthContext()
  const { data: routes, isLoading: routesLoading } = useRoutes()
  const { data: datatypes } = useDatatypes()
  const { data: contentData, isLoading: contentLoading } = useContentData()
  const { data: contentFields } = useContentFields()
  const createRoute = useCreateRoute()
  const createContent = useCreateContentData()
  const deleteContentField = useDeleteContentField()
  const deleteContentData = useDeleteContentData()
  const deleteRoute = useDeleteRoute()

  const [createOpen, setCreateOpen] = useState(false)
  const [createTitle, setCreateTitle] = useState('')
  const [createSlug, setCreateSlug] = useState('')
  const [selectedDatatypeId, setSelectedDatatypeId] = useState('')
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<RouteType | null>(null)
  const [deleting, setDeleting] = useState(false)

  const [creating, setCreating] = useState(false)

  async function handleCreate() {
    const now = new Date().toISOString()
    setCreating(true)
    const route = await createRoute.mutateAsync({
      slug: createSlug as Slug,
      title: createTitle,
      status: 1,
      author_id: (user?.user_id ?? null) as RouteType['author_id'],
      date_created: now,
      date_modified: now,
    })
    await createContent.mutateAsync({
      route_id: route.route_id,
      datatype_id: selectedDatatypeId as DatatypeID,
      parent_id: null,
      first_child_id: null,
      next_sibling_id: null,
      prev_sibling_id: null,
      author_id: (user?.user_id ?? null) as UserID | null,
      status: 'draft' as ContentStatus,
      date_created: now,
      date_modified: now,
    })
    setCreateOpen(false)
    setCreateTitle('')
    setCreateSlug('')
    setSelectedDatatypeId('')
    setCreating(false)
  }

  async function handleDelete() {
    if (!deleteTarget) return
    setDeleting(true)
    const entries = (contentData ?? []).filter(
      (cd) => cd.route_id === deleteTarget.route_id,
    )
    for (const entry of entries) {
      const fields = (contentFields ?? []).filter(
        (cf) => cf.content_data_id === entry.content_data_id,
      )
      for (const field of fields) {
        await deleteContentField.mutateAsync(field.content_field_id)
      }
      await deleteContentData.mutateAsync(entry.content_data_id)
    }
    deleteRoute.mutate(deleteTarget.route_id, {
      onSuccess: () => {
        setDeleteOpen(false)
        setDeleteTarget(null)
        setDeleting(false)
      },
      onError: () => {
        setDeleting(false)
      },
    })
  }

  const columns = useMemo<ColumnDef<RouteType>[]>(
    () => [
      {
        accessorKey: 'title',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title="Title" />
        ),
        cell: ({ row }) => (
          <button
            type="button"
            className="font-medium hover:underline"
            onClick={() =>
              navigate({ to: '/content/$title', params: { title: encodeURIComponent(row.original.title.toLowerCase()) } })
            }
          >
            {row.getValue('title')}
          </button>
        ),
      },
      {
        accessorKey: 'slug',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title="Slug" />
        ),
        cell: ({ row }) => (
          <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-sm">
            {row.getValue('slug')}
          </code>
        ),
      },
      {
        accessorKey: 'status',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title="Status" />
        ),
        cell: ({ row }) => {
          const status = row.getValue<number>('status')
          return status === 1 ? (
            <Badge variant="default">Active</Badge>
          ) : (
            <Badge variant="secondary">Inactive</Badge>
          )
        },
      },
      {
        accessorKey: 'date_created',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title="Created" />
        ),
        cell: ({ row }) => formatDate(row.getValue('date_created')),
      },
      {
        id: 'actions',
        cell: ({ row }) => (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                variant="destructive"
                onClick={() => {
                  setDeleteTarget(row.original)
                  setDeleteOpen(true)
                }}
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        ),
      },
    ],
    [navigate],
  )

  if (routesLoading || contentLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    )
  }

  const data = routes ?? []

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Content</h1>
          <p className="text-muted-foreground">
            Browse and manage content entries organized by their route structure.
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Content
        </Button>
      </div>

      {data.length === 0 ? (
        <EmptyState
          icon={FileText}
          title="No content routes"
          description="Create a new route to start organizing content."
          action={
            <Button onClick={() => setCreateOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Create Content
            </Button>
          }
        />
      ) : (
        <DataTable
          columns={columns}
          data={data}
          searchKey="title"
          searchPlaceholder="Search by title..."
        />
      )}

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Content</DialogTitle>
            <DialogDescription>
              Define a new route for your content.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Title</Label>
              <Input
                placeholder="e.g. Home"
                value={createTitle}
                onChange={(e) => setCreateTitle(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label>Slug</Label>
              <Input
                placeholder="e.g. /home"
                value={createSlug}
                onChange={(e) => setCreateSlug(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label>Datatype</Label>
              <Select value={selectedDatatypeId} onValueChange={setSelectedDatatypeId}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Select a datatype" />
                </SelectTrigger>
                <SelectContent>
                  {datatypes?.filter((dt) => dt.type === 'ROOT').map((dt) => (
                    <SelectItem key={dt.datatype_id} value={dt.datatype_id}>
                      {dt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setCreateOpen(false)}
              disabled={creating}
            >
              Cancel
            </Button>
            <Button
              onClick={handleCreate}
              disabled={!createTitle || !createSlug || !selectedDatatypeId || creating}
            >
              {creating ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title="Delete Route"
        description={`Are you sure you want to delete the route "${deleteTarget?.title}" and all its content data? This action cannot be undone.`}
        onConfirm={handleDelete}
        loading={deleting}
        variant="destructive"
      />
    </div>
  )
}
