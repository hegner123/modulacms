import type { ContentNode } from '@modulacms/admin-sdk'
import { Input } from '@/components/ui/input'
import { FieldRenderer } from '@/components/fields/field-renderer'
import type { MergedField } from './build-merged-fields'

type PageFieldsProps = {
  node: ContentNode
  mergedFields: MergedField[]
  getFieldValue: (contentDataId: string, fieldId: string, mergedFields: MergedField[]) => string
  setFieldValue: (contentDataId: string, fieldId: string, value: string) => void
}

const TITLE_LIKE_TYPES = new Set(['text', 'slug'])

export function PageFields({
  node,
  mergedFields,
  getFieldValue,
  setFieldValue,
}: PageFieldsProps) {
  const contentDataId = node.datatype.content.content_data_id

  if (mergedFields.length === 0) return null

  return (
    <div className="space-y-4">
      {mergedFields.map((field, index) => {
        const value = getFieldValue(contentDataId, field.fieldId, mergedFields)

        if (index === 0 && TITLE_LIKE_TYPES.has(field.type)) {
          return (
            <Input
              key={field.fieldId}
              value={value}
              onChange={(e) => setFieldValue(contentDataId, field.fieldId, e.target.value)}
              placeholder={field.label}
              className="border-none bg-transparent text-3xl font-bold shadow-none placeholder:text-muted-foreground/40 focus-visible:ring-0"
            />
          )
        }

        return (
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
            value={value}
            onChange={(v) => setFieldValue(contentDataId, field.fieldId, v)}
          />
        )
      })}
    </div>
  )
}
