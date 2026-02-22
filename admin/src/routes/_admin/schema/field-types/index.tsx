import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import type { ColumnDef } from '@tanstack/react-table'
import type {
  FieldTypeInfo,
  AdminFieldTypeInfo,
} from '@modulacms/admin-sdk'
import {
  List,
  MoreHorizontal,
  Plus,
  Pencil,
  Trash2,
  Loader2,
} from 'lucide-react'
import { DataTable } from '@/components/data-table/data-table'
import { DataTableColumnHeader } from '@/components/data-table/column-header'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { EmptyState } from '@/components/shared/empty-state'
import { Button } from '@/components/ui/button'
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
import {
  useFieldTypes,
  useCreateFieldType,
  useUpdateFieldType,
  useDeleteFieldType,
  useAdminFieldTypes,
  useCreateAdminFieldType,
  useUpdateAdminFieldType,
  useDeleteAdminFieldType,
} from '@/queries/field-types'

export const Route = createFileRoute('/_admin/schema/field-types/')({
  component: FieldTypesPage,
})

function FieldTypesPage() {
  const { data: fieldTypes, isLoading: ftLoading } = useFieldTypes()
  const { data: adminFieldTypes, isLoading: aftLoading } = useAdminFieldTypes()

  const isLoading = ftLoading || aftLoading

  if (isLoading) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Field Types</h1>
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      <FieldTypesSection fieldTypes={fieldTypes ?? []} />
      <AdminFieldTypesSection adminFieldTypes={adminFieldTypes ?? []} />
    </div>
  )
}

// ---------------------------------------------------------------------------
// Section 1: Field Types (public content fields)
// ---------------------------------------------------------------------------

