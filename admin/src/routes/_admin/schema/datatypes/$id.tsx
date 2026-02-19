import { useState } from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { useForm } from 'react-hook-form'
import { ArrowLeft, Plus, Trash2, GripVertical } from 'lucide-react'
import {
  useDatatype,
  useUpdateDatatype,
  useDatatypeFieldsByDatatype,
  useDeleteDatatypeField,
} from '@/queries/datatypes'
import { useFields } from '@/queries/fields'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'

export const Route = createFileRoute('/_admin/schema/datatypes/$id')({
  component: DatatypeDetailPage,
})

type EditFormValues = {
  label: string
  type: string
}

function DatatypeDetailPage() {
  const { id } = Route.useParams()
  const { data: datatype, isLoading } = useDatatype(id)
  const updateDatatype = useUpdateDatatype()
  const { data: datatypeFields, isLoading: fieldsLoading } = useDatatypeFieldsByDatatype(id)
  const { data: allFields } = useFields()
  const deleteDatatypeField = useDeleteDatatypeField()

  const [removeTarget, setRemoveTarget] = useState<{ id: string; label: string } | null>(null)
  const [removeOpen, setRemoveOpen] = useState(false)

  const {
    register,
    handleSubmit,
    formState: { isDirty },
    reset,
  } = useForm<EditFormValues>({
    values: datatype
      ? { label: datatype.label, type: datatype.type }
      : { label: '', type: '' },
  })

  function getFieldLabel(fieldId: string): string {
    const field = allFields?.find((f) => f.field_id === fieldId)
    return field?.label ?? fieldId
  }

  function getFieldType(fieldId: string): string {
    const field = allFields?.find((f) => f.field_id === fieldId)
    return field?.type ?? 'unknown'
  }

  function onSave(values: EditFormValues) {
    if (!datatype) return
    updateDatatype.mutate(
      {
        datatype_id: datatype.datatype_id,
        parent_id: datatype.parent_id,
        label: values.label,
        type: values.type,
        author_id: datatype.author_id,
        date_created: datatype.date_created,
        date_modified: new Date().toISOString(),
      },
      {
        onSuccess: (updated) => {
          reset({ label: updated.label, type: updated.type })
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

  if (isLoading) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Datatype Detail</h1>
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  if (!datatype) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Datatype Not Found</h1>
        <p className="text-muted-foreground">
          The requested datatype could not be found.
        </p>
        <Button variant="outline" asChild>
          <Link to="/schema/datatypes">Back to Datatypes</Link>
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" asChild>
          <Link to="/schema/datatypes">
            <ArrowLeft className="h-4 w-4" />
          </Link>
        </Button>
        <div>
          <h1 className="text-2xl font-bold">{datatype.label}</h1>
          <p className="text-sm text-muted-foreground">
            Edit datatype properties and manage assigned fields.
          </p>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Properties</CardTitle>
          <CardDescription>Update the datatype label and type.</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(onSave)} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="edit-label">Label</Label>
              <Input
                id="edit-label"
                {...register('label', { required: true })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-type">Type</Label>
              <Input
                id="edit-type"
                {...register('type', { required: true })}
              />
            </div>
            <Button
              type="submit"
              disabled={!isDirty || updateDatatype.isPending}
            >
              {updateDatatype.isPending ? 'Saving...' : 'Save Changes'}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Assigned Fields</CardTitle>
              <CardDescription>
                Fields attached to this datatype, in display order.
              </CardDescription>
            </div>
            <Button size="sm" asChild>
              <Link to="/schema/fields" search={{ datatype: id }}>
                <Plus className="mr-2 h-4 w-4" />
                Add Field
              </Link>
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {fieldsLoading ? (
            <p className="text-sm text-muted-foreground">Loading fields...</p>
          ) : datatypeFields && datatypeFields.length > 0 ? (
            <div className="space-y-2">
              {datatypeFields
                .sort((a, b) => a.sort_order - b.sort_order)
                .map((df) => (
                  <div
                    key={df.id}
                    className="group flex items-center justify-between rounded-md border px-4 py-3 transition-colors hover:bg-accent hover:text-accent-foreground"
                  >
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
            <p className="py-4 text-center text-sm text-muted-foreground">
              No fields assigned to this datatype yet.
            </p>
          )}
        </CardContent>
      </Card>

      <ConfirmDialog
        open={removeOpen}
        onOpenChange={setRemoveOpen}
        title="Remove Field"
        description={`Are you sure you want to remove "${removeTarget?.label}" from this datatype?`}
        onConfirm={handleRemoveField}
        loading={deleteDatatypeField.isPending}
        variant="destructive"
      />
    </div>
  )
}
