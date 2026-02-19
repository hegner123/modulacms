import { FieldRenderer } from '@/components/fields/field-renderer'
import type { MergedField } from './build-merged-fields'

type BlockFieldEditorProps = {
  contentDataId: string
  mergedFields: MergedField[]
  getFieldValue: (contentDataId: string, fieldId: string, mergedFields: MergedField[]) => string
  setFieldValue: (contentDataId: string, fieldId: string, value: string) => void
}

export function BlockFieldEditor({
  contentDataId,
  mergedFields,
  getFieldValue,
  setFieldValue,
}: BlockFieldEditorProps) {
  if (mergedFields.length === 0) {
    return (
      <p className="px-4 py-3 text-sm text-muted-foreground">
        No fields defined for this datatype.
      </p>
    )
  }

  return (
    <div className="space-y-4 px-4 pb-4">
      {mergedFields.map((field) => (
        <FieldRenderer
          key={field.fieldId}
          field={{
            field_id: field.fieldId,
            label: field.label,
            type: field.type,
            validation: field.validation,
            ui_config: field.ui_config,
            data: field.data,
          }}
          value={getFieldValue(contentDataId, field.fieldId, mergedFields)}
          onChange={(v) => setFieldValue(contentDataId, field.fieldId, v)}
        />
      ))}
    </div>
  )
}
