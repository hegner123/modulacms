import { useState, useCallback } from 'react'
import type {
  ContentNode,
  ContentFieldID,
  ContentID,
  FieldID,
  RouteID,
  UserID,
} from '@modulacms/admin-sdk'
import { useCreateContentField, useUpdateContentField } from '@/queries/content'
import { useAuthContext } from '@/lib/auth'
import type { MergedField } from './build-merged-fields'

type LocalEdits = Record<string, Record<string, string>>

type BlockEditorState = {
  saving: boolean
  getFieldValue: (contentDataId: string, fieldId: string, mergedFields: MergedField[]) => string
  setFieldValue: (contentDataId: string, fieldId: string, value: string) => void
  isBlockDirty: (contentDataId: string, mergedFields: MergedField[]) => boolean
  isDirty: (allBlocks: Array<{ contentDataId: string; mergedFields: MergedField[] }>) => boolean
  dirtyCount: (allBlocks: Array<{ contentDataId: string; mergedFields: MergedField[] }>) => number
  saveBlock: (node: ContentNode, mergedFields: MergedField[]) => Promise<void>
  saveAll: (blocks: Array<{ node: ContentNode; mergedFields: MergedField[] }>) => Promise<void>
  saveAllForce: (blocks: Array<{ node: ContentNode; mergedFields: MergedField[] }>) => Promise<void>
  clearBlockEdits: (contentDataId: string) => void
}

export function useBlockEditorState(): BlockEditorState {
  const { user } = useAuthContext()
  const createContentField = useCreateContentField()
  const updateContentField = useUpdateContentField()
  const [localEdits, setLocalEdits] = useState<LocalEdits>({})
  const [saving, setSaving] = useState(false)

  const getFieldValue = useCallback(
    (contentDataId: string, fieldId: string, mergedFields: MergedField[]): string => {
      const edit = localEdits[contentDataId]?.[fieldId]
      if (edit !== undefined) return edit
      const mf = mergedFields.find((f) => f.fieldId === fieldId)
      return mf?.value ?? ''
    },
    [localEdits],
  )

  const setFieldValue = useCallback(
    (contentDataId: string, fieldId: string, value: string) => {
      setLocalEdits((prev) => ({
        ...prev,
        [contentDataId]: {
          ...prev[contentDataId],
          [fieldId]: value,
        },
      }))
    },
    [],
  )

  const isBlockDirty = useCallback(
    (contentDataId: string, mergedFields: MergedField[]): boolean => {
      const edits = localEdits[contentDataId]
      if (!edits) return false
      for (const fieldId of Object.keys(edits)) {
        const mf = mergedFields.find((f) => f.fieldId === fieldId)
        if (edits[fieldId] !== (mf?.value ?? '')) return true
      }
      return false
    },
    [localEdits],
  )

  const isDirty = useCallback(
    (allBlocks: Array<{ contentDataId: string; mergedFields: MergedField[] }>): boolean => {
      return allBlocks.some((b) => isBlockDirty(b.contentDataId, b.mergedFields))
    },
    [isBlockDirty],
  )

  const dirtyCount = useCallback(
    (allBlocks: Array<{ contentDataId: string; mergedFields: MergedField[] }>): number => {
      return allBlocks.filter((b) => isBlockDirty(b.contentDataId, b.mergedFields)).length
    },
    [isBlockDirty],
  )

  const clearBlockEdits = useCallback((contentDataId: string) => {
    setLocalEdits((prev) => {
      const next = { ...prev }
      delete next[contentDataId]
      return next
    })
  }, [])

  const saveBlock = useCallback(
    async (node: ContentNode, mergedFields: MergedField[]) => {
      const contentDataId = node.datatype.content.content_data_id
      const edits = localEdits[contentDataId]
      if (!edits) return

      setSaving(true)
      const now = new Date().toISOString()
      const promises: Promise<unknown>[] = []

      for (const field of mergedFields) {
        const currentValue = edits[field.fieldId]
        if (currentValue === undefined || currentValue === field.value) continue

        if (field.contentField) {
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
      clearBlockEdits(contentDataId)
      setSaving(false)
    },
    [localEdits, user, createContentField, updateContentField, clearBlockEdits],
  )

  const saveAll = useCallback(
    async (blocks: Array<{ node: ContentNode; mergedFields: MergedField[] }>) => {
      setSaving(true)
      for (const block of blocks) {
        const contentDataId = block.node.datatype.content.content_data_id
        if (!isBlockDirty(contentDataId, block.mergedFields)) continue
        await saveBlock(block.node, block.mergedFields)
      }
      setSaving(false)
    },
    [isBlockDirty, saveBlock],
  )

  const saveBlockForce = useCallback(
    async (node: ContentNode, mergedFields: MergedField[]) => {
      const contentDataId = node.datatype.content.content_data_id
      const edits = localEdits[contentDataId]
      const now = new Date().toISOString()
      const promises: Promise<unknown>[] = []

      for (const field of mergedFields) {
        const localValue = edits?.[field.fieldId]
        const currentValue = localValue !== undefined ? localValue : field.value

        // The tree response includes stub entries for schema fields without
        // saved values. These stubs have contentField set but an empty
        // content_field_id. Treat them the same as missing content fields.
        const hasPersistedField = field.contentField && field.contentField.content_field_id

        if (hasPersistedField) {
          promises.push(
            updateContentField.mutateAsync({
              content_field_id: field.contentField!.content_field_id as ContentFieldID,
              route_id: (node.datatype.content.route_id ?? null) as RouteID | null,
              content_data_id: node.datatype.content.content_data_id as ContentID | null,
              field_id: field.fieldId as FieldID | null,
              field_value: currentValue,
              author_id: (user?.user_id ?? null) as UserID | null,
              date_created: field.contentField!.date_created,
              date_modified: now,
            }),
          )
        } else if (currentValue !== '') {
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

      if (promises.length > 0) {
        await Promise.all(promises)
      }
      clearBlockEdits(contentDataId)
    },
    [localEdits, user, createContentField, updateContentField, clearBlockEdits],
  )

  const saveAllForce = useCallback(
    async (blocks: Array<{ node: ContentNode; mergedFields: MergedField[] }>) => {
      setSaving(true)
      for (const block of blocks) {
        await saveBlockForce(block.node, block.mergedFields)
      }
      setSaving(false)
    },
    [saveBlockForce],
  )

  return {
    saving,
    getFieldValue,
    setFieldValue,
    isBlockDirty,
    isDirty,
    dirtyCount,
    saveBlock,
    saveAll,
    saveAllForce,
    clearBlockEdits,
  }
}

export type { MergedField, LocalEdits }
