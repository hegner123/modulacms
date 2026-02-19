import { useState, useMemo, useRef } from 'react'
import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import type { Datatype, DatatypeID, FieldID, FieldType } from '@modulacms/admin-sdk'
import {
  FormInput,
  Plus,
  Trash2,
  GripVertical,
  Search,
} from 'lucide-react'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { EmptyState } from '@/components/shared/empty-state'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Textarea } from '@/components/ui/textarea'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'
import { useAuthContext } from '@/lib/auth'
import { useFields, useCreateField } from '@/queries/fields'
import {
  useDatatypes,
  useDeleteDatatype,
  useDatatypeFields,
  useDatatypeFieldsByDatatype,
  useCreateDatatypeField,
  useDeleteDatatypeField,
} from '@/queries/datatypes'

export const Route = createFileRoute('/_admin/schema/fields/')({
  component: FieldsPage,
  validateSearch: (search: Record<string, unknown>) => ({
    datatype:
      typeof search.datatype === 'string' ? search.datatype : undefined,
  }),
})

const FIELD_TYPES: FieldType[] = [
  'text',
  'textarea',
  'number',
  'date',
  'datetime',
  'boolean',
  'select',
  'media',
  'relation',
  'json',
  'richtext',
  'slug',
  'email',
  'url',
]

