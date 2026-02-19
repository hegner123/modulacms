import type { ContentNode, Field, DatatypeField } from '@modulacms/admin-sdk'
import type { FieldType } from '@/components/fields/field-renderer'

export type MergedField = {
  fieldId: string
  label: string
  type: FieldType
  value: string
  validation: string
  ui_config: string
  data: string
  contentField: ContentNode['fields'][number]['content'] | null
}

export function buildMergedFields(
  node: ContentNode,
  datatypeFields: DatatypeField[] | undefined,
  allFields: Field[] | undefined,
): MergedField[] {
  const fieldDefMap = new Map<string, Field>()
  if (allFields) {
    for (const f of allFields) {
      fieldDefMap.set(f.field_id, f)
    }
  }

  const contentFieldMap = new Map<string, ContentNode['fields'][number]>()
  for (const nf of node.fields) {
    contentFieldMap.set(nf.info.field_id, nf)
  }

  if (!datatypeFields) {
    return node.fields.map((nf) => {
      const def = fieldDefMap.get(nf.info.field_id)
      return {
        fieldId: nf.info.field_id,
        label: nf.info.label,
        type: nf.info.type as FieldType,
        value: nf.content.field_value ?? '',
        validation: def?.validation ?? nf.info.validation ?? '',
        ui_config: def?.ui_config ?? nf.info.ui_config ?? '',
        data: def?.data ?? nf.info.data ?? '',
        contentField: nf.content,
      }
    })
  }

  const result: MergedField[] = []
  const seen = new Set<string>()

  const sorted = [...datatypeFields].sort((a, b) => a.sort_order - b.sort_order)

  for (const df of sorted) {
    seen.add(df.field_id)
    const existing = contentFieldMap.get(df.field_id)
    const fieldDef = fieldDefMap.get(df.field_id)

    if (existing) {
      result.push({
        fieldId: df.field_id,
        label: existing.info.label,
        type: existing.info.type as FieldType,
        value: existing.content.field_value ?? '',
        validation: fieldDef?.validation ?? existing.info.validation ?? '',
        ui_config: fieldDef?.ui_config ?? existing.info.ui_config ?? '',
        data: fieldDef?.data ?? existing.info.data ?? '',
        contentField: existing.content,
      })
    } else if (fieldDef) {
      result.push({
        fieldId: df.field_id,
        label: fieldDef.label,
        type: fieldDef.type as FieldType,
        value: '',
        validation: fieldDef.validation ?? '',
        ui_config: fieldDef.ui_config ?? '',
        data: fieldDef.data ?? '',
        contentField: null,
      })
    }
  }

  for (const nf of node.fields) {
    if (seen.has(nf.info.field_id)) continue
    const def = fieldDefMap.get(nf.info.field_id)
    result.push({
      fieldId: nf.info.field_id,
      label: nf.info.label,
      type: nf.info.type as FieldType,
      value: nf.content.field_value ?? '',
      validation: def?.validation ?? nf.info.validation ?? '',
      ui_config: def?.ui_config ?? nf.info.ui_config ?? '',
      data: def?.data ?? nf.info.data ?? '',
      contentField: nf.content,
    })
  }

  return result
}
