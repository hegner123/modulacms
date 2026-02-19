import { useState, useMemo, useRef } from 'react'
import type {
  ContentNode,
  NodeField,
  Field,
  ContentFieldID,
  ContentID,
  FieldID,
  RouteID,
  UserID,
} from '@modulacms/admin-sdk'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { useAuthContext } from '@/lib/auth'
import { useDatatypeFieldsByDatatype } from '@/queries/datatypes'
import { useFields } from '@/queries/fields'
import { useCreateContentField, useUpdateContentField } from '@/queries/content'

type MergedField = {
  fieldId: string
  label: string
  type: string
  value: string
  /** The existing content field record, if any. */
  contentField: NodeField['content'] | null
}

type NodeEditorProps = {
  node: ContentNode
}

export function NodeEditor({ node }: NodeEditorProps) {
  const { user } = useAuthContext()
  const datatypeId = node.datatype.info.datatype_id
  const { data: datatypeFields } = useDatatypeFieldsByDatatype(datatypeId)
  const { data: allFields } = useFields()
  const createContentField = useCreateContentField()
  const updateContentField = useUpdateContentField()

  // Build a lookup from field_id -> Field definition
  const fieldDefMap = useMemo(() => {
    const map = new Map<string, Field>()
    if (!allFields) return map
    for (const f of allFields) {
      map.set(f.field_id, f)
    }
    return map
  }, [allFields])

  // Build a lookup from field_id -> content value (from tree response)
  const contentFieldMap = useMemo(() => {
    const map = new Map<string, NodeField>()
    for (const nf of node.fields) {
      map.set(nf.info.field_id, nf)
    }
    return map
  }, [node.fields])

  // Merge schema-defined fields with existing content values
  const mergedFields = useMemo<MergedField[]>(() => {
    // If datatype field assignments haven't loaded yet, fall back to node.fields
    if (!datatypeFields) {
      return node.fields.map((nf) => ({
        fieldId: nf.info.field_id,
        label: nf.info.label,
        type: nf.info.type,
        value: nf.content.field_value ?? '',
        contentField: nf.content,
      }))
    }

    const result: MergedField[] = []
    const seen = new Set<string>()

    // Walk through schema-assigned fields in sort order
    for (const df of datatypeFields.sort((a, b) => a.sort_order - b.sort_order)) {
      seen.add(df.field_id)
      const existing = contentFieldMap.get(df.field_id)
      const fieldDef = fieldDefMap.get(df.field_id)

      if (existing) {
        result.push({
          fieldId: df.field_id,
          label: existing.info.label,
          type: existing.info.type,
          value: existing.content.field_value ?? '',
          contentField: existing.content,
        })
      } else if (fieldDef) {
        result.push({
          fieldId: df.field_id,
          label: fieldDef.label,
          type: fieldDef.type,
          value: '',
          contentField: null,
        })
      }
    }

    // Include any content fields not in schema (orphaned)
    for (const nf of node.fields) {
      if (!seen.has(nf.info.field_id)) {
        result.push({
          fieldId: nf.info.field_id,
          label: nf.info.label,
          type: nf.info.type,
          value: nf.content.field_value ?? '',
          contentField: nf.content,
        })
      }
    }

    return result
  }, [datatypeFields, node.fields, contentFieldMap, fieldDefMap])

  // Local edits: only stores fields the user has actually changed.
  // Server values come from mergedFields; local edits override them.
  const [localEdits, setLocalEdits] = useState<Record<string, string>>({})
  const [saving, setSaving] = useState(false)

  // Reset local edits when switching to a different node
  const nodeId = node.datatype.content.content_data_id
  const prevNodeIdRef = useRef(nodeId)
  if (prevNodeIdRef.current !== nodeId) {
    prevNodeIdRef.current = nodeId
    setLocalEdits({})
  }

  function getValue(fieldId: string): string {
    if (fieldId in localEdits) return localEdits[fieldId]
    const mf = mergedFields.find((f) => f.fieldId === fieldId)
    return mf?.value ?? ''
  }

  function updateValue(fieldId: string, value: string) {
    setLocalEdits((prev) => ({ ...prev, [fieldId]: value }))
  }

  // Check if any local edits differ from server values
  const isDirty = useMemo(() => {
    for (const fieldId of Object.keys(localEdits)) {
      const mf = mergedFields.find((f) => f.fieldId === fieldId)
      if (localEdits[fieldId] !== (mf?.value ?? '')) return true
    }
    return false
  }, [localEdits, mergedFields])

  async function handleSave() {
    setSaving(true)
    const now = new Date().toISOString()
    const promises: Promise<unknown>[] = []

    for (const field of mergedFields) {
      const currentValue = getValue(field.fieldId)
      if (currentValue === field.value) continue

      if (field.contentField) {
        // Update existing content field
        promises.push(
          updateContentField.mutateAsync({
            content_field_id: field.contentField.content_field_id as ContentFieldID,
            route_id: (node.datatype.content.route_id ?? null) as RouteID | null,
            content_data_id: node.datatype.content.content_data_id as ContentID | null,
            field_id: field.fieldId as FieldID | null,
            field_value: currentValue,
            author_id: (user?.user_id ?? null) as UserID | null,
            date_created: field.contentField.date_created,
            date_modified: now,
          }),
        )
      } else {
        // Create new content field
        promises.push(
          createContentField.mutateAsync({
            route_id: (node.datatype.content.route_id ?? null) as RouteID | null,
            content_data_id: node.datatype.content.content_data_id as ContentID | null,
            field_id: field.fieldId as FieldID | null,
            field_value: currentValue,
            author_id: (user?.user_id ?? null) as UserID | null,
            date_created: now,
            date_modified: now,
          }),
        )
      }
    }

    await Promise.all(promises)
    setLocalEdits({})
    setSaving(false)
  }

  return (
    <div className="space-y-6">
      <div>
        <div className="flex items-center gap-3">
          <h2 className="text-xl font-semibold">{node.datatype.info.label}</h2>
          <Badge variant="secondary">{node.datatype.info.type}</Badge>
        </div>
        <p className="mt-1 font-mono text-xs text-muted-foreground">
          {node.datatype.content.content_data_id}
        </p>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Fields</CardTitle>
            {mergedFields.length > 0 && (
              <Button
                size="sm"
                onClick={handleSave}
                disabled={!isDirty || saving}
              >
                {saving ? 'Saving...' : 'Save'}
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent>
          {mergedFields.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              No fields defined for this datatype.
            </p>
          ) : (
            <div className="space-y-5">
              {mergedFields.map((field) => (
                <FieldInput
                  key={field.fieldId}
                  field={field}
                  value={getValue(field.fieldId)}
                  onChange={(v) => updateValue(field.fieldId, v)}
                />
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function FieldInput({
  field,
  value,
  onChange,
}: {
  field: MergedField
  value: string
  onChange: (value: string) => void
}) {
  const inputId = `field-${field.fieldId}`

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <Label htmlFor={inputId}>{field.label}</Label>
        <Badge variant="outline" className="text-xs">
          {field.type}
        </Badge>
      </div>
      {renderInput(field.type, inputId, value, onChange)}
    </div>
  )
}

function renderInput(
  type: string,
  id: string,
  value: string,
  onChange: (value: string) => void,
) {
  switch (type) {
    case 'textarea':
    case 'richtext':
      return (
        <Textarea
          id={id}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          rows={5}
        />
      )
    case 'json':
      return (
        <Textarea
          id={id}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          rows={6}
          className="font-mono text-sm"
        />
      )
    case 'boolean':
      return (
        <div className="flex items-center gap-2 pt-1">
          <Checkbox
            id={id}
            checked={value === 'true'}
            onCheckedChange={(checked) => onChange(String(!!checked))}
          />
          <Label htmlFor={id} className="font-normal text-muted-foreground">
            {value === 'true' ? 'Yes' : 'No'}
          </Label>
        </div>
      )
    case 'number':
      return (
        <Input
          id={id}
          type="number"
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
      )
    case 'date':
      return (
        <Input
          id={id}
          type="date"
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
      )
    case 'datetime':
      return (
        <Input
          id={id}
          type="datetime-local"
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
      )
    case 'email':
      return (
        <Input
          id={id}
          type="email"
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
      )
    case 'url':
      return (
        <Input
          id={id}
          type="url"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder="https://"
        />
      )
    default:
      // text, slug, select, media, relation, and anything unknown
      return (
        <Input
          id={id}
          type="text"
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
      )
  }
}
