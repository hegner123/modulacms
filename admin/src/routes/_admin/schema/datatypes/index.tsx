import { useState, useRef } from 'react'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import type { ColumnDef } from '@tanstack/react-table'
import type { Datatype, DatatypeID, HealReport } from '@modulacms/admin-sdk'
import { Blocks, Plus, Trash2, HeartPulse } from 'lucide-react'
import { DataTable } from '@/components/data-table/data-table'
import { DataTableColumnHeader } from '@/components/data-table/column-header'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { EmptyState } from '@/components/shared/empty-state'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useDatatypes, useCreateDatatype, useDeleteDatatype, useContentHeal } from '@/queries/datatypes'
import { useAuthContext } from '@/lib/auth'
import { formatDate } from '@/lib/utils'

export const Route = createFileRoute('/_admin/schema/datatypes/')({
  component: DatatypesPage,
})

function DatatypesPage() {
  const { user } = useAuthContext()
  const navigate = useNavigate()
  const { data: datatypes, isLoading } = useDatatypes()
  const createDatatype = useCreateDatatype()
  const deleteDatatype = useDeleteDatatype()

  const [createOpen, setCreateOpen] = useState(false)
  const [createLabel, setCreateLabel] = useState('')
  const [createType, setCreateType] = useState('')
  const [createParentId, setCreateParentId] = useState<string>('')

  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<Datatype | null>(null)

  // Bulk delete state
  const [bulkDeleteOpen, setBulkDeleteOpen] = useState(false)
  const bulkDeleteTargets = useRef<Datatype[]>([])
  const [bulkDeleting, setBulkDeleting] = useState(false)

  // Content heal state
  const contentHeal = useContentHeal()
  const [healOpen, setHealOpen] = useState(false)
  const [healReport, setHealReport] = useState<HealReport | null>(null)

  function handleBulkDelete() {
    const targets = bulkDeleteTargets.current
    if (targets.length === 0) return
    setBulkDeleting(true)
    let completed = 0
    for (const dt of targets) {
      deleteDatatype.mutate(dt.datatype_id as DatatypeID, {
        onSettled: () => {
          completed++
          if (completed === targets.length) {
            setBulkDeleting(false)
            setBulkDeleteOpen(false)
            bulkDeleteTargets.current = []
          }
        },
      })
    }
  }

  function handleCreate() {
    const now = new Date().toISOString()
    createDatatype.mutate(
      {
        label: createLabel,
        type: createType,
        parent_id: (createParentId && createParentId !== 'none') ? createParentId : null,
        author_id: user?.user_id ?? null,
        date_created: now,
        date_modified: now,
      } as Parameters<typeof createDatatype.mutate>[0],
      {
        onSuccess: () => {
          setCreateOpen(false)
          setCreateLabel('')
          setCreateType('')
          setCreateParentId('')
        },
      },
    )
  }

  function handleDelete() {
    if (!deleteTarget) return
    deleteDatatype.mutate(deleteTarget.datatype_id, {
      onSuccess: () => {
        setDeleteOpen(false)
        setDeleteTarget(null)
      },
    })
  }

  function handleHeal(dryRun: boolean) {
    contentHeal.mutate(dryRun, {
      onSuccess: (report) => {
        setHealReport(report)
      },
    })
  }

  const totalRepairs = healReport
    ? healReport.content_data_repairs.length + healReport.content_field_repairs.length + healReport.missing_fields.length
    : 0

  const columns: ColumnDef<Datatype>[] = [
    {
      accessorKey: 'label',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Label" />
      ),
      cell: ({ row }) => (
        <span className="font-medium">{row.original.label}</span>
      ),
    },
    {
      accessorKey: 'type',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Type" />
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
        const datatype = row.original
        return (
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-muted-foreground hover:text-destructive"
            onClick={(e) => {
              e.stopPropagation()
              setDeleteTarget(datatype)
              setDeleteOpen(true)
            }}
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        )
      },
    },
  ]

  if (isLoading) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Datatypes</h1>
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Datatypes</h1>
          <p className="text-muted-foreground">
            Define and manage the content schemas used across your site.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            onClick={() => {
              setHealReport(null)
              setHealOpen(true)
            }}
          >
            <HeartPulse className="mr-2 h-4 w-4" />
            Heal Content
          </Button>
          <Button onClick={() => setCreateOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Create Datatype
          </Button>
        </div>
      </div>

      {datatypes && datatypes.length > 0 ? (
        <DataTable
          columns={columns}
          data={datatypes}
          searchKey="label"
          searchPlaceholder="Search datatypes..."
          onRowClick={(dt) => navigate({ to: '/schema/datatypes/$id', params: { id: dt.datatype_id } })}
          enableRowSelection
          getRowId={(row) => row.datatype_id}
          selectionActions={(rows) => (
            <Button
              variant="destructive"
              size="sm"
              onClick={() => {
                bulkDeleteTargets.current = rows
                setBulkDeleteOpen(true)
              }}
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Delete ({rows.length})
            </Button>
          )}
        />
      ) : (
        <EmptyState
          icon={Blocks}
          title="No datatypes yet"
          description="Create your first datatype to define a content schema."
          action={
            <Button onClick={() => setCreateOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Create Datatype
            </Button>
          }
        />
      )}

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Datatype</DialogTitle>
            <DialogDescription>
              Define a new content type for your schema.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="create-label">Label</Label>
              <Input
                id="create-label"

                value={createLabel}
                onChange={(e) => setCreateLabel(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="create-type">Type</Label>
              <Input
                id="create-type"

                value={createType}
                onChange={(e) => setCreateType(e.target.value)}
              />
              <p className="text-xs text-muted-foreground">
                Use <code className="rounded bg-muted px-1 py-0.5">ROOT</code> to register this datatype as a root-level content type.
              </p>
            </div>
            <div className="space-y-2">
              <Label>Parent</Label>
              <Select value={createParentId} onValueChange={setCreateParentId}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="None" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">None</SelectItem>
                  {datatypes?.map((dt) => (
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
              disabled={createDatatype.isPending}
            >
              Cancel
            </Button>
            <Button
              onClick={handleCreate}
              disabled={!createLabel.trim() || !createType.trim() || createDatatype.isPending}
            >
              {createDatatype.isPending ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title="Delete Datatype"
        description={`Are you sure you want to delete "${deleteTarget?.label}"? This action cannot be undone.`}
        onConfirm={handleDelete}
        loading={deleteDatatype.isPending}
        variant="destructive"
      />

      <ConfirmDialog
        open={bulkDeleteOpen}
        onOpenChange={setBulkDeleteOpen}
        title="Delete Datatypes"
        description={`Are you sure you want to delete ${bulkDeleteTargets.current.length} datatype(s)? This action cannot be undone.`}
        onConfirm={handleBulkDelete}
        loading={bulkDeleting}
        variant="destructive"
      />

      <Dialog open={healOpen} onOpenChange={setHealOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Content Heal</DialogTitle>
            <DialogDescription>
              Scan for malformed IDs and create missing content fields for incomplete field sets.
            </DialogDescription>
          </DialogHeader>

          {healReport ? (
            <div className="space-y-4">
              <div className="flex items-center gap-2 text-sm">
                {healReport.dry_run ? (
                  <span className="rounded bg-yellow-500/10 px-2 py-0.5 text-yellow-500">Dry Run</span>
                ) : (
                  <span className="rounded bg-green-500/10 px-2 py-0.5 text-green-500">Applied</span>
                )}
                <span className="text-muted-foreground">
                  Scanned {healReport.content_data_scanned} data rows, {healReport.content_field_scanned} field rows
                </span>
              </div>

              {totalRepairs === 0 ? (
                <p className="text-sm text-muted-foreground">No malformed IDs found. Everything looks healthy.</p>
              ) : (
                <div className="space-y-3">
                  {healReport.content_data_repairs.length > 0 && (
                    <div>
                      <h4 className="mb-1 text-sm font-medium">Content Data Repairs ({healReport.content_data_repairs.length})</h4>
                      <div className="max-h-48 overflow-auto rounded border">
                        <table className="w-full text-xs">
                          <thead className="sticky top-0 bg-muted">
                            <tr>
                              <th className="px-2 py-1 text-left">Row ID</th>
                              <th className="px-2 py-1 text-left">Column</th>
                              <th className="px-2 py-1 text-left">Old</th>
                              <th className="px-2 py-1 text-left">New</th>
                            </tr>
                          </thead>
                          <tbody>
                            {healReport.content_data_repairs.map((r, i) => (
                              <tr key={i} className="border-t">
                                <td className="px-2 py-1 font-mono">{r.row_id}</td>
                                <td className="px-2 py-1">{r.column}</td>
                                <td className="px-2 py-1 font-mono text-destructive">{r.old_value}</td>
                                <td className="px-2 py-1 font-mono text-green-500">{r.new_value}</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}
                  {healReport.content_field_repairs.length > 0 && (
                    <div>
                      <h4 className="mb-1 text-sm font-medium">Content Field Repairs ({healReport.content_field_repairs.length})</h4>
                      <div className="max-h-48 overflow-auto rounded border">
                        <table className="w-full text-xs">
                          <thead className="sticky top-0 bg-muted">
                            <tr>
                              <th className="px-2 py-1 text-left">Row ID</th>
                              <th className="px-2 py-1 text-left">Column</th>
                              <th className="px-2 py-1 text-left">Old</th>
                              <th className="px-2 py-1 text-left">New</th>
                            </tr>
                          </thead>
                          <tbody>
                            {healReport.content_field_repairs.map((r, i) => (
                              <tr key={i} className="border-t">
                                <td className="px-2 py-1 font-mono">{r.row_id}</td>
                                <td className="px-2 py-1">{r.column}</td>
                                <td className="px-2 py-1 font-mono text-destructive">{r.old_value}</td>
                                <td className="px-2 py-1 font-mono text-green-500">{r.new_value}</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}
                  {healReport.missing_fields.length > 0 && (
                    <div>
                      <h4 className="mb-1 text-sm font-medium">Missing Fields ({healReport.missing_fields.length})</h4>
                      <div className="max-h-48 overflow-auto rounded border">
                        <table className="w-full text-xs">
                          <thead className="sticky top-0 bg-muted">
                            <tr>
                              <th className="px-2 py-1 text-left">Content Data ID</th>
                              <th className="px-2 py-1 text-left">Field ID</th>
                              <th className="px-2 py-1 text-left">Status</th>
                            </tr>
                          </thead>
                          <tbody>
                            {healReport.missing_fields.map((m, i) => (
                              <tr key={i} className="border-t">
                                <td className="px-2 py-1 font-mono">{m.content_data_id}</td>
                                <td className="px-2 py-1 font-mono">{m.field_id}</td>
                                <td className="px-2 py-1">
                                  {m.created ? (
                                    <span className="text-green-500">Created</span>
                                  ) : healReport.dry_run ? (
                                    <span className="text-yellow-500">Pending</span>
                                  ) : (
                                    <span className="text-destructive">Failed</span>
                                  )}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}
                </div>
              )}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">
              Run a dry run first to preview repairs, then apply to fix malformed IDs.
            </p>
          )}

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setHealOpen(false)}
              disabled={contentHeal.isPending}
            >
              Close
            </Button>
            <Button
              variant="outline"
              onClick={() => handleHeal(true)}
              disabled={contentHeal.isPending}
            >
              {contentHeal.isPending ? 'Scanning...' : 'Dry Run'}
            </Button>
            <Button
              onClick={() => handleHeal(false)}
              disabled={contentHeal.isPending || (healReport !== null && totalRepairs === 0)}
            >
              {contentHeal.isPending ? 'Healing...' : 'Apply Heal'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
