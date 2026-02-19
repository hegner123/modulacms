import { createFileRoute, Link } from '@tanstack/react-router'
import { useForm, Controller } from 'react-hook-form'
import type { FieldType } from '@modulacms/admin-sdk'
import { ArrowLeft } from 'lucide-react'
import { useField, useUpdateField } from '@/queries/fields'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

export const Route = createFileRoute('/_admin/schema/fields/$id')({
  component: FieldDetailPage,
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

type EditFormValues = {
  label: string
  type: FieldType
  data: string
  validation: string
  ui_config: string
}

function FieldDetailPage() {
  const { id } = Route.useParams()
  const { data: field, isLoading } = useField(id)
  const updateField = useUpdateField()

  const {
    register,
    handleSubmit,
    control,
    formState: { isDirty },
    reset,
  } = useForm<EditFormValues>({
    values: field
      ? {
          label: field.label,
          type: field.type,
          data: field.data,
          validation: field.validation,
          ui_config: field.ui_config,
        }
      : {
          label: '',
          type: 'text' as FieldType,
          data: '{}',
          validation: '{}',
          ui_config: '{}',
        },
  })

  function onSave(values: EditFormValues) {
    if (!field) return
    updateField.mutate(
      {
        field_id: field.field_id,
        parent_id: field.parent_id,
        label: values.label,
        data: values.data,
        validation: values.validation,
        ui_config: values.ui_config,
        type: values.type,
        author_id: field.author_id,
        date_created: field.date_created,
        date_modified: new Date().toISOString(),
      },
      {
        onSuccess: (updated) => {
          reset({
            label: updated.label,
            type: updated.type,
            data: updated.data,
            validation: updated.validation,
            ui_config: updated.ui_config,
          })
        },
      },
    )
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Field Detail</h1>
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  if (!field) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Field Not Found</h1>
        <p className="text-muted-foreground">
          The requested field could not be found.
        </p>
        <Button variant="outline" asChild>
          <Link to="/schema/fields" search={{ datatype: undefined }}>Back to Fields</Link>
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" asChild>
          <Link to="/schema/fields" search={{ datatype: undefined }}>
            <ArrowLeft className="h-4 w-4" />
          </Link>
        </Button>
        <div>
          <h1 className="text-2xl font-bold">{field.label}</h1>
          <p className="text-sm text-muted-foreground">
            Edit field properties, validation, and UI configuration.
          </p>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Field Properties</CardTitle>
          <CardDescription>
            Update the field label, type, and configuration.
          </CardDescription>
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
              <Controller
                name="type"
                control={control}
                rules={{ required: true }}
                render={({ field: controllerField }) => (
                  <Select
                    value={controllerField.value}
                    onValueChange={controllerField.onChange}
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
                )}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-data">Data (JSON)</Label>
              <Textarea
                id="edit-data"
                {...register('data')}
                rows={4}
                className="font-mono text-sm"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-validation">Validation (JSON)</Label>
              <Textarea
                id="edit-validation"
                {...register('validation')}
                rows={4}
                className="font-mono text-sm"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-ui-config">UI Config (JSON)</Label>
              <Textarea
                id="edit-ui-config"
                {...register('ui_config')}
                rows={4}
                className="font-mono text-sm"
              />
            </div>

            <Button
              type="submit"
              disabled={!isDirty || updateField.isPending}
            >
              {updateField.isPending ? 'Saving...' : 'Save Changes'}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