function FieldsPage() {
  const { user } = useAuthContext()
  const navigate = useNavigate()
  const { datatype: selectedDatatypeId } = Route.useSearch()

  const { data: datatypes, isLoading: datatypesLoading } = useDatatypes()
  const { data: allFields } = useFields()
  const { data: allDatatypeFields } = useDatatypeFields()
  const { data: selectedDtFields, isLoading: dtFieldsLoading } =
    useDatatypeFieldsByDatatype(selectedDatatypeId ?? '')

  const createField = useCreateField()
  const createDatatypeField = useCreateDatatypeField()
  const deleteDatatypeField = useDeleteDatatypeField()
  const deleteDatatype = useDeleteDatatype()

  const [filter, setFilter] = useState('')

  // Delete datatype confirmation
  const [deleteDtOpen, setDeleteDtOpen] = useState(false)
  const [deleteDtTarget, setDeleteDtTarget] = useState<Datatype | null>(null)

  // Unified add field dialog
  const [addFieldOpen, setAddFieldOpen] = useState(false)
  const [addFieldTab, setAddFieldTab] = useState<'existing' | 'new'>('new')
  const [selectedFieldId, setSelectedFieldId] = useState('')

  // Create new field form state
  const [createLabel, setCreateLabel] = useState('')
  const [createType, setCreateType] = useState<FieldType>('text')
  const [createData, setCreateData] = useState('{}')
  const [createValidation, setCreateValidation] = useState('{}')
  const [createUiConfig, setCreateUiConfig] = useState('{}')

  // Remove field confirmation
  const [removeOpen, setRemoveOpen] = useState(false)
  const [removeTarget, setRemoveTarget] = useState<{
    id: string
    label: string
  } | null>(null)

  // Bulk field selection
  const [selectedFieldDfIds, setSelectedFieldDfIds] = useState<Set<string>>(new Set())
  const [bulkRemoveOpen, setBulkRemoveOpen] = useState(false)
  const bulkRemoveTargets = useRef<string[]>([])
  const [bulkRemoving, setBulkRemoving] = useState(false)

  function toggleFieldSelection(dfId: string) {
    setSelectedFieldDfIds((prev) => {
      const next = new Set(prev)
      if (next.has(dfId)) {
        next.delete(dfId)
      } else {
        next.add(dfId)
      }
      return next
    })
  }

  function toggleAllFields() {
    if (!selectedDtFields) return
    const allIds = selectedDtFields.map((df) => df.id)
    if (selectedFieldDfIds.size === allIds.length) {
      setSelectedFieldDfIds(new Set())
    } else {
      setSelectedFieldDfIds(new Set(allIds))
    }
  }

  function handleBulkRemoveFields() {
    const targets = bulkRemoveTargets.current
    if (targets.length === 0) return
    setBulkRemoving(true)
    let completed = 0
    for (const id of targets) {
      deleteDatatypeField.mutate(id, {
        onSettled: () => {
          completed++
          if (completed === targets.length) {
            setBulkRemoving(false)
            setBulkRemoveOpen(false)
            bulkRemoveTargets.current = []
            setSelectedFieldDfIds(new Set())
          }
        },
      })
    }
  }

  // Compute field counts per datatype
  const fieldCountMap = useMemo(() => {
    const map = new Map<string, number>()
    if (!allDatatypeFields) return map
    for (const df of allDatatypeFields) {
      map.set(df.datatype_id, (map.get(df.datatype_id) ?? 0) + 1)
    }
    return map
  }, [allDatatypeFields])

  // Filter datatypes by search
  const filteredDatatypes = useMemo(() => {
    if (!datatypes) return []
    if (!filter.trim()) return datatypes
    const lower = filter.toLowerCase()
    return datatypes.filter((dt) => dt.label.toLowerCase().includes(lower))
  }, [datatypes, filter])

  // Selected datatype object
  const selectedDatatype = useMemo(() => {
    if (!selectedDatatypeId || !datatypes) return null
    return datatypes.find((dt) => dt.datatype_id === selectedDatatypeId) ?? null
  }, [datatypes, selectedDatatypeId])

  // Fields assigned to selected datatype — compute set of assigned field IDs
  const assignedFieldIds = useMemo(() => {
    if (!selectedDtFields) return new Set<string>()
    return new Set(selectedDtFields.map((df) => df.field_id))
  }, [selectedDtFields])

  // Available fields for "Add Field" dropdown (not already assigned)
  const availableFields = useMemo(() => {
    if (!allFields) return []
    return allFields.filter((f) => !assignedFieldIds.has(f.field_id))
  }, [allFields, assignedFieldIds])

  function getFieldLabel(fieldId: string): string {
    const field = allFields?.find((f) => f.field_id === fieldId)
    return field?.label ?? fieldId
  }

  function getFieldType(fieldId: string): string {
    const field = allFields?.find((f) => f.field_id === fieldId)
    return field?.type ?? 'unknown'
  }

  function selectDatatype(id: string) {
    setSelectedFieldDfIds(new Set())
    navigate({
      to: '/schema/fields',
      search: { datatype: id },
    })
  }

  function handleAddField() {
    if (!selectedFieldId || !selectedDatatypeId) return
    const nextSortOrder = selectedDtFields ? selectedDtFields.length : 0
    createDatatypeField.mutate(
      {
        datatype_id: selectedDatatypeId as DatatypeID,
        field_id: selectedFieldId as FieldID,
        sort_order: nextSortOrder,
      },
      {
        onSuccess: () => {
          setAddFieldOpen(false)
          setSelectedFieldId('')
        },
      },
    )
  }

  function resetCreateForm() {
    setCreateLabel('')
    setCreateType('text')
    setCreateData('{}')
    setCreateValidation('{}')
    setCreateUiConfig('{}')
  }

  function handleCreateField() {
    if (!selectedDatatypeId) return
    const now = new Date().toISOString()
    createField.mutate(
      {
        label: createLabel,
        data: createData,
        validation: createValidation,
        ui_config: createUiConfig,
        type: createType,
        parent_id: null,
        author_id: user?.user_id ?? null,
        date_created: now,
        date_modified: now,
      } as Parameters<typeof createField.mutate>[0],
      {
        onSuccess: (newField) => {
          // Auto-assign the new field to the selected datatype
          const nextSortOrder = selectedDtFields ? selectedDtFields.length : 0
          createDatatypeField.mutate(
            {
              datatype_id: selectedDatatypeId as DatatypeID,
              field_id: newField.field_id,
              sort_order: nextSortOrder,
            },
            {
              onSuccess: () => {
                setAddFieldOpen(false)
                resetCreateForm()
              },
            },
          )
        },
      },
    )
  }

  function handleRemoveField() {
    if (!removeTarget) return
    deleteDatatypeField.mutate(removeTarget.id, {
      onSuccess: () => {
        setRemoveOpen(false)
        setRemoveTarget(null)
      },
    })
  }

  function handleDeleteDatatype() {
    if (!deleteDtTarget) return
    deleteDatatype.mutate(deleteDtTarget.datatype_id, {
      onSuccess: () => {
        setDeleteDtOpen(false)
        setDeleteDtTarget(null)
        if (selectedDatatypeId === deleteDtTarget.datatype_id) {
          navigate({ to: '/schema/fields', search: { datatype: undefined } })
        }
      },
    })
  }

  if (datatypesLoading) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Fields</h1>
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">Fields</h1>
        <p className="text-muted-foreground">
          Manage field assignments per datatype.
        </p>
      </div>

      <div className="flex rounded-lg border" style={{ minHeight: '500px' }}>
        {/* Left panel — datatype sidebar */}
        <div className="w-[280px] shrink-0 border-r">
          <div className="border-b p-3">
            <h2 className="mb-2 text-sm font-semibold text-muted-foreground">
              Datatypes
            </h2>
            <div className="relative">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Filter..."
                value={filter}
                onChange={(e) => setFilter(e.target.value)}
                className="pl-8"
              />
            </div>
          </div>
          <ScrollArea className="h-[440px]">
            <div className="p-2">
              {filteredDatatypes.length > 0 ? (
                filteredDatatypes.map((dt) => (
                  <div
                    key={dt.datatype_id}
                    className={cn(
                      'group flex w-full items-center justify-between rounded-md px-3 py-2 text-sm transition-colors hover:bg-accent hover:text-accent-foreground',
                      selectedDatatypeId === dt.datatype_id && 'bg-primary text-primary-foreground',
                    )}
                  >
                    <button
                      onClick={() => selectDatatype(dt.datatype_id)}
                      className="flex min-w-0 flex-1 items-center gap-2 text-left"
                    >
                      <span className="truncate font-medium">{dt.label}</span>
                      <Badge variant="secondary" className="shrink-0">
                        {fieldCountMap.get(dt.datatype_id) ?? 0}
                      </Badge>
                    </button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6 shrink-0 opacity-0 group-hover:opacity-100"
                      onClick={(e) => {
                        e.stopPropagation()
                        setDeleteDtTarget(dt)
                        setDeleteDtOpen(true)
                      }}
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </div>
                ))
              ) : (
                <p className="px-3 py-4 text-center text-sm text-muted-foreground">
                  No datatypes found.
                </p>
              )}
            </div>
          </ScrollArea>
        </div>

        {/* Right panel — fields for selected datatype */}
        <div className="flex flex-1 flex-col">
          {selectedDatatype ? (
            <>
              <div className="flex items-center justify-between border-b p-4">
                <div className="flex items-center gap-3">
                  <h2 className="text-lg font-semibold">
                    Fields for "{selectedDatatype.label}"
                  </h2>
                  {selectedFieldDfIds.size > 0 && (
                    <div className="flex items-center gap-2">
                      <span className="text-sm text-muted-foreground">
                        {selectedFieldDfIds.size} selected
                      </span>
                      <Button
                        variant="destructive"
                        size="sm"
                        onClick={() => {
                          bulkRemoveTargets.current = Array.from(selectedFieldDfIds)
                          setBulkRemoveOpen(true)
                        }}
                      >
                        <Trash2 className="mr-2 h-4 w-4" />
                        Remove ({selectedFieldDfIds.size})
                      </Button>
                    </div>
                  )}
                </div>
                <Button size="sm" onClick={() => setAddFieldOpen(true)}>
                  <Plus className="mr-2 h-4 w-4" />
                  Add Field
                </Button>
              </div>
              <div className="flex-1 p-4">
                {dtFieldsLoading ? (
                  <p className="text-sm text-muted-foreground">
                    Loading fields...
                  </p>
                ) : selectedDtFields && selectedDtFields.length > 0 ? (
                  <div className="space-y-2">
                    <div className="flex items-center gap-3 px-4 py-1">
                      <Checkbox
                        checked={
                          selectedFieldDfIds.size === selectedDtFields.length ||
                          (selectedFieldDfIds.size > 0 && 'indeterminate')
                        }
                        onCheckedChange={toggleAllFields}
                        aria-label="Select all fields"
                      />
                      <span className="text-xs text-muted-foreground">Select all</span>
                    </div>
                    {selectedDtFields
                      .sort((a, b) => a.sort_order - b.sort_order)
                      .map((df) => (
                        <div
                          key={df.id}
                          className="group flex items-center justify-between rounded-md border px-4 py-3 transition-colors hover:bg-accent hover:text-accent-foreground"
                        >
                          <div className="flex min-w-0 flex-1 items-center gap-3">
                            <Checkbox
                              checked={selectedFieldDfIds.has(df.id)}
                              onCheckedChange={() => toggleFieldSelection(df.id)}
                              onClick={(e) => e.stopPropagation()}
                              aria-label={`Select ${getFieldLabel(df.field_id)}`}
                            />
                            <Link
                              to="/schema/fields/$id"
                              params={{ id: df.field_id }}
                              className="flex min-w-0 flex-1 items-center gap-3"
                            >
                              <GripVertical className="h-4 w-4 text-muted-foreground" />
                              <span className="text-sm font-medium">
                                {getFieldLabel(df.field_id)}
                              </span>
                              <Badge variant="secondary">
                                {getFieldType(df.field_id)}
                              </Badge>
                            </Link>
                          </div>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 shrink-0 text-muted-foreground opacity-0 hover:text-destructive group-hover:opacity-100"
                            onClick={() => {
                              setRemoveTarget({
                                id: df.id,
                                label: getFieldLabel(df.field_id),
                              })
                              setRemoveOpen(true)
                            }}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      ))}
                  </div>
                ) : (
                  <EmptyState
                    icon={FormInput}
                    title="No fields assigned"
                    description="Add an existing field or create a new one to assign it to this datatype."
                  />
                )}
              </div>
            </>
          ) : (
            <div className="flex flex-1 items-center justify-center">
              <EmptyState
                icon={FormInput}
                title="Select a datatype"
                description="Choose a datatype from the sidebar to manage its fields."
              />
            </div>
          )}
        </div>
      </div>

      {/* Add field dialog — existing or new */}
      <Dialog open={addFieldOpen} onOpenChange={setAddFieldOpen}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>Add Field</DialogTitle>
            <DialogDescription>
              Assign an existing field or create a new one for "{selectedDatatype?.label}".
            </DialogDescription>
          </DialogHeader>
          <Tabs value={addFieldTab} onValueChange={(v) => setAddFieldTab(v as 'existing' | 'new')}>
            <TabsList className="w-full">
              <TabsTrigger value="new" className="flex-1">Create New</TabsTrigger>
              <TabsTrigger value="existing" className="flex-1">Existing</TabsTrigger>
            </TabsList>
            <TabsContent value="existing" className="space-y-4 pt-2">
              <div className="space-y-2">
                <Label>Field</Label>
                <Select value={selectedFieldId} onValueChange={setSelectedFieldId}>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select a field..." />
                  </SelectTrigger>
                  <SelectContent>
                    {availableFields.map((field) => (
                      <SelectItem key={field.field_id} value={field.field_id}>
                        {field.label} ({field.type})
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                {availableFields.length === 0 && (
                  <p className="text-sm text-muted-foreground">
                    All fields are already assigned. Switch to "Create New" to make one.
                  </p>
                )}
              </div>
              <DialogFooter>
                <Button
                  variant="outline"
                  onClick={() => setAddFieldOpen(false)}
                  disabled={createDatatypeField.isPending}
                >
                  Cancel
                </Button>
                <Button
                  onClick={handleAddField}
                  disabled={!selectedFieldId || createDatatypeField.isPending}
                >
                  {createDatatypeField.isPending ? 'Adding...' : 'Add Field'}
                </Button>
              </DialogFooter>
            </TabsContent>
            <TabsContent value="new" className="space-y-4 pt-2">
              <div className="space-y-2">
                <Label htmlFor="create-field-label">Label</Label>
                <Input
                  id="create-field-label"
                  placeholder="e.g. Title, Body, Featured Image"
                  value={createLabel}
                  onChange={(e) => setCreateLabel(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="create-field-type">Type</Label>
                <Select
                  value={createType}
                  onValueChange={(value) => setCreateType(value as FieldType)}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {FIELD_TYPES.map((ft) => (
                      <SelectItem key={ft} value={ft}>
                        {ft}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="create-field-data">Data (JSON)</Label>
                <Textarea
                  id="create-field-data"
                  value={createData}
                  onChange={(e) => setCreateData(e.target.value)}
                  rows={3}
                  className="font-mono text-sm"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="create-field-validation">
                  Validation (JSON)
                </Label>
                <Textarea
                  id="create-field-validation"
                  value={createValidation}
                  onChange={(e) => setCreateValidation(e.target.value)}
                  rows={3}
                  className="font-mono text-sm"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="create-field-ui-config">UI Config (JSON)</Label>
                <Textarea
                  id="create-field-ui-config"
                  value={createUiConfig}
                  onChange={(e) => setCreateUiConfig(e.target.value)}
                  rows={3}
                  className="font-mono text-sm"
                />
              </div>
              <DialogFooter>
                <Button
                  variant="outline"
                  onClick={() => setAddFieldOpen(false)}
                  disabled={createField.isPending}
                >
                  Cancel
                </Button>
                <Button
                  onClick={handleCreateField}
                  disabled={!createLabel.trim() || createField.isPending}
                >
                  {createField.isPending ? 'Creating...' : 'Create & Assign'}
                </Button>
              </DialogFooter>
            </TabsContent>
          </Tabs>
        </DialogContent>
      </Dialog>

      {/* Remove field confirmation */}
      <ConfirmDialog
        open={removeOpen}
        onOpenChange={setRemoveOpen}
        title="Remove Field"
        description={`Are you sure you want to remove "${removeTarget?.label}" from this datatype? The field itself will not be deleted.`}
        onConfirm={handleRemoveField}
        loading={deleteDatatypeField.isPending}
        variant="destructive"
      />

      {/* Delete datatype confirmation */}
      <ConfirmDialog
        open={deleteDtOpen}
        onOpenChange={setDeleteDtOpen}
        title="Delete Datatype"
        description={`Are you sure you want to delete "${deleteDtTarget?.label}"? This action cannot be undone.`}
        onConfirm={handleDeleteDatatype}
        loading={deleteDatatype.isPending}
        variant="destructive"
      />

      {/* Bulk remove fields confirmation */}
      <ConfirmDialog
        open={bulkRemoveOpen}
        onOpenChange={setBulkRemoveOpen}
        title="Remove Fields"
        description={`Are you sure you want to remove ${bulkRemoveTargets.current.length} field(s) from this datatype? The fields themselves will not be deleted.`}
        onConfirm={handleBulkRemoveFields}
        loading={bulkRemoving}
        variant="destructive"
      />
    </div>
  )
}