function FieldTypesSection({ fieldTypes }: { fieldTypes: FieldTypeInfo[] }) {
  const createFieldType = useCreateFieldType()
  const updateFieldType = useUpdateFieldType()
  const deleteFieldType = useDeleteFieldType()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<'create' | 'edit'>('create')
  const [dialogType, setDialogType] = useState('')
  const [dialogLabel, setDialogLabel] = useState('')
  const [dialogEditTarget, setDialogEditTarget] = useState<FieldTypeInfo | null>(null)

  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<FieldTypeInfo | null>(null)

  function openCreate() {
    setDialogMode('create')
    setDialogType('')
    setDialogLabel('')
    setDialogEditTarget(null)
    setDialogOpen(true)
  }

  function openEdit(ft: FieldTypeInfo) {
    setDialogMode('edit')
    setDialogType(ft.type)
    setDialogLabel(ft.label)
    setDialogEditTarget(ft)
    setDialogOpen(true)
  }

  function handleSave() {
    if (dialogMode === 'create') {
      createFieldType.mutate(
        { type: dialogType.trim(), label: dialogLabel.trim() },
        { onSuccess: () => setDialogOpen(false) },
      )
    } else if (dialogEditTarget) {
      updateFieldType.mutate(
        {
          field_type_id: dialogEditTarget.field_type_id,
          type: dialogType.trim(),
          label: dialogLabel.trim(),
        },
        { onSuccess: () => setDialogOpen(false) },
      )
    }
  }

  function handleDelete() {
    if (!deleteTarget) return
    deleteFieldType.mutate(deleteTarget.field_type_id, {
      onSuccess: () => {
        setDeleteOpen(false)
        setDeleteTarget(null)
      },
    })
  }

  const isSaving = createFieldType.isPending || updateFieldType.isPending

  const columns: ColumnDef<FieldTypeInfo>[] = [
    {
      accessorKey: 'type',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Type" />
      ),
      cell: ({ row }) => (
        <code className="rounded bg-muted px-1.5 py-0.5 text-sm">
          {row.original.type}
        </code>
      ),
    },
    {
      accessorKey: 'label',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Label" />
      ),
    },
    {
      id: 'actions',
      enableHiding: false,
      cell: ({ row }) => {
        const ft = row.original
        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <MoreHorizontal className="h-4 w-4" />
                <span className="sr-only">Open menu</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => openEdit(ft)}>
                <Pencil className="mr-2 h-4 w-4" />
                Edit
              </DropdownMenuItem>
              <DropdownMenuItem
                variant="destructive"
                onClick={() => {
                  setDeleteTarget(ft)
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
            <h1 className="text-2xl font-bold">Field Types</h1>
            <p className="text-muted-foreground">
              Manage the available field types for public content fields.
            </p>
          </div>
          <Button onClick={openCreate}>
            <Plus className="mr-2 h-4 w-4" />
            Create Field Type
          </Button>
        </div>

        {fieldTypes.length > 0 ? (
          <DataTable
            columns={columns}
            data={fieldTypes}
            searchKey="type"
            searchPlaceholder="Search field types..."
          />
        ) : (
          <EmptyState
            icon={List}
            title="No field types yet"
            description="Create your first field type to define available types for content fields."
            action={
              <Button onClick={openCreate}>
                <Plus className="mr-2 h-4 w-4" />
                Create Field Type
              </Button>
            }
          />
        )}
      </div>

      {/* Create / Edit Field Type Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {dialogMode === 'create' ? 'Create Field Type' : 'Edit Field Type'}
            </DialogTitle>
            <DialogDescription>
              {dialogMode === 'create'
                ? 'Define a new field type for content fields.'
                : 'Update the field type key and label.'}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="ft-type">Type</Label>
              <Input
                id="ft-type"
                placeholder="e.g. text, number, media"
                value={dialogType}
                onChange={(e) => setDialogType(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="ft-label">Label</Label>
              <Input
                id="ft-label"
                placeholder="e.g. Text, Number, Media"
                value={dialogLabel}
                onChange={(e) => setDialogLabel(e.target.value)}
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDialogOpen(false)}
              disabled={isSaving}
            >
              Cancel
            </Button>
            <Button
              onClick={handleSave}
              disabled={!dialogType.trim() || !dialogLabel.trim() || isSaving}
            >
              {isSaving ? (
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
        title="Delete Field Type"
        description={`Are you sure you want to delete the field type "${deleteTarget?.label}"? This action cannot be undone.`}
        onConfirm={handleDelete}
        loading={deleteFieldType.isPending}
        variant="destructive"
      />
    </>
  )
}

// ---------------------------------------------------------------------------
// Section 2: Admin Field Types (admin content fields)
// ---------------------------------------------------------------------------

function AdminFieldTypesSection({
  adminFieldTypes,
}: {
  adminFieldTypes: AdminFieldTypeInfo[]
}) {
  const createAdminFieldType = useCreateAdminFieldType()
  const updateAdminFieldType = useUpdateAdminFieldType()
  const deleteAdminFieldType = useDeleteAdminFieldType()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<'create' | 'edit'>('create')
  const [dialogType, setDialogType] = useState('')
  const [dialogLabel, setDialogLabel] = useState('')
  const [dialogEditTarget, setDialogEditTarget] =
    useState<AdminFieldTypeInfo | null>(null)

  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<AdminFieldTypeInfo | null>(
    null,
  )

  function openCreate() {
    setDialogMode('create')
    setDialogType('')
    setDialogLabel('')
    setDialogEditTarget(null)
    setDialogOpen(true)
  }

  function openEdit(aft: AdminFieldTypeInfo) {
    setDialogMode('edit')
    setDialogType(aft.type)
    setDialogLabel(aft.label)
    setDialogEditTarget(aft)
    setDialogOpen(true)
  }

  function handleSave() {
    if (dialogMode === 'create') {
      createAdminFieldType.mutate(
        { type: dialogType.trim(), label: dialogLabel.trim() },
        { onSuccess: () => setDialogOpen(false) },
      )
    } else if (dialogEditTarget) {
      updateAdminFieldType.mutate(
        {
          admin_field_type_id: dialogEditTarget.admin_field_type_id,
          type: dialogType.trim(),
          label: dialogLabel.trim(),
        },
        { onSuccess: () => setDialogOpen(false) },
      )
    }
  }

  function handleDelete() {
    if (!deleteTarget) return
    deleteAdminFieldType.mutate(deleteTarget.admin_field_type_id, {
      onSuccess: () => {
        setDeleteOpen(false)
        setDeleteTarget(null)
      },
    })
  }

  const isSaving =
    createAdminFieldType.isPending || updateAdminFieldType.isPending

  const columns: ColumnDef<AdminFieldTypeInfo>[] = [
    {
      accessorKey: 'type',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Type" />
      ),
      cell: ({ row }) => (
        <code className="rounded bg-muted px-1.5 py-0.5 text-sm">
          {row.original.type}
        </code>
      ),
    },
    {
      accessorKey: 'label',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Label" />
      ),
    },
    {
      id: 'actions',
      enableHiding: false,
      cell: ({ row }) => {
        const aft = row.original
        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <MoreHorizontal className="h-4 w-4" />
                <span className="sr-only">Open menu</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => openEdit(aft)}>
                <Pencil className="mr-2 h-4 w-4" />
                Edit
              </DropdownMenuItem>
              <DropdownMenuItem
                variant="destructive"
                onClick={() => {
                  setDeleteTarget(aft)
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
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Admin Field Types</CardTitle>
              <CardDescription>
                Manage the available field types for admin content fields.
              </CardDescription>
            </div>
            <Button size="sm" onClick={openCreate}>
              <Plus className="mr-2 h-4 w-4" />
              Create Admin Field Type
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {adminFieldTypes.length > 0 ? (
            <DataTable
              columns={columns}
              data={adminFieldTypes}
              searchKey="type"
              searchPlaceholder="Search admin field types..."
            />
          ) : (
            <EmptyState
              icon={List}
              title="No admin field types yet"
              description="Create your first admin field type to define available types for admin content fields."
              action={
                <Button onClick={openCreate}>
                  <Plus className="mr-2 h-4 w-4" />
                  Create Admin Field Type
                </Button>
              }
            />
          )}
        </CardContent>
      </Card>

      {/* Create / Edit Admin Field Type Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {dialogMode === 'create'
                ? 'Create Admin Field Type'
                : 'Edit Admin Field Type'}
            </DialogTitle>
            <DialogDescription>
              {dialogMode === 'create'
                ? 'Define a new field type for admin content fields.'
                : 'Update the admin field type key and label.'}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="aft-type">Type</Label>
              <Input
                id="aft-type"
                placeholder="e.g. text, number, media"
                value={dialogType}
                onChange={(e) => setDialogType(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="aft-label">Label</Label>
              <Input
                id="aft-label"
                placeholder="e.g. Text, Number, Media"
                value={dialogLabel}
                onChange={(e) => setDialogLabel(e.target.value)}
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDialogOpen(false)}
              disabled={isSaving}
            >
              Cancel
            </Button>
            <Button
              onClick={handleSave}
              disabled={!dialogType.trim() || !dialogLabel.trim() || isSaving}
            >
              {isSaving ? (
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
        title="Delete Admin Field Type"
        description={`Are you sure you want to delete the admin field type "${deleteTarget?.label}"? This action cannot be undone.`}
        onConfirm={handleDelete}
        loading={deleteAdminFieldType.isPending}
        variant="destructive"
      />
    </>
  )
}
